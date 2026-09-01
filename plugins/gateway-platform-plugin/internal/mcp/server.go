package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type ToolServer struct {
	client     *Client
	consoleURL string
	runtime    Runtime
}

func NewToolServer(client *Client, consoleURL string) *ToolServer {
	return &ToolServer{client: client, consoleURL: consoleURL}
}

func NewToolServerWithRuntime(client *Client, consoleURL string, runtime Runtime) *ToolServer {
	return &ToolServer{client: client, consoleURL: consoleURL, runtime: runtime}
}

func (s *ToolServer) Build() *mcpserver.MCPServer {
	server := mcpserver.NewMCPServer(
		"gateway-platform",
		"0.1.7",
		mcpserver.WithToolCapabilities(true),
	)
	registerHealthTool(server, s)
	registerConsoleInfoTool(server, s)
	registerListRoutesTool(server, s)
	registerListKeysTool(server, s)
	registerListReferencesTool(server, s)
	return server
}

func (s *ToolServer) Health(ctx context.Context) (map[string]any, error) {
	if s.runtime.Client != nil {
		if err := s.runtime.EnsurePlatform(ctx); err != nil {
			return nil, err
		}
	}
	status, err := s.client.Health()
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"ok":                status.OK,
		"console_url":       s.consoleURL,
		"data_dir":          status.DataDir,
		"database_path":     status.DatabasePath,
		"database_writable": status.DatabaseWritable,
		"version":           status.Version,
	}, nil
}

func registerHealthTool(server *mcpserver.MCPServer, tools *ToolServer) {
	server.AddTool(mcp.NewTool("health_check", mcp.WithDescription("检查 gateway platform 本地平台健康状态")), func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		body, err := tools.Health(ctx)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mustJSONResult(body), nil
	})
}

func registerConsoleInfoTool(server *mcpserver.MCPServer, tools *ToolServer) {
	server.AddTool(mcp.NewTool("open_console_info", mcp.WithDescription("返回 gateway platform Web Console 地址")), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return mustJSONResult(map[string]any{"console_url": tools.consoleURL}), nil
	})
}

func registerListRoutesTool(server *mcpserver.MCPServer, tools *ToolServer) {
	server.AddTool(mcp.NewTool("list_routes", mcp.WithDescription("列出本地 gateway routes")), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		routes, err := tools.client.ListRoutes()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mustJSONResult(routes), nil
	})
}

func registerListKeysTool(server *mcpserver.MCPServer, tools *ToolServer) {
	server.AddTool(mcp.NewTool("list_keys", mcp.WithDescription("列出 key 配置，敏感值已脱敏")), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		keys, err := tools.client.ListKeys()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mustJSONResult(keys), nil
	})
}

func registerListReferencesTool(server *mcpserver.MCPServer, tools *ToolServer) {
	server.AddTool(mcp.NewTool("list_references", mcp.WithDescription("列出缺失引用与未使用 key")), func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		report, err := tools.client.ListReferences()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mustJSONResult(report), nil
	})
}

func mustJSONResult(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return mcp.NewToolResultError(err.Error())
	}
	return mcp.NewToolResultText(string(data))
}
