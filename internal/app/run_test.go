package app

import (
	"context"
	"errors"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type mockRunner struct {
	resp    core.Response
	execute func(req core.RunRequest) (core.Response, error)
	err     error
}

func (m *mockRunner) Execute(ctx context.Context, req core.RunRequest) (core.Response, error) {
	if m.execute != nil {
		return m.execute(req)
	}
	return m.resp, m.err
}

type mockSelector struct {
	collection core.Collection
	request    core.Request
	err        error
}

func (m *mockSelector) SelectCollection(collections []core.Collection) (core.Collection, error) {
	if m.err != nil {
		return core.Collection{}, m.err
	}
	return m.collection, nil
}

func (m *mockSelector) SelectRequest(requests []core.Request) (core.Request, error) {
	if m.err != nil {
		return core.Request{}, m.err
	}
	return m.request, nil
}

type runSpyPresenter struct {
	calledShowResponse bool
	calledShowError    bool
	receivedResponse   core.Response
	receivedOpts       core.PresentOpts
	receivedErrorMsg   string
}

func (s *runSpyPresenter) ShowResponse(resp core.Response, opts core.PresentOpts) error {
	s.calledShowResponse = true
	s.receivedResponse = resp
	s.receivedOpts = opts
	return nil
}

func (s *runSpyPresenter) ShowRequestDetails(core.Request) error { return nil }

func (s *runSpyPresenter) ShowCollections([]core.Collection) error { return nil }

func (s *runSpyPresenter) ShowRequests([]core.Request) error { return nil }

func (s *runSpyPresenter) ShowError(msg string) error {
	s.calledShowError = true
	s.receivedErrorMsg = msg
	return nil
}

func TestRunWithExplicitParams(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
		resolved:    request,
	}
	runner := &mockRunner{resp: core.Response{StatusCode: 200, StatusText: "OK", Body: `{"ok":true}`, Duration: 100}}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "My API",
		RequestName:    "Get Users",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !presenter.calledShowResponse {
		t.Error("expected ShowResponse to be called")
	}
	if presenter.receivedResponse.StatusCode != 200 {
		t.Errorf("expected status code 200, got %d", presenter.receivedResponse.StatusCode)
	}
}

func TestRunWithVarsAndEnv(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	var capturedReq core.RunRequest
	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
		resolved:    request,
	}
	runner := &mockRunner{
		execute: func(req core.RunRequest) (core.Response, error) {
			capturedReq = req
			return core.Response{StatusCode: 200, StatusText: "OK", Body: `{}`}, nil
		},
	}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "My API",
		RequestName:    "Get Users",
		Env:            "staging",
		Variables:      []core.Variable{{Key: "token", Value: "abc123"}},
		Raw:            true,
		Verbose:        true,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedReq.Env != "staging" {
		t.Errorf("expected env 'staging', got %q", capturedReq.Env)
	}
	if len(capturedReq.Variables) != 1 || capturedReq.Variables[0].Key != "token" {
		t.Errorf("expected variable token, got %v", capturedReq.Variables)
	}
	if !presenter.receivedOpts.Raw {
		t.Error("expected Raw option to be true")
	}
	if !presenter.receivedOpts.Verbose {
		t.Error("expected Verbose option to be true")
	}
}

func TestRunRequestNotFound(t *testing.T) {
	catalog := &mockCatalog{resolveErr: errors.New("request not found")}
	runner := &mockRunner{}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "My API",
		RequestName:    "Missing",
	})
	if err != nil {
		t.Fatal("expected error to be handled by presenter, got non-nil return")
	}
	if !presenter.calledShowError {
		t.Error("expected ShowError to be called")
	}
}

func TestRunCollectionNotFound(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
		resolved:    request,
	}
	runner := &mockRunner{}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "Unknown API",
		RequestName:    "Get Users",
	})
	if err != nil {
		t.Fatal("expected error to be handled by presenter, got non-nil return")
	}
	if !presenter.calledShowError {
		t.Error("expected ShowError to be called")
	}
}

func TestRunRunnerError(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
		resolved:    request,
	}
	runner := &mockRunner{err: errors.New("bru failed")}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "My API",
		RequestName:    "Get Users",
	})
	if err != nil {
		t.Fatal("expected error to be handled by presenter, got non-nil return")
	}
	if !presenter.calledShowError {
		t.Error("expected ShowError to be called")
	}
}

func TestRunWithInteractiveSelection(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
	}
	runner := &mockRunner{resp: core.Response{StatusCode: 200, StatusText: "OK", Body: `{}`}}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{
		collection: collection,
		request:    request,
	}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !presenter.calledShowResponse {
		t.Error("expected ShowResponse to be called")
	}
}

func TestRunWithPartialInteractiveSelection(t *testing.T) {
	collection := core.Collection{Name: "My API", Path: "/path/api"}
	request := core.Request{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com", Path: "/path/api/users.bru"}

	catalog := &mockCatalog{
		collections: []core.Collection{collection},
		requests:    []core.Request{request},
	}
	runner := &mockRunner{resp: core.Response{StatusCode: 200, StatusText: "OK", Body: `{}`}}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{
		request: request,
	}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{
		CollectionName: "My API",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !presenter.calledShowResponse {
		t.Error("expected ShowResponse to be called")
	}
}

func TestRunInteractiveSelectionError(t *testing.T) {
	catalog := &mockCatalog{
		collections: []core.Collection{{Name: "My API", Path: "/path/api"}},
	}
	runner := &mockRunner{}
	presenter := &runSpyPresenter{}
	selector := &mockSelector{err: errors.New("selection cancelled")}

	app := NewRunApp(catalog, runner, presenter, selector)

	err := app.Run(context.Background(), RunParams{})
	if err != nil {
		t.Fatal("expected error to be handled by presenter, got non-nil return")
	}
	if !presenter.calledShowError {
		t.Error("expected ShowError to be called")
	}
}
