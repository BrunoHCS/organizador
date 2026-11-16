package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var initialMigration = []string{
	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS app_meta (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`,
}

// Open creates the on-disk database, applies initial migrations, and enables
// durable rollback-journal settings required for a removable drive.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}
	database, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	database.SetMaxOpenConns(1)
	if err := migrate(database); err != nil {
		_ = database.Close()
		return nil, err
	}
	return database, nil
}

func migrate(database *sql.DB) error {
	for _, statement := range []string{"PRAGMA journal_mode=DELETE", "PRAGMA synchronous=FULL"} {
		if _, err := database.Exec(statement); err != nil {
			return fmt.Errorf("apply %q: %w", statement, err)
		}
	}
	var exists int
	err := database.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&exists)
	if err != nil {
		return fmt.Errorf("check migration table: %w", err)
	}
	if exists > 0 {
		var applied int
		if err := database.QueryRow("SELECT count(*) FROM schema_migrations WHERE version=1").Scan(&applied); err != nil {
			return fmt.Errorf("read migration state: %w", err)
		}
		if applied > 0 {
			return nil
		}
	}
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin migration: %w", err)
	}
	for _, statement := range initialMigration {
		if _, err := tx.Exec(statement); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply initial migration: %w", err)
		}
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations(version, name) VALUES (1, 'initial')"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record initial migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit initial migration: %w", err)
	}
	return nil
}
