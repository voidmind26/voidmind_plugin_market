package codeindexplugin_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestGoplsRouterSmoke(t *testing.T) {
	if os.Getenv("CODE_INDEX_GOPLS_SMOKE") != "1" {
		t.Skip("set CODE_INDEX_GOPLS_SMOKE=1 to run the real gopls router smoke test")
	}
	if _, err := exec.LookPath("gopls"); err != nil {
		t.Fatalf("gopls is not available on PATH: %v", err)
	}

	projectRoot := os.Getenv("CODE_INDEX_GOPLS_SMOKE_ROOT")
	expectGoWork := os.Getenv("CODE_INDEX_GOPLS_SMOKE_GOWORK") == "1"
	if expectGoWork {
		projectRoot = createSmokeGoWork(t)
	} else if projectRoot == "" {
		var err error
		projectRoot, err = os.Getwd()
		if err != nil {
			t.Fatalf("get plugin root: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	command := "go"
	arguments := []string{"run", "./cmd/code-index-mcp"}
	env := append(os.Environ(), "GOWORK=off")
	if os.Getenv("CODE_INDEX_GOPLS_SMOKE_BINARY") == "1" {
		command = "./bin/code-index-mcp"
		arguments = nil
		env = nil
	}
	c, err := client.NewStdioMCPClient(command, env, arguments...)
	if err != nil {
		t.Fatalf("start code-index MCP: %v", err)
	}
	defer c.Close()

	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{Name: "code-index-smoke", Version: "0.2.0"}
	if _, err := c.Initialize(ctx, initRequest); err != nil {
		t.Fatalf("initialize code-index MCP: %v", err)
	}

	tools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		t.Fatalf("list routed tools: %v", err)
	}
	available := make(map[string]bool, len(tools.Tools))
	for _, tool := range tools.Tools {
		available[tool.Name] = true
	}
	for _, name := range []string{
		"build_code_index",
		"refresh_code_index",
		"search_code_index",
		"get_code_index_status",
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
		if !available[name] {
			t.Errorf("routed MCP tool %q is missing", name)
		}
	}

	listResult := callTool(t, ctx, c, "list_go_workspaces", map[string]any{"project_root": projectRoot})
	var listed struct {
		Count      int `json:"count"`
		Workspaces []struct {
			Root          string `json:"root"`
			SessionActive bool   `json:"session_active"`
		} `json:"workspaces"`
	}
	if err := json.Unmarshal([]byte(resultText(t, listResult)), &listed); err != nil {
		t.Fatalf("decode list_go_workspaces result: %v", err)
	}
	if listed.Count == 0 || len(listed.Workspaces) == 0 {
		t.Fatalf("no Go workspaces discovered under %s", projectRoot)
	}
	if expected := os.Getenv("CODE_INDEX_GOPLS_SMOKE_WORKSPACES"); expected != "" {
		want, err := strconv.Atoi(expected)
		if err != nil {
			t.Fatalf("invalid CODE_INDEX_GOPLS_SMOKE_WORKSPACES: %v", err)
		}
		if listed.Count != want {
			t.Fatalf("workspace count = %d, want %d", listed.Count, want)
		}
	}

	workspaceResult := callTool(t, ctx, c, "go_workspace", map[string]any{
		"project_root": listed.Workspaces[0].Root,
	})
	workspaceText := resultText(t, workspaceResult)
	if !strings.Contains(workspaceText, listed.Workspaces[0].Root) {
		t.Fatalf("go_workspace did not identify routed root %s: %s", listed.Workspaces[0].Root, workspaceText)
	}
	if expectGoWork && !strings.Contains(workspaceText, "go workspace defined by") {
		t.Fatalf("go_workspace ignored the routed go.work: %s", workspaceText)
	}
	if os.Getenv("CODE_INDEX_GOPLS_SMOKE_FANOUT") == "1" {
		callTool(t, ctx, c, "go_search", map[string]any{
			"project_root": projectRoot,
			"query":        "main",
		})
		activeResult := callTool(t, ctx, c, "list_go_workspaces", map[string]any{"project_root": projectRoot})
		if err := json.Unmarshal([]byte(resultText(t, activeResult)), &listed); err != nil {
			t.Fatalf("decode active workspace list: %v", err)
		}
		for _, ws := range listed.Workspaces {
			if !ws.SessionActive {
				t.Errorf("gopls session was not activated for %s", ws.Root)
			}
		}
	}
	t.Logf("discovered %d workspaces under %s; routed gopls to %s", listed.Count, projectRoot, listed.Workspaces[0].Root)
}

func createSmokeGoWork(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	module := filepath.Join(root, "module")
	if err := os.MkdirAll(module, 0o755); err != nil {
		t.Fatalf("create smoke module: %v", err)
	}
	files := map[string]string{
		filepath.Join(root, "go.work"):   "go 1.24.2\n\nuse ./module\n",
		filepath.Join(module, "go.mod"):  "module example.com/smoke\n\ngo 1.24.2\n",
		filepath.Join(module, "main.go"): "package smoke\n\nfunc Ready() bool { return true }\n",
	}
	for path, content := range files {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write smoke workspace file %s: %v", path, err)
		}
	}
	return root
}

func callTool(t *testing.T, ctx context.Context, c *client.Client, name string, arguments map[string]any) *mcp.CallToolResult {
	t.Helper()
	request := mcp.CallToolRequest{}
	request.Params.Name = name
	request.Params.Arguments = arguments
	result, err := c.CallTool(ctx, request)
	if err != nil {
		t.Fatalf("call %s: %v", name, err)
	}
	if result.IsError {
		t.Fatalf("%s returned an error: %s", name, resultText(t, result))
	}
	return result
}

func resultText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("tool result has no content")
	}
	content, ok := mcp.AsTextContent(result.Content[0])
	if !ok {
		t.Fatalf("tool returned unexpected content: %+v", result.Content)
	}
	return content.Text
}
