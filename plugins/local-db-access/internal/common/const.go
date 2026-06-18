package common

// DatabaseType 数据库类型枚举
type DatabaseType string

const (
	DatabaseTypeMySQL     DatabaseType = "mysql"
	DatabaseTypeSQLite    DatabaseType = "sqlite"
	DatabaseTypeSQLServer DatabaseType = "sqlserver"
)
