package db

import (
	"path/filepath"
	"testing"
)

func TestOpenCreatesCatalogAndAppliesDurabilitySettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "catalogo.db")
	database, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()

	var mode string
	if err := database.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "delete" {
		t.Fatalf("journal_mode = %q, want delete", mode)
	}
	var syncMode int
	if err := database.QueryRow("PRAGMA synchronous").Scan(&syncMode); err != nil {
		t.Fatal(err)
	}
	if syncMode != 2 {
		t.Fatalf("synchronous = %d, want FULL (2)", syncMode)
	}
	var migrationCount int
	if err := database.QueryRow("SELECT count(*) FROM schema_migrations WHERE version=1").Scan(&migrationCount); err != nil {
		t.Fatal(err)
	}
	if migrationCount != 1 {
		t.Fatalf("initial migration rows = %d, want 1", migrationCount)
	}
}

func TestOpenDoesNotReapplyMigration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalogo.db")
	first, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := first.Exec("INSERT INTO app_meta(key, value) VALUES ('keep', 'yes')"); err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	var value string
	if err := second.QueryRow("SELECT value FROM app_meta WHERE key='keep'").Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "yes" {
		t.Fatalf("value = %q, want yes", value)
	}
}
