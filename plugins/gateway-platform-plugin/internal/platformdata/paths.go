package platformdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DataDirEnv       = "GATEWAY_PLATFORM_DATA_DIR"
	DatabaseFileName = "gateway-platform.db"
	LogFileName      = "gateway-platform-http.log"
)

type Paths struct {
	Root     string
	Database string
	Log      string
}

func Prepare(legacyRoot string) (Paths, error) {
	root, err := resolveRoot()
	if err != nil {
		return Paths{}, err
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return Paths{}, fmt.Errorf("create gateway platform data directory: %w", err)
	}
	if err := os.Chmod(root, 0o700); err != nil {
		return Paths{}, fmt.Errorf("secure gateway platform data directory: %w", err)
	}

	paths := Paths{
		Root:     root,
		Database: filepath.Join(root, DatabaseFileName),
		Log:      filepath.Join(root, LogFileName),
	}
	legacyDatabase := filepath.Join(legacyRoot, DatabaseFileName)
	if err := migrateDatabase(legacyDatabase, paths.Database); err != nil {
		return Paths{}, err
	}
	if err := SecureDatabaseFile(paths.Database); err != nil {
		return Paths{}, err
	}
	return paths, nil
}

func CheckWritable(root string) error {
	file, err := os.CreateTemp(root, ".gateway-platform-write-check-*")
	if err != nil {
		return fmt.Errorf("data directory is not writable: %w", err)
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("close data directory write check: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("remove data directory write check: %w", err)
	}
	return nil
}

func CheckDatabaseWritable(ctx context.Context, db *sql.DB) error {
	connection, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire database write check connection: %w", err)
	}
	defer connection.Close()

	if _, err := connection.ExecContext(ctx, "BEGIN IMMEDIATE"); err != nil {
		if isDatabaseBusy(err) {
			return nil
		}
		return fmt.Errorf("begin database write check: %w", err)
	}
	rolledBack := false
	defer func() {
		if !rolledBack {
			_, _ = connection.ExecContext(context.Background(), "ROLLBACK")
		}
	}()

	if _, err := connection.ExecContext(ctx, "UPDATE routes SET updated_at = updated_at WHERE 0"); err != nil {
		if isDatabaseBusy(err) {
			return nil
		}
		return fmt.Errorf("execute database write check: %w", err)
	}
	if _, err := connection.ExecContext(ctx, "ROLLBACK"); err != nil {
		return fmt.Errorf("rollback database write check: %w", err)
	}
	rolledBack = true
	return nil
}

func isDatabaseBusy(err error) bool {
	var sqliteError interface{ Code() int }
	if !errors.As(err, &sqliteError) {
		return false
	}
	const (
		sqliteBusy   = 5
		sqliteLocked = 6
	)
	primaryCode := sqliteError.Code() & 0xff
	return primaryCode == sqliteBusy || primaryCode == sqliteLocked
}

func SecureDatabaseFile(path string) error {
	if err := os.Chmod(path, 0o600); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("secure gateway platform database: %w", err)
	}
	return nil
}

func resolveRoot() (string, error) {
	if configured := os.Getenv(DataDirEnv); configured != "" {
		root, err := filepath.Abs(configured)
		if err != nil {
			return "", fmt.Errorf("resolve %s: %w", DataDirEnv, err)
		}
		return filepath.Clean(root), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(home, "CodexData", "gateway-platform-plugin"), nil
}

func migrateDatabase(source, target string) error {
	if filepath.Clean(source) == filepath.Clean(target) {
		return nil
	}
	if _, err := os.Stat(target); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect gateway platform database: %w", err)
	}

	sourceFile, err := os.Open(source)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("open legacy gateway platform database: %w", err)
	}
	defer sourceFile.Close()

	temporary, err := os.CreateTemp(filepath.Dir(target), ".gateway-platform-migration-*")
	if err != nil {
		return fmt.Errorf("create database migration file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)

	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure database migration file: %w", err)
	}
	if _, err := io.Copy(temporary, sourceFile); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("copy legacy gateway platform database: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync migrated gateway platform database: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close migrated gateway platform database: %w", err)
	}

	if err := os.Link(temporaryPath, target); err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return fmt.Errorf("install migrated gateway platform database: %w", err)
	}
	return nil
}
