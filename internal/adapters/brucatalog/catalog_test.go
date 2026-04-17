package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestFindCollections(t *testing.T) {
	ymlPath := filepath.Join("testdata", "sample_yml_collection")
	badPath := filepath.Join("testdata", "nonexistent")

	cat := NewCatalog([]string{ymlPath, badPath})

	cols, err := cat.FindCollections()
	if err != nil {
		t.Fatalf("FindCollections() returned error: %v", err)
	}

	if len(cols) != 1 {
		t.Fatalf("FindCollections() returned %d collections, want 1", len(cols))
	}

	if cols[0].Name != "Sample YML Collection" {
		t.Errorf("expected to find 'Sample YML Collection', got %q", cols[0].Name)
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
	ymlPath := filepath.Join("testdata", "sample_yml_collection")
	cat := NewCatalog([]string{ymlPath})

	reqs, err := cat.FindRequests("Sample YML Collection")
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
			if r.Collection != "Sample YML Collection" {
				t.Errorf("Collection = %q, want %q", r.Collection, "Sample YML Collection")
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
	ymlPath := filepath.Join("testdata", "sample_yml_collection")
	cat := NewCatalog([]string{ymlPath})

	req, err := cat.ResolveRequest("Sample YML Collection", "Get Users")
	if err != nil {
		t.Fatalf("ResolveRequest() returned error: %v", err)
	}

	if req.Name != "Get Users" {
		t.Errorf("Name = %q, want %q", req.Name, "Get Users")
	}
	if req.Method != core.MethodGet {
		t.Errorf("Method = %q, want %q", req.Method, core.MethodGet)
	}
	if req.Collection != "Sample YML Collection" {
		t.Errorf("Collection = %q, want %q", req.Collection, "Sample YML Collection")
	}
}

func TestResolveRequestNotFound(t *testing.T) {
	ymlPath := filepath.Join("testdata", "sample_yml_collection")
	cat := NewCatalog([]string{ymlPath})

	_, err := cat.ResolveRequest("Sample YML Collection", "Nonexistent Request")
	if err == nil {
		t.Error("ResolveRequest with nonexistent request should return error, got nil")
	}
}
