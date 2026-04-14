package app

import (
	"context"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type RunApp struct {
	catalog   core.Catalog
	runner    core.Runner
	presenter core.Presenter
	selector  core.Selector
}

type RunParams struct {
	CollectionName string
	RequestName    string
	Env            string
	Variables      []core.Variable
	Raw            bool
	Verbose        bool
}

func NewRunApp(catalog core.Catalog, runner core.Runner, presenter core.Presenter, selector core.Selector) *RunApp {
	return &RunApp{catalog: catalog, runner: runner, presenter: presenter, selector: selector}
}

func (a *RunApp) Run(ctx context.Context, params RunParams) error {
	var collection core.Collection
	var request core.Request
	var err error

	if params.CollectionName == "" || params.RequestName == "" {
		collection, request, err = a.interactiveSelection(params.CollectionName)
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}
	} else {
		request, err = a.catalog.ResolveRequest(params.CollectionName, params.RequestName)
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}
		collections, err := a.catalog.FindCollections()
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}
		for _, c := range collections {
			if c.Name == params.CollectionName {
				collection = c
				break
			}
		}
	}

	runReq := core.RunRequest{
		CollectionPath: collection.Path,
		RequestPath:    request.Path,
		Env:            params.Env,
		Variables:      params.Variables,
	}

	response, err := a.runner.Execute(ctx, runReq)
	if err != nil {
		return a.presenter.ShowError(err.Error())
	}

	return a.presenter.ShowResponse(response, core.PresentOpts{
		Raw:     params.Raw,
		Verbose: params.Verbose,
	})
}

func (a *RunApp) interactiveSelection(collectionName string) (core.Collection, core.Request, error) {
	var collection core.Collection

	if collectionName == "" {
		collections, err := a.catalog.FindCollections()
		if err != nil {
			return core.Collection{}, core.Request{}, err
		}
		collection, err = a.selector.SelectCollection(collections)
		if err != nil {
			return core.Collection{}, core.Request{}, err
		}
	} else {
		collections, err := a.catalog.FindCollections()
		if err != nil {
			return core.Collection{}, core.Request{}, err
		}
		for _, c := range collections {
			if c.Name == collectionName {
				collection = c
				break
			}
		}
	}

	requests, err := a.catalog.FindRequests(collection.Name)
	if err != nil {
		return collection, core.Request{}, err
	}

	request, err := a.selector.SelectRequest(requests)
	if err != nil {
		return collection, core.Request{}, err
	}

	return collection, request, nil
}
