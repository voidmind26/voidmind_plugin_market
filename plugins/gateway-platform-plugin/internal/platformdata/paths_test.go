package platformdata

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestPrepareUsesVisibleDefaultDirectory(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(DataDirEnv, "")

	paths, err := Prepare(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	expected := filepath.Join(home, "CodexData", "gateway-platform-plugin")
	if paths.Root != expected {
		t.Fatalf("expected root %q, got %q", expected, paths.Root)
	}
	if paths.Database != filepath.Join(expected, DatabaseFileName) {
		t.Fatalf("unexpected database path: %q", paths.Database)
	}
	if paths.Log != filepath.Join(expected, LogFileName) {
		t.Fatalf("unexpected log path: %q", paths.Log)
	}

	info, err := os.Stat(paths.Root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("expected data directory mode 0700, got %o", info.Mode().Perm())
	}
}

func TestPrepareMigratesLegacyDatabaseWithoutOverwritingTarget(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(DataDirEnv, "")
	legacyRoot := t.TempDir()
	legacyPath := filepath.Join(legacyRoot, DatabaseFileName)
	if err := os.WriteFile(legacyPath, []byte("legacy"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := Prepare(legacyRoot)
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "legacy" {
		t.Fatalf("expected migrated content, got %q", content)
	}
	info, err := os.Stat(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("expected database mode 0600, got %o", info.Mode().Perm())
	}

	if err := os.WriteFile(paths.Database, []byte("current"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacyPath, []byte("new-legacy"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Prepare(legacyRoot); err != nil {
		t.Fatal(err)
	}
	content, err = os.ReadFile(paths.Database)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "current" {
		t.Fatalf("existing target was overwritten: %q", content)
	}
}

func TestCheckWritableRejectsMissingDirectory(t *testing.T) {
	err := CheckWritable(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("expected missing directory to be reported as not writable")
	}
}

func TestCheckDatabaseWritableRejectsReadOnlyConnection(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), DatabaseFileName)
	writable, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := writable.Exec(`CREATE TABLE routes (id INTEGER PRIMARY KEY, updated_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	if err := writable.Close(); err != nil {
		t.Fatal(err)
	}

	readOnly, err := sql.Open("sqlite", "file:"+databasePath+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer readOnly.Close()
	if err := CheckDatabaseWritable(context.Background(), readOnly); err == nil {
		t.Fatal("expected read-only database connection to fail the write check")
	}
}

func TestCheckDatabaseWritableAcceptsBusyWritableDatabase(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), DatabaseFileName)
	first, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	if _, err := first.Exec(`CREATE TABLE routes (id INTEGER PRIMARY KEY, updated_at TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	connection, err := first.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if _, err := connection.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatal(err)
	}
	defer connection.ExecContext(context.Background(), "ROLLBACK")

	second, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	if err := CheckDatabaseWritable(context.Background(), second); err != nil {
		t.Fatalf("busy writable database should remain healthy: %v", err)
	}
}
