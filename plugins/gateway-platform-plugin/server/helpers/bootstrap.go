package helpers

import (
	"database/sql"

	"gateway-platform-plugin/server/data/sqlite"
)

type App struct {
	DB *sql.DB
}

func MustBootstrap() *App {
	db, err := sqlite.Open(sqlite.DataSource("."))
	if err != nil {
		panic(err)
	}
	if err := sqlite.InitSchema(db); err != nil {
		panic(err)
	}
	return &App{DB: db}
}
