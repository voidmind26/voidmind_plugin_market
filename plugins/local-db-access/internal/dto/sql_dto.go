package dto

import (
	"time"
	"local-db-access/internal/common"
)

// QueryResult 查询结果
type QueryResult struct {
	Success      bool                     `json:"success"`
	Database     string                   `json:"sql"`
	DatabaseName string                   `json:"database_name"`
	QueryTime    int64                    `json:"query_time_ms"`
	RowCount     int                      `json:"row_count"`
	TotalRows    int                      `json:"total_rows,omitempty"`
	Columns      []string                 `json:"columns"`
	Data         []map[string]interface{} `json:"data"`
	Limited      bool                     `json:"limited"`
	ErrorMessage string                   `json:"error_message,omitempty"`
	QueryID      string                   `json:"query_id,omitempty"`
}

// WriteResult 受限写入结果
 type WriteResult struct {
	Success      bool   `json:"success"`
	DatabaseName string `json:"database_name"`
	SQLType      string `json:"sql_type"`
	RowsAffected int64  `json:"rows_affected"`
	LastInsertID int64  `json:"last_insert_id,omitempty"`
	Message      string `json:"message"`
	ExecutedSQL  string `json:"sql"`
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Type            common.DatabaseType    `json:"type"`
	Version         string                 `json:"version"`
	CurrentDatabase string                 `json:"current_database"`
	Host            string                 `json:"host"`
	Port            int                    `json:"port"`
	User            string                 `json:"user"`
	TableCount      int                    `json:"table_count"`
	Tables          []TableSummary         `json:"tables"`
	SizeInfo        map[string]interface{} `json:"size_info,omitempty"`
	Status          string                 `json:"status"`
	LastPing        time.Time              `json:"last_ping"`
}

// TableSummary 表摘要信息
type TableSummary struct {
	Name       string                 `json:"name"`
	RowCount   int64                  `json:"row_count,omitempty"`
	SizeMB     float64                `json:"size_mb,omitempty"`
	Engine     string                 `json:"engine,omitempty"`
	CreateTime *time.Time             `json:"create_time,omitempty"`
	UpdateTime *time.Time             `json:"update_time,omitempty"`
	Comment    string                 `json:"comment,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// TableInfo 表详细信息
type TableInfo struct {
	Name            string                 `json:"name"`
	CreateStatement string                 `json:"create_statement,omitempty"`
	Columns         []ColumnInfo           `json:"columns"`
	Indices         []IndexInfo            `json:"indices,omitempty"`
	Constraints     []ConstraintInfo       `json:"constraints,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Nullable     bool        `json:"nullable"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	IsPrimaryKey bool        `json:"is_primary_key"`
	IsUnique     bool        `json:"is_unique"`
	Comment      string      `json:"comment,omitempty"`
	Extra        string      `json:"extra,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Type    string   `json:"type,omitempty"`
}

// ConstraintInfo 约束信息
type ConstraintInfo struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Columns    []string `json:"columns"`
	References string   `json:"references,omitempty"`
}

// DatabaseConnectionInfo 数据库连接信息
type DatabaseConnectionInfo struct {
	Name        string              `json:"name"`
	Type        common.DatabaseType `json:"type"`
	DisplayName string              `json:"display_name"`
	Host        string              `json:"host"`
	Port        int                 `json:"port"`
	Database    string              `json:"sql"`
	User        string              `json:"user"`
	Status      string              `json:"status"`
	Version     string              `json:"version,omitempty"`
	ConnectedAt time.Time           `json:"connected_at"`
}
