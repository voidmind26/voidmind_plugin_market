package codeindexplugin_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGoplsMCPSmoke(t *testing.T) {
	if os.Getenv("CODE_INDEX_GOPLS_SMOKE") != "1" {
		t.Skip("set CODE_INDEX_GOPLS_SMOKE=1 to run the real gopls MCP smoke test")
	}
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Fatalf("gopls is not available on PATH: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	c, err := client.NewStdioMCPClient("gopls", nil, "mcp")
	if err != nil {
		t.Fatalf("start gopls MCP: %v", err)
	}
	defer c.Close()

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "code-index-smoke", Version: "0.2.0"}
	if _, err := c.Initialize(ctx, initRequest); err != nil {
		t.Fatalf("initialize gopls MCP: %v", err)
	}

	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list gopls tools: %v", err)
	}
	available := make(map[string]bool, len(tools.Tools))
	for _, tool := range tools.Tools {
		available[tool.Name] = true
	}
	for _, name := range []string{
		"go_workspace",
		"go_search",
		"go_file_context",
		"go_package_api",
		"go_symbol_references",
		"go_diagnostics",
		"go_vulncheck",
		"go_rename_symbol",
	} {
		if !available[name] {
			t.Errorf("gopls MCP tool %q is missing", name)
		}
	}

	workspaceRequest := mcp.CallToolRequest{}
	workspaceRequest.Params.Name = "go_workspace"
	workspaceRequest.Params.Arguments = map[string]any{}
	workspace, err := c.CallTool(ctx, workspaceRequest)
	if err != nil {
		t.Fatalf("call go_workspace: %v", err)
	}
	if workspace.IsError {
		t.Fatalf("go_workspace returned an error: %+v", workspace.Content)
	}
	if len(workspace.Content) == 0 {
		t.Fatal("go_workspace returned no content")
	}
	content, ok := mcp.AsTextContent(workspace.Content[0])
	if !ok {
		t.Fatalf("go_workspace returned unexpected content: %+v", workspace.Content)
	}
	if !strings.Contains(content.Text, "code-index-plugin") {
		t.Fatalf("go_workspace did not identify the plugin test module: %s", content.Text)
	}
	t.Logf("go_workspace: %s", content.Text)
}
