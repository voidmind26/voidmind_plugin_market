package gopls

import (
	"context"
	"fmt"
	"strings"

	goplsrouter "code-index-plugin/internal/gopls/router"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type Handler struct {
	router *goplsrouter.Router
}

func NewHandler(router *goplsrouter.Router) *Handler {
	return &Handler{router: router}
}

func (h *Handler) RegisterTools(server *mcpserver.MCPServer) {
	readOnly := []mcp.ToolOption{
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	}

	server.AddTool(mcp.NewTool(
		"list_go_workspaces",
		append(readOnly,
			mcp.WithDescription("发现父目录下可独立启动 gopls 的 go.work 或 go.mod 工作区，并报告会话状态"),
			mcp.WithString("project_root", mcp.Required(), mcp.Description("单仓目录或包含多个 Go 仓库的父目录")),
		)...,
	), h.ListGoWorkspaces)
	server.AddTool(mcp.NewTool(
		"go_workspace",
		append(readOnly,
			mcp.WithDescription("按 project_root 路由并汇总 gopls 的 Go workspace 摘要"),
			mcp.WithString("project_root", mcp.Required(), mcp.Description("单仓目录或包含多个 Go 仓库的父目录")),
		)...,
	), h.GoWorkspace)
	server.AddTool(mcp.NewTool(
		"go_search",
		append(readOnly,
			mcp.WithDescription("在一个或多个 Go workspace 中模糊搜索符号，并合并去重结果"),
			mcp.WithString("project_root", mcp.Required(), mcp.Description("单仓目录或包含多个 Go 仓库的父目录")),
			mcp.WithString("query", mcp.Required(), mcp.Description("符号模糊搜索词")),
		)...,
	), h.GoSearch)
	server.AddTool(mcp.NewTool(
		"go_file_context",
		append(readOnly,
			mcp.WithDescription("按绝对文件路径选择最近的 Go workspace，并总结文件的跨文件依赖"),
			mcp.WithString("file", mcp.Required(), mcp.Description("Go 文件绝对路径")),
		)...,
	), h.GoFileContext)
	server.AddTool(mcp.NewTool(
		"go_package_api",
		append(readOnly,
			mcp.WithDescription("在明确的单个 Go workspace 中查询包公开 API"),
			mcp.WithString("project_root", mcp.Required(), mcp.Description("具体 Go workspace 根目录")),
			mcp.WithArray("packagePaths", mcp.Required(), mcp.Description("要查询的 Go 包导入路径"), mcp.Items(map[string]any{"type": "string"})),
		)...,
	), h.GoPackageAPI)
	server.AddTool(mcp.NewTool(
		"go_symbol_references",
		append(readOnly,
			mcp.WithDescription("按文件所在 Go workspace 查询包级符号、字段或方法的引用"),
			mcp.WithString("file", mcp.Required(), mcp.Description("包含符号的 Go 文件绝对路径")),
			mcp.WithString("symbol", mcp.Required(), mcp.Description("符号或限定符号名")),
		)...,
	), h.GoSymbolReferences)
	server.AddTool(mcp.NewTool(
		"go_diagnostics",
		append(readOnly,
			mcp.WithDescription("按 project_root 或绝对文件列表路由并汇总 Go 诊断；多仓文件会自动分组"),
			mcp.WithString("project_root", mcp.Description("单仓目录或包含多个 Go 仓库的父目录；未传 files 时必填")),
			mcp.WithArray("files", mcp.Description("需要额外诊断的活动 Go 文件绝对路径"), mcp.Items(map[string]any{"type": "string"})),
		)...,
	), h.GoDiagnostics)
	server.AddTool(mcp.NewTool(
		"go_vulncheck",
		append(readOnly,
			mcp.WithDescription("在明确的单个 Go workspace 中执行依赖漏洞检查"),
			mcp.WithString("project_root", mcp.Required(), mcp.Description("具体 Go workspace 根目录")),
			mcp.WithString("dir", mcp.Description("workspace 内执行漏洞检查的目录")),
			mcp.WithString("pattern", mcp.Description("包模式，默认 ./...")),
		)...,
	), h.GoVulncheck)
	server.AddTool(mcp.NewTool(
		"go_rename_symbol",
		mcp.WithDescription("按文件所在 Go workspace 生成符号重命名编辑；不会自动应用编辑"),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithString("file", mcp.Required(), mcp.Description("包含符号的 Go 文件绝对路径")),
		mcp.WithString("symbol", mcp.Required(), mcp.Description("符号或限定符号名")),
		mcp.WithString("new_name", mcp.Required(), mcp.Description("新符号名")),
	), h.GoRenameSymbol)
}

func (h *Handler) ListGoWorkspaces(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectRoot, err := requiredString(request, "project_root")
	if err != nil {
		return toolError(err), nil
	}
	result, err := h.router.ListWorkspaces(projectRoot)
	if err != nil {
		return toolError(err), nil
	}
	return goplsrouter.JSONResult(result)
}

func (h *Handler) GoWorkspace(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectRoot, err := requiredString(request, "project_root")
	if err != nil {
		return toolError(err), nil
	}
	return h.callOrToolError(h.router.CallFanout(ctx, "go_workspace", projectRoot, request.GetArguments()))
}

func (h *Handler) GoSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectRoot, err := requiredString(request, "project_root")
	if err != nil {
		return toolError(err), nil
	}
	if _, err := requiredString(request, "query"); err != nil {
		return toolError(err), nil
	}
	return h.callOrToolError(h.router.CallFanout(ctx, "go_search", projectRoot, request.GetArguments()))
}

