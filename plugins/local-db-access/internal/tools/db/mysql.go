package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"local-db-access/internal/common"
	"local-db-access/internal/config"
	"local-db-access/internal/dto"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLQueryTool MySQL查询工具
type MySQLQueryTool struct {
	db         *sql.DB
	config     *config.DatabaseConfig
	name       string
	queryIDGen *sync.Map
}

// NewMySQLQueryTool 创建MySQL查询工具实例
func NewMySQLQueryTool(name string, dbConfig *config.DatabaseConfig) (*MySQLQueryTool, error) {
	if dbConfig == nil {
		return nil, errors.New("MySQL配置不能为空")
	}

	if dbConfig.QueryLimits == nil {
		dbConfig.QueryLimits = defaultQueryLimits()
	}

	dsn := buildMySQLDSN(dbConfig)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("连接数据库失败: %w", err)
	}

	if dbConfig.MaxOpenConn > 0 {
		db.SetMaxOpenConns(dbConfig.MaxOpenConn)
	} else {
		db.SetMaxOpenConns(10)
	}

	if dbConfig.MaxIdleConn > 0 {
		db.SetMaxIdleConns(dbConfig.MaxIdleConn)
	} else {
		db.SetMaxIdleConns(5)
	}

	if dbConfig.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(dbConfig.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(5 * time.Minute)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("数据库连接测试失败: %w", err)
	}

	return &MySQLQueryTool{
		db:         db,
		config:     dbConfig,
		name:       name,
		queryIDGen: &sync.Map{},
	}, nil
}

func (t *MySQLQueryTool) Type() common.DatabaseType {
	return common.DatabaseTypeMySQL
}

func (t *MySQLQueryTool) Name() string {
	return t.name
}

func (t *MySQLQueryTool) ConnectionInfo() *dto.DatabaseConnectionInfo {
	return &dto.DatabaseConnectionInfo{
		Name:        t.name,
		Type:        common.DatabaseTypeMySQL,
		Host:        t.config.Host,
		Port:        t.config.Port,
		Database:    t.config.Database,
		User:        t.config.User,
		Status:      "connected",
		ConnectedAt: time.Now(),
	}
}

func (t *MySQLQueryTool) IsReadOnlyQuery(sqlText string) bool {
	return isReadOnlyQuery(sqlText)
}

func (t *MySQLQueryTool) ExecuteQuery(ctx context.Context, sqlStr string, params []interface{}, limit int) (*dto.QueryResult, error) {
	startTime := time.Now()

	if !t.IsReadOnlyQuery(sqlStr) {
		return &dto.QueryResult{
			Success:      false,
			ErrorMessage: "只允许执行SELECT、SHOW、DESCRIBE、EXPLAIN等只读操作",
		}, nil
	}

	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(t.config.QueryLimits.MaxTimeSec)*time.Second)
	defer cancel()

	if limit > 0 && strings.HasPrefix(strings.ToLower(strings.TrimSpace(sqlStr)), "select") {
		sqlLower := strings.ToLower(sqlStr)
		if !strings.Contains(sqlLower, " limit ") {
			sqlStr = strings.TrimSuffix(strings.TrimSpace(sqlStr), ";")
			sqlStr += fmt.Sprintf(" LIMIT %d", limit)
		}
	}

	var rows *sql.Rows
	var err error
	if len(params) > 0 {
		rows, err = t.db.QueryContext(queryCtx, sqlStr, params...)
	} else {
		rows, err = t.db.QueryContext(queryCtx, sqlStr)
	}
	if err != nil {
		return &dto.QueryResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("查询执行失败: %v", err),
		}, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return &dto.QueryResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("获取列信息失败: %v", err),
		}, err
	}

	var results []map[string]interface{}
	colCount := len(columns)
	values := make([]interface{}, colCount)
	valuePtrList := make([]interface{}, colCount)
	for i := range values {
		valuePtrList[i] = &values[i]
	}

	rowCount := 0
	totalRows := 0
	for rows.Next() {
		totalRows++
		if rowCount >= t.config.QueryLimits.MaxRows {
			break
		}

		if err := rows.Scan(valuePtrList...); err != nil {
			return &dto.QueryResult{
				Success:      false,
				ErrorMessage: fmt.Sprintf("扫描行数据失败: %v", err),
			}, err
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			switch v := val.(type) {
			case []byte:
				if len(v) > 0 && (v[0] == '{' || v[0] == '[') {
					var jsonData interface{}
					if json.Unmarshal(v, &jsonData) == nil {
						rowData[col] = jsonData
					} else {
						rowData[col] = string(v)
					}
				} else {
					rowData[col] = string(v)
				}
			case time.Time:
				rowData[col] = v.Format("2006-01-02 15:04:05")
			case nil:
				rowData[col] = nil
			default:
				rowData[col] = val
			}
		}

		results = append(results, rowData)
		rowCount++
		if limit > 0 && rowCount >= limit {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return &dto.QueryResult{
			Success:      false,
			ErrorMessage: fmt.Sprintf("遍历结果集失败: %v", err),
		}, err
	}

	limited := rowCount >= t.config.QueryLimits.MaxRows || (limit > 0 && rowCount >= limit)
	executionTime := time.Since(startTime).Milliseconds()

	return &dto.QueryResult{
		Success:   true,
		Database:  t.config.Database,
		QueryTime: executionTime,
		RowCount:  rowCount,
		TotalRows: totalRows,
		Columns:   columns,
		Data:      results,
		Limited:   limited,
		QueryID:   generateQueryID(),
	}, nil
}

