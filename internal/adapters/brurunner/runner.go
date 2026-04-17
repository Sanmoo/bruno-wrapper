package brurunner

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type runner struct {
	bruPath string
}

func NewRunner(bruPath string) core.Runner {
	return &runner{bruPath: bruPath}
}

func FindBru() (string, error) {
	path, err := exec.LookPath("bru")
	if err != nil {
		return "", fmt.Errorf("bru binary not found in PATH: %w", err)
	}
	return path, nil
}

func (r *runner) Execute(ctx context.Context, req core.RunRequest) (core.RunResult, error) {
	reportFile, err := os.CreateTemp("", "bruwrapper-report-*.json")
	if err != nil {
		return core.RunResult{}, fmt.Errorf("creating temp file: %w", err)
	}
	reportPath := reportFile.Name()
	reportFile.Close()
	defer os.Remove(reportPath)

	args := []string{"run", req.RequestPath, "--reporter-json", reportPath}
	if req.Env != "" {
		args = append(args, "--env", req.Env)
	}
	for _, v := range req.Variables {
		args = append(args, "--env-var", fmt.Sprintf("%s=%s", v.Key, v.Value))
	}

	cmd := exec.CommandContext(ctx, r.bruPath, args...)
	cmd.Dir = req.CollectionPath
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	execErr := cmd.Run()

	reportData, readErr := os.ReadFile(reportPath)
	if readErr != nil {
		if execErr != nil {
			return core.RunResult{}, fmt.Errorf("bru run failed: %w", execErr)
		}
		return core.RunResult{}, fmt.Errorf("reading report file: %w", readErr)
	}

	report, parseErr := parseReport(reportData)
	if parseErr != nil {
		if execErr != nil {
			return core.RunResult{}, fmt.Errorf("bru run failed: %w", execErr)
		}
		return core.RunResult{}, fmt.Errorf("parsing report: %w", parseErr)
	}

	if len(report.Results) == 0 {
		return core.RunResult{}, fmt.Errorf("bru report contains no results")
	}

	result := report.Results[0]

	if result.Status == "error" || result.Status == "skipped" {
		if execErr != nil {
			return core.RunResult{}, fmt.Errorf("bru run failed (%s): %w", result.Status, execErr)
		}
		return core.RunResult{}, fmt.Errorf("bru run status: %s", result.Status)
	}

	statusCode, err := result.Response.StatusCode()
	if err != nil {
		return core.RunResult{}, fmt.Errorf("parsing status code: %w", err)
	}

	var body string
	if result.Response.Data != nil {
		switch d := result.Response.Data.(type) {
		case string:
			body = d
		default:
			b, err := json.MarshalIndent(d, "", "  ")
			if err != nil {
				body = fmt.Sprintf("%v", d)
			} else {
				body = string(b)
			}
		}
	}

	responseHeaders := result.Response.Headers
	if responseHeaders == nil {
		responseHeaders = map[string]string{}
	}

	requestHeaders := result.Request.Headers
	if requestHeaders == nil {
		requestHeaders = map[string]string{}
	}

	return core.RunResult{
		Request: core.RequestMeta{
			Headers: requestHeaders,
		},
		Response: core.Response{
			StatusCode: statusCode,
			StatusText: result.Response.StatusText,
			Headers:    responseHeaders,
			Body:       body,
			Duration:   result.Response.ResponseTime,
		},
	}, nil
}
