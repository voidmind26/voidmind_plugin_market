package helpers

import (
	"os"
	"path/filepath"
	"testing"

	"gateway-platform-plugin/internal/platformdata"
	"gateway-platform-plugin/server/data/sqlite"
)

func TestBootstrapMigratesLegacyDatabaseToVisibleDataDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(platformdata.DataDirEnv, "")
	legacyRoot := t.TempDir()
	legacyDB, err := sqlite.Open(sqlite.DataSource(legacyRoot))
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlite.InitSchema(legacyDB); err != nil {
		t.Fatal(err)
	}
	if _, err := legacyDB.Exec(`INSERT INTO routes (name, enabled, local_path, upstream_url, timeout_ms, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "ship", 1, "/ship", "https://ship.example", 30000, "legacy", "now", "now"); err != nil {
		t.Fatal(err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatal(err)
	}

	app, err := Bootstrap(legacyRoot)
	if err != nil {
		t.Fatal(err)
	}
	defer app.DB.Close()

	expectedRoot := filepath.Join(home, "CodexData", "gateway-platform-plugin")
	if app.DataDir != expectedRoot {
		t.Fatalf("expected data directory %q, got %q", expectedRoot, app.DataDir)
	}
	var count int
	if err := app.DB.QueryRow(`SELECT COUNT(1) FROM routes WHERE name = ?`, "ship").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected migrated route, got count %d", count)
	}
	info, err := os.Stat(app.DatabasePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected database mode 0600, got %o", info.Mode().Perm())
	}
}
