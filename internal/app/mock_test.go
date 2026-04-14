package app

import "github.com/sanmoo/bruwrapper/internal/core"

type mockCatalog struct {
	collections []core.Collection
	requests    []core.Request
	resolved    core.Request
	resolveErr  error
	findErr     error
}

func (m *mockCatalog) FindCollections() ([]core.Collection, error) {
	return m.collections, m.findErr
}

func (m *mockCatalog) FindRequests(string) ([]core.Request, error) {
	return m.requests, m.findErr
}

func (m *mockCatalog) ResolveRequest(_, _ string) (core.Request, error) {
	return m.resolved, m.resolveErr
}

type mockPresenter struct {
	calledShowCollections bool
	calledShowRequests    bool
	receivedCollections   []core.Collection
	receivedRequests      []core.Request
	err                   error
}

func (m *mockPresenter) ShowResponse(core.Response, core.PresentOpts) error {
	return nil
}

func (m *mockPresenter) ShowRequestDetails(core.Request) error {
	return nil
}

func (m *mockPresenter) ShowCollections(collections []core.Collection) error {
	m.calledShowCollections = true
	m.receivedCollections = collections
	return m.err
}

func (m *mockPresenter) ShowRequests(requests []core.Request) error {
	m.calledShowRequests = true
	m.receivedRequests = requests
	return m.err
}

func (m *mockPresenter) ShowError(string) error {
	return nil
}