func (h *Handler) GoFileContext(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.callFile(ctx, "go_file_context", request)
}

func (h *Handler) GoPackageAPI(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectRoot, err := requiredString(request, "project_root")
	if err != nil {
		return toolError(err), nil
	}
	var params struct {
		PackagePaths []string `json:"packagePaths"`
	}
	if err := request.BindArguments(&params); err != nil || len(params.PackagePaths) == 0 {
		return toolError(fmt.Errorf("packagePaths 至少需要一个包路径")), nil
	}
	return h.callOrToolError(h.router.CallSingle(ctx, "go_package_api", projectRoot, request.GetArguments()))
}

func (h *Handler) GoSymbolReferences(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if _, err := requiredString(request, "symbol"); err != nil {
		return toolError(err), nil
	}
	return h.callFile(ctx, "go_symbol_references", request)
}

func (h *Handler) GoDiagnostics(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var params struct {
		ProjectRoot string   `json:"project_root"`
		Files       []string `json:"files"`
	}
	if err := request.BindArguments(&params); err != nil {
		return toolError(fmt.Errorf("解析 go_diagnostics 参数失败: %w", err)), nil
	}
	if strings.TrimSpace(params.ProjectRoot) == "" && len(params.Files) == 0 {
		return toolError(fmt.Errorf("project_root 和 files 至少需要提供一个")), nil
	}
	return h.callOrToolError(h.router.Diagnostics(ctx, params.ProjectRoot, params.Files, request.GetArguments()))
}

func (h *Handler) GoVulncheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectRoot, err := requiredString(request, "project_root")
	if err != nil {
		return toolError(err), nil
	}
	dir := request.GetString("dir", "")
	return h.callOrToolError(h.router.Vulncheck(ctx, projectRoot, dir, request.GetArguments()))
}

func (h *Handler) GoRenameSymbol(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	for _, name := range []string{"symbol", "new_name"} {
		if _, err := requiredString(request, name); err != nil {
			return toolError(err), nil
		}
	}
	return h.callFile(ctx, "go_rename_symbol", request)
}

func (h *Handler) callFile(ctx context.Context, toolName string, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := requiredString(request, "file")
	if err != nil {
		return toolError(err), nil
	}
	return h.callOrToolError(h.router.CallFile(ctx, toolName, file, request.GetArguments()))
}

func (h *Handler) callOrToolError(result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
	if err != nil {
		return toolError(err), nil
	}
	return result, nil
}

func requiredString(request mcp.CallToolRequest, name string) (string, error) {
	value, err := request.RequireString(name)
	if err != nil || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("参数 %s 必须是非空字符串", name)
	}
	return value, nil
}

func toolError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}
