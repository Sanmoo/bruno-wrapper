package brucatalog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestParseBruFile(t *testing.T) {
	path := filepath.Join("testdata", "get_users.bru")
	req, err := ParseBruFile(path)
	if err != nil {
		t.Fatalf("ParseBruFile(%q) returned error: %v", path, err)
	}

	if req.Name != "Get Users" {
		t.Errorf("Name = %q, want %q", req.Name, "Get Users")
	}
	if req.Method != core.MethodGet {
		t.Errorf("Method = %q, want %q", req.Method, core.MethodGet)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q, want %q", req.URL, "https://api.example.com/users")
	}
	if req.Path != path {
		t.Errorf("Path = %q, want %q", req.Path, path)
	}

	expectedHeaders := map[string]string{
		"Authorization": "Bearer {{token}}",
		"Content-Type":  "application/json",
	}
	for k, v := range expectedHeaders {
		if req.Headers[k] != v {
			t.Errorf("Headers[%q] = %q, want %q", k, req.Headers[k], v)
		}
	}
}

func TestParseBruFileWithBody(t *testing.T) {
	path := filepath.Join("testdata", "create_user.bru")
	req, err := ParseBruFile(path)
	if err != nil {
		t.Fatalf("ParseBruFile(%q) returned error: %v", path, err)
	}

	if req.Name != "Create User" {
		t.Errorf("Name = %q, want %q", req.Name, "Create User")
	}
	if req.Method != core.MethodPost {
		t.Errorf("Method = %q, want %q", req.Method, core.MethodPost)
	}
	if req.Body == "" {
		t.Error("Body is empty, expected non-empty body")
	}
}

func TestParseBruFileNotFound(t *testing.T) {
	_, err := ParseBruFile("nonexistent.bru")
	if err == nil {
		t.Error("ParseBruFile with nonexistent file should return error, got nil")
	}
}

func TestParseDisabledHeaders(t *testing.T) {
	dir := t.TempDir()
	bruContent := `meta {
  name: Disabled Headers Test
  type: http
  seq: 1
}

get {
  url: https://api.example.com/test
}

headers {
  ~Authorization: Bearer {{token}}
  Content-Type: application/json
}
`
	path := filepath.Join(dir, "disabled_headers.bru")
	if err := os.WriteFile(path, []byte(bruContent), 0644); err != nil {
		t.Fatal(err)
	}

	req, err := ParseBruFile(path)
	if err != nil {
		t.Fatalf("ParseBruFile returned error: %v", err)
	}

	if _, ok := req.Headers["Authorization"]; ok {
		t.Error("Disabled header '~Authorization' should be skipped")
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Content-Type = %q, want %q", req.Headers["Content-Type"], "application/json")
	}
}
