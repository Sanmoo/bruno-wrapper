package app

import (
	"errors"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestListCollections(t *testing.T) {
	collections := []core.Collection{
		{Name: "col1", Path: "/path/col1"},
		{Name: "col2", Path: "/path/col2"},
	}
	catalog := &mockCatalog{collections: collections}
	presenter := &mockPresenter{}
	app := NewListApp(catalog, presenter)

	err := app.ListCollections()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !presenter.calledShowCollections {
		t.Error("expected ShowCollections to be called")
	}
	if len(presenter.receivedCollections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(presenter.receivedCollections))
	}
	if presenter.receivedCollections[0].Name != "col1" {
		t.Errorf("expected col1, got %s", presenter.receivedCollections[0].Name)
	}
	if presenter.receivedCollections[1].Name != "col2" {
		t.Errorf("expected col2, got %s", presenter.receivedCollections[1].Name)
	}
}

func TestListCollectionsError(t *testing.T) {
	catalog := &mockCatalog{findErr: errors.New("catalog error")}
	presenter := &mockPresenter{}
	app := NewListApp(catalog, presenter)

	err := app.ListCollections()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "catalog error" {
		t.Errorf("expected 'catalog error', got '%s'", err.Error())
	}
	if presenter.calledShowCollections {
		t.Error("expected ShowCollections not to be called")
	}
}

func TestListRequests(t *testing.T) {
	requests := []core.Request{
		{Name: "req1", Method: core.MethodGet, Collection: "col1"},
		{Name: "req2", Method: core.MethodPost, Collection: "col1"},
	}
	catalog := &mockCatalog{requests: requests}
	presenter := &mockPresenter{}
	app := NewListApp(catalog, presenter)

	err := app.ListRequests("col1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !presenter.calledShowRequests {
		t.Error("expected ShowRequests to be called")
	}
	if len(presenter.receivedRequests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(presenter.receivedRequests))
	}
	if presenter.receivedRequests[0].Name != "req1" {
		t.Errorf("expected req1, got %s", presenter.receivedRequests[0].Name)
	}
	if presenter.receivedRequests[1].Name != "req2" {
		t.Errorf("expected req2, got %s", presenter.receivedRequests[1].Name)
	}
}

func TestListRequestsError(t *testing.T) {
	catalog := &mockCatalog{findErr: errors.New("catalog error")}
	presenter := &mockPresenter{}
	app := NewListApp(catalog, presenter)

	err := app.ListRequests("col1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "catalog error" {
		t.Errorf("expected 'catalog error', got '%s'", err.Error())
	}
	if presenter.calledShowRequests {
		t.Error("expected ShowRequests not to be called")
	}
}
