# bruwrapper

A CLI wrapper for [Bruno](https://www.usebruno.com/) that enables ad-hoc API consumption with better UX.

## Why bruwrapper?

The Bruno CLI (`bru`) is great for contract testing, but painful for ad-hoc API consumption:

- You **must** be inside the collection directory to run requests
- There's **no easy way** to select a request by name — you need the file path
- Passing variables is verbose (`--env-var key=value` for each variable)
- Output is **summary-only** to the console — response details go to report files

bruwrapper solves all of this.

## Features

- **Directory-agnostic** — run requests from anywhere, configure collection paths in `~/.bruwrapper.yaml`
- **Interactive selection** — fuzzy finder to browse collections and requests
- **Simple variable override** — `-v token=abc123` instead of `--env-var token=abc123`
- **Pretty-printed JSON output** — no need to pipe through `jq`
- **Dual format support** — works with both classic `.bru` and OpenCollection `.yml` formats
- **Verbose mode** — show request and response headers with `--verbose`

## Installation

```bash
go install github.com/sanmoo/bruwrapper@latest
```

Requires [bru CLI](https://github.com/usebruno/bruno) installed:

```bash
npm i -g @usebruno/cli
```

## Quick Start

1. Create `~/.bruwrapper.yaml` with your collection paths:

```yaml
collections:
  - ~/projects/my-api
  - ~/work/other-api
```

2. List your collections:

```bash
bruwrapper list
```

3. Run a request:

```bash
bruwrapper run -c "My API" -r "Get Users"
```

## Usage

### List collections

```bash
bruwrapper list
```

```
Collections:
  My API (/home/user/projects/my-api)
  Other API (/home/user/projects/other-api)
```

### List requests in a collection

```bash
bruwrapper list -c "My API"
```

```
Requests:
  GET     Get Users
  POST    Create User
  DELETE  Delete User
```

### Show request details (without executing)

```bash
bruwrapper show -c "My API" -r "Get Users"
```

```
Method: GET
URL:    https://api.example.com/users

Headers:
  Authorization: ***
  Content-Type: application/json
```

### Run a request

```bash
bruwrapper run -c "My API" -r "Get Users"
```

```
Status: 200 OK
Time:   142ms

{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret"
}
```

### Run with variable overrides

```bash
bruwrapper run -c "My API" -r "Get Users" -v token=abc123
```

### Run with environment

```bash
bruwrapper run -c "My API" -r "Get Users" -e staging
```

### Verbose output (headers + body)

```bash
bruwrapper run -c "My API" -r "Get Users" --verbose
```

```
Status: 200 OK
Time:   142ms

Response Headers:
  Content-Type: application/json; charset=utf-8
  X-RateLimit-Remaining: 59

{
  "id": 1,
  "name": "Leanne Graham"
}
```

### Raw output (no pretty-print)

```bash
bruwrapper run -c "My API" -r "Get Users" --raw
```

```
{"id":1,"name":"Leanne Graham"}
```

### Interactive mode

Run without `-c` and/or `-r` to open the fuzzy finder:

```bash
bruwrapper run
```

## Configuration

### Config file: `~/.bruwrapper.yaml`

```yaml
collections:
  - ~/projects/my-api
  - ~/work/other-api
  - /absolute/path/to/collection
```

Paths starting with `~` are expanded to your home directory.

### Config file location

By default, bruwrapper looks for `~/.bruwrapper.yaml`. You can override this with:

- `--config` flag: `bruwrapper --config /path/to/config.yaml list`
- `BRUWRAPPER_CONFIG` env var: `BRUWRAPPER_CONFIG=/path/to/config.yaml bruwrapper list`

## Architecture

bruwrapper follows hexagonal (ports & adapters) architecture:

```
core/    → Domain models and port interfaces (zero external dependencies)
app/     → Use cases orchestrating domain via ports
adapters/→ Concrete implementations:
            brucatalog/   → .bru/.yml parsers + filesystem scanner
            brurunner/    → bru CLI subprocess + JSON report parser
            terminal/     → Pretty-print output
            interactive/  → Bubbletea fuzzy finder
            yamlconfig/   → Config file loader
cmd/     → Cobra CLI commands + dependency injection wiring
```

## Built with Superpowers

This project was built using the [Superpowers](https://github.com/obra/superpowers) methodology with [OpenCode](https://opencode.ai) — an agentic workflow that guided the design, planning, and implementation of bruwrapper from spec to working code.

## Development

```bash
go build -o bruwrapper .
go test ./...
go vet ./...
```

## License

MIT