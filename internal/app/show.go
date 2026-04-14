package app

import "github.com/sanmoo/bruwrapper/internal/core"

type ShowApp struct {
	catalog   core.Catalog
	presenter core.Presenter
}

func NewShowApp(catalog core.Catalog, presenter core.Presenter) *ShowApp {
	return &ShowApp{catalog: catalog, presenter: presenter}
}

func (a *ShowApp) ShowRequestDetails(collectionName, requestName string) error {
	req, err := a.catalog.ResolveRequest(collectionName, requestName)
	if err != nil {
		return err
	}
	return a.presenter.ShowRequestDetails(req)
}
