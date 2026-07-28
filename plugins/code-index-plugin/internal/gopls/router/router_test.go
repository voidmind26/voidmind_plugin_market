package router

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"code-index-plugin/internal/gopls/session"
	"code-index-plugin/internal/gopls/workspace"

	"github.com/mark3labs/mcp-go/mcp"
)

type fakeSessions struct {
	mu      sync.Mutex
	clients map[string]*recordingClient
	active  map[string]bool
	invalid []string
}

func (s *fakeSessions) Get(_ context.Context, root string) (session.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active[root] = true
	return s.clients[root], nil
}

func (s *fakeSessions) Active(root string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active[root]
}

func (s *fakeSessions) Invalidate(root string, _ session.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invalid = append(s.invalid, root)
}

type recordingClient struct {
	mu       sync.Mutex
	response string
	err      error
	requests []mcp.CallToolRequest
}

func (c *recordingClient) CallTool(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.requests = append(c.requests, request)
	if c.err != nil {
		return nil, c.err
	}
	return mcp.NewToolResultText(c.response), nil
}

func (c *recordingClient) Close() error { return nil }

func TestCallFanoutSearchMergesAndDeduplicates(t *testing.T) {
	root, moduleA, moduleB, _, _ := makeMultiRepo(t)
	duplicate := "Top symbol matches:\n\tRun (Function in `/shared/run.go`)\n"
	sessions := &fakeSessions{
		clients: map[string]*recordingClient{
			moduleA: {response: duplicate},
			moduleB: {response: duplicate},
		},
		active: map[string]bool{},
	}
	router := New(workspace.NewResolver(), sessions)

	result, err := router.CallFanout(context.Background(), "go_search", root, map[string]any{
		"project_root": root,
		"query":        "Run",
	})
	if err != nil {
		t.Fatalf("CallFanout() error = %v", err)
	}
	text := resultText(t, result)
	if strings.Count(text, "/shared/run.go") != 1 {
		t.Fatalf("search result was not deduplicated: %s", text)
	}
	for _, client := range sessions.clients {
		if len(client.requests) != 1 {
			t.Fatalf("client requests = %d, want 1", len(client.requests))
		}
		args := client.requests[0].GetArguments()
		if _, exists := args["project_root"]; exists {
			t.Fatal("routing-only project_root was forwarded upstream")
		}
	}
}

func TestDiagnosticsGroupsFilesByWorkspace(t *testing.T) {
	root, moduleA, moduleB, fileA, fileB := makeMultiRepo(t)
	clientA := &recordingClient{response: "No diagnostics."}
	clientB := &recordingClient{response: "No diagnostics."}
	router := New(workspace.NewResolver(), &fakeSessions{
		clients: map[string]*recordingClient{moduleA: clientA, moduleB: clientB},
		active:  map[string]bool{},
	})

	result, err := router.Diagnostics(context.Background(), root, []string{fileA, fileB}, map[string]any{
		"project_root": root,
		"files":        []string{fileA, fileB},
	})
	if err != nil {
		t.Fatalf("Diagnostics() error = %v", err)
	}
	if result.IsError {
		t.Fatalf("Diagnostics() returned tool error: %s", resultText(t, result))
	}
	assertSingleDiagnosticFile(t, clientA, fileA)
	assertSingleDiagnosticFile(t, clientB, fileB)
}

func TestCallSingleRejectsMultiRepoParent(t *testing.T) {
	root, moduleA, moduleB, _, _ := makeMultiRepo(t)
	router := New(workspace.NewResolver(), &fakeSessions{
		clients: map[string]*recordingClient{moduleA: {}, moduleB: {}},
		active:  map[string]bool{},
	})

	_, err := router.CallSingle(context.Background(), "go_package_api", root, nil)
	if err == nil || !strings.Contains(err.Error(), "明确的单个 Go workspace") {
		t.Fatalf("CallSingle() error = %v, want explicit workspace guidance", err)
	}
}

func TestCallInvalidatesFailedSession(t *testing.T) {
	_, moduleA, _, _, _ := makeMultiRepo(t)
	client := &recordingClient{err: errors.New("transport closed")}
	sessions := &fakeSessions{
		clients: map[string]*recordingClient{moduleA: client},
		active:  map[string]bool{},
	}
	router := New(workspace.NewResolver(), sessions)

	_, err := router.CallSingle(context.Background(), "go_workspace", moduleA, nil)
	if err == nil {
		t.Fatal("CallSingle() accepted an upstream transport failure")
	}
	if len(sessions.invalid) != 1 || sessions.invalid[0] != moduleA {
		t.Fatalf("invalidated sessions = %v, want [%s]", sessions.invalid, moduleA)
	}

	client.err = context.Canceled
	_, _ = router.CallSingle(context.Background(), "go_workspace", moduleA, nil)
	if len(sessions.invalid) != 1 {
		t.Fatalf("canceled call invalidated session: %v", sessions.invalid)
	}
}

func makeMultiRepo(t *testing.T) (root, moduleA, moduleB, fileA, fileB string) {
	t.Helper()
	root = t.TempDir()
	moduleA = filepath.Join(root, "repo-a")
	moduleB = filepath.Join(root, "repo-b")
	for _, module := range []string{moduleA, moduleB} {
		if err := os.MkdirAll(module, 0o755); err != nil {
			t.Fatalf("MkdirAll(%s): %v", module, err)
		}
		if err := os.WriteFile(filepath.Join(module, "go.mod"), []byte("module example.com/test\n"), 0o644); err != nil {
			t.Fatalf("WriteFile(go.mod): %v", err)
		}
	}
	fileA = filepath.Join(moduleA, "a.go")
	fileB = filepath.Join(moduleB, "b.go")
	if err := os.WriteFile(fileA, []byte("package a\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(a.go): %v", err)
	}
	if err := os.WriteFile(fileB, []byte("package b\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(b.go): %v", err)
	}
	var err error
	root, err = workspace.Normalize(root)
	if err != nil {
		t.Fatalf("Normalize(root): %v", err)
	}
	moduleA, err = workspace.Normalize(moduleA)
	if err != nil {
		t.Fatalf("Normalize(moduleA): %v", err)
	}
	moduleB, err = workspace.Normalize(moduleB)
	if err != nil {
		t.Fatalf("Normalize(moduleB): %v", err)
	}
	fileA = filepath.Join(moduleA, "a.go")
	fileB = filepath.Join(moduleB, "b.go")
	return root, moduleA, moduleB, fileA, fileB
}

func assertSingleDiagnosticFile(t *testing.T, client *recordingClient, want string) {
	t.Helper()
	if len(client.requests) != 1 {
		t.Fatalf("diagnostics requests = %d, want 1", len(client.requests))
	}
	files, ok := client.requests[0].GetArguments()["files"].([]string)
	if !ok || len(files) != 1 || files[0] != want {
		t.Fatalf("diagnostics files = %#v, want [%s]", client.requests[0].GetArguments()["files"], want)
	}
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("tool result has no content")
	}
	text, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("unexpected tool content: %+v", result.Content[0])
	}
	return text.Text
}
