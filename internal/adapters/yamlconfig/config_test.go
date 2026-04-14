package yamlconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".bruwrapper.yaml")
	content := "collections:\n  - ~/projects/api\n  - ~/projects/web\n  - /absolute/path\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := New(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	home, _ := os.UserHomeDir()
	expected := []string{
		filepath.Join(home, "projects/api"),
		filepath.Join(home, "projects/web"),
		"/absolute/path",
	}
	if len(cfg.CollectionPaths) != len(expected) {
		t.Fatalf("expected %d paths, got %d", len(expected), len(cfg.CollectionPaths))
	}
	for i, p := range cfg.CollectionPaths {
		if p != expected[i] {
			t.Errorf("path[%d]: expected %q, got %q", i, expected[i], p)
		}
	}
}

func TestLoadMissingConfig(t *testing.T) {
	loader := New("/nonexistent/.bruwrapper.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestLoadEmptyCollections(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".bruwrapper.yaml")
	content := "collections: []\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	loader := New(cfgPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	var empty core.Config
	if len(cfg.CollectionPaths) != len(empty.CollectionPaths) {
		t.Fatalf("expected empty collections, got %v", cfg.CollectionPaths)
	}
}
