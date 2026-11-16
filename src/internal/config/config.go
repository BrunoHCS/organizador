package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const libraryRootEnv = "BIBLIOTECA_ROOT"

type fileConfig struct {
	LibraryRoot     string   `json:"libraryRoot"`
	ExcludedFolders []string `json:"excludedFolders"`
}

// Settings contains absolute paths derived from the location of the executable.
type Settings struct {
	AppDir          string
	LibraryRoot     string
	DataDir         string
	DatabasePath    string
	ExcludedFolders []string
}

// Load resolves the app's portable data folder and the configured library root.
// BIBLIOTECA_ROOT takes precedence over data/config.json. When neither is set,
// the app directory itself is the initial library root.
func Load() (Settings, error) {
	executable, err := os.Executable()
	if err != nil {
		return Settings{}, fmt.Errorf("resolve executable: %w", err)
	}
	appDir, err := filepath.Abs(filepath.Dir(executable))
	if err != nil {
		return Settings{}, fmt.Errorf("resolve app directory: %w", err)
	}
	dataDir := filepath.Join(appDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return Settings{}, fmt.Errorf("create data directory: %w", err)
	}

	var saved fileConfig
	configPath := filepath.Join(dataDir, "config.json")
	contents, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return Settings{}, fmt.Errorf("read config: %w", err)
	}
	if err == nil {
		if err := json.Unmarshal(contents, &saved); err != nil {
			return Settings{}, fmt.Errorf("decode config: %w", err)
		}
	}

	root, err := resolveLibraryRoot(appDir, os.Getenv(libraryRootEnv), saved.LibraryRoot)
	if err != nil {
		return Settings{}, err
	}
	return Settings{AppDir: appDir, LibraryRoot: root, DataDir: dataDir, DatabasePath: filepath.Join(dataDir, "catalogo.db"), ExcludedFolders: saved.ExcludedFolders}, nil
}

func resolveLibraryRoot(appDir, environmentRoot, configuredRoot string) (string, error) {
	root := strings.TrimSpace(environmentRoot)
	if root == "" {
		root = strings.TrimSpace(configuredRoot)
	}
	if root == "" {
		root = appDir
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve library root: %w", err)
	}
	return filepath.Clean(absolute), nil
}
