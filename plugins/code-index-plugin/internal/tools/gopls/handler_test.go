package gopls

import (
	"context"
	"encoding/json"
	"testing"

	goplsrouter "code-index-plugin/internal/gopls/router"
	"code-index-plugin/internal/gopls/session"
	"code-index-plugin/internal/gopls/workspace"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type noSessionPool struct{}

func (noSessionPool) Get(context.Context, string) (session.Client, error) { return nil, nil }
func (noSessionPool) Active(string) bool                                  { return false }

func TestRegisterToolsExposesRoutedGoplsSchemas(t *testing.T) {
	server := mcpserver.NewMCPServer("test", "1.0.0")
	handler := NewHandler(goplsrouter.New(workspace.NewResolver(), noSessionPool{}))
	handler.RegisterTools(server)

	response := server.HandleMessage(context.Background(), json.RawMessage(`{
		"jsonrpc":"2.0",
		"id":1,
		"method":"tools/list"
	}`))
	jsonResponse, ok := response.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("unexpected response type %T", response)
	}
	result, ok := jsonResponse.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("unexpected result type %T", jsonResponse.Result)
	}
	if len(result.Tools) != 9 {
		t.Fatalf("gopls tool count = %d, want 9", len(result.Tools))
	}

	for _, name := range []string{
		"list_go_workspaces",
		"go_workspace",
		"go_search",
		"go_file_context",
		"go_package_api",
		"go_symbol_references",
		"go_diagnostics",
		"go_vulncheck",
		"go_rename_symbol",
	} {
		if findTool(result.Tools, name) == nil {
			t.Errorf("tool %s is missing", name)
		}
	}

	search := findTool(result.Tools, "go_search")
	if search == nil {
		t.Fatal("go_search is missing")
	}
	if !contains(search.InputSchema.Required, "project_root") || !contains(search.InputSchema.Required, "query") {
		t.Fatalf("go_search required fields = %v", search.InputSchema.Required)
	}
	diagnostics := findTool(result.Tools, "go_diagnostics")
	if diagnostics == nil || diagnostics.InputSchema.Properties["files"] == nil {
		t.Fatal("go_diagnostics files schema is missing")
	}
}

func TestDiagnosticsRequiresRootOrFiles(t *testing.T) {
	handler := NewHandler(goplsrouter.New(workspace.NewResolver(), noSessionPool{}))
	result, err := handler.GoDiagnostics(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Arguments: map[string]any{}},
	})
	if err != nil {
		t.Fatalf("GoDiagnostics() error = %v", err)
	}
	if !result.IsError {
		t.Fatal("GoDiagnostics() accepted empty routing arguments")
	}
}

func findTool(tools []mcp.Tool, name string) *mcp.Tool {
	for index := range tools {
		if tools[index].Name == name {
			return &tools[index]
		}
	}
	return nil
}

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
