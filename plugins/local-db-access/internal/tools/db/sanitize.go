package db

import (
	"fmt"
	"os"
	"strings"
	"time"

	"local-db-access/internal/config"
)

func defaultQueryLimits() *config.QueryLimits {
	return &config.QueryLimits{
		MaxRows:     1000,
		MaxTimeSec:  30,
		MaxRowLimit: 10000,
	}
}

func buildMySQLDSN(dbConfig *config.DatabaseConfig) string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Database)

	var params []string
	if dbConfig.Charset != "" {
		params = append(params, "charset="+dbConfig.Charset)
	}
	if dbConfig.Timeout > 0 {
		params = append(params, fmt.Sprintf("timeout=%ds", dbConfig.Timeout))
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn
}

func sanitizeTableName(tableName string) string {
	tableName = strings.TrimSpace(tableName)
	tableName = strings.ReplaceAll(tableName, "`", "")
	tableName = strings.ReplaceAll(tableName, "'", "")
	tableName = strings.ReplaceAll(tableName, "\"", "")
	tableName = strings.ReplaceAll(tableName, ";", "")
	tableName = strings.ReplaceAll(tableName, "--", "")
	return tableName
}

func generateQueryID() string {
	return fmt.Sprintf("qry_%d_%d", time.Now().UnixNano(), os.Getpid())
}

func countHealthy(results map[string]error) int {
	count := 0
	for _, err := range results {
		if err == nil {
			count++
		}
	}
	return count
}
