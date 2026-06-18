package db

import (
	"context"
	"local-db-access/internal/common"
	"local-db-access/internal/dto"
)

// DatabaseQueryTool 数据库查询工具接口
type DatabaseQueryTool interface {
	Type() common.DatabaseType
	Name() string
	IsReadOnlyQuery(sql string) bool
	ExecuteQuery(ctx context.Context, sql string, params []interface{}, limit int) (*dto.QueryResult, error)
	ExecuteWriteQuery(ctx context.Context, sql string, params []interface{}) (*dto.WriteResult, error)
	GetDatabaseInfo(ctx context.Context) (*dto.DatabaseInfo, error)
	GetTableInfo(ctx context.Context, tableName string) (*dto.TableInfo, error)
	GetTableData(ctx context.Context, tableName string, limit int) (*dto.QueryResult, error)
	HealthCheck(ctx context.Context) error
	ConnectionInfo() *dto.DatabaseConnectionInfo
	Close() error
}
