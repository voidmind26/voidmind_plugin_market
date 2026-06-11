package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway-platform-plugin/server/data/sqlite"
	"gateway-platform-plugin/server/helpers"
)

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	db, err := sqlite.OpenTestDB(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.InitSchema(db); err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(NewRouter(&helpers.App{DB: db}))
}

func httpGet(t *testing.T, url string) *http.Response {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func httpPostJSON(t *testing.T, url string, body string) *http.Response {
	t.Helper()
	resp, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func httpPutJSON(t *testing.T, url string, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func httpDelete(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func assertStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		t.Fatalf("expected status %d, got %d", expected, resp.StatusCode)
	}
}

func assertBodyContains(t *testing.T, resp *http.Response, expected string) {
	t.Helper()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), expected) {
		t.Fatalf("expected body to contain %q, got %q", expected, string(body))
	}
}

func createRouteAndKeyFixtures(t *testing.T, server *httptest.Server) {
	t.Helper()
	assertStatus(t, httpPostJSON(t, server.URL+"/api/routes", `{"name":"ship","enabled":true,"local_path":"/ship","upstream_url":"https://ship.example/mcp","timeout_ms":30000,"description":"ship"}`), http.StatusCreated)
	assertStatus(t, httpPostJSON(t, server.URL+"/api/keys", `{"name":"ips-token","value":"abc","description":"ips"}`), http.StatusCreated)
}

func TestRoutesCRUD(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()

	createBody := `{"name":"ship","enabled":true,"local_path":"/ship","upstream_url":"https://ship.example/mcp","timeout_ms":30000,"description":"ship"}`
	resp := httpPostJSON(t, server.URL+"/api/routes", createBody)
	assertStatus(t, resp, http.StatusCreated)

	resp = httpGet(t, server.URL+"/api/routes")
	assertStatus(t, resp, http.StatusOK)
	assertBodyContains(t, resp, "ship")

	updateBody := `{"name":"ship-v2","enabled":false,"local_path":"/ship-v2","upstream_url":"https://ship.example/mcp","timeout_ms":1000,"description":"updated"}`
	resp = httpPutJSON(t, server.URL+"/api/routes/1", updateBody)
	assertStatus(t, resp, http.StatusOK)
	assertBodyContains(t, resp, "ship-v2")

	resp = httpDelete(t, server.URL+"/api/routes/1")
	assertStatus(t, resp, http.StatusNoContent)
}

func TestKeysCRUD(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()
	resp := httpPostJSON(t, server.URL+"/api/keys", `{"name":"ips-token","value":"abc","description":"ips"}`)
	assertStatus(t, resp, http.StatusCreated)
	resp = httpPutJSON(t, server.URL+"/api/keys/1", `{"name":"ips-token","value":"xyz","description":"updated"}`)
	assertStatus(t, resp, http.StatusOK)
	resp = httpDelete(t, server.URL+"/api/keys/1")
	assertStatus(t, resp, http.StatusNoContent)
}

func TestRouteRewritesCRUD(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()
	createRouteAndKeyFixtures(t, server)
	resp := httpPostJSON(t, server.URL+"/api/routes/1/rewrites", `{"rewrite_type":"header","target_name":"Authorization","key_id":1,"template":"Bearer {{value}}","ordering":1}`)
	assertStatus(t, resp, http.StatusCreated)
	assertBodyContains(t, resp, "Authorization")
	resp = httpPutJSON(t, server.URL+"/api/routes/1/rewrites/1", `{"rewrite_type":"cookie","target_name":"ZYBIPSCAS","key_id":1,"template":"{{value}}","ordering":1}`)
	assertStatus(t, resp, http.StatusOK)
	assertBodyContains(t, resp, "ZYBIPSCAS")
	resp = httpDelete(t, server.URL+"/api/routes/1/rewrites/1")
	assertStatus(t, resp, http.StatusNoContent)
}

func TestGatewayPrefixRouteMatchesSubpath(t *testing.T) {
	server := newTestServer(t)
	defer server.Close()
	createRouteAndKeyFixtures(t, server)
	resp := httpGet(t, server.URL+"/gateway/ship/health")
	if resp.StatusCode == http.StatusNotFound {
		t.Fatal("expected prefix route to match subpath, got 404")
	}
}
