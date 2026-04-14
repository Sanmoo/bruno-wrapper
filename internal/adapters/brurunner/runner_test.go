package brurunner

import (
	"encoding/json"
	"testing"
)

func TestParseReport(t *testing.T) {
	raw := `{
		"summary": {"totalRequests": 1, "passedRequests": 1},
		"results": [
			{
				"test": {"filename": "health/check.bru"},
				"request": {"method": "GET", "url": "https://api.example.com/health", "headers": {"Accept": "application/json"}, "data": null},
				"response": {"status": 200, "statusText": "OK", "headers": {"Content-Type": "application/json"}, "data": {"healthy": true}, "responseTime": 142},
				"error": null,
				"status": "pass",
				"name": "Health Check",
				"path": "health/check.bru"
			}
		]
	}`

	report, err := parseReport([]byte(raw))
	if err != nil {
		t.Fatalf("parseReport() error: %v", err)
	}

	if len(report.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(report.Results))
	}

	r := report.Results[0]
	if r.Name != "Health Check" {
		t.Errorf("Name = %q, want %q", r.Name, "Health Check")
	}
	if r.Status != "pass" {
		t.Errorf("Status = %q, want %q", r.Status, "pass")
	}
	if r.Path != "health/check.bru" {
		t.Errorf("Path = %q, want %q", r.Path, "health/check.bru")
	}
	if r.Test.Filename != "health/check.bru" {
		t.Errorf("Test.Filename = %q, want %q", r.Test.Filename, "health/check.bru")
	}

	code, err := r.Response.StatusCode()
	if err != nil {
		t.Fatalf("StatusCode() error: %v", err)
	}
	if code != 200 {
		t.Errorf("StatusCode() = %d, want 200", code)
	}
	if r.Response.StatusText != "OK" {
		t.Errorf("StatusText = %q, want %q", r.Response.StatusText, "OK")
	}
	if r.Request.Method != "GET" {
		t.Errorf("Request.Method = %q, want %q", r.Request.Method, "GET")
	}
	if r.Request.URL != "https://api.example.com/health" {
		t.Errorf("Request.URL = %q, want %q", r.Request.URL, "https://api.example.com/health")
	}
	if r.Response.ResponseTime != 142 {
		t.Errorf("ResponseTime = %d, want 142", r.Response.ResponseTime)
	}
}

func TestParseReportInvalidJSON(t *testing.T) {
	_, err := parseReport([]byte(`{invalid json`))
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestParseResponseStatusInt(t *testing.T) {
	resp := BruReportResponse{
		Status: json.RawMessage(`200`),
	}
	code, err := resp.StatusCode()
	if err != nil {
		t.Fatalf("StatusCode() error: %v", err)
	}
	if code != 200 {
		t.Errorf("StatusCode() = %d, want 200", code)
	}
}

func TestParseResponseStatusString(t *testing.T) {
	resp := BruReportResponse{
		Status: json.RawMessage(`"error"`),
	}
	code, err := resp.StatusCode()
	if err == nil {
		t.Error("expected error for string status, got nil")
	}
	if code != 0 {
		t.Errorf("StatusCode() = %d, want 0 for error status", code)
	}
}