func (t *MySQLQueryTool) ExecuteWriteQuery(ctx context.Context, sqlStr string, params []interface{}) (*dto.WriteResult, error) {
	sqlType, ok := detectSQLAction(sqlStr)
	if !ok || !isAllowedWriteQuery(sqlStr) {
		return &dto.WriteResult{
			Success:      false,
			DatabaseName: t.name,
			ExecutedSQL:  sqlStr,
			Message:      "不支持的写入语句，仅允许 INSERT/UPDATE/CREATE/ALTER，禁用 DELETE/DROP/TRUNCATE",
		}, nil
	}

	queryTimeout := 30
	if t.config != nil && t.config.QueryLimits != nil && t.config.QueryLimits.MaxTimeSec > 0 {
		queryTimeout = t.config.QueryLimits.MaxTimeSec
	}
	queryCtx, cancel := context.WithTimeout(ctx, time.Duration(queryTimeout)*time.Second)
	defer cancel()

	var (
		result sql.Result
		err    error
	)
	if len(params) > 0 {
		result, err = t.db.ExecContext(queryCtx, sqlStr, params...)
	} else {
		result, err = t.db.ExecContext(queryCtx, sqlStr)
	}
	if err != nil {
		return &dto.WriteResult{
			Success:      false,
			DatabaseName: t.name,
			SQLType:      sqlType,
			ExecutedSQL:  sqlStr,
			Message:      fmt.Sprintf("写入执行失败: %v", err),
		}, err
	}

	rowsAffected, _ := result.RowsAffected()
	lastInsertID, _ := result.LastInsertId()
	return &dto.WriteResult{
		Success:      true,
		DatabaseName: t.name,
		SQLType:      sqlType,
		RowsAffected: rowsAffected,
		LastInsertID: lastInsertID,
		ExecutedSQL:  sqlStr,
		Message:      "ok",
	}, nil
}

func (t *MySQLQueryTool) GetDatabaseInfo(ctx context.Context) (*dto.DatabaseInfo, error) {
	versionResult, err := t.ExecuteQuery(ctx,
		"SELECT VERSION() as version, DATABASE() as current_db, @@version_comment as version_comment",
		nil, 1)
	if err != nil || !versionResult.Success {
		return nil, fmt.Errorf("获取数据库版本失败: %v", err)
	}

	var version, currentDB string
	if len(versionResult.Data) > 0 {
		version = fmt.Sprintf("%v", versionResult.Data[0]["version"])
		currentDB = fmt.Sprintf("%v", versionResult.Data[0]["current_db"])
	}

	tablesResult, err := t.ExecuteQuery(ctx, "SHOW TABLE STATUS", nil, 1000)
	if err != nil {
		return nil, fmt.Errorf("获取表信息失败: %v", err)
	}

	var tables []dto.TableSummary
	for _, row := range tablesResult.Data {
		tableName := fmt.Sprintf("%v", row["Name"])
		engine := fmt.Sprintf("%v", row["Engine"])
		comment := fmt.Sprintf("%v", row["Comment"])

		var rowCount int64
		if rows, ok := row["Rows"]; ok {
			switch v := rows.(type) {
			case int64:
				rowCount = v
			case float64:
				rowCount = int64(v)
			case string:
				if i, convErr := strconv.ParseInt(v, 10, 64); convErr == nil {
					rowCount = i
				}
			}
		}

		var sizeMB float64
		if dataLength, ok := row["Data_length"]; ok {
			switch v := dataLength.(type) {
			case int64:
				sizeMB = float64(v) / 1024 / 1024
			case float64:
				sizeMB = v / 1024 / 1024
			}
		}

		tables = append(tables, dto.TableSummary{
			Name:     tableName,
			RowCount: rowCount,
			SizeMB:   sizeMB,
			Engine:   engine,
			Comment:  comment,
		})
	}

	sizeResult, err := t.ExecuteQuery(ctx, `
		SELECT
			table_schema as db_name,
			SUM(data_length + index_length) / 1024 / 1024 as total_size_mb,
			SUM(data_length) / 1024 / 1024 as data_size_mb,
			SUM(index_length) / 1024 / 1024 as index_size_mb
		FROM information_schema.tables
		WHERE table_schema = ?
		GROUP BY table_schema`, []interface{}{t.config.Database}, 1)

	var sizeInfo map[string]interface{}
	if sizeResult != nil && sizeResult.Success && len(sizeResult.Data) > 0 {
		sizeInfo = sizeResult.Data[0]
	}

	return &dto.DatabaseInfo{
		Name:            t.name,
		Type:            common.DatabaseTypeMySQL,
		Version:         version,
		CurrentDatabase: currentDB,
		Host:            t.config.Host,
		Port:            t.config.Port,
		User:            t.config.User,
		TableCount:      len(tables),
		Tables:          tables,
		SizeInfo:        sizeInfo,
		Status:          "connected",
		LastPing:        time.Now(),
	}, nil
}

