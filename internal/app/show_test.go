package app

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestShowRequestDetails(t *testing.T) {
	req := core.Request{
		Name:   "GetUsers",
		Method: core.MethodGet,
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Accept": "application/json",
		},
	}

	catalog := &mockCatalog{resolved: req}
	var buf bytes.Buffer
	presenter := terminal.NewPresenter(&buf)
	app := NewShowApp(catalog, presenter)

	err := app.ShowRequestDetails("my-collection", "GetUsers")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected output to contain method GET, got %q", output)
	}
	if !strings.Contains(output, "https://api.example.com/users") {
		t.Errorf("expected output to contain URL, got %q", output)
	}
	if !strings.Contains(output, "Accept") {
		t.Errorf("expected output to contain header Accept, got %q", output)
	}
}

func TestShowRequestNotFound(t *testing.T) {
	catalog := &mockCatalog{resolveErr: errors.New("request not found")}
	var buf bytes.Buffer
	presenter := terminal.NewPresenter(&buf)
	app := NewShowApp(catalog, presenter)

	err := app.ShowRequestDetails("my-collection", "Missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "request not found" {
		t.Errorf("expected 'request not found' error, got %v", err)
	}
}
