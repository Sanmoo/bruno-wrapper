# bruwrapper Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI that wraps the Bruno `bru` CLI, providing better UX for ad-hoc API consumption: directory-agnostic execution, interactive request selection, easy variable overriding, and pretty-printed JSON output.

**Architecture:** Hexagonal/Clean Architecture with core domain (models + ports), app use cases, and adapters for each concern (filesystem catalog, bru subprocess runner, terminal output, interactive selection, config). The domain has zero external dependencies. Adapters implement ports with concrete technology.

**Tech Stack:** Go 1.26, Cobra CLI, Bubbletea + Bubbles (interactive UI), Lipgloss (styling), gopkg.in/yaml.v3

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd /home/sanmoo/dev/github.com/Sanmoo/bruno-wrapper
go mod init github.com/sanmoo/bruwrapper
```

- [ ] **Step 2: Create directory structure**

```bash
mkdir -p cmd internal/core internal/app internal/adapters/brucatalog internal/adapters/brurunner internal/adapters/terminal internal/adapters/interactive internal/adapters/yamlconfig
```

- [ ] **Step 3: Install Cobra dependency**

```bash
go get github.com/spf13/cobra@latest
```

- [ ] **Step 4: Create `cmd/root.go`**

```go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bruwrapper",
	Short: "A CLI wrapper for Bruno API client",
	Long:  "bruwrapper wraps the Bruno CLI (bru) providing better UX for ad-hoc API consumption.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 5: Create `main.go`**

```go
package main

import "github.com/sanmoo/bruwrapper/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 6: Verify it builds and runs**

```bash
go build -o bruwrapper . && ./bruwrapper --help
```

Expected: help output showing `bruwrapper` usage.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "feat: project scaffold with cobra root command"
```

---

### Task 2: Core Domain Models

**Files:**
- Create: `internal/core/model.go`

- [ ] **Step 1: Create domain model types**

```go
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
	Vars       map[string]string
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/core/...
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: core domain models"
```

---

### Task 3: Core Domain Ports

**Files:**
- Create: `internal/core/ports.go`

- [ ] **Step 1: Define port interfaces**

```go
package core

import "context"

type Catalog interface {
	FindCollections() ([]Collection, error)
	FindRequests(collectionName string) ([]Request, error)
	ResolveRequest(collectionName, requestName string) (Request, error)
}

type Runner interface {
	Execute(ctx context.Context, req RunRequest) (Response, error)
}

type Presenter interface {
	ShowResponse(resp Response, opts PresentOpts) error
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/core/...
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: core port interfaces"
```

---

### Task 4: Config Adapter (yamlconfig)

**Files:**
- Create: `internal/adapters/yamlconfig/config.go`
- Create: `internal/adapters/yamlconfig/config_test.go`

- [ ] **Step 1: Write failing test for config loading**

```go
package yamlconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestLoadValidConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".bruwrapper.yaml")
	content := `collections:
  - ~/projects/myapi
  - ~/work/other-api
  - /absolute/path/to/collection
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := New(configPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.CollectionPaths) != 3 {
		t.Fatalf("expected 3 collection paths, got %d", len(cfg.CollectionPaths))
	}
	if cfg.CollectionPaths[0] != "~/projects/myapi" {
		t.Errorf("expected ~/projects/myapi, got %s", cfg.CollectionPaths[0])
	}
	if cfg.CollectionPaths[2] != "/absolute/path/to/collection" {
		t.Errorf("expected /absolute/path/to/collection, got %s", cfg.CollectionPaths[2])
	}
}

func TestLoadMissingConfig(t *testing.T) {
	loader := New("/nonexistent/path/.bruwrapper.yaml")
	_, err := loader.Load()
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoadEmptyCollections(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".bruwrapper.yaml")
	content := `collections: []
`
	err := os.WriteFile(configPath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := New(configPath)
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.CollectionPaths) != 0 {
		t.Fatalf("expected 0 collection paths, got %d", len(cfg.CollectionPaths))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/adapters/yamlconfig/... -v
```

Expected: compilation error (New doesn't exist yet).

- [ ] **Step 3: Implement config loader**

```go
package yamlconfig

import (
	"fmt"
	"os"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

type yamlConfig struct {
	CollectionPaths []string `yaml:"collections"`
}

type configLoader struct {
	path string
}

func New(path string) core.ConfigLoader {
	return &configLoader{path: path}
}

func (l *configLoader) Load() (core.Config, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return core.Config{}, fmt.Errorf("config file not found at %s — create it with your collection paths", l.path)
		}
		return core.Config{}, fmt.Errorf("reading config file: %w", err)
	}

	var yc yamlConfig
	if err := yaml.Unmarshal(data, &yc); err != nil {
		return core.Config{}, fmt.Errorf("parsing config file: %w", err)
	}

	return core.Config{
		CollectionPaths: yc.CollectionPaths,
	}, nil
}

func DefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return home + "/.bruwrapper.yaml"
}
```

- [ ] **Step 4: Install YAML dependency**

```bash
go get gopkg.in/yaml.v3
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/adapters/yamlconfig/... -v
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: config adapter with yaml loading"
```

---

### Task 5: Bru File Parser (brucatalog)

**Files:**
- Create: `internal/adapters/brucatalog/bru_parser.go`
- Create: `internal/adapters/brucatalog/bru_parser_test.go`
- Create: `internal/adapters/brucatalog/testdata/get_users.bru`
- Create: `internal/adapters/brucatalog/testdata/create_user.bru`

- [ ] **Step 1: Create test fixture files**

`internal/adapters/brucatalog/testdata/get_users.bru`:
```
meta {
  name: Get Users
  type: http
  seq: 1
}

get {
  url: https://api.example.com/users
  headers: {
    Authorization: Bearer {{token}}
    Content-Type: application/json
  }
}

headers {
  Authorization: Bearer {{token}}
  Content-Type: application/json
}
```

`internal/adapters/brucatalog/testdata/create_user.bru`:
```
meta {
  name: Create User
  type: http
  seq: 2
}

post {
  url: https://api.example.com/users
}

headers {
  Content-Type: application/json
}

body {
  {
    "name": "{{userName}}"
  }
}

