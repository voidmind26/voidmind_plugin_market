package server

import (
	"fmt"

	"local-db-access/internal/config"
	dbtool "local-db-access/internal/tools/db"

	"github.com/mark3labs/mcp-go/server"
)

func New() (*server.MCPServer, error) {
	dbManager := dbtool.NewDatabaseManager()
	for alias, dbConfig := range config.Conf.Databases {
		if !dbConfig.Enabled {
			dbManager.AddConfigOnly(alias, dbConfig)
			continue
		}
		if err := dbManager.AddDatabase(alias, dbConfig); err != nil {
			return nil, fmt.Errorf("添加数据库 %s 失败: %w", alias, err)
		}
	}

	s := server.NewMCPServer(
		"测试环境数据库只读查询工具",
		"0.3.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	handler := dbtool.NewMCPHandler(dbManager)
	handler.RegisterTools(s)
	return s, nil
}
