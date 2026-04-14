# Design Doc: bruwrapper

## Visão Geral

`bruwrapper` é uma CLI em Go que envolve o `bru` CLI do Bruno, resolvendo suas limitações para consumo ad-hoc de APIs: obrigação de estar no diretório da collection, falta de seleção fácil de requests, variáveis verbosas, e output não formatado.

## Abordagem: Wrapper do `bru` CLI

Wrapper em Go que:

- Descobre collections via config file (`~/.bruwrapper.yaml`)
- Parseia `.bru` / `.yml` superficialmente para catálogo e seleção
- Oferece modo interativo (fuzzy finder) e modo por flags
- Shella out para `bru run` com CWD no diretório da collection
- Usa `--reporter-json` em temp file para capturar detalhes da resposta
- Formata output como pretty-print JSON no stdout

## CLI Interface

```
bruwrapper run [-c <collection>] [-r <request>] [-e <env>] [-v key=value...] [--raw] [--verbose]
bruwrapper list [-c <collection>]
bruwrapper show -c <collection> -r <request>
```

### Flags do `run`

| Flag               | Tipo     | Descrição                                                        |
| ------------------ | -------- | ---------------------------------------------------------------- |
| `-c, --collection` | string   | Nome da collection (obrigatório se não interativo)              |
| `-r, --request`    | string   | Nome do request (obrigatório se não interativo)                  |
| `-e, --env`        | string   | Ambiente Bruno (mapeado para `--env` do bru)                     |
| `-v, --var`        | string[] | Variável override (repeatável). Mapeado para `--env-var` do bru  |
| `--raw`            | bool     | Output crua sem pretty-print                                     |
| `--verbose`        | bool     | Mostra headers de request e response                             |

### Modo interativo

`bruwrapper run` sem `-c` e/ou `-r` abre fuzzy finder.

### Comando `list`

- Sem `-c`: lista collections configuradas
- Com `-c`: lista requests da collection exibindo method + name

### Comando `show`

Mostra método, URL, headers, e corpo do request ANTES de executar.

## Configuração

Arquivo `~/.bruwrapper.yaml`:

```yaml
collections:
  - ~/projects/myapi
  - ~/work/other-api
  - /absolute/path/to/collection
```

Cada path deve conter `bruno.json` (formato clássico) ou `opencollection.yml` (formato YAML v3+). O nome da collection vem do campo `name` nesses arquivos.

## Parser `.bru` / `.yml` (Superficial)

Não reimplementa execução. O parser lê apenas:

### Formato `.bru` clássico

- `meta { name, type, seq, tags }` — identificação do request
- Bloco de método HTTP (`get { url: ... }`, `post { url: ... }`, etc.)
- `headers { key: value }` — para exibição no `show`
- `body { ... }` — para exibição no `show`
- `vars:* { key: value }` — para documentação, não para resolução
- `{{varName}}` — detecção de variáveis para exibição

### Formato `.yml` (OpenCollection v3+)

- Campos correspondentes em YAML
- `meta.name`, `http.method`, `http.url`, etc.

O parser detecta o formato da collection automaticamente (presença de `bruno.json` vs `opencollection.yml`).

## Arquitetura (Hexagonal / Clean Architecture)

```
bruwrapper/
├── cmd/                          # Entry point (adapter: CLI)
│   ├── root.go
│   ├── run.go
│   ├── list.go
│   └── show.go
├── internal/
│   ├── core/                     # Domínio — zero dependências externas
│   │   ├── model.go              # Entidades: Collection, Request, Variable, Response
│   │   └── ports.go              # Interfaces: Catalog, Runner, Presenter, Selector, ConfigLoader
│   ├── app/                      # Casos de uso (orquestram domínio via ports)
│   │   ├── run.go                # Run use case
│   │   ├── list.go               # List use case
│   │   └── show.go                # Show use case
│   └── adapters/                 # Implementações concretas das ports
│       ├── brucatalog/           # Port: Catalog → filesystem scanner + .bru/.yml parser
│       │   ├── bru_parser.go
│       │   └── yml_parser.go
│       ├── brurunner/            # Port: Runner → bru CLI subprocess + JSON report parser
│       │   └── runner.go
│       ├── terminal/             # Port: Presenter → stdout pretty-print
│       │   └── formatter.go
│       ├── interactive/          # Port: Selector → bubbletea fuzzy finder
│       │   └── selector.go
│       └── yamlconfig/          # Port: ConfigLoader → ~/.bruwrapper.yaml
│           └── config.go
├── main.go                       # Wiring: injeção de dependências
└── go.mod
```