vars:pre-request {
  userName: defaultName
}
```

- [ ] **Step 2: Write failing tests**

```go
package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestParseBruFile(t *testing.T) {
	path := filepath.Join("testdata", "get_users.bru")
	req, err := ParseBruFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Name != "Get Users" {
		t.Errorf("expected name 'Get Users', got %s", req.Name)
	}
	if req.Method != core.MethodGet {
		t.Errorf("expected method GET, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("expected URL 'https://api.example.com/users', got %s", req.URL)
	}
	if req.Headers["Authorization"] != "Bearer {{token}}" {
		t.Errorf("expected Authorization header with variable, got %s", req.Headers["Authorization"])
	}
}

func TestParseBruFileWithBody(t *testing.T) {
	path := filepath.Join("testdata", "create_user.bru")
	req, err := ParseBruFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Name != "Create User" {
		t.Errorf("expected name 'Create User', got %s", req.Name)
	}
	if req.Method != core.MethodPost {
		t.Errorf("expected method POST, got %s", req.Method)
	}
	if req.Body == "" {
		t.Error("expected body to be non-empty")
	}
	if req.Vars["userName"] != "defaultName" {
		t.Errorf("expected var userName=defaultName, got %s", req.Vars["userName"])
	}
}

func TestParseBruFileNotFound(t *testing.T) {
	_, err := ParseBruFile("nonexistent.bru")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestParseBru
```

Expected: compilation error (ParseBruFile doesn't exist yet).

- [ ] **Step 4: Implement bru parser**

```go
package brucatalog

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type bruBlock struct {
	name    string
	content string
}

func ParseBruFile(path string) (core.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Request{}, fmt.Errorf("reading bru file %s: %w", path, err)
	}

	blocks, err := parseBlocks(string(data))
	if err != nil {
		return core.Request{}, fmt.Errorf("parsing bru blocks from %s: %w", path, err)
	}

	req := core.Request{Path: path}

	for _, block := range blocks {
		switch block.name {
		case "meta":
			parseMetaBlock(block.content, &req)
		case "get", "post", "put", "patch", "delete", "options", "head":
			req.Method = core.RequestMethod(strings.ToUpper(block.name))
			parseMethodBlock(block.content, &req)
		case "headers":
			req.Headers = parseDictionaryBlock(block.content)
		case "body":
			req.Body = strings.TrimSpace(block.content)
		case "body:text", "body:xml", "body:json", "body:graphql", "body:graphql:vars":
			req.Body = strings.TrimSpace(block.content)
		case "body:form-urlencoded", "body:multipart-form":
			req.Body = strings.TrimSpace(block.content)
		case "vars:pre-request", "vars:post-response":
			req.Vars = parseDictionaryBlock(block.content)
		}
	}

	return req, nil
}

func parseBlocks(content string) ([]bruBlock, error) {
	var blocks []bruBlock
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentName string
	var currentContent strings.Builder
	braceDepth := 0

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if braceDepth == 0 {
			if strings.Contains(trimmed, "{") && !strings.Contains(trimmed, "{{") {
				parts := strings.SplitN(trimmed, "{", 2)
				currentName = strings.TrimSpace(parts[0])
				braceDepth = 1
				afterBrace := ""
				if len(parts) > 1 {
					afterBrace = strings.TrimPrefix(parts[1], " ")
				}
				if strings.Contains(afterBrace, "}") && !strings.Contains(afterBrace, "{{") {
					currentContent.WriteString(afterBrace)
					blocks = append(blocks, bruBlock{name: currentName, content: currentContent.String()})
					currentContent.Reset()
					currentName = ""
					braceDepth = 0
				} else {
					currentContent.WriteString(afterBrace)
					currentContent.WriteString("\n")
				}
			}
		} else {
			openBraces := strings.Count(line, "{") - strings.Count(line, "{{")
			closeBraces := strings.Count(line, "}") - strings.Count(line, "}}")
			braceDepth += openBraces - closeBraces

			if braceDepth <= 0 {
				lastClose := strings.LastIndex(line, "}")
				if lastClose > 0 {
					currentContent.WriteString(line[:lastClose])
				}
				blocks = append(blocks, bruBlock{name: currentName, content: currentContent.String()})
				currentContent.Reset()
				currentName = ""
				braceDepth = 0
			} else {
				currentContent.WriteString(line)
				currentContent.WriteString("\n")
			}
		}
	}

	return blocks, nil
}

func parseMetaBlock(content string, req *core.Request) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name:") {
			req.Name = strings.TrimSpace(strings.TrimPrefix(line, "name:"))
		}
	}
}

func parseMethodBlock(content string, req *core.Request) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "url:") {
			req.URL = strings.TrimSpace(strings.TrimPrefix(line, "url:"))
		}
	}
}

func parseDictionaryBlock(content string) map[string]string {
	result := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "~") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				result[key] = value
			}
		}
	}
	return result
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestParseBru
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: bru file parser for catalog discovery"
```

---

### Task 6: YML (OpenCollection) Parser (brucatalog)

**Files:**
- Create: `internal/adapters/brucatalog/yml_parser.go`
- Create: `internal/adapters/brucatalog/yml_parser_test.go`
- Create: `internal/adapters/brucatalog/testdata/create_user.yml`

- [ ] **Step 1: Create test fixture**

`internal/adapters/brucatalog/testdata/create_user.yml`:
```yaml
meta:
  name: Create User
  type: http
  seq: 2
http:
  method: POST
  url: https://api.example.com/users
headers:
  Content-Type: application/json
body:
  type: json
  content: |
    {
      "name": "{{userName}}"
    }
