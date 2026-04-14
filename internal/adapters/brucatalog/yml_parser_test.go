package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestParseYMLFile(t *testing.T) {
	path := filepath.Join("testdata", "create_user.yml")
	req, err := ParseYMLFile(path)
	if err != nil {
		t.Fatalf("ParseYMLFile(%q) returned error: %v", path, err)
	}

	if req.Name != "Create User" {
		t.Errorf("Name = %q, want %q", req.Name, "Create User")
	}
	if req.Method != core.MethodPost {
		t.Errorf("Method = %q, want %q", req.Method, core.MethodPost)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q, want %q", req.URL, "https://api.example.com/users")
	}
	if req.Path != path {
		t.Errorf("Path = %q, want %q", req.Path, path)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Headers[Content-Type] = %q, want %q", req.Headers["Content-Type"], "application/json")
	}
	if req.Body == "" {
		t.Error("Body is empty, expected non-empty body")
	}
}

func TestParseYMLFileNotFound(t *testing.T) {
	_, err := ParseYMLFile("nonexistent.yml")
	if err == nil {
		t.Error("ParseYMLFile with nonexistent file should return error, got nil")
	}
}
