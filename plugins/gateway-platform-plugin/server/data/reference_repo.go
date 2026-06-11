package data

import (
	"database/sql"
)

type MissingReference struct {
	Route      string `json:"route"`
	Key        string `json:"key"`
	Type       string `json:"type"`
	TargetName string `json:"target_name"`
}

type UnusedKey struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

func ListMissingReferences(db *sql.DB) ([]MissingReference, error) {
	rows, err := db.Query(`
		SELECT r.name, COALESCE(k.name, ''), rr.rewrite_type, rr.target_name
		FROM route_rewrites rr
		JOIN routes r ON r.id = rr.route_id
		LEFT JOIN keys k ON k.id = rr.key_id
		WHERE k.id IS NULL
		ORDER BY rr.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MissingReference
	for rows.Next() {
		var item MissingReference
		if err := rows.Scan(&item.Route, &item.Key, &item.Type, &item.TargetName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func ListUnusedKeys(db *sql.DB) ([]UnusedKey, error) {
	rows, err := db.Query(`
		SELECT k.name, k.description
		FROM keys k
		LEFT JOIN route_rewrites rr ON rr.key_id = k.id
		WHERE rr.id IS NULL
		ORDER BY k.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UnusedKey
	for rows.Next() {
		var item UnusedKey
		if err := rows.Scan(&item.Key, &item.Description); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func KeyInUse(db *sql.DB, id int64) (bool, error) {
	row := db.QueryRow(`SELECT COUNT(1) FROM route_rewrites WHERE key_id = ?`, id)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
