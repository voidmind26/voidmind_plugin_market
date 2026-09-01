package mcp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestEnsurePlatformReturnsImmediatelyWhenHealthy(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_writable":true}`))
	}))
	defer api.Close()

	runtime := Runtime{
		BaseURL: api.URL,
		Client:  NewClient(api.URL, api.Client()),
		Start: func(context.Context) error {
			t.Fatal("start should not be called when platform is healthy")
			return nil
		},
		PollInterval:    10 * time.Millisecond,
		Timeout:         100 * time.Millisecond,
		ExpectedDataDir: "/tmp/gateway",
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
}

func TestEnsurePlatformStartsWhenUnhealthy(t *testing.T) {
	healthy := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !healthy {
			http.Error(w, "booting", http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_writable":true}`))
	}))
	defer api.Close()

	started := false
	runtime := Runtime{
		BaseURL: api.URL,
		Client:  NewClient(api.URL, api.Client()),
		Start: func(context.Context) error {
			started = true
			healthy = true
			return nil
		},
		PollInterval:    10 * time.Millisecond,
		Timeout:         200 * time.Millisecond,
		ExpectedDataDir: "/tmp/gateway",
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
	if !started {
		t.Fatal("expected start to be called")
	}
}

func TestEnsurePlatformStartsWhenDataDirectoryDoesNotMatch(t *testing.T) {
	dataDir := "/tmp/legacy"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"` + dataDir + `","database_writable":true}`))
	}))
	defer api.Close()

	started := false
	stopped := false
	runtime := Runtime{
		BaseURL:         api.URL,
		Client:          NewClient(api.URL, api.Client()),
		ExpectedDataDir: "/tmp/gateway",
		Stop: func(context.Context, *HealthStatus) error {
			stopped = true
			return nil
		},
		Start: func(context.Context) error {
			if !stopped {
				t.Fatal("expected mismatched instance to stop before replacement starts")
			}
			started = true
			dataDir = "/tmp/gateway"
			return nil
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      200 * time.Millisecond,
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
	if !started {
		t.Fatal("expected mismatched instance to trigger a restart")
	}
}

func TestEnsurePlatformRestartsWhenVersionIsOlder(t *testing.T) {
	version := "1.0.1+codex.20260801000000"
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_writable":true,"version":"` + version + `"}`))
	}))
	defer api.Close()

	stopped := false
	started := false
	runtime := Runtime{
		BaseURL:         api.URL,
		Client:          NewClient(api.URL, api.Client()),
		ExpectedDataDir: "/tmp/gateway",
		ExpectedVersion: "1.0.1+codex.20260901000000",
		Stop: func(_ context.Context, status *HealthStatus) error {
			if status == nil || status.Version != version {
				t.Fatalf("unexpected health status passed to stop: %#v", status)
			}
			stopped = true
			return nil
		},
		Start: func(context.Context) error {
			if !stopped {
				t.Fatal("expected stale version to stop before replacement starts")
			}
			started = true
			version = "1.0.1+codex.20260901000000"
			return nil
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      200 * time.Millisecond,
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
	if !started {
		t.Fatal("expected stale version to trigger a restart")
	}
}

func TestEnsurePlatformReusesNewerVersion(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_writable":true,"version":"1.0.1+codex.20260902000000"}`))
	}))
	defer api.Close()

	runtime := Runtime{
		Client:          NewClient(api.URL, api.Client()),
		ExpectedDataDir: "/tmp/gateway",
		ExpectedVersion: "1.0.1+codex.20260901000000",
		Stop: func(context.Context, *HealthStatus) error {
			t.Fatal("older MCP must not stop a newer HTTP service")
			return nil
		},
		Start: func(context.Context) error {
			t.Fatal("older MCP must not replace a newer HTTP service")
			return nil
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      100 * time.Millisecond,
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
}

func TestCompareGatewayVersions(t *testing.T) {
	tests := []struct {
		running  string
		expected string
		want     int
	}{
		{"1.0.1+codex.20260901000000", "1.0.1+codex.20260901000000", 0},
		{"1.0.1+codex.20260902000000", "1.0.1+codex.20260901000000", 1},
		{"1.0.1+codex.20260831000000", "1.0.1+codex.20260901000000", -1},
		{"1.1.0+codex.20260101000000", "1.0.1+codex.20260901000000", 1},
	}
	for _, test := range tests {
		got, ok := compareGatewayVersions(test.running, test.expected)
		if !ok || got != test.want {
			t.Fatalf("compareGatewayVersions(%q, %q) = %d, %v; want %d, true", test.running, test.expected, got, ok, test.want)
		}
	}
	if _, ok := compareGatewayVersions("dev", "1.0.1+codex.20260901000000"); ok {
		t.Fatal("expected unversioned build to be incomparable")
	}
}

func TestGatewayPlatformCommandGuard(t *testing.T) {
	for _, command := range []string{
		"/tmp/gateway-platform-http",
		"/var/folders/go-build/exe/gateway-platform-plugin",
	} {
		if !isGatewayPlatformCommand(command) {
			t.Fatalf("expected gateway command to be accepted: %q", command)
		}
	}
	if isGatewayPlatformCommand("python -m http.server 18787") {
		t.Fatal("expected unrelated listener command to be rejected")
	}
}

func TestStopHTTPPlatformCommandStopsReportedGatewayProcess(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestGatewayPlatformHTTPHelperProcess$", "--", "gateway-platform-http")
	cmd.Env = append(os.Environ(),
		"GO_WANT_GATEWAY_HTTP_HELPER=1",
		"GATEWAY_HTTP_HELPER_PORT="+strconv.Itoa(port),
	)
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}
	})

	deadline := time.Now().Add(2 * time.Second)
	for !portIsListening(port) && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if !portIsListening(port) {
		t.Fatal("helper HTTP process did not start listening")
	}

	status := &HealthStatus{
		PID:            cmd.Process.Pid,
		ExecutablePath: "/tmp/gateway-platform-http",
	}
	if err := StopHTTPPlatformCommand(port)(context.Background(), status); err != nil {
		t.Fatal(err)
	}
	if portIsListening(port) {
		t.Fatalf("expected helper HTTP process to release port %d", port)
	}
	_ = cmd.Wait()
}

func TestGatewayPlatformHTTPHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_GATEWAY_HTTP_HELPER") != "1" {
		return
	}
	port, err := strconv.Atoi(os.Getenv("GATEWAY_HTTP_HELPER_PORT"))
	if err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	for {
		connection, err := listener.Accept()
		if err != nil {
			return
		}
		_ = connection.Close()
	}
}

func TestEnsurePlatformReturnsStartError(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unhealthy", http.StatusServiceUnavailable)
	}))
	defer api.Close()

	runtime := Runtime{
		BaseURL: api.URL,
		Client:  NewClient(api.URL, api.Client()),
		Start: func(context.Context) error {
			return errors.New("boom")
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      50 * time.Millisecond,
	}

	err := runtime.EnsurePlatform(context.Background())
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom error, got %v", err)
	}
}

func TestStartHTTPPlatformCommandUsesStableDataDirectory(t *testing.T) {
	root := t.TempDir()
	executable := filepath.Join(root, "gateway-platform-http")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\npwd\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir := filepath.Join(root, "visible-data")
	logPath := filepath.Join(dataDir, "gateway-platform-http.log")

	if err := StartHTTPPlatformCommand(executable, logPath)(context.Background()); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		content, err := os.ReadFile(logPath)
		if err == nil && samePath(strings.TrimSpace(string(content)), dataDir) {
			info, statErr := os.Stat(logPath)
			if statErr != nil {
				t.Fatal(statErr)
			}
			if info.Mode().Perm() != 0o600 {
				t.Fatalf("expected log mode 0600, got %o", info.Mode().Perm())
			}
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	content, err := os.ReadFile(logPath)
	t.Fatalf("HTTP process did not run from stable data directory %q: content=%q err=%v", dataDir, content, err)
}
