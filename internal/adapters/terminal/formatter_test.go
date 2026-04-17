package terminal

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestShowResponse(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	resp := core.Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"id":1,"name":"Leanne Graham"}`,
		Duration:   142,
	}

	err := p.ShowResponse(core.RunResult{Response: resp}, core.PresentOpts{})
	if err != nil {
		t.Fatalf("ShowResponse returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Status: 200 OK") {
		t.Errorf("expected status line 'Status: 200 OK', got:\n%s", output)
	}
	if !strings.Contains(output, "Time:   142ms") {
		t.Errorf("expected time line 'Time:   142ms', got:\n%s", output)
	}
	if !strings.Contains(output, `"id": 1`) {
		t.Errorf("expected pretty-printed JSON with indented keys, got:\n%s", output)
	}
	if !strings.Contains(output, `"name": "Leanne Graham"`) {
		t.Errorf("expected pretty-printed JSON with name field, got:\n%s", output)
	}
}

func TestShowResponseRaw(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	resp := core.Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       `{"id":1,"name":"Leanne Graham"}`,
		Duration:   142,
	}

	err := p.ShowResponse(core.RunResult{Response: resp}, core.PresentOpts{Raw: true})
	if err != nil {
		t.Fatalf("ShowResponse returned error: %v", err)
	}

	output := buf.String()

	if strings.Contains(output, "Status:") {
		t.Errorf("raw mode should not contain status line, got:\n%s", output)
	}
	if strings.Contains(output, "Time:") {
		t.Errorf("raw mode should not contain time line, got:\n%s", output)
	}
	if !strings.Contains(output, `{"id":1,"name":"Leanne Graham"}`) {
		t.Errorf("raw mode should output raw body, got:\n%s", output)
	}
}

func TestShowResponseVerbose(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	result := core.RunResult{
		Request: core.RequestMeta{Headers: map[string]string{}},
		Response: core.Response{
			StatusCode: 200,
			StatusText: "OK",
			Headers: map[string]string{
				"Content-Type":          "application/json",
				"X-RateLimit-Remaining": "59",
			},
			Body:     `{"id": 1}`,
			Duration: 142,
		},
	}

	err := p.ShowResponse(result, core.PresentOpts{Verbose: true})
	if err != nil {
		t.Fatalf("ShowResponse returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Response Headers:") {
		t.Errorf("verbose mode should contain 'Response Headers:', got:\n%s", output)
	}
	if !strings.Contains(output, "Content-Type: application/json") {
		t.Errorf("verbose mode should list Content-Type header, got:\n%s", output)
	}
	if !strings.Contains(output, "X-RateLimit-Remaining: 59") {
		t.Errorf("verbose mode should list X-RateLimit-Remaining header, got:\n%s", output)
	}
}

func TestShowResponseVerboseShowsRequestHeaders(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	result := core.RunResult{
		Request: core.RequestMeta{
			Headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			},
		},
		Response: core.Response{
			StatusCode: 200,
			StatusText: "OK",
			Headers: map[string]string{
				"Content-Type":          "application/json",
				"X-RateLimit-Remaining": "59",
			},
			Body:     `{"id": 1}`,
			Duration: 142,
		},
	}

	err := p.ShowResponse(result, core.PresentOpts{Verbose: true})
	if err != nil {
		t.Fatalf("ShowResponse returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Request Headers:") {
		t.Errorf("verbose mode should contain 'Request Headers:', got:\n%s", output)
	}

	if !strings.Contains(output, "Accept: application/json") {
		t.Errorf("verbose mode should list Accept header, got:\n%s", output)
	}

	if !strings.Contains(output, "Authorization: ***") {
		t.Errorf("verbose mode should mask Authorization header, got:\n%s", output)
	}
	if strings.Contains(output, "secret-token") {
		t.Errorf("verbose mode should not expose secret token, got:\n%s", output)
	}

	if !strings.Contains(output, "Response Headers:") {
		t.Errorf("verbose mode should contain 'Response Headers:', got:\n%s", output)
	}

	if !strings.Contains(output, "X-RateLimit-Remaining: 59") {
		t.Errorf("verbose mode should list X-RateLimit-Remaining header unmasked, got:\n%s", output)
	}

	requestIdx := strings.Index(output, "Request Headers:")
	responseIdx := strings.Index(output, "Response Headers:")
	if requestIdx >= responseIdx {
		t.Errorf("request headers should appear before response headers, got request at %d and response at %d", requestIdx, responseIdx)
	}
}

func TestShowCollections(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	collections := []core.Collection{
		{Name: "My API", Path: "/home/user/myapi"},
		{Name: "Other API", Path: "/home/user/other"},
	}

	err := p.ShowCollections(collections)
	if err != nil {
		t.Fatalf("ShowCollections returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Collections:") {
		t.Errorf("expected 'Collections:' header, got:\n%s", output)
	}
	if !strings.Contains(output, "My API (/home/user/myapi)") {
		t.Errorf("expected 'My API (/home/user/myapi)', got:\n%s", output)
	}
	if !strings.Contains(output, "Other API (/home/user/other)") {
		t.Errorf("expected 'Other API (/home/user/other)', got:\n%s", output)
	}
}

func TestShowRequests(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	requests := []core.Request{
		{Name: "Get Users", Method: core.MethodGet},
		{Name: "Create User", Method: core.MethodPost},
	}

	err := p.ShowRequests(requests)
	if err != nil {
		t.Fatalf("ShowRequests returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Requests:") {
		t.Errorf("expected 'Requests:' header, got:\n%s", output)
	}
	if !strings.Contains(output, "GET    Get Users") {
		t.Errorf("expected 'GET    Get Users', got:\n%s", output)
	}
	if !strings.Contains(output, "POST   Create User") {
		t.Errorf("expected 'POST   Create User', got:\n%s", output)
	}
}

func TestShowError(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	err := p.ShowError("something went wrong")
	if err != nil {
		t.Fatalf("ShowError returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Error: something went wrong\n") {
		t.Errorf("expected 'Error: something went wrong\\n', got:\n%s", output)
	}
}

func TestShowRequestDetails(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	req := core.Request{
		Name:   "Get Users",
		Method: core.MethodGet,
		URL:    "https://api.example.com/users",
		Headers: map[string]string{
			"Authorization": "Bearer secret-token",
			"Content-Type":  "application/json",
		},
	}

	err := p.ShowRequestDetails(req)
	if err != nil {
		t.Fatalf("ShowRequestDetails returned error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Method: GET") {
		t.Errorf("expected 'Method: GET', got:\n%s", output)
	}
	if !strings.Contains(output, "URL:    https://api.example.com/users") {
		t.Errorf("expected 'URL:    https://api.example.com/users', got:\n%s", output)
	}
	if !strings.Contains(output, "Authorization: ***") {
		t.Errorf("expected Authorization header to be masked, got:\n%s", output)
	}
	if !strings.Contains(output, "Content-Type: application/json") {
		t.Errorf("expected Content-Type header unmasked, got:\n%s", output)
	}
}
