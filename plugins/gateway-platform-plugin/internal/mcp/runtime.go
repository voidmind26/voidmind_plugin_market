package mcp

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"
)

type Runtime struct {
	BaseURL      string
	Client       *Client
	Start        func(context.Context) error
	PollInterval time.Duration
	Timeout      time.Duration
}

func (r Runtime) EnsurePlatform(ctx context.Context) error {
	ok, err := r.Client.HealthCheck()
	if err == nil && ok {
		return nil
	}
	if r.Start == nil {
		return fmt.Errorf("platform unavailable: %w", err)
	}
	if err := r.Start(ctx); err != nil {
		return err
	}

	timeout := time.NewTimer(r.Timeout)
	ticker := time.NewTicker(r.PollInterval)
	defer timeout.Stop()
	defer ticker.Stop()

	for {
		ok, healthErr := r.Client.HealthCheck()
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

func StartHTTPPlatformCommand(pluginRoot string) func(context.Context) error {
	return func(ctx context.Context) error {
		logPath := filepath.Join(pluginRoot, "gateway-platform-http.log")
		cmd := exec.CommandContext(ctx, "bash", "-lc", "cd \""+pluginRoot+"\" && GOWORK=off go run . >\""+logPath+"\" 2>&1 &")
		return cmd.Run()
	}
}
