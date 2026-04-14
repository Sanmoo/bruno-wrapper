package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestFindCollections(t *testing.T) {
	bruPath := filepath.Join("testdata", "sample_bru_collection")
	ymlPath := filepath.Join("testdata", "sample_yml_collection")
	badPath := filepath.Join("testdata", "nonexistent")

	cat := NewCatalog([]string{bruPath, ymlPath, badPath})

	cols, err := cat.FindCollections()
	if err != nil {
		t.Fatalf("FindCollections() returned error: %v", err)
	}

	if len(cols) != 2 {
		t.Fatalf("FindCollections() returned %d collections, want 2", len(cols))
	}

	names := map[string]bool{}
	for _, c := range cols {
		names[c.Name] = true
	}
	if !names["Sample Bru Collection"] {
		t.Error("expected to find 'Sample Bru Collection'")
	}
	if !names["Sample YML Collection"] {
		t.Error("expected to find 'Sample YML Collection'")
	}
}

func TestFindCollectionsEmpty(t *testing.T) {
	cat := NewCatalog([]string{})
	cols, err := cat.FindCollections()
	if err != nil {
		t.Fatalf("FindCollections() with empty paths returned error: %v", err)
	}
	if len(cols) != 0 {
		t.Errorf("FindCollections() with empty paths returned %d collections, want 0", len(cols))
	}
}

func TestFindRequests(t *testing.T) {
	bruPath := filepath.Join("testdata", "sample_bru_collection")
	cat := NewCatalog([]string{bruPath})

	reqs, err := cat.FindRequests("Sample Bru Collection")
	if err != nil {
		t.Fatalf("FindRequests() returned error: %v", err)
	}

	if len(reqs) == 0 {
		t.Fatal("FindRequests() returned 0 requests, want at least 1")
	}

	found := false
	for _, r := range reqs {
		if r.Name == "Get Users" {
			found = true
			if r.Method != core.MethodGet {
				t.Errorf("Method = %q, want %q", r.Method, core.MethodGet)
			}
			if r.URL != "https://api.example.com/users" {
				t.Errorf("URL = %q, want %q", r.URL, "https://api.example.com/users")
			}
			if r.Collection != "Sample Bru Collection" {
				t.Errorf("Collection = %q, want %q", r.Collection, "Sample Bru Collection")
			}
			break
		}
	}
	if !found {
		t.Error("expected to find request named 'Get Users'")
	}
}

func TestFindRequestsCollectionNotFound(t *testing.T) {
	cat := NewCatalog([]string{})
	_, err := cat.FindRequests("Nonexistent")
	if err == nil {
		t.Error("FindRequests with nonexistent collection should return error, got nil")
	}
}

func TestResolveRequest(t *testing.T) {
	bruPath := filepath.Join("testdata", "sample_bru_collection")
	cat := NewCatalog([]string{bruPath})

	req, err := cat.ResolveRequest("Sample Bru Collection", "Get Users")
	if err != nil {
		t.Fatalf("ResolveRequest() returned error: %v", err)
	}

	if req.Name != "Get Users" {
		t.Errorf("Name = %q, want %q", req.Name, "Get Users")
	}
	if req.Method != core.MethodGet {
		t.Errorf("Method = %q, want %q", req.Method, core.MethodGet)
	}
	if req.Collection != "Sample Bru Collection" {
		t.Errorf("Collection = %q, want %q", req.Collection, "Sample Bru Collection")
	}
}

func TestResolveRequestNotFound(t *testing.T) {
	bruPath := filepath.Join("testdata", "sample_bru_collection")
	cat := NewCatalog([]string{bruPath})

	_, err := cat.ResolveRequest("Sample Bru Collection", "Nonexistent Request")
	if err == nil {
		t.Error("ResolveRequest with nonexistent request should return error, got nil")
	}
}
