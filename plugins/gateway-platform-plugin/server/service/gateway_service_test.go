package service

import (
	"net/http"
	"testing"

	"gateway-platform-plugin/server/models"
)

func TestGatewayInjectsRewriteValues(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com/mcp", nil)
	if err != nil {
		t.Fatal(err)
	}
	rewrites := []RewriteLike{
		{RewriteType: "header", TargetName: "Authorization", KeyID: 1, Template: "Bearer {{value}}"},
		{RewriteType: "query", TargetName: "token", KeyID: 1, Template: "{{value}}"},
		{RewriteType: "cookie", TargetName: "ZYBIPSCAS", KeyID: 1, Template: "{{value}}"},
	}
	keys := map[int64]KeyLike{
		1: {ID: 1, Name: "ips-token", Value: "abc"},
	}

	if err := ApplyRewritesLike(req, rewrites, keys); err != nil {
		t.Fatal(err)
	}

	if got := req.Header.Get("Authorization"); got != "Bearer abc" {
		t.Fatalf("expected Authorization header, got %q", got)
	}
	if got := req.URL.Query().Get("token"); got != "abc" {
		t.Fatalf("expected token query, got %q", got)
	}
	cookies := req.Cookies()
	if len(cookies) != 1 || cookies[0].Name != "ZYBIPSCAS" || cookies[0].Value != "abc" {
		t.Fatalf("expected injected cookie, got %+v", cookies)
	}
}

func TestGatewayMatchesExactPrefixRoot(t *testing.T) {
	routes := []models.Route{{ID: 1, Enabled: true, LocalPath: "ship", UpstreamURL: "https://example.com/mcp"}}
	route, restPath := MatchRouteByPrefix(routes, "ship")
	if route == nil {
		t.Fatal("expected route to match")
	}
	if restPath != "" {
		t.Fatalf("expected empty restPath, got %q", restPath)
	}
}

func TestGatewayMatchesPrefixWithRestPath(t *testing.T) {
	routes := []models.Route{{ID: 1, Enabled: true, LocalPath: "ship", UpstreamURL: "https://example.com/mcp"}}
	route, restPath := MatchRouteByPrefix(routes, "ship/health")
	if route == nil {
		t.Fatal("expected route to match")
	}
	if restPath != "/health" {
		t.Fatalf("expected /health, got %q", restPath)
	}
}

func TestGatewayPrefersLongestPrefix(t *testing.T) {
	routes := []models.Route{
		{ID: 1, Enabled: true, LocalPath: "ship", UpstreamURL: "https://example.com/mcp"},
		{ID: 2, Enabled: true, LocalPath: "ship-admin", UpstreamURL: "https://admin.example.com/mcp"},
	}
	route, restPath := MatchRouteByPrefix(routes, "ship-admin/health")
	if route == nil || route.ID != 2 {
		t.Fatal("expected longest prefix route to match")
	}
	if restPath != "/health" {
		t.Fatalf("expected /health, got %q", restPath)
	}
}

func TestJoinUpstreamURLWithRestPath(t *testing.T) {
	got := JoinUpstreamURL("https://ship.yukework.com/ship-mcp/mcp", "/health")
	want := "https://ship.yukework.com/ship-mcp/mcp/health"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
