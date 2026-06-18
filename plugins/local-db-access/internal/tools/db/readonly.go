package db

import "strings"

var readOnlyPrefixes = []string{"select", "show", "desc", "describe", "explain"}
var allowedWritePrefixes = []string{"insert", "update", "create", "alter"}
var blockedWritePrefixes = []string{"delete", "drop", "truncate"}

func normalizeSQL(sqlText string) string {
	return strings.TrimSpace(strings.ToLower(sqlText))
}

func splitSQLStatements(sqlText string) []string {
	parts := strings.Split(sqlText, ";")
	statements := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			statements = append(statements, part)
		}
	}
	return statements
}

func detectSQLAction(sqlText string) (string, bool) {
	normalized := normalizeSQL(sqlText)
	if normalized == "" {
		return "", false
	}
	statements := splitSQLStatements(normalized)
	if len(statements) != 1 {
		return "", false
	}
	fields := strings.Fields(statements[0])
	if len(fields) == 0 {
		return "", false
	}
	return fields[0], true
}

func isReadOnlyQuery(sqlText string) bool {
	action, ok := detectSQLAction(sqlText)
	if !ok {
		return false
	}
	for _, prefix := range readOnlyPrefixes {
		if action == prefix {
			return true
		}
	}
	return false
}

func isAllowedWriteQuery(sqlText string) bool {
	action, ok := detectSQLAction(sqlText)
	if !ok {
		return false
	}
	for _, blocked := range blockedWritePrefixes {
		if action == blocked {
			return false
		}
	}
	for _, allowed := range allowedWritePrefixes {
		if action == allowed {
			return true
		}
	}
	return false
}
