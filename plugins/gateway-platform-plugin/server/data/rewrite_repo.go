package data

import (
	"database/sql"

	"gateway-platform-plugin/server/models"
)

func ListRewrites(db *sql.DB, routeID int64) ([]models.RouteRewrite, error) {
	rows, err := db.Query(`SELECT id, route_id, rewrite_type, target_name, key_id, template, ordering FROM route_rewrites WHERE route_id = ? ORDER BY ordering, id`, routeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.RouteRewrite
	for rows.Next() {
		var item models.RouteRewrite
		if err := rows.Scan(&item.ID, &item.RouteID, &item.RewriteType, &item.TargetName, &item.KeyID, &item.Template, &item.Ordering); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func CreateRewrite(db *sql.DB, item models.RouteRewrite) (*models.RouteRewrite, error) {
	res, err := db.Exec(`INSERT INTO route_rewrites(route_id, rewrite_type, target_name, key_id, template, ordering) VALUES(?,?,?,?,?,?)`, item.RouteID, item.RewriteType, item.TargetName, item.KeyID, item.Template, item.Ordering)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetRewrite(db, id)
}

func GetRewrite(db *sql.DB, id int64) (*models.RouteRewrite, error) {
	row := db.QueryRow(`SELECT id, route_id, rewrite_type, target_name, key_id, template, ordering FROM route_rewrites WHERE id = ?`, id)
	var item models.RouteRewrite
	if err := row.Scan(&item.ID, &item.RouteID, &item.RewriteType, &item.TargetName, &item.KeyID, &item.Template, &item.Ordering); err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdateRewrite(db *sql.DB, id int64, item models.RouteRewrite) (*models.RouteRewrite, error) {
	_, err := db.Exec(`UPDATE route_rewrites SET rewrite_type=?, target_name=?, key_id=?, template=?, ordering=? WHERE id=?`, item.RewriteType, item.TargetName, item.KeyID, item.Template, item.Ordering, id)
	if err != nil {
		return nil, err
	}
	return GetRewrite(db, id)
}

func DeleteRewrite(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM route_rewrites WHERE id = ?`, id)
	return err
}
