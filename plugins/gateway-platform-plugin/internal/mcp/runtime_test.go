package mcp

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnsurePlatformReturnsImmediatelyWhenHealthy(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer api.Close()

	runtime := Runtime{
		BaseURL: api.URL,
		Client:  NewClient(api.URL, api.Client()),
		Start: func(context.Context) error {
			t.Fatal("start should not be called when platform is healthy")
			return nil
		},
		PollInterval: 10 * time.Millisecond,
		Timeout:      100 * time.Millisecond,
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
		_, _ = w.Write([]byte(`{"ok":true}`))
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
		PollInterval: 10 * time.Millisecond,
		Timeout:      200 * time.Millisecond,
	}

	if err := runtime.EnsurePlatform(context.Background()); err != nil {
		t.Fatalf("EnsurePlatform returned error: %v", err)
	}
	if !started {
		t.Fatal("expected start to be called")
	}
}

func TestEnsurePlatformReturnsStartError(t *testing.T) {
	runtime := Runtime{
		BaseURL: "http://127.0.0.1:18787",
		Client:  NewClient("http://127.0.0.1:18787", &http.Client{Timeout: 10 * time.Millisecond}),
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