```

- [ ] **Step 2: Write failing tests**

```go
package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestParseYMLFile(t *testing.T) {
	path := filepath.Join("testdata", "create_user.yml")
	req, err := ParseYMLFile(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Name != "Create User" {
		t.Errorf("expected name 'Create User', got %s", req.Name)
	}
	if req.Method != core.MethodPost {
		t.Errorf("expected method POST, got %s", req.Method)
	}
	if req.URL != "https://api.example.com/users" {
		t.Errorf("expected URL 'https://api.example.com/users', got %s", req.URL)
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("expected Content-Type header, got %s", req.Headers["Content-Type"])
	}
}

func TestParseYMLFileNotFound(t *testing.T) {
	_, err := ParseYMLFile("nonexistent.yml")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestParseYML
```

Expected: compilation error (ParseYMLFile doesn't exist yet).

- [ ] **Step 4: Implement YML parser**

```go
package brucatalog

import (
	"fmt"
	"os"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
	"gopkg.in/yaml.v3"
)

type openCollectionRequest struct {
	Meta struct {
		Name string `yaml:"name"`
		Type string `yaml:"type"`
		Seq  int    `yaml:"seq"`
	} `yaml:"meta"`
	HTTP struct {
		Method string `yaml:"method"`
		URL    string `yaml:"url"`
	} `yaml:"http"`
	Headers map[string]string `yaml:"headers"`
	Body   struct {
		Type    string `yaml:"type"`
		Content string `yaml:"content"`
	} `yaml:"body"`
}

func ParseYMLFile(path string) (core.Request, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return core.Request{}, fmt.Errorf("reading yml file %s: %w", path, err)
	}

	var ocr openCollectionRequest
	if err := yaml.Unmarshal(data, &ocr); err != nil {
		return core.Request{}, fmt.Errorf("parsing yml file %s: %w", path, err)
	}

	req := core.Request{
		Name:    ocr.Meta.Name,
		Method:  core.RequestMethod(strings.ToUpper(ocr.HTTP.Method)),
		URL:     ocr.HTTP.URL,
		Headers: ocr.Headers,
		Body:    strings.TrimSpace(ocr.Body.Content),
		Path:    path,
	}

	if req.Headers == nil {
		req.Headers = make(map[string]string)
	}

	return req, nil
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestParseYML
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: yml parser for OpenCollection format"
```

---

### Task 7: Catalog Adapter (filesystem scanner)

**Files:**
- Create: `internal/adapters/brucatalog/catalog.go`
- Create: `internal/adapters/brucatalog/catalog_test.go`
- Create: `internal/adapters/brucatalog/testdata/sample_bru_collection/bruno.json`
- Create: `internal/adapters/brucatalog/testdata/sample_bru_collection/users/get_users.bru`
- Create: `internal/adapters/brucatalog/testdata/sample_yml_collection/opencollection.yml`

- [ ] **Step 1: Create test fixtures**

`internal/adapters/brucatalog/testdata/sample_bru_collection/bruno.json`:
```json
{
  "version": "1",
  "name": "Sample Bru Collection",
  "type": "collection"
}
```

`internal/adapters/brucatalog/testdata/sample_bru_collection/users/get_users.bru`:
```
meta {
  name: Get Users
  type: http
  seq: 1
}

get {
  url: https://api.example.com/users
}

headers {
  Authorization: Bearer {{token}}
}
```

`internal/adapters/brucatalog/testdata/sample_yml_collection/opencollection.yml`:
```yaml
version: "1"
name: Sample YML Collection
type: collection
```

- [ ] **Step 2: Write failing tests**

```go
package brucatalog

import (
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestFindCollections(t *testing.T) {
	bruPath := filepath.Join("testdata", "sample_bru_collection")
	ymlPath := filepath.Join("testdata", "sample_yml_collection")

	cat := NewCatalog([]string{bruPath, ymlPath, "/nonexistent"})
	collections, err := cat.FindCollections()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(collections) != 2 {
		t.Fatalf("expected 2 collections, got %d", len(collections))
	}

	foundBru := false
	foundYML := false
	for _, c := range collections {
		if c.Name == "Sample Bru Collection" {
			foundBru = true
		}
		if c.Name == "Sample YML Collection" {
			foundYML = true
		}
	}
	if !foundBru {
		t.Error("expected to find Sample Bru Collection")
	}
	if !foundYML {
		t.Error("expected to find Sample YML Collection")
	}
}

func TestFindRequests(t *testing.T) {
	bruPath, _ := filepath.Abs(filepath.Join("testdata", "sample_bru_collection"))
	cat := NewCatalog([]string{bruPath})

	requests, err := cat.FindRequests("Sample Bru Collection")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(requests) == 0 {
		t.Fatal("expected at least 1 request, got 0")
	}

	found := false
	for _, r := range requests {
		if r.Name == "Get Users" {
			found = true
			if r.Method != core.MethodGet {
				t.Errorf("expected GET method, got %s", r.Method)
			}
		}
	}
	if !found {
		t.Error("expected to find 'Get Users' request")
	}
}

func TestResolveRequest(t *testing.T) {
	bruPath, _ := filepath.Abs(filepath.Join("testdata", "sample_bru_collection"))
	cat := NewCatalog([]string{bruPath})

	req, err := cat.ResolveRequest("Sample Bru Collection", "Get Users")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.Name != "Get Users" {
		t.Errorf("expected 'Get Users', got %s", req.Name)
	}
	if req.Method != core.MethodGet {
		t.Errorf("expected GET, got %s", req.Method)
	}
}

func TestFindCollectionsEmpty(t *testing.T) {
	cat := NewCatalog([]string{})
	collections, err := cat.FindCollections()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(collections) != 0 {
		t.Errorf("expected 0 collections, got %d", len(collections))
	}
}

func TestFindRequestsCollectionNotFound(t *testing.T) {
	cat := NewCatalog([]string{})
	_, err := cat.FindRequests("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent collection")
	}
}

func TestResolveRequestNotFound(t *testing.T) {
	bruPath, _ := filepath.Abs(filepath.Join("testdata", "sample_bru_collection"))
	cat := NewCatalog([]string{bruPath})

	_, err := cat.ResolveRequest("Sample Bru Collection", "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent request")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestFind
```

Expected: compilation error (NewCatalog doesn't exist yet).

- [ ] **Step 4: Implement catalog**

```go
package brucatalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type catalog struct {
	paths []string
}

func NewCatalog(paths []string) core.Catalog {
	return &catalog{paths: paths}
}

func (c *catalog) FindCollections() ([]core.Collection, error) {
	var collections []core.Collection
	for _, p := range c.paths {
		absPath, err := expandPath(p)
		if err != nil {
			continue
		}

		name, format, err := detectCollection(absPath)
		if err != nil {
			continue
		}

		collections = append(collections, core.Collection{
			Name: name,
			Path: absPath,
		})
		_ = format
	}
	return collections, nil
}

func (c *catalog) FindRequests(collectionName string) ([]core.Request, error) {
	collPath, err := c.resolveCollectionPath(collectionName)
	if err != nil {
		return nil, err
	}

	_, format, err := detectCollection(collPath)
	if err != nil {
		return nil, err
	}

	ext := ".bru"
	if format == core.FormatYML {
		ext = ".yml"
	}

	var requests []core.Request
	err = filepath.Walk(collPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ext) {
			if ext == ".bru" {
				req, parseErr := ParseBruFile(path)
				if parseErr != nil {
					return nil
				}
				req.Collection = collectionName
				requests = append(requests, req)
			} else {
				req, parseErr := ParseYMLFile(path)
				if parseErr != nil {
					return nil
				}
				req.Collection = collectionName
				requests = append(requests, req)
			}
		}
		return nil
	})
	return requests, err
}

func (c *catalog) ResolveRequest(collectionName, requestName string) (core.Request, error) {
	requests, err := c.FindRequests(collectionName)
	if err != nil {
		return core.Request{}, err
	}
	for _, r := range requests {
		if r.Name == requestName {
			return r, nil
		}
	}
	return core.Request{}, fmt.Errorf("request %q not found in collection %q", requestName, collectionName)
}

func (c *catalog) resolveCollectionPath(name string) (string, error) {
	collections, err := c.FindCollections()
	if err != nil {
		return "", err
	}
	for _, coll := range collections {
		if coll.Name == name {
			return coll.Path, nil
		}
	}
	return "", fmt.Errorf("collection %q not found in config", name)
}

func detectCollection(dirPath string) (string, core.CollectionFormat, error) {
	brunoJSON := filepath.Join(dirPath, "bruno.json")
	if _, err := os.Stat(brunoJSON); err == nil {
		data, err := os.ReadFile(brunoJSON)
		if err != nil {
			return "", "", err
		}
		var meta struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(data, &meta); err != nil {
			return "", "", err
		}
		return meta.Name, core.FormatBru, nil
	}

	ocYML := filepath.Join(dirPath, "opencollection.yml")
	if _, err := os.Stat(ocYML); err == nil {
		data, err := os.ReadFile(ocYML)
		if err != nil {
			return "", "", err
		}
		var meta struct {
			Name string `yaml:"name"`
		}
		if err := yamlUnmarshal(data, &meta); err != nil {
			return "", "", err
		}
		return meta.Name, core.FormatYML, nil
	}

	return "", "", fmt.Errorf("no bruno.json or opencollection.yml found in %s", dirPath)
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Abs(path)
}

func yamlUnmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
```

Note: needs `gopkg.in/yaml.v3` import. Add it to imports.

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./internal/adapters/brucatalog/... -v -run TestFind
```

- [ ] **Step 6: Run all brucatalog tests**

```bash
go test ./internal/adapters/brucatalog/... -v
```

- [ ] **Step 7: Fix any test failures and iterate**

- [ ] **Step 8: Commit**

```bash
git add -A && git commit -m "feat: catalog adapter for filesystem scanning"
```

---

### Task 8: Bru Runner Adapter (brurunner)

**Files:**
- Create: `internal/adapters/brurunner/runner.go`
- Create: `internal/adapters/brurunner/runner_test.go`
- Create: `internal/adapters/brurunner/report.go`

- [ ] **Step 1: Define JSON report structs and parser**

```go
package brurunner

type BruReport struct {
	Summary BruReportSummary `json:"summary"`
	Results []BruReportResult `json:"results"`
}

type BruReportSummary struct {
	TotalRequests   int `json:"totalRequests"`
	PassedRequests  int `json:"passedRequests"`
	FailedRequests  int `json:"failedRequests"`
	ErrorRequests   int `json:"errorRequests"`
	SkippedRequests int `json:"skippedRequests"`
}

type BruReportResult struct {
	Test     BruReportTest     `json:"test"`
	Request  BruReportRequest  `json:"request"`
	Response BruReportResponse `json:"response"`
	Status   string            `json:"status"`
	Error    string            `json:"error"`
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
	Status       int               `json:"status"`
	StatusText   string            `json:"statusText"`
	Headers      map[string]string `json:"headers"`
	Data         interface{}       `json:"data"`
	ResponseTime int64             `json:"responseTime"`
}

func parseReport(data []byte) (*BruReport, error) {
	var report BruReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("parsing bru report: %w", err)
	}
	return &report, nil
}
```

Note: add `encoding/json` and `fmt` to imports.

- [ ] **Step 2: Write failing test for runner**

```go
package brurunner

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestParseReport(t *testing.T) {
	report := BruReport{
		Summary: BruReportSummary{TotalRequests: 1, PassedRequests: 1},
		Results: []BruReportResult{
			{
				Name:   "Get Users",
				Status: "pass",
				Request: BruReportRequest{
					Method:  "GET",
					URL:     "https://api.example.com/users",
					Headers: map[string]string{"Content-Type": "application/json"},
				},
				Response: BruReportResponse{
					Status:       200,
					StatusText:   "OK",
					ResponseTime: 142,
					Headers:      map[string]string{"content-type": "application/json"},
					Data:         map[string]interface{}{"id": float64(1), "name": "Leanne Graham"},
				},
			},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := parseReport(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed.Summary.TotalRequests != 1 {
		t.Errorf("expected 1 total request, got %d", parsed.Summary.TotalRequests)
	}
	if len(parsed.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(parsed.Results))
	}
	if parsed.Results[0].Name != "Get Users" {
		t.Errorf("expected 'Get Users', got %s", parsed.Results[0].Name)
	}
	if parsed.Results[0].Response.Status != 200 {
		t.Errorf("expected 200, got %d", parsed.Results[0].Response.Status)
	}
}

func TestParseReportInvalidJSON(t *testing.T) {
	_, err := parseReport([]byte("not json"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExecuteIntegration(t *testing.T) {
	if os.Getenv("BRUWRAPPER_TEST_BRU") == "" {
		t.Skip("Skipping integration test (set BRUWRAPPER_TEST_BRU=1 to run)")
	}

	dir := t.TempDir()
	reportPath := filepath.Join(dir, "report.json")

	runner := NewRunner("bru")
	_, err := runner.Execute(context.Background(), core.RunRequest{
		CollectionPath: os.Getenv("BRUWRAPPER_TEST_COLLECTION"),
		RequestPath:    os.Getenv("BRUWRAPPER_TEST_REQUEST"),
	})
	_ = reportPath
	if err != nil {
		t.Logf("bru execution failed (expected if no collection configured): %v", err)
	}
}
```

- [ ] **Step 3: Run test to verify report parsing works**

```bash
go test ./internal/adapters/brurunner/... -v -run TestParseReport
```

Expected: compilation error (parseReport needs to be properly imported/scoped). Fix as needed.

- [ ] **Step 4: Implement runner**

```go
package brurunner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
		return "", fmt.Errorf("bru CLI not found. Install it with: npm i -g @usebruno/cli\nSee https://www.usebruno.com for more information")
	}
	return path, nil
}

func (r *runner) Execute(ctx context.Context, req core.RunRequest) (core.Response, error) {
	tmpFile, err := os.CreateTemp("", "bruwrapper-report-*.json")
	if err != nil {
		return core.Response{}, fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	args := []string{"run", req.RequestPath}
	if req.Env != "" {
		args = append(args, "--env", req.Env)
	}
	for _, v := range req.Variables {
		args = append(args, "--env-var", fmt.Sprintf("%s=%s", v.Key, v.Value))
	}
	args = append(args, "--reporter-json", tmpPath)

	cmd := exec.CommandContext(ctx, r.bruPath, args...)
	cmd.Dir = req.CollectionPath
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	reportData, readErr := os.ReadFile(tmpPath)

	if readErr != nil {
		if err != nil {
			return core.Response{}, fmt.Errorf("bru run failed and report unavailable: %w", err)
		}
		return core.Response{}, fmt.Errorf("reading bru report: %w", readErr)
	}

	report, parseErr := parseReport(reportData)
	if parseErr != nil {
		if err != nil {
			return core.Response{}, fmt.Errorf("bru run failed (exit code %d) and report parse failed", cmd.ProcessState.ExitCode())
		}
		return core.Response{}, fmt.Errorf("parsing bru report: %w", parseErr)
	}

	if len(report.Results) == 0 {
		return core.Response{}, fmt.Errorf("no results in bru report")
	}

	result := report.Results[0]

	var bodyStr string
	if result.Response.Data != nil {
		bodyBytes, jsonErr := json.MarshalIndent(result.Response.Data, "", "  ")
		if jsonErr == nil {
			bodyStr = string(bodyBytes)
		} else {
			bodyStr = fmt.Sprintf("%v", result.Response.Data)
		}
	}

	responseHeaders := make(map[string]string)
	if result.Response.Headers != nil {
		responseHeaders = result.Response.Headers
	}

	response := core.Response{
		StatusCode: result.Response.Status,
		StatusText: result.Response.StatusText,
		Headers:   responseHeaders,
		Body:      bodyStr,
		Duration:   result.Response.ResponseTime,
	}

	if result.Status == "error" && result.Error != "" {
		return response, fmt.Errorf("request error: %s", result.Error)
	}

	return response, nil
}
```

Note: add required imports. Also add a helper to convert report path to absolute.

- [ ] **Step 5: Run all runner tests**

```bash
go test ./internal/adapters/brurunner/... -v
```

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: bru runner adapter with report parsing"
```

---

### Task 9: Terminal Presenter Adapter

**Files:**
- Create: `internal/adapters/terminal/formatter.go`
- Create: `internal/adapters/terminal/formatter_test.go`

- [ ] **Step 1: Write failing tests for presenter**

```go
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
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       `{"id": 1, "name": "Leanne Graham"}`,
		Duration:   142,
	}

	err := p.ShowResponse(resp, core.PresentOpts{Raw: false, Verbose: false})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Status: 200 OK") {
		t.Errorf("expected output to contain 'Status: 200 OK', got %s", output)
	}
	if !strings.Contains(output, "Time:") {
		t.Errorf("expected output to contain 'Time:', got %s", output)
	}
}

