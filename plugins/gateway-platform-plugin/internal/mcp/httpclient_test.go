package mcp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientHealthCheck(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/health" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true,"data_dir":"/tmp/gateway","database_path":"/tmp/gateway/gateway-platform.db","database_writable":true}`))
	}))
	defer api.Close()

	client := NewClient(api.URL, api.Client())
	ok, err := client.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected health check to be true")
	}
}

func TestClientHealthCheckRejectsLegacyResponse(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer api.Close()

	client := NewClient(api.URL, api.Client())
	ok, err := client.HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck returned error: %v", err)
	}
	if ok {
		t.Fatal("expected legacy health response without writable data directory to be rejected")
	}
}

func TestClientListKeysMasksValues(t *testing.T) {
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/keys" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":1,"name":"ips-token","value":"secret","description":"ips"}]`))
	}))
	defer api.Close()

	client := NewClient(api.URL, api.Client())
	keys, err := client.ListKeys()
	if err != nil {
		t.Fatalf("ListKeys returned error: %v", err)
	}
	if keys[0].Value != "***" {
		t.Fatalf("expected masked value, got %q", keys[0].Value)
	}
}
