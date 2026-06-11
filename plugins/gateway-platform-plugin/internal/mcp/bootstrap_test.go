package mcp

import (
	"context"
	"testing"
)

func TestEnsurePlatformAtStartupCallsRuntime(t *testing.T) {
	called := false
	runtime := Runtime{
		Start: func(context.Context) error {
			called = true
			return nil
		},
	}

	if err := EnsurePlatformAtStartup(context.Background(), runtime); err != nil {
		t.Fatalf("EnsurePlatformAtStartup returned error: %v", err)
	}
	if !called {
		t.Fatal("expected runtime start to be called during MCP startup")
	}
}
