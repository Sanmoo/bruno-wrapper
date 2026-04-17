package core

type Collection struct {
	Name string
	Path string
}

type RequestMethod string

const (
	MethodGet     RequestMethod = "GET"
	MethodPost    RequestMethod = "POST"
	MethodPut     RequestMethod = "PUT"
	MethodPatch   RequestMethod = "PATCH"
	MethodDelete  RequestMethod = "DELETE"
	MethodOptions RequestMethod = "OPTIONS"
	MethodHead    RequestMethod = "HEAD"
)

type Request struct {
	Name       string
	Method     RequestMethod
	URL        string
	Headers    map[string]string
	Body       string
	Collection string
	Path       string
}

type Variable struct {
	Key   string
	Value string
}

type RunRequest struct {
	CollectionPath string
	RequestPath    string
	Env            string
	Variables      []Variable
}

type Response struct {
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
	Duration   int64
}

type RunResult struct {
	Request  RequestMeta
	Response Response
}

type RequestMeta struct {
	Headers map[string]string
}

type Config struct {
	CollectionPaths []string
}

type PresentOpts struct {
	Raw     bool
	Verbose bool
}

type CollectionFormat string

const (
	FormatBru CollectionFormat = "bru"
	FormatYML CollectionFormat = "yml"
)
