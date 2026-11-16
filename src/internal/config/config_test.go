package config

import (
	"path/filepath"
	"testing"
)

func TestLoadUsesExecutableDirectoryForPortableData(t *testing.T) {
	settings, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(settings.DataDir) != "data" || filepath.Dir(settings.DataDir) != settings.AppDir {
		t.Fatalf("data directory should be beside the executable: %+v", settings)
	}
	if settings.DatabasePath != filepath.Join(settings.DataDir, "catalogo.db") {
		t.Fatalf("unexpected database path: %s", settings.DatabasePath)
	}
}

func TestLoadUsesConfiguredLibraryRoot(t *testing.T) {
	want := filepath.Join(t.TempDir(), "biblioteca")
	got, err := resolveLibraryRoot(t.TempDir(), want, "")
	if err != nil {
		t.Fatal(err)
	}
	absolute, err := filepath.Abs(want)
	if err != nil {
		t.Fatal(err)
	}
	if got != filepath.Clean(absolute) {
		t.Fatalf("LibraryRoot = %q, want %q", got, filepath.Clean(absolute))
	}
}

func TestEnvironmentRootOverridesSavedRoot(t *testing.T) {
	got, err := resolveLibraryRoot(t.TempDir(), "env-library", "saved-library")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs("env-library")
	if got != filepath.Clean(want) {
		t.Fatalf("LibraryRoot = %q, want %q", got, filepath.Clean(want))
	}
}
