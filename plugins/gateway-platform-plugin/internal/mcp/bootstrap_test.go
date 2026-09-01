package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnsurePlatformAtStartupUsesRuntimeHealthCheck(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_writable":true}`))
	}))
	defer api.Close()

	runtime := Runtime{
		Client: NewClient(api.URL, api.Client()),
		Start: func(context.Context) error {
			t.Fatal("healthy platform should not be started again")
			return nil
		},
		PollInterval:    10 * time.Millisecond,
		Timeout:         100 * time.Millisecond,
		ExpectedDataDir: "/tmp/gateway",
	}

	if err := EnsurePlatformAtStartup(context.Background(), runtime); err != nil {
		t.Fatalf("EnsurePlatformAtStartup returned error: %v", err)
	}
}
