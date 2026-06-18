package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"local-db-access/internal/config"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPHandler MCP处理器
type MCPHandler struct {
	dbManager *DatabaseManager
}

func NewMCPHandler(dbManager *DatabaseManager) *MCPHandler {
	return &MCPHandler{dbManager: dbManager}
}

func (h *MCPHandler) RegisterTools(s *server.MCPServer) {
	queryTool := mcp.NewTool("execute_query",
		mcp.WithDescription("执行只读SQL查询（SELECT、SHOW、DESCRIBE、EXPLAIN）"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_name", mcp.Required(), mcp.Description("数据库连接别名")),
		mcp.WithString("sql", mcp.Required(), mcp.Description("SQL查询语句")),
		mcp.WithNumber("limit", mcp.Description("限制返回的行数（0表示无限制，默认100）"), mcp.MinLength(0), mcp.MaxLength(10000)),
		mcp.WithArray("params",
			mcp.Description("SQL参数化查询的参数列表"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	)
	writeTool := mcp.NewTool("execute_write_query",
		mcp.WithDescription("执行测试数据库受限写入 SQL（支持 INSERT、UPDATE、CREATE、ALTER，禁用 DELETE、DROP、TRUNCATE）"),
		mcp.WithString("database_name", mcp.Required(), mcp.Description("数据库连接别名")),
		mcp.WithString("sql", mcp.Required(), mcp.Description("单条 SQL 语句")),
		mcp.WithArray("params",
			mcp.Description("SQL参数化查询的参数列表"),
			mcp.Items(map[string]any{"type": "string"}),
		),
	)
	listTool := mcp.NewTool("list_databases",
		mcp.WithDescription("列出全部已配置的数据库连接（包含未启用的），以及每个连接下的全部表名"),
		mcp.WithReadOnlyHintAnnotation(true),
	)
	infoTool := mcp.NewTool("get_database_info",
		mcp.WithDescription("获取MySQL数据库的基本信息"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_name", mcp.Required(), mcp.Description("数据库连接别名")),
	)
	describeTool := mcp.NewTool("describe_table",
		mcp.WithDescription("查看MySQL表的列信息"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_name", mcp.Required(), mcp.Description("数据库连接别名")),
		mcp.WithString("table_name", mcp.Required(), mcp.Description("表名")),
	)
	tableDataTool := mcp.NewTool("get_table_data",
		mcp.WithDescription("获取表数据（带分页）"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_name", mcp.Required(), mcp.Description("数据库连接别名")),
		mcp.WithString("table_name", mcp.Required(), mcp.Description("表名")),
		mcp.WithNumber("limit", mcp.Description("限制返回的行数（默认100）"), mcp.MinLength(1), mcp.MaxLength(1000)),
	)
	healthTool := mcp.NewTool("health_check",
		mcp.WithDescription("检查数据库连接状态"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithString("database_name", mcp.Description("数据库连接别名（为空则检查全部）")),
	)
	reloadTool := mcp.NewTool("reload_config",
		mcp.WithDescription("从磁盘重新读取配置文件并热更新数据库连接，无需重启服务"),
	)
	initTool, err := newInitDatabasesTool()
	if err != nil {
		panic(err)
	}

	s.AddTool(queryTool, h.handleExecuteQuery)
	s.AddTool(writeTool, h.handleExecuteWriteQuery)
	s.AddTool(listTool, h.handleListDatabases)
	s.AddTool(infoTool, h.handleGetDatabaseInfo)
	s.AddTool(describeTool, h.handleDescribeTable)
	s.AddTool(tableDataTool, h.handleGetTableData)
	s.AddTool(healthTool, h.handleHealthCheck)
	s.AddTool(reloadTool, h.handleReloadConfig)
	s.AddTool(initTool, h.handleInitDatabases)
}

func newInitDatabasesTool() (mcp.Tool, error) {
	schema := json.RawMessage(`{
  "type": "object",
  "properties": {
    "overwrite": {"type": "boolean", "description": "是否覆盖现有插件内置配置"},
    "default_database": {"type": "string", "description": "默认数据库别名,可选"},
    "databases": {
      "type": "object",
      "description": "数据库连接映射,key 为连接别名",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "type": {"type": "string"},
          "enabled": {"type": "boolean"},
          "host": {"type": "string"},
          "port": {"type": "integer"},
          "user": {"type": "string"},
          "password": {"type": "string"},
          "database": {"type": "string"},
          "charset": {"type": "string"},
          "timeout": {"type": "integer"}
        },
        "required": ["type", "enabled", "host", "port", "user", "password", "database"]
      }
    }
  },
  "required": ["overwrite", "databases"]
}`)
	return mcp.NewToolWithRawSchema("init_databases", "初始化或覆盖插件内部数据库连接配置（写入 internal/config/config.yaml，之后执行 reload_config 热加载无需重启）", schema), nil
}

func (h *MCPHandler) handleExecuteQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, err := request.RequireString("database_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sqlStr, err := request.RequireString("sql")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	limit := 100
	if limitVal, reqErr := request.RequireFloat("limit"); reqErr == nil && limitVal > 0 {
		limit = int(limitVal)
	}
	var params []interface{}
	if paramsVal, reqErr := request.RequireStringSlice("params"); reqErr == nil {
		for _, param := range paramsVal {
			params = append(params, param)
		}
	}

	dbTool, err := h.dbManager.GetDatabase(dbName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
	}
	result, err := dbTool.ExecuteQuery(ctx, sqlStr, params, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("查询执行失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func (h *MCPHandler) handleExecuteWriteQuery(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, err := request.RequireString("database_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	sqlStr, err := request.RequireString("sql")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	var params []interface{}
	if paramsVal, reqErr := request.RequireStringSlice("params"); reqErr == nil {
		for _, param := range paramsVal {
			params = append(params, param)
		}
	}

	dbTool, err := h.dbManager.GetDatabase(dbName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
	}
	result, err := dbTool.ExecuteWriteQuery(ctx, sqlStr, params)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("写入执行失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

// listDatabasesEntry 列库工具的单条输出
type listDatabasesEntry struct {
	Alias       string   `json:"alias"`
	Host        string   `json:"host"`
	Port        int      `json:"port"`
	Database    string   `json:"database"`
	Enabled     bool     `json:"enabled"`
	IsDefault   bool     `json:"is_default,omitempty"`
	TableCount  int      `json:"table_count"`
	Tables      []string `json:"tables"`
	ErrorReason string   `json:"error,omitempty"`
}

func (h *MCPHandler) handleListDatabases(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 从配置源获取全部连接（包含未启用的）
	allAliases := h.dbManager.AllAliases()
	entries := make([]listDatabasesEntry, 0, len(allAliases))

	for _, alias := range allAliases {
		cfg, _ := h.dbManager.GetConfig(alias)
		entry := listDatabasesEntry{Alias: alias, IsDefault: alias == config.Conf.DefaultDatabase}
		if cfg != nil {
			entry.Host = cfg.Host
			entry.Port = cfg.Port
			entry.Database = cfg.Database
			entry.Enabled = cfg.Enabled
		}
		// 只有启用的连接才尝试获取表列表
		if cfg != nil && cfg.Enabled {
			tables, err := h.dbManager.ListTables(ctx, alias)
			if err != nil {
				entry.ErrorReason = err.Error()
			} else {
				entry.Tables = tables
				entry.TableCount = len(tables)
			}
		}
		entries = append(entries, entry)
	}

	result := map[string]interface{}{
		"default_database": config.Conf.DefaultDatabase,
		"count":            len(entries),
		"databases":        entries,
		"server_time":      time.Now().Format(time.RFC3339),
	}
	return toolResultJSON(result)
}

func (h *MCPHandler) handleGetDatabaseInfo(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, err := request.RequireString("database_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dbTool, err := h.dbManager.GetDatabase(dbName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
	}
	info, err := dbTool.GetDatabaseInfo(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库信息失败: %v", err)), nil
	}
	return toolResultJSON(info)
}

func (h *MCPHandler) handleDescribeTable(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, err := request.RequireString("database_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tableName, err := request.RequireString("table_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	dbTool, err := h.dbManager.GetDatabase(dbName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
	}
	info, err := dbTool.GetTableInfo(ctx, tableName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取表信息失败: %v", err)), nil
	}
	return toolResultJSON(info)
}

func (h *MCPHandler) handleGetTableData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, err := request.RequireString("database_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	tableName, err := request.RequireString("table_name")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	limit := 100
	if limitVal, reqErr := request.RequireFloat("limit"); reqErr == nil && limitVal > 0 {
		limit = int(limitVal)
	}
	dbTool, err := h.dbManager.GetDatabase(dbName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
	}
	result, err := dbTool.GetTableData(ctx, tableName, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("获取表数据失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func (h *MCPHandler) handleHealthCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dbName, _ := request.RequireString("database_name")
	var results map[string]error
	if dbName == "" {
		results = h.dbManager.HealthCheckAll(ctx)
	} else {
		dbTool, err := h.dbManager.GetDatabase(dbName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("获取数据库失败: %v", err)), nil
		}
		results = map[string]error{dbName: dbTool.HealthCheck(ctx)}
	}

	healthStatus := make(map[string]string)
	for name, err := range results {
		if err == nil {
			healthStatus[name] = "healthy"
		} else {
			healthStatus[name] = fmt.Sprintf("unhealthy: %v", err)
		}
	}

	result := map[string]interface{}{
		"timestamp":    time.Now().Format(time.RFC3339),
		"health_check": healthStatus,
		"total":        len(results),
		"healthy":      countHealthy(results),
		"unhealthy":    len(results) - countHealthy(results),
	}
	return toolResultJSON(result)
}

func (h *MCPHandler) handleInitDatabases(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var input config.InitDatabasesInput
	if err := request.BindArguments(&input); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("解析初始化参数失败: %v", err)), nil
	}
	result, err := config.WriteInternalConfig(&input)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("写入数据库配置失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func (h *MCPHandler) handleReloadConfig(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := h.dbManager.ReloadConfig()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("配置重载失败: %v", err)), nil
	}
	return toolResultJSON(result)
}

func toolResultJSON(v interface{}) (*mcp.CallToolResult, error) {
	jsonResult, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("结果格式化失败: %v", err)), nil
	}
	return mcp.NewToolResultText(string(jsonResult)), nil
}
