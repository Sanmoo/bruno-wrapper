package app

import "github.com/sanmoo/bruwrapper/internal/core"

type ListApp struct {
	catalog   core.Catalog
	presenter core.Presenter
}

func NewListApp(catalog core.Catalog, presenter core.Presenter) *ListApp {
	return &ListApp{catalog: catalog, presenter: presenter}
}

func (a *ListApp) ListCollections() error {
	collections, err := a.catalog.FindCollections()
	if err != nil {
		return err
	}
	return a.presenter.ShowCollections(collections)
}

func (a *ListApp) ListRequests(collectionName string) error {
	requests, err := a.catalog.FindRequests(collectionName)
	if err != nil {
		return err
	}
	return a.presenter.ShowRequests(requests)
}
