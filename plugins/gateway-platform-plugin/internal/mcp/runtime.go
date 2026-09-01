package mcp

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Runtime struct {
	BaseURL         string
	Client          *Client
	Start           func(context.Context) error
	Stop            func(context.Context, *HealthStatus) error
	PollInterval    time.Duration
	Timeout         time.Duration
	ExpectedDataDir string
	ExpectedVersion string
}

func (r Runtime) EnsurePlatform(ctx context.Context) error {
	status, ok, err := r.health()
	if err == nil && ok {
		return nil
	}
	if r.Stop != nil {
		if stopErr := r.Stop(ctx, status); stopErr != nil {
			return stopErr
		}
	}
	if r.Start == nil {
		if err != nil {
			return fmt.Errorf("platform unavailable: %w", err)
		}
		return fmt.Errorf("platform unavailable")
	}
	if err := r.Start(ctx); err != nil {
		return err
	}

	timeout := time.NewTimer(r.Timeout)
	ticker := time.NewTicker(r.PollInterval)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		_, ok, healthErr := r.health()
		if healthErr == nil && ok {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			if healthErr != nil {
				return fmt.Errorf("platform health check timeout: %w", healthErr)
			}
			return fmt.Errorf("platform health check timeout")
		case <-ticker.C:
		}
	}
}

func (r Runtime) health() (*HealthStatus, bool, error) {
	status, err := r.Client.Health()
	if err != nil {
		return status, false, err
	}
	if !status.OK || !status.DatabaseWritable || status.DataDir == "" {
		return status, false, nil
	}
	if r.ExpectedDataDir != "" && !samePath(status.DataDir, r.ExpectedDataDir) {
		return status, false, nil
	}
	if r.ExpectedVersion != "" {
		comparison, comparable := compareGatewayVersions(status.Version, r.ExpectedVersion)
		if !comparable {
			return status, false, nil
		}
		if comparison < 0 {
			return status, false, nil
		}
	}
	return status, true, nil
}

func compareGatewayVersions(running, expected string) (int, bool) {
	runningVersion, runningOK := parseGatewayVersion(running)
	expectedVersion, expectedOK := parseGatewayVersion(expected)
	if !runningOK || !expectedOK {
		return 0, false
	}
	for index := range runningVersion.base {
		if runningVersion.base[index] < expectedVersion.base[index] {
			return -1, true
		}
		if runningVersion.base[index] > expectedVersion.base[index] {
			return 1, true
		}
	}
	if runningVersion.cachebuster < expectedVersion.cachebuster {
		return -1, true
	}
	if runningVersion.cachebuster > expectedVersion.cachebuster {
		return 1, true
	}
	return 0, true
}

type gatewayVersion struct {
	base        [3]uint64
	cachebuster uint64
}

func parseGatewayVersion(version string) (gatewayVersion, bool) {
	baseAndCache := strings.SplitN(version, "+codex.", 2)
	if len(baseAndCache) != 2 {
		return gatewayVersion{}, false
	}
	baseParts := strings.Split(baseAndCache[0], ".")
	if len(baseParts) != 3 {
		return gatewayVersion{}, false
	}
	parsed := gatewayVersion{}
	for index, part := range baseParts {
		value, err := strconv.ParseUint(part, 10, 64)
		if err != nil {
			return gatewayVersion{}, false
		}
		parsed.base[index] = value
	}
	cachebuster, err := strconv.ParseUint(baseAndCache[1], 10, 64)
	if err != nil {
		return gatewayVersion{}, false
	}
	parsed.cachebuster = cachebuster
	return parsed, true
}

func samePath(left, right string) bool {
	leftInfo, leftErr := os.Stat(left)
	rightInfo, rightErr := os.Stat(right)
	if leftErr == nil && rightErr == nil {
		return os.SameFile(leftInfo, rightInfo)
	}
	return filepath.Clean(left) == filepath.Clean(right)
}

func StartHTTPPlatformCommand(executablePath, logPath string) func(context.Context) error {
	return func(context.Context) error {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
			return fmt.Errorf("create gateway platform log directory: %w", err)
		}
		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return fmt.Errorf("open gateway platform log: %w", err)
		}
		if err := logFile.Chmod(0o600); err != nil {
			_ = logFile.Close()
			return fmt.Errorf("secure gateway platform log: %w", err)
		}
		cmd := exec.Command(executablePath)
		cmd.Dir = filepath.Dir(logPath)
		cmd.Env = os.Environ()
		cmd.Stdout = logFile
		cmd.Stderr = logFile
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			_ = logFile.Close()
			return fmt.Errorf("start gateway platform HTTP service: %w", err)
		}
		go func() {
			_ = cmd.Wait()
			_ = logFile.Close()
		}()
		return nil
	}
}

func StopHTTPPlatformCommand(port int) func(context.Context, *HealthStatus) error {
	return func(ctx context.Context, status *HealthStatus) error {
		pids, err := gatewayPlatformPIDs(ctx, port, status)
		if err != nil {
			return err
		}
		for _, pid := range pids {
			process, err := os.FindProcess(pid)
			if err != nil {
				return fmt.Errorf("find gateway platform HTTP process %d: %w", pid, err)
			}
			if err := process.Signal(os.Interrupt); err != nil {
				return fmt.Errorf("stop gateway platform HTTP process %d: %w", pid, err)
			}
		}
		if len(pids) == 0 {
			return nil
		}

		deadline := time.NewTimer(2 * time.Second)
		ticker := time.NewTicker(50 * time.Millisecond)
		defer deadline.Stop()
		defer ticker.Stop()
		for {
			if !portIsListening(port) {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-deadline.C:
				return fmt.Errorf("gateway platform HTTP process did not release port %d", port)
			case <-ticker.C:
			}
		}
	}
}

func gatewayPlatformPIDs(ctx context.Context, port int, status *HealthStatus) ([]int, error) {
	if !portIsListening(port) {
		return nil, nil
	}

	output, err := exec.CommandContext(
		ctx,
		"lsof",
		"-nP",
		"-t",
		fmt.Sprintf("-iTCP:%d", port),
		"-sTCP:LISTEN",
	).Output()
	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect process listening on port %d: %w", port, err)
	}

	var pids []int
	reportedPIDFound := status == nil || status.PID <= 0
	for _, field := range strings.Fields(string(output)) {
		pid, err := strconv.Atoi(field)
		if err != nil {
			return nil, fmt.Errorf("parse listener pid %q: %w", field, err)
		}
		command, err := processCommand(ctx, pid)
		if err != nil {
			return nil, err
		}
		if !isGatewayPlatformCommand(command) {
			return nil, fmt.Errorf("port %d is occupied by a non-gateway process: %s", port, command)
		}
		if status != nil && pid == status.PID {
			reportedPIDFound = true
		}
		pids = append(pids, pid)
	}
	if !reportedPIDFound {
		return nil, fmt.Errorf("gateway platform health PID %d does not own port %d", status.PID, port)
	}
	return pids, nil
}

func processCommand(ctx context.Context, pid int) (string, error) {
	output, err := exec.CommandContext(ctx, "ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return "", fmt.Errorf("inspect listener process %d: %w", pid, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func isGatewayPlatformCommand(command string) bool {
	command = strings.ToLower(command)
	return strings.Contains(command, "gateway-platform-http") ||
		strings.Contains(command, "gateway-platform-plugin")
}

func portIsListening(port int) bool {
	connection, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 50*time.Millisecond)
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}
