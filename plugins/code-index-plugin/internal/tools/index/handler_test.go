package index

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"code-index-plugin/internal/index/scanner"
	indexservice "code-index-plugin/internal/index/service"
	"code-index-plugin/internal/index/storage"
	mcp "github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

func TestRegisterToolsUsesSingleSchemaDefinition(t *testing.T) {
	handler := NewHandler(nil)
	s := mcpserver.NewMCPServer("test-server", "1.0.0")
	handler.RegisterTools(s)

	resp := s.HandleMessage(context.Background(), []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`))
	jsonResp, ok := resp.(mcp.JSONRPCResponse)
	if !ok {
		t.Fatalf("expected JSONRPCResponse, got %T", resp)
	}

	result, ok := jsonResp.Result.(mcp.ListToolsResult)
	if !ok {
		t.Fatalf("expected ListToolsResult, got %T", jsonResp.Result)
	}

	if len(result.Tools) != 4 {
		t.Fatalf("expected 4 tools, got %d", len(result.Tools))
	}

	searchTool := findToolByName(t, result.Tools, "search_code_index")
	searchProp := mustToolProperty(t, searchTool, "limit")

	if got := searchProp["type"]; got != "number" {
		t.Fatalf("expected limit type number, got %#v", got)
	}
	if got := searchProp["minimum"]; got != 1.0 {
		t.Fatalf("expected minimum=1, got %#v", got)
	}
	if got := searchProp["maximum"]; got != 100.0 {
		t.Fatalf("expected maximum=100, got %#v", got)
	}
	if _, exists := searchProp["minLength"]; exists {
		t.Fatalf("did not expect minLength in number schema")
	}
	if _, exists := searchProp["maxLength"]; exists {
		t.Fatalf("did not expect maxLength in number schema")
	}
}

func TestDecodeSearchRequestRejectsFractionalLimit(t *testing.T) {
	_, err := decodeSearchRequest(mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: map[string]any{
				"query": "payment callback",
				"limit": 3.5,
			},
		},
	})
	if err == nil {
		t.Fatal("expected fractional limit to return error")
	}
}

func TestToolsReturnUnimplementedErrorWhenServiceMissing(t *testing.T) {
	handler := NewHandler(nil)
	ctx := context.Background()
	tests := []struct {
		name    string
		call    func() (*mcp.CallToolResult, error)
		message string
	}{
		{
			name: "build",
			call: func() (*mcp.CallToolResult, error) {
				return handler.BuildCodeIndex(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"project_root": "/tmp/project"}}})
			},
			message: "build_code_index 尚未实现：索引服务未接入",
		},
		{
			name: "refresh",
			call: func() (*mcp.CallToolResult, error) {
				return handler.RefreshCodeIndex(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"project_root": "/tmp/project"}}})
			},
			message: "refresh_code_index 尚未实现：索引服务未接入",
		},
		{
			name: "search",
			call: func() (*mcp.CallToolResult, error) {
				return handler.SearchCodeIndex(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"query": "payment"}}})
			},
			message: "search_code_index 尚未实现：索引服务未接入",
		},
		{
			name: "status",
			call: func() (*mcp.CallToolResult, error) {
				return handler.GetCodeIndexStatus(ctx, mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{"project_root": "/tmp/project"}}})
			},
			message: "get_code_index_status 尚未实现：索引服务未接入",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.call()
			if err != nil {
				t.Fatalf("unexpected handler error: %v", err)
			}
			if !result.IsError {
				t.Fatal("expected tool result to be marked as error")
			}
			if len(result.Content) != 1 {
				t.Fatalf("expected one content item, got %d", len(result.Content))
			}
			text, ok := mcp.AsTextContent(result.Content[0])
			if !ok {
				t.Fatalf("expected text content, got %T", result.Content[0])
			}
			if text.Text != tt.message {
				t.Fatalf("expected error message %q, got %q", tt.message, text.Text)
			}
		})
	}
}

func TestSearchAndStatusUseRealServiceResults(t *testing.T) {
	root := t.TempDir()
	mustWriteProjectFile(t, filepath.Join(root, "service", "payment.go"), "package service\n\nfunc HandlePaymentCallback() error {\n\treturn nil\n}\n")
	mustWriteProjectFile(t, filepath.Join(root, "docs", "payment.md"), "# payment callback\n")

	svc := indexservice.New(storage.New(), scanner.New(indexservice.DefaultOptions()))
	if _, err := svc.Build(context.Background(), indexservice.BuildRequest{ProjectRoot: root}); err != nil {
		t.Fatalf("Build returned error: %v", err)
	}

	handler := NewHandler(svc)

	searchResult, err := handler.SearchCodeIndex(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"query":            "payment callback",
		"project_root":     root,
		"prefer_deep_hits": true,
		"limit":            5,
	}}})
	if err != nil {
		t.Fatalf("SearchCodeIndex returned error: %v", err)
	}
	if searchResult.IsError {
		t.Fatalf("expected search result not to be error: %+v", searchResult)
	}
	searchText, ok := mcp.AsTextContent(searchResult.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", searchResult.Content[0])
	}
	var searchPayload SearchResult
	if err := json.Unmarshal([]byte(searchText.Text), &searchPayload); err != nil {
		t.Fatalf("unmarshal search payload: %v", err)
	}
	if searchPayload.ProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, searchPayload.ProjectRoot)
	}
	if searchPayload.ResultCount == 0 || len(searchPayload.Results) == 0 {
		t.Fatalf("expected structured search hits, got %+v", searchPayload)
	}
	if searchPayload.Results[0].Kind != "symbol" {
		t.Fatalf("expected first hit to prefer deep symbol result, got %+v", searchPayload.Results[0])
	}

	statusResult, err := handler.GetCodeIndexStatus(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"project_root": root,
	}}})
	if err != nil {
		t.Fatalf("GetCodeIndexStatus returned error: %v", err)
	}
	if statusResult.IsError {
		t.Fatalf("expected status result not to be error: %+v", statusResult)
	}
	statusText, ok := mcp.AsTextContent(statusResult.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", statusResult.Content[0])
	}
	var statusPayload StatusResult
	if err := json.Unmarshal([]byte(statusText.Text), &statusPayload); err != nil {
		t.Fatalf("unmarshal status payload: %v", err)
	}
	if !statusPayload.Ready || statusPayload.FileCount == 0 {
		t.Fatalf("expected ready status with counts, got %+v", statusPayload)
	}
}

func TestStatusReturnsNotReadyWhenIndexMissing(t *testing.T) {
	root := t.TempDir()
	handler := NewHandler(indexservice.New(storage.New(), scanner.New(indexservice.DefaultOptions())))

	statusResult, err := handler.GetCodeIndexStatus(context.Background(), mcp.CallToolRequest{Params: mcp.CallToolParams{Arguments: map[string]any{
		"project_root": root,
	}}})
	if err != nil {
		t.Fatalf("GetCodeIndexStatus returned error: %v", err)
	}
	if statusResult.IsError {
		t.Fatalf("expected status result not to be error when index is missing: %+v", statusResult)
	}
	statusText, ok := mcp.AsTextContent(statusResult.Content[0])
	if !ok {
		t.Fatalf("expected text content, got %T", statusResult.Content[0])
	}
	var statusPayload StatusResult
	if err := json.Unmarshal([]byte(statusText.Text), &statusPayload); err != nil {
		t.Fatalf("unmarshal status payload: %v", err)
	}
	if statusPayload.Ready {
		t.Fatalf("expected ready=false when index is missing, got %+v", statusPayload)
	}
	if statusPayload.ProjectRoot != root {
		t.Fatalf("expected project root %q, got %q", root, statusPayload.ProjectRoot)
	}
}

func findToolByName(t *testing.T, tools []mcp.Tool, name string) mcp.Tool {
	t.Helper()
	for _, tool := range tools {
		if tool.Name == name {
			return tool
		}
	}
	t.Fatalf("tool %q not found", name)
	return mcp.Tool{}
}

func mustToolProperty(t *testing.T, tool mcp.Tool, key string) map[string]any {
	t.Helper()
	data, err := json.Marshal(tool)
	if err != nil {
		t.Fatalf("marshal tool: %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal tool: %v", err)
	}
	inputSchema, ok := payload["inputSchema"].(map[string]any)
	if !ok {
		t.Fatalf("inputSchema missing or invalid: %#v", payload["inputSchema"])
	}
	properties, ok := inputSchema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties missing or invalid: %#v", inputSchema["properties"])
	}
	prop, ok := properties[key].(map[string]any)
	if !ok {
		t.Fatalf("property %q missing or invalid: %#v", key, properties[key])
	}
	return prop
}

func mustWriteProjectFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}
}