func (t *MySQLQueryTool) GetTableInfo(ctx context.Context, tableName string) (*dto.TableInfo, error) {
	tableName = sanitizeTableName(tableName)
	columnsResult, err := t.ExecuteQuery(ctx, fmt.Sprintf("SHOW FULL COLUMNS FROM `%s`", tableName), nil, 100)
	if err != nil || !columnsResult.Success {
		return nil, fmt.Errorf("获取表结构失败: %v", err)
	}

	var columns []dto.ColumnInfo
	for _, row := range columnsResult.Data {
		column := dto.ColumnInfo{
			Name:     fmt.Sprintf("%v", row["Field"]),
			Type:     fmt.Sprintf("%v", row["Type"]),
			Nullable: fmt.Sprintf("%v", row["Null"]) == "YES",
			Comment:  fmt.Sprintf("%v", row["Comment"]),
			Extra:    fmt.Sprintf("%v", row["Extra"]),
		}
		if key, ok := row["Key"]; ok {
			keyStr := fmt.Sprintf("%v", key)
			column.IsPrimaryKey = keyStr == "PRI"
			column.IsUnique = keyStr == "UNI"
		}
		if defaultVal, ok := row["Default"]; ok && defaultVal != nil {
			column.DefaultValue = defaultVal
		}
		columns = append(columns, column)
	}

	createResult, err := t.ExecuteQuery(ctx, fmt.Sprintf("SHOW CREATE TABLE `%s`", tableName), nil, 1)
	var createStatement string
	if createResult != nil && createResult.Success && len(createResult.Data) > 0 {
		if stmt, ok := createResult.Data[0]["Create Table"]; ok {
			createStatement = fmt.Sprintf("%v", stmt)
		}
	}

	indexResult, err := t.ExecuteQuery(ctx, fmt.Sprintf("SHOW INDEX FROM `%s`", tableName), nil, 100)
	var indices []dto.IndexInfo
	if indexResult != nil && indexResult.Success {
		indexMap := make(map[string]*dto.IndexInfo)
		for _, row := range indexResult.Data {
			indexName := fmt.Sprintf("%v", row["Key_name"])
			columnName := fmt.Sprintf("%v", row["Column_name"])
			nonUnique := fmt.Sprintf("%v", row["Non_unique"]) == "1"
			if indexName == "PRIMARY" {
				continue
			}
			if idx, exists := indexMap[indexName]; exists {
				idx.Columns = append(idx.Columns, columnName)
			} else {
				indexMap[indexName] = &dto.IndexInfo{
					Name:    indexName,
					Columns: []string{columnName},
					Unique:  !nonUnique,
				}
			}
		}
		for _, idx := range indexMap {
			indices = append(indices, *idx)
		}
	}

	return &dto.TableInfo{
		Name:            tableName,
		CreateStatement: createStatement,
		Columns:         columns,
		Indices:         indices,
	}, nil
}

func (t *MySQLQueryTool) GetTableData(ctx context.Context, tableName string, limit int) (*dto.QueryResult, error) {
	tableName = sanitizeTableName(tableName)
	querySQL := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	return t.ExecuteQuery(ctx, querySQL, nil, limit)
}

func (t *MySQLQueryTool) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return t.db.PingContext(ctx)
}

func (t *MySQLQueryTool) Close() error {
	if t.db != nil {
		return t.db.Close()
	}
	return nil
}
