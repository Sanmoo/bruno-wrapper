package brurunner

import (
	"encoding/json"
	"fmt"
)

type BruReport struct {
	Summary BruReportSummary  `json:"summary"`
	Results []BruReportResult `json:"results"`
}

type BruIterationReport struct {
	IterationIndex int               `json:"iterationIndex"`
	Results        []BruReportResult `json:"results"`
	Summary        BruReportSummary  `json:"summary"`
}

type BruReportSummary struct {
	TotalRequests  int `json:"totalRequests"`
	PassedRequests int `json:"passedRequests"`
}

type BruReportResult struct {
	Test     BruReportTest     `json:"test"`
	Request  BruReportRequest  `json:"request"`
	Response BruReportResponse `json:"response"`
	Error    interface{}       `json:"error"`
	Status   string            `json:"status"`
	Name     string            `json:"name"`
	Path     string            `json:"path"`
}

type BruReportTest struct {
	Filename string `json:"filename"`
}

type BruReportRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Data    interface{}       `json:"data"`
}

type BruReportResponse struct {
	Status       json.RawMessage   `json:"status"`
	StatusText   string            `json:"statusText"`
	Headers      map[string]string `json:"headers"`
	Data         interface{}       `json:"data"`
	ResponseTime int64             `json:"responseTime"`
}

func (r *BruReportResponse) StatusCode() (int, error) {
	var intStatus int
	if err := json.Unmarshal(r.Status, &intStatus); err == nil {
		return intStatus, nil
	}
	var strStatus string
	if err := json.Unmarshal(r.Status, &strStatus); err == nil {
		return 0, fmt.Errorf("non-numeric status: %s", strStatus)
	}
	return 0, fmt.Errorf("invalid status value")
}

func parseReport(data []byte) (*BruReport, error) {
	var iterations []BruIterationReport
	if err := json.Unmarshal(data, &iterations); err == nil && len(iterations) > 0 {
		iter := iterations[0]
		return &BruReport{
			Summary: iter.Summary,
			Results: iter.Results,
		}, nil
	}

	var report BruReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing bru report: %w", err)
	}
	return &report, nil
}
