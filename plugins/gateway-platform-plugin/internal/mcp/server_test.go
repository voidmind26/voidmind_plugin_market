package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBuildRegistersReadOnlyTools(t *testing.T) {
	srv := NewToolServer(NewClient("http://127.0.0.1:18787", nil), "http://127.0.0.1:18787/app").Build()
	tools := srv.ListTools()
	if len(tools) != 5 {
		t.Fatalf("expected 5 tools, got %d", len(tools))
	}
	for _, name := range []string{"health_check", "open_console_info", "list_routes", "list_keys", "list_references"} {
		if _, ok := tools[name]; !ok {
			t.Fatalf("expected tool %q to be registered", name)
		}
	}
}

func TestHealthToolUsesHTTPAPI(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/health":
			_, _ = w.Write([]byte(`{"ok":true}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer api.Close()

	client := NewClient(api.URL, api.Client())
	srv := NewToolServer(client, api.URL+"/app")

	body, err := srv.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if body["console_url"] != api.URL+"/app" {
		t.Fatalf("unexpected console url: %#v", body)
	}
}

func TestHealthEnsuresPlatformBeforeCheckingHealth(t *testing.T) {
	healthy := false
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			http.NotFound(w, r)
			return
		}
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

	srv := NewToolServer(NewClient(api.URL, api.Client()), api.URL+"/app")
	srv.runtime = runtime

	body, err := srv.Health(context.Background())
	if err != nil {
		t.Fatalf("Health returned error: %v", err)
	}
	if !started {
		t.Fatal("expected runtime start to be called before health response")
	}
	if body["ok"] != true {
		t.Fatalf("expected ok=true, got %#v", body)
	}
}
