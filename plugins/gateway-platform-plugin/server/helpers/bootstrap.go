package helpers

import (
	"database/sql"
	"fmt"

	"gateway-platform-plugin/internal/platformdata"
	"gateway-platform-plugin/server/data/sqlite"
)

type App struct {
	DB           *sql.DB
	DataDir      string
	DatabasePath string
}

func MustBootstrap() *App {
	app, err := Bootstrap(".")
	if err != nil {
		panic(err)
	}
	return app
}

func Bootstrap(legacyRoot string) (*App, error) {
	paths, err := platformdata.Prepare(legacyRoot)
	if err != nil {
		return nil, err
	}
	db, err := sqlite.Open(paths.Database)
	if err != nil {
		return nil, fmt.Errorf("open gateway platform database: %w", err)
	}
	if err := sqlite.InitSchema(db); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("initialize gateway platform database: %w", err)
	}
	if err := platformdata.SecureDatabaseFile(paths.Database); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &App{DB: db, DataDir: paths.Root, DatabasePath: paths.Database}, nil
}
