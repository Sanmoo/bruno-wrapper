package core

import "context"

type Catalog interface {
	FindCollections() ([]Collection, error)
	FindRequests(collectionName string) ([]Request, error)
	ResolveRequest(collectionName, requestName string) (Request, error)
}

type Runner interface {
	Execute(ctx context.Context, req RunRequest) (RunResult, error)
}

type Presenter interface {
	ShowResponse(result RunResult, opts PresentOpts) error
	ShowRequestDetails(req Request) error
	ShowCollections(collections []Collection) error
	ShowRequests(requests []Request) error
	ShowError(msg string) error
}

type Selector interface {
	SelectCollection(collections []Collection) (Collection, error)
	SelectRequest(requests []Request) (Request, error)
}

type ConfigLoader interface {
	Load() (Config, error)
}