### Princípios

- **`core/`** contém só interfaces (ports) e entidades de domínio. Zero imports de libs externas. Não sabe nada sobre `bru`, `bubbletea`, ou YAML.
- **`app/`** orquestra os use cases usando as interfaces de `core/`. Chama `Catalog.Find()`, `Runner.Execute()`, `Presenter.Show()`.
- **`adapters/`** implementam as portas com tecnologia concreta. Trocar `brurunner` por um HTTP client nativo é plugar um adapter novo sem tocar em `core/` ou `app/`.
- **`main.go`** faz o wiring (dependency injection) — decide qual adapter usar para cada port.

### Port Interfaces (core/ports.go)

```go
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
}

type Selector interface {
    SelectCollection(collections []Collection) (Collection, error)
    SelectRequest(requests []Request) (Request, error)
}

type ConfigLoader interface {
    Load() (Config, error)
}
```

### Dependências Go

- `github.com/spf13/cobra` — CLI framework
- `github.com/charmbracelet/bubbletea` + `bubbles` — UI interativa
- `github.com/charmbracelet/lipgloss` — Styling
- `gopkg.in/yaml.v3` — YAML parsing (config + OpenCollection)

## Fluxo de Execução (comando `run`)

```
1. Carregar config → lista de diretórios de collections
2. Se -c e -r fornecidos:
   a. Resolver collection name → path no filesystem
   b. Resolver request name → caminho do .bru/.yml file
3. Se -c ou -r ausente:
   a. Construir catálogo das collections
   b. Abrir fuzzy finder interativo
   c. Usuário seleciona collection → request
4. Construir comando bru:
   - CWD = diretório da collection
   - bru run <request-file> --env <env> --env-var key=value ... --reporter-json <tmpfile>
5. Executar bru como subprocess
6. Parsear JSON report do tempfile
7. Extrair: status code, status text, response headers, response body, tempo
8. Formatar output:
   - Pretty-print JSON body (usando encoding/json com indent)
   - ou raw output se --raw
   - incluir headers se --verbose
9. Limpar tempfile
10. Retornar exit code do bru (0 = sucesso, 1 = falha)
```

## Tratamento de Erros

| Condição                  | Comportamento                                                             |
| ------------------------- | ------------------------------------------------------------------------- |
| `bru` não encontrado      | Erro claro: "bru CLI not found. Install: npm i -g @usebruno/cli" com link |
| Collection não encontrada | Lista collections disponíveis no config                                   |
| Request não encontrado    | Lista requests da collection selecionada                                  |
| `bru run` falha           | Mostra stderr + exit code                                                 |
| JSON report inválido      | Fallback para stdout/stdout do bru                                        |
| Arquivo de config ausente | Erro com instrução para criar ~/.bruwrapper.yaml                          |

## Formato de Output

### Padrão (pretty-print JSON)

```
Status: 200 OK
Time:   142ms

{
  "id": 1,
  "name": "Leanne Graham",
  "username": "Bret"
}
```

### Verbose (headers + body)

```
Status: 200 OK
Time:   142ms

Request Headers:
  Authorization: Bearer ***
  Content-Type: application/json

Response Headers:
  Content-Type: application/json; charset=utf-8
  X-RateLimit-Remaining: 59

{
  "id": 1,
  "name": "Leanne Graham"
}
```

### Raw (sem formatação)

```
{"id":1,"name":"Leanne Graham"}
```

## Suporte a Formatos Bruno

- **Clássico (.bru)**: `bruno.json` como marcador de collection, arquivos `.bru` para requests
- **OpenCollection (.yml)**: `opencollection.yml` como marcador, arquivos `.yml` para requests

O parser detecta o formato pelo arquivo marcador presente no diretório da collection.

## Fora do Escopo (MVP)

- Execução de scripts pre/post-request (delegado ao `bru`)
- Importação/criação de collections
- Salvamento de histórico de execuções
- Autenticação OAuth interativa (post-MVP)
- Suporte a GraphQL (delegado ao `bru`)
- Watch mode / auto-rerun