func TestShowResponseRaw(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	resp := core.Response{
		StatusCode: 200,
		StatusText: "OK",
		Body:       `{"id":1}`,
		Duration:   50,
	}

	err := p.ShowResponse(resp, core.PresentOpts{Raw: true, Verbose: false})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, `{"id":1}`) {
		t.Errorf("expected raw JSON output, got %s", output)
	}
	if strings.Contains(output, "Status:") {
		t.Errorf("raw output should not contain Status header, got %s", output)
	}
}

func TestShowResponseVerbose(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	resp := core.Response{
		StatusCode: 200,
		StatusText: "OK",
		Headers:    map[string]string{"content-type": "application/json"},
		Body:       `{"id": 1}`,
		Duration:   50,
	}

	err := p.ShowResponse(resp, core.PresentOpts{Raw: false, Verbose: true})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Response Headers") {
		t.Errorf("verbose output should contain 'Response Headers', got %s", output)
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
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "My API") {
		t.Errorf("expected output to contain 'My API', got %s", output)
	}
	if !strings.Contains(output, "Other API") {
		t.Errorf("expected output to contain 'Other API', got %s", output)
	}
}

func TestShowRequests(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	requests := []core.Request{
		{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com/users"},
		{Name: "Create User", Method: core.MethodPost, URL: "https://api.example.com/users"},
	}

	err := p.ShowRequests(requests)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected output to contain 'GET', got %s", output)
	}
	if !strings.Contains(output, "POST") {
		t.Errorf("expected output to contain 'POST', got %s", output)
	}
}

