package sqlite

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func Open(dbPath string) (*sql.DB, error) {
	return sql.Open("sqlite", dbPath)
}

func DataSource(root string) string {
	return filepath.Join(root, "gateway-platform.db")
}

func OpenTestDB(root string) (*sql.DB, error) {
	return Open(filepath.Join(root, "test.db"))
}

func TableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	row := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type='table' AND name=?`, table)
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count > 0
}
