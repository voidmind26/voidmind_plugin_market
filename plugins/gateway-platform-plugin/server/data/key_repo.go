package data

import (
	"database/sql"
	"errors"
	"time"

	"gateway-platform-plugin/server/models"
)

var ErrKeyInUse = errors.New("key is still referenced by routes")

func ListKeys(db *sql.DB) ([]models.Key, error) {
	rows, err := db.Query(`SELECT id, name, value, description, source, created_at, updated_at FROM keys ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []models.Key
	for rows.Next() {
		var item models.Key
		if err := rows.Scan(&item.ID, &item.Name, &item.Value, &item.Description, &item.Source, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func GetKey(db *sql.DB, id int64) (*models.Key, error) {
	row := db.QueryRow(`SELECT id, name, value, description, source, created_at, updated_at FROM keys WHERE id = ?`, id)
	var item models.Key
	if err := row.Scan(&item.ID, &item.Name, &item.Value, &item.Description, &item.Source, &item.CreatedAt, &item.UpdatedAt); err != nil {
		return nil, err
	}
	return &item, nil
}

func CreateKey(db *sql.DB, item models.Key) (*models.Key, error) {
	now := time.Now().Format(time.RFC3339)
	if item.Source == "" {
		item.Source = "manual"
	}
	res, err := db.Exec(`INSERT INTO keys(name, value, description, source, created_at, updated_at) VALUES(?,?,?,?,?,?)`, item.Name, item.Value, item.Description, item.Source, now, now)
	if err != nil {
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetKey(db, id)
}

func UpdateKey(db *sql.DB, id int64, item models.Key) (*models.Key, error) {
	now := time.Now().Format(time.RFC3339)
	if item.Source == "" {
		item.Source = "manual"
	}
	_, err := db.Exec(`UPDATE keys SET name=?, value=?, description=?, source=?, updated_at=? WHERE id=?`, item.Name, item.Value, item.Description, item.Source, now, id)
	if err != nil {
		return nil, err
	}
	return GetKey(db, id)
}

func DeleteKey(db *sql.DB, id int64) error {
	inUse, err := KeyInUse(db, id)
	if err != nil {
		return err
	}
	if inUse {
		return ErrKeyInUse
	}
	_, err = db.Exec(`DELETE FROM keys WHERE id = ?`, id)
	return err
}