func TestShowError(t *testing.T) {
	var buf bytes.Buffer
	p := NewPresenter(&buf)

	err := p.ShowError("something went wrong")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "something went wrong") {
		t.Errorf("expected output to contain error message, got %s", output)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/adapters/terminal/... -v
```

Expected: compilation error.

- [ ] **Step 3: Implement presenter**

```go
package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type presenter struct {
	w io.Writer
}

func NewPresenter(w io.Writer) core.Presenter {
	return &presenter{w: w}
}

func (p *presenter) ShowResponse(resp core.Response, opts core.PresentOpts) error {
	if opts.Raw {
		fmt.Fprintln(p.w, resp.Body)
		return nil
	}

	fmt.Fprintf(p.w, "Status: %d %s\n", resp.StatusCode, resp.StatusText)
	fmt.Fprintf(p.w, "Time:   %dms\n\n", resp.Duration)

	if opts.Verbose && resp.Headers != nil {
		fmt.Fprintln(p.w, "Response Headers:")
		keys := sortedKeys(resp.Headers)
		for _, k := range keys {
			fmt.Fprintf(p.w, "  %s: %s\n", k, resp.Headers[k])
		}
		fmt.Fprintln(p.w)
	}

	body := resp.Body
	if body != "" {
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(body), &jsonBody); err == nil {
			pretty, err := json.MarshalIndent(jsonBody, "", "  ")
			if err == nil {
				fmt.Fprintln(p.w, string(pretty))
			} else {
				fmt.Fprintln(p.w, body)
			}
		} else {
			fmt.Fprintln(p.w, body)
		}
	}

	return nil
}

func (p *presenter) ShowRequestDetails(req core.Request) error {
	fmt.Fprintf(p.w, "Method: %s\n", req.Method)
	fmt.Fprintf(p.w, "URL:    %s\n", req.URL)
	if len(req.Headers) > 0 {
		fmt.Fprintln(p.w, "\nHeaders:")
		keys := sortedKeys(req.Headers)
		for _, k := range keys {
			fmt.Fprintf(p.w, "  %s: %s\n", k, maskSensitive(k, req.Headers[k]))
		}
	}
	if req.Body != "" {
		fmt.Fprintln(p.w, "\nBody:")
		var jsonBody interface{}
		if err := json.Unmarshal([]byte(req.Body), &jsonBody); err == nil {
			pretty, err := json.MarshalIndent(jsonBody, "", "  ")
			if err == nil {
				fmt.Fprintln(p.w, string(pretty))
			} else {
				fmt.Fprintln(p.w, req.Body)
			}
		} else {
			fmt.Fprintln(p.w, req.Body)
		}
	}
	return nil
}

