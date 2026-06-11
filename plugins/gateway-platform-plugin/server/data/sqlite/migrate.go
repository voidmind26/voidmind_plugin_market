package sqlite

import "database/sql"

func InitSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS routes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			enabled INTEGER NOT NULL,
			local_path TEXT NOT NULL UNIQUE,
			upstream_url TEXT NOT NULL,
			timeout_ms INTEGER NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			value TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT 'manual',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS route_rewrites (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			route_id INTEGER NOT NULL,
			rewrite_type TEXT NOT NULL,
			target_name TEXT NOT NULL,
			key_id INTEGER NOT NULL,
			template TEXT NOT NULL,
			ordering INTEGER NOT NULL,
			FOREIGN KEY(route_id) REFERENCES routes(id),
			FOREIGN KEY(key_id) REFERENCES keys(id)
		);`,
	}
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}
