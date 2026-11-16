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

var catalogMigration = []string{
	`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY,
		path TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		extension TEXT NOT NULL,
		type TEXT NOT NULL,
		size INTEGER NOT NULL,
		mtime_ns INTEGER NOT NULL,
		partial_hash TEXT NOT NULL,
		removed INTEGER NOT NULL DEFAULT 0,
		indexed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE INDEX IF NOT EXISTS idx_items_type_path ON items(type, path)`,
	`CREATE INDEX IF NOT EXISTS idx_items_removed ON items(removed)`,
	`CREATE TABLE IF NOT EXISTS scan_runs (
		id INTEGER PRIMARY KEY,
		started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
		finished_at TEXT,
		visited INTEGER NOT NULL DEFAULT 0,
		indexed INTEGER NOT NULL DEFAULT 0,
		errors INTEGER NOT NULL DEFAULT 0,
		complete INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS scan_seen (run_id INTEGER NOT NULL, path TEXT NOT NULL, PRIMARY KEY(run_id,path))`,
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
			return migrateCatalog(database)
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
	return migrateCatalog(database)
}

func migrateCatalog(database *sql.DB) error {
	var applied int
	if err := database.QueryRow("SELECT count(*) FROM schema_migrations WHERE version=2").Scan(&applied); err != nil {
		return fmt.Errorf("read catalog migration state: %w", err)
	}
	if applied > 0 {
		return nil
	}
	tx, err := database.Begin()
	if err != nil {
		return fmt.Errorf("begin catalog migration: %w", err)
	}
	for _, statement := range catalogMigration {
		if _, err := tx.Exec(statement); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply catalog migration: %w", err)
		}
	}
	if _, err := tx.Exec("INSERT INTO schema_migrations(version, name) VALUES (2, 'catalog')"); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record catalog migration: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit catalog migration: %w", err)
	}
	return nil
}