func (p *presenter) ShowCollections(collections []core.Collection) error {
	if len(collections) == 0 {
		fmt.Fprintln(p.w, "No collections found.")
		fmt.Fprintln(p.w, "Add collection paths to ~/.bruwrapper.yaml")
		return nil
	}
	fmt.Fprintln(p.w, "Collections:")
	for _, c := range collections {
		fmt.Fprintf(p.w, "  %s (%s)\n", c.Name, c.Path)
	}
	return nil
}

func (p *presenter) ShowRequests(requests []core.Request) error {
	if len(requests) == 0 {
		fmt.Fprintln(p.w, "No requests found.")
		return nil
	}
	fmt.Fprintln(p.w, "Requests:")
	for _, r := range requests {
		fmt.Fprintf(p.w, "  %-6s %s\n", r.Method, r.Name)
	}
	return nil
}

func (p *presenter) ShowError(msg string) error {
	fmt.Fprintf(p.w, "Error: %s\n", msg)
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func maskSensitive(key string, value string) string {
	lower := strings.ToLower(key)
	if strings.Contains(lower, "auth") || strings.Contains(lower, "token") || strings.Contains(lower, "key") || strings.Contains(lower, "secret") || strings.Contains(lower, "password") {
		return "***"
	}
	return value
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/adapters/terminal/... -v
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: terminal presenter adapter"
```

---

### Task 10: Interactive Selector Adapter

**Files:**
- Create: `internal/adapters/interactive/selector.go`

- [ ] **Step 1: Install bubbletea dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
```

- [ ] **Step 2: Implement selector with bubbletea**

```go
package interactive

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type collectionItem core.Collection
type requestItem core.Request

func (i collectionItem) FilterValue() string { return i.Name }
func (i requestItem) FilterValue() string { return i.Name }

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5"))
	selectedItem = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

type selectorModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m selectorModel) Init() tea.Cmd {
	return nil
}

func (m selectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if item, ok := m.list.SelectedItem().(delegateItem); ok {
				m.choice = item.title
			}
			m.quitting = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectorModel) View() string {
	if m.quitting {
		return ""
	}
	return "\n" + m.list.View()
}

type delegateItem struct {
	title string
	desc  string
}

func (d delegateItem) FilterValue() string { return d.title }

type selector struct{}

func NewSelector() core.Selector {
	return &selector{}
}

func (s *selector) SelectCollection(collections []core.Collection) (core.Collection, error) {
	items := make([]list.Item, len(collections))
	for i, c := range collections {
		items[i] = delegateItem{title: c.Name, desc: c.Path}
	}

	l := list.New(items, list.NewDefaultDelegate([]rune{}), 0, 0)
	l.Title = "Select a Collection"
	l.Styles.Title = titleStyle

	m := selectorModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return core.Collection{}, fmt.Errorf("running selector: %w", err)
	}

	fm := result.(selectorModel)
	if fm.choice == "" {
		return core.Collection{}, fmt.Errorf("no collection selected")
	}

	for _, c := range collections {
		if c.Name == fm.choice {
			return c, nil
		}
	}
	return core.Collection{}, fmt.Errorf("collection %q not found", fm.choice)
}

func (s *selector) SelectRequest(requests []core.Request) (core.Request, error) {
	items := make([]list.Item, len(requests))
	for i, r := range requests {
		desc := string(r.Method) + " " + r.URL
		items[i] = delegateItem{title: r.Name, desc: desc}
	}

	l := list.New(items, list.NewDefaultDelegate([]rune{}), 0, 0)
	l.Title = "Select a Request"
	l.Styles.Title = titleStyle

	m := selectorModel{list: l}
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return core.Request{}, fmt.Errorf("running selector: %w", err)
	}

	fm := result.(selectorModel)
	if fm.choice == "" {
		return core.Request{}, fmt.Errorf("no request selected")
	}

	for _, r := range requests {
		if r.Name == fm.choice {
			return r, nil
		}
	}
	return core.Request{}, fmt.Errorf("request %q not found", fm.choice)
}
```

- [ ] **Step 3: Verify it compiles**

```bash
go build ./internal/adapters/interactive/...
```

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: interactive selector with bubbletea"
```

---

### Task 11: App Use Case — List

**Files:**
- Create: `internal/app/list.go`
- Create: `internal/app/list_test.go`

- [ ] **Step 1: Write failing test for List use case**

```go
package app

import (
	"bytes"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/core"
)

type mockCatalog struct {
	collections []core.Collection
	requests    []core.Request
	err         error
}

func (m *mockCatalog) FindCollections() ([]core.Collection, error) {
	return m.collections, m.err
}

func (m *mockCatalog) FindRequests(collectionName string) ([]core.Request, error) {
	return m.requests, m.err
}

func (m *mockCatalog) ResolveRequest(collectionName, requestName string) (core.Request, error) {
	return core.Request{}, m.err
}

func TestListCollections(t *testing.T) {
	catalog := &mockCatalog{
		collections: []core.Collection{
			{Name: "My API", Path: "/home/user/myapi"},
			{Name: "Other", Path: "/home/user/other"},
		},
	}

	var buf bytes.Buffer
	presenter := terminal.NewPresenter(&buf)

	listApp := NewListApp(catalog, nil, presenter)
	err := listApp.ListCollections()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("My API")) {
		t.Errorf("expected output to contain 'My API', got %s", output)
	}
}

func TestListRequests(t *testing.T) {
	catalog := &mockCatalog{
		requests: []core.Request{
			{Name: "Get Users", Method: core.MethodGet, URL: "https://api.example.com/users"},
			{Name: "Create User", Method: core.MethodPost, URL: "https://api.example.com/users"},
		},
	}

	var buf bytes.Buffer
	presenter := terminal.NewPresenter(&buf)

	listApp := NewListApp(catalog, nil, presenter)
	err := listApp.ListRequests("My API")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("GET")) {
		t.Errorf("expected output to contain 'GET', got %s", output)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/app/... -v -run TestList
```

Expected: compilation error.

- [ ] **Step 3: Implement List use case**

```go
package app

import (
	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/core"
)

type ListApp struct {
	catalog   core.Catalog
	selector  core.Selector
	presenter core.Presenter
}

func NewListApp(catalog core.Catalog, selector core.Selector, presenter core.Presenter) *ListApp {
	return &ListApp{
		catalog:   catalog,
		selector:  selector,
		presenter: presenter,
	}
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
```

Note: the test imports `terminal` adapter directly; this is for unit testing only. In production, main.go wires everything.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/app/... -v -run TestList
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: list use case"
```

---

### Task 12: App Use Case — Show

**Files:**
- Create: `internal/app/show.go`

- [ ] **Step 1: Implement Show use case**

```go
package app

type ShowApp struct {
	catalog   core.Catalog
	presenter core.Presenter
}

func NewShowApp(catalog core.Catalog, presenter core.Presenter) *ShowApp {
	return &ShowApp{
		catalog:   catalog,
		presenter: presenter,
	}
}

func (a *ShowApp) ShowRequestDetails(collectionName, requestName string) error {
	req, err := a.catalog.ResolveRequest(collectionName, requestName)
	if err != nil {
		return err
	}
	return a.presenter.ShowRequestDetails(req)
}
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/app/...
```

- [ ] **Step 3: Write a simple test**

```go
// Add to internal/app/show_test.go
package app

import (
	"bytes"
	"testing"

	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/core"
)

func TestShowRequestDetails(t *testing.T) {
	catalog := &mockCatalog{
		requests: []core.Request{},
		err: nil,
	}

	// Override ResolveRequest for this test
	resolvedReq := core.Request{
		Name:    "Get Users",
		Method:  core.MethodGet,
		URL:     "https://api.example.com/users",
		Headers: map[string]string{"Authorization": "Bearer token"},
	}

	var buf bytes.Buffer
	presenter := terminal.NewPresenter(&buf)

	showApp := NewShowApp(&resolveMock{req: resolvedReq}, presenter)
	err := showApp.ShowRequestDetails("My API", "Get Users")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	output := buf.String()
	if !bytes.Contains([]byte(output), []byte("GET")) {
		t.Errorf("expected output to contain 'GET', got %s", output)
	}
	if !bytes.Contains([]byte(output), []byte("https://api.example.com/users")) {
		t.Errorf("expected output to contain URL, got %s", output)
	}
}

type resolveMock struct {
	req core.Request
	err error
}

func (m *resolveMock) FindCollections() ([]core.Collection, error) { return nil, nil }
func (m *resolveMock) FindRequests(collectionName string) ([]core.Request, error) { return nil, nil }
func (m *resolveMock) ResolveRequest(collectionName, requestName string) (core.Request, error) {
	return m.req, m.err
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/app/... -v
```

- [ ] **Step 5: Commit**

```bash
git add -A && git commit -m "feat: show use case"
```

---

### Task 13: App Use Case — Run

**Files:**
- Create: `internal/app/run.go`

- [ ] **Step 1: Implement Run use case**

```go
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

func NewRunApp(catalog core.Catalog, runner core.Runner, presenter core.Presenter, selector core.Selector) *RunApp {
	return &RunApp{
		catalog:   catalog,
		runner:    runner,
		presenter: presenter,
		selector:  selector,
	}
}

type RunParams struct {
	CollectionName string
	RequestName    string
	Env            string
	Variables      []core.Variable
	Raw            bool
	Verbose        bool
}

func (a *RunApp) Run(ctx context.Context, params RunParams) error {
	var collection core.Collection
	var request core.Request
	var err error

	if params.CollectionName == "" || params.RequestName == "" {
		if params.CollectionName == "" {
			collections, err := a.catalog.FindCollections()
			if err != nil {
				return a.presenter.ShowError(err.Error())
			}
			collection, err = a.selector.SelectCollection(collections)
			if err != nil {
				return err
			}
		} else {
			colls, err := a.catalog.FindCollections()
			if err != nil {
				return a.presenter.ShowError(err.Error())
			}
			for _, c := range colls {
				if c.Name == params.CollectionName {
					collection = c
					break
				}
			}
		}

		requests, err := a.catalog.FindRequests(collection.Name)
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}
		request, err = a.selector.SelectRequest(requests)
		if err != nil {
			return err
		}
	} else {
		request, err = a.catalog.ResolveRequest(params.CollectionName, params.RequestName)
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}

		colls, err := a.catalog.FindCollections()
		if err != nil {
			return a.presenter.ShowError(err.Error())
		}
		for _, c := range colls {
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
```

- [ ] **Step 2: Verify it compiles**

```bash
go build ./internal/app/...
```

- [ ] **Step 3: Commit**

```bash
git add -A && git commit -m "feat: run use case"
```

---

### Task 14: CLI Commands (Wiring)

**Files:**
- Create: `cmd/run.go`
- Create: `cmd/list.go`
- Create: `cmd/show.go`

- [ ] **Step 1: Create run command**

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sanmoo/bruwrapper/internal/adapters/brucatalog"
	"github.com/sanmoo/bruwrapper/internal/adapters/brurunner"
	"github.com/sanmoo/bruwrapper/internal/adapters/interactive"
	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/adapters/yamlconfig"
	"github.com/sanmoo/bruwrapper/internal/app"
	"github.com/sanmoo/bruwrapper/internal/core"
)

var (
	runCollection string
	runRequest    string
	runEnv        string
	runVars       []string
	runRaw        bool
	runVerbose    bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a Bruno request",
	Long:  "Run a request from a Bruno collection. Opens interactive selection if -c and -r are not provided.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, catalog, runner, presenter, selector, err := wireUp()
		if err != nil {
			return err
		}
		_ = cfg
		runApp := app.NewRunApp(catalog, runner, presenter, selector)

		var vars []core.Variable
		for _, v := range runVars {
			parts := strings.SplitN(v, "=", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid variable format %q, expected key=value", v)
			}
			vars = append(vars, core.Variable{Key: parts[0], Value: parts[1]})
		}

		return runApp.Run(context.Background(), app.RunParams{
			CollectionName: runCollection,
			RequestName:    runRequest,
			Env:            runEnv,
			Variables:      vars,
			Raw:            runRaw,
			Verbose:        runVerbose,
		})
	},
}

func init() {
	runCmd.Flags().StringVarP(&runCollection, "collection", "c", "", "Collection name")
	runCmd.Flags().StringVarP(&runRequest, "request", "r", "", "Request name")
	runCmd.Flags().StringVarP(&runEnv, "env", "e", "", "Environment name")
	runCmd.Flags().StringArrayVarP(&runVars, "var", "v", nil, "Variable override (key=value, repeatable)")
	runCmd.Flags().BoolVar(&runRaw, "raw", false, "Output raw response without pretty-print")
	runCmd.Flags().BoolVar(&runVerbose, "verbose", false, "Show request and response headers")
	rootCmd.AddCommand(runCmd)
}
```

- [ ] **Step 2: Create list command**

```go
package cmd

import (
	"github.com/spf13/cobra"
)

var listCollection string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections or requests",
	Long:  "List available collections, or list requests in a specific collection.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, catalog, _, presenter, _, err := wireUp()
		if err != nil {
			return err
		}
		listApp := app.NewListApp(catalog, nil, presenter)

		if listCollection == "" {
			return listApp.ListCollections()
		}
		return listApp.ListRequests(listCollection)
	},
}

