package data

import (
	"database/sql"
	"time"

	"gateway-platform-plugin/server/models"
)

func ListRoutes(db *sql.DB) ([]models.Route, error) {
	rows, err := db.Query(`SELECT id, name, enabled, local_path, upstream_url, timeout_ms, description, created_at, updated_at FROM routes ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Route
	for rows.Next() {
		var item models.Route
		var enabled int
		if err := rows.Scan(&item.ID, &item.Name, &enabled, &item.LocalPath, &item.UpstreamURL, &item.TimeoutMS, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Enabled = enabled == 1
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetRoute(db *sql.DB, id int64) (*models.Route, error) {
	row := db.QueryRow(`SELECT id, name, enabled, local_path, upstream_url, timeout_ms, description, created_at, updated_at FROM routes WHERE id = ?`, id)
	var item models.Route
	var enabled int
	if err := row.Scan(&item.ID, &item.Name, &enabled, &item.LocalPath, &item.UpstreamURL, &item.TimeoutMS, &item.Description, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	item.Enabled = enabled == 1
	return &item, nil
}

func CreateRoute(db *sql.DB, item models.Route) (*models.Route, error) {
	now := time.Now().Format(time.RFC3339)
	res, err := db.Exec(`INSERT INTO routes(name, enabled, local_path, upstream_url, timeout_ms, description, created_at, updated_at) VALUES(?,?,?,?,?,?,?,?)`, item.Name, boolToInt(item.Enabled), item.LocalPath, item.UpstreamURL, item.TimeoutMS, item.Description, now, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetRoute(db, id)
}

func UpdateRoute(db *sql.DB, id int64, item models.Route) (*models.Route, error) {
	now := time.Now().Format(time.RFC3339)
	_, err := db.Exec(`UPDATE routes SET name=?, enabled=?, local_path=?, upstream_url=?, timeout_ms=?, description=?, updated_at=? WHERE id=?`, item.Name, boolToInt(item.Enabled), item.LocalPath, item.UpstreamURL, item.TimeoutMS, item.Description, now, id)
	if err != nil {
		return nil, err
	}
	return GetRoute(db, id)
}

func DeleteRoute(db *sql.DB, id int64) error {
	if _, err := db.Exec(`DELETE FROM route_rewrites WHERE route_id = ?`, id); err != nil {
		return err
	}
	_, err := db.Exec(`DELETE FROM routes WHERE id = ?`, id)
	return err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
