package index

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	indexservice "code-index-plugin/internal/index/service"
	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterTools(s *mcpserver.MCPServer) {
	s.AddTool(mcp.NewTool(
		"build_code_index",
		mcp.WithDescription("构建当前项目代码索引"),
		mcp.WithString("project_root", mcp.Description("要构建索引的项目根目录，默认当前工作目录")),
		mcp.WithArray("deep_index_paths",
			mcp.Description("需要深度索引的相对路径列表"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	), h.BuildCodeIndex)
	s.AddTool(mcp.NewTool(
		"refresh_code_index",
		mcp.WithDescription("增量刷新当前项目代码索引"),
		mcp.WithString("project_root", mcp.Description("要刷新索引的项目根目录，默认当前工作目录")),
		mcp.WithArray("deep_index_paths",
			mcp.Description("需要深度索引的相对路径列表"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	), h.RefreshCodeIndex)
	s.AddTool(mcp.NewTool(
		"search_code_index",
		mcp.WithDescription("搜索当前项目代码索引"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("query", mcp.Required(), mcp.Description("搜索关键词或短语")),
		mcp.WithString("project_root", mcp.Description("要搜索索引的项目根目录，默认当前工作目录")),
		mcp.WithString("path_prefix", mcp.Description("只在指定相对路径前缀下搜索")),
		mcp.WithBoolean("prefer_deep_hits", mcp.Description("是否优先返回 symbol/chunk 命中")),
		mcp.WithNumber("limit", mcp.Description("返回结果数量上限，默认 10"), mcp.Min(1), mcp.Max(100), mcp.MultipleOf(1)),
	), h.SearchCodeIndex)
	s.AddTool(mcp.NewTool(
		"get_code_index_status",
		mcp.WithDescription("获取当前项目代码索引状态"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("project_root", mcp.Description("要查询索引状态的项目根目录，默认当前工作目录")),
	), h.GetCodeIndexStatus)
}

func (h *Handler) BuildCodeIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parsed := decodeBuildRequest(req)
	if h.svc == nil {
		return toolUnimplementedError("build_code_index"), nil
	}
	result, err := h.svc.Build(ctx, parsed)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("构建代码索引失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func (h *Handler) RefreshCodeIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parsed := decodeRefreshRequest(req)
	if h.svc == nil {
		return toolUnimplementedError("refresh_code_index"), nil
	}
	result, err := h.svc.Refresh(ctx, parsed)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("刷新代码索引失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func (h *Handler) SearchCodeIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parsed, err := decodeSearchRequest(req)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	if h.svc == nil {
		return toolUnimplementedError("search_code_index"), nil
	}
	result, svcErr := h.svc.Search(ctx, indexservice.SearchRequest{
		Query:          parsed.Query,
		ProjectRoot:    parsed.ProjectRoot,
		Limit:          parsed.Limit,
		PreferDeepHits: parsed.PreferDeepHits,
		PathPrefix:     parsed.PathPrefix,
	})
	if svcErr != nil {
		return mcp.NewToolResultError(fmt.Sprintf("搜索代码索引失败: %v", svcErr)), nil
	}
	return toolResultJSON(result)
}

func (h *Handler) GetCodeIndexStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	parsed := decodeStatusRequest(req)
	if h.svc == nil {
		return toolUnimplementedError("get_code_index_status"), nil
	}
	result, err := h.svc.Status(ctx, indexservice.StatusRequest{ProjectRoot: parsed.ProjectRoot})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取代码索引状态失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func decodeBuildRequest(req mcp.CallToolRequest) BuildRequest {
	projectRoot, _ := req.RequireString("project_root")
	deepIndexPaths, _ := req.RequireStringSlice("deep_index_paths")
	return BuildRequest{
		ProjectRoot:    projectRoot,
		DeepIndexPaths: deepIndexPaths,
	}
}

func decodeRefreshRequest(req mcp.CallToolRequest) RefreshRequest {
	projectRoot, _ := req.RequireString("project_root")
	deepIndexPaths, _ := req.RequireStringSlice("deep_index_paths")
	return RefreshRequest{
		ProjectRoot:    projectRoot,
		DeepIndexPaths: deepIndexPaths,
	}
}

func decodeSearchRequest(req mcp.CallToolRequest) (SearchRequest, error) {
	query, err := req.RequireString("query")
	if err != nil {
		return SearchRequest{}, fmt.Errorf("缺少 query 参数: %w", err)
	}
	projectRoot, _ := req.RequireString("project_root")
	pathPrefix, _ := req.RequireString("path_prefix")
	preferDeepHits := false
	if rawPrefer, exists := req.GetArguments()["prefer_deep_hits"]; exists {
		flag, ok := rawPrefer.(bool)
		if !ok {
			return SearchRequest{}, fmt.Errorf("prefer_deep_hits 参数必须是布尔值")
		}
		preferDeepHits = flag
	}
	limit := 10
	if rawLimit, exists := req.GetArguments()["limit"]; exists {
		parsedLimit, parseErr := parseLimit(rawLimit)
		if parseErr != nil {
			return SearchRequest{}, parseErr
		}
		limit = parsedLimit
	}
	return SearchRequest{
		Query:          query,
		ProjectRoot:    projectRoot,
		Limit:          limit,
		PreferDeepHits: preferDeepHits,
		PathPrefix:     pathPrefix,
	}, nil
}

func decodeStatusRequest(req mcp.CallToolRequest) StatusRequest {
	projectRoot, _ := req.RequireString("project_root")
	return StatusRequest{ProjectRoot: projectRoot}
}

func toolResultJSON(payload any) (*mcp.CallToolResult, error) {
	jsonResult, err := json.Marshal(payload)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("结果格式化失败: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}

func toolUnimplementedError(toolName string) *mcp.CallToolResult {
	return mcp.NewToolResultError(fmt.Sprintf("%s 尚未实现：索引服务未接入", toolName))
}

func parseLimit(raw any) (int, error) {
	switch v := raw.(type) {
	case int:
		return validateLimit(v)
	case int8:
		return validateLimit(int(v))
	case int16:
		return validateLimit(int(v))
	case int32:
		return validateLimit(int(v))
	case int64:
		return validateLimit(int(v))
	case uint:
		return validateLimit(int(v))
	case uint8:
		return validateLimit(int(v))
	case uint16:
		return validateLimit(int(v))
	case uint32:
		return validateLimit(int(v))
	case uint64:
		return validateLimit(int(v))
	case float32:
		return parseFloatLimit(float64(v))
	case float64:
		return parseFloatLimit(v)
	case string:
		parsedFloat, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("limit 参数必须是 1 到 100 的整数")
		}
		return parseFloatLimit(parsedFloat)
	default:
		return 0, fmt.Errorf("limit 参数必须是 1 到 100 的整数")
	}
}

func parseFloatLimit(limit float64) (int, error) {
	if math.Trunc(limit) != limit {
		return 0, fmt.Errorf("limit 参数必须是整数，不能为小数")
	}
	return validateLimit(int(limit))
}

func validateLimit(limit int) (int, error) {
	if limit < 1 || limit > 100 {
		return 0, fmt.Errorf("limit 参数必须在 1 到 100 之间")
	}
	return limit, nil
}