func init() {
	listCmd.Flags().StringVarP(&listCollection, "collection", "c", "", "Collection name to list requests from")
	rootCmd.AddCommand(listCmd)
}
```

- [ ] **Step 3: Create show command**

```go
package cmd

import (
	"github.com/spf13/cobra"
)

var (
	showCollection string
	showRequest    string
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show request details before executing",
	Long:  "Show the method, URL, headers, and body of a request without executing it.",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, catalog, _, presenter, _, err := wireUp()
		if err != nil {
			return err
		}
		showApp := app.NewShowApp(catalog, presenter)
		return showApp.ShowRequestDetails(showCollection, showRequest)
	},
}

func init() {
	showCmd.Flags().StringVarP(&showCollection, "collection", "c", "", "Collection name (required)")
	showCmd.Flags().StringVarP(&showRequest, "request", "r", "", "Request name (required)")
	showCmd.MarkFlagRequired("collection")
	showCmd.MarkFlagRequired("request")
	rootCmd.AddCommand(showCmd)
}
```

- [ ] **Step 4: Create wireUp helper**

```go
package cmd

import (
	"fmt"

	"github.com/sanmoo/bruwrapper/internal/adapters/brucatalog"
	"github.com/sanmoo/bruwrapper/internal/adapters/brurunner"
	"github.com/sanmoo/bruwrapper/internal/adapters/interactive"
	"github.com/sanmoo/bruwrapper/internal/adapters/terminal"
	"github.com/sanmoo/bruwrapper/internal/adapters/yamlconfig"
	"github.com/sanmoo/bruwrapper/internal/core"
)

func wireUp() (core.Config, core.Catalog, core.Runner, core.Presenter, core.Selector, error) {
	cfgLoader := yamlconfig.New(yamlconfig.DefaultConfigPath())
	cfg, err := cfgLoader.Load()
	if err != nil {
		return core.Config{}, nil, nil, nil, nil, fmt.Errorf("loading config: %w\n\nCreate ~/.bruwrapper.yaml with your collection paths", err)
	}

	catalog := brucatalog.NewCatalog(cfg.CollectionPaths)

	bruPath, err := brurunner.FindBru()
	if err != nil {
		return core.Config{}, nil, nil, nil, nil, err
	}
	runner := brurunner.NewRunner(bruPath)

	presenter := terminal.NewPresenter(terminal.NewStdWriter())
	selector := interactive.NewSelector()

	return cfg, catalog, runner, presenter, selector, nil
}
```

Also add to `terminal/formatter.go` a `NewStdWriter()` function:

```go
type stdWriter struct{}

func NewStdWriter() *stdWriter { return &stdWriter{} }
func (s *stdWriter) Write(p []byte) (n int, err error) { return os.Stdout.Write(p) }
func (s *stdWriter) String() string { return "stdout" }
```

And update `NewPresenter` to accept `io.Writer` (it already does via the interface).

- [ ] **Step 5: Verify it compiles**

```bash
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "feat: CLI commands with wiring"
```

---

### Task 15: Main.go + DI Wiring Cleanup

**Files:**
- Modify: `main.go`
- Modify: `internal/adapters/terminal/formatter.go` (add StdoutWriter)

- [ ] **Step 1: Update main.go**

```go
package main

import "github.com/sanmoo/bruwrapper/cmd"

func main() {
	cmd.Execute()
}
```

- [ ] **Step 2: Add StdoutWriter to terminal package**

Add to `internal/adapters/terminal/formatter.go` an `os` import and:

```go
type StdoutWriter struct{}

func NewStdoutWriter() *StdoutWriter { return &StdoutWriter{} }
func (s *StdoutWriter) Write(p []byte) (n int, err error) { return os.Stdout.Write(p) }
```

- [ ] **Step 3: Build and verify**

```bash
go build -o bruwrapper . && ./bruwrapper --help
```

Expected: help output showing `run`, `list`, `show` commands.

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "feat: main entry point with DI wiring"
```

---

### Task 16: Integration Test & Bug Fixes

**Files:**
- Create: `test/integration_test.go` (integration test using a real Bruno collection)

- [ ] **Step 1: Create a sample Bruno collection for testing**

```bash
mkdir -p test/fixtures/sample-collection
```

Create `test/fixtures/sample-collection/bruno.json`:
```json
{
  "version": "1",
  "name": "Test Collection",
  "type": "collection"
}
```

Create `test/fixtures/sample-collection/health.bru`:
```
meta {
  name: Health Check
  type: http
  seq: 1
}

get {
  url: https://httpbin.org/get
}

headers {
  Accept: application/json
}
```

- [ ] **Step 2: Add config fixture**

Create `test/fixtures/.bruwrapper.yaml`:
```yaml
collections:
  - ./sample-collection
```

- [ ] **Step 3: Verify bruwrapper list works**

```bash
BRUWRAPPER_CONFIG=test/fixtures/.bruwrapper.yaml ./bruwrapper list
```

Expected: lists "Test Collection".

- [ ] **Step 4: Verify bruwrapper list -c works**

```bash
BRUWRAPPER_CONFIG=test/fixtures/.bruwrapper.yaml ./bruwrapper list -c "Test Collection"
```

Expected: lists "GET Health Check".

- [ ] **Step 5: Run full test suite**

```bash
go test ./... -v
```

- [ ] **Step 6: Fix any compilation or test failures**

Address issues found during integration testing.

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "feat: integration test fixtures and bug fixes"
```

---

### Task 17: Final Polish & go vet

- [ ] **Step 1: Run go vet**

```bash
go vet ./...
```

- [ ] **Step 2: Fix any go vet issues**

- [ ] **Step 3: Run full test suite one more time**

```bash
go test ./... -v
```

- [ ] **Step 4: Build final binary**

```bash
go build -o bruwrapper .
```

- [ ] **Step 5: Verify all commands work**

```bash
./bruwrapper --help
./bruwrapper list --help
./bruwrapper run --help
./bruwrapper show --help
```

- [ ] **Step 6: Final commit**

```bash
git add -A && git commit -m "chore: final polish and cleanup"
```