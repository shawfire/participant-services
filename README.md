# Participant Services — Go, GraphQL and Cloud-Native Middleware Lab

A hands-on, **synthetic** participant-services platform built to practise the engineering skills relevant to the NDIA EL2 Lead Developer interview (RFQ LH-07533). This is an independent learning project, **not an NDIA system** and not a claim of NDIA production experience.

> **Status (27 October 2026):** The Go HTTP health endpoint and the gqlgen starter Todo mutation/query have been run successfully. The repository was created as private and the initial HTTP server commit was pushed. The GraphQL mutation/query and the new `TestCreateTodo` unit test passed, including `go test -race ./...`. The GraphQL and test commits/pushes have not yet been confirmed; the participant service was supplied as a next step but has not been verified.

## Project roadmap

- [x] Verify local development tools and GitHub authentication
- [x] Initialise Go module and private GitHub repository
- [x] Implement and verify `/healthz`
- [x] Install gqlgen and generate a starter GraphQL API
- [x] Replace generated resolver panics with working in-memory implementations
- [x] Verify a Todo mutation and query
- [x] Run `TestCreateTodo` and `go test -race ./...` successfully
- [ ] Verify participant service implementation and tests
- [ ] Commit and push the GraphQL starter implementation
- [ ] Replace Todo with participant and plan domain services
- [ ] Separate **Portal** and **Mobile** GraphQL schemas and endpoints
- [ ] Add request-scoped authorisation, context cancellation, timeouts and concurrency-safe storage
- [ ] Add mock MuleSoft/Salesforce integration, retries and idempotency
- [ ] Add PostgreSQL, automated tests, logging and tracing
- [ ] Containerise, deploy locally on Kubernetes, then evaluate AWS EKS deployment
- [ ] Add GitHub Actions CI/CD and production troubleshooting exercises

## Architecture target (planned)

```text
Participant Portal  ──> /graphql/portal ──┐
                                         ├──> Shared Go domain/services
Participant Mobile  ──> /graphql/mobile ──┘           │
                                                     ├──> PostgreSQL
                                                     └──> Mock MuleSoft -> Mock Salesforce

Deployment target: AWS EKS (planned)
```

The Portal and Mobile GraphQL APIs will have **separate consumer schemas** but reuse common business logic, authorisation and integration services. The current gqlgen Todo scaffold is temporary.

## Design documentation

See **[DESIGN.md](DESIGN.md)** for the proposed architecture, API boundaries, data flows, security model, reliability patterns, operational considerations and architectural decisions. Planned design is distinguished from implemented functionality.

## Reproducible development journal

<details>
<summary><strong>1. Check development tools on an Apple Silicon Mac</strong></summary>

Run:

```bash
go version
git --version
docker --version
gh --version
brew --version
which go
which git
which docker
which gh
```

Observed during setup:

```text
Go:       go1.27.3 darwin/arm64 (after update)
Git:      Apple Git 2.39.5 initially; Homebrew Git 2.56.0_1 installed
Docker:   Docker CLI 25.0.3, /usr/local/bin/docker (upgrade not yet verified)
GitHub:   gh 2.101.0
Homebrew: /opt/homebrew
```

**Important:** Installing Homebrew Git does not automatically guarantee that `git` resolves to it. Check:

```bash
/opt/homebrew/bin/git --version
which git
type -a git
echo "$PATH"
```

The existing `~/.zshrc` already prepends `/opt/homebrew/bin`. If needed, reload it and clear zsh's command cache:

```bash
source ~/.zshrc
rehash
which git
git --version
```

Do not broadly trust unrelated Homebrew taps merely to silence warnings. For Docker Desktop, inspect `docker version` (client and server) and update Docker Desktop itself if necessary; do not assume `brew upgrade docker` updates Docker Desktop.

</details>

<details>
<summary><strong>2. Authenticate GitHub and initialise the repository</strong></summary>

The project was created under:

```text
/Users/shawfire/Documents/jobs/NDIA_LH-07533_Lead_Principal_Developer/participant-services
```

Commands:

```bash
cd /Users/shawfire/Documents/jobs/NDIA_LH-07533_Lead_Principal_Developer
gh auth status
mkdir -p participant-services
cd participant-services
git init -b main
go mod init github.com/shawfire/participant-services
pwd
go version
git status
cat go.mod
```

Initial `go.mod`:

```go
module github.com/shawfire/participant-services

go 1.27.3
```

Create and attach a **private** GitHub repository:

```bash
gh repo create shawfire/participant-services \
  --private \
  --source=. \
  --remote=origin
```

Repository: <https://github.com/shawfire/participant-services>

If the browser displays **404**, verify you are signed in as `shawfire`: GitHub intentionally returns 404 to unauthorised viewers of private repositories.

```bash
gh repo view shawfire/participant-services \
  --json name,visibility,url \
  --jq '{name,visibility,url}'
gh repo view shawfire/participant-services --web
```

Observed result: `visibility` was `PRIVATE`.

</details>

<details>
<summary><strong>3. Add the first Go HTTP server and health endpoint</strong></summary>

Create the entry point:

```bash
mkdir -p cmd/api
cat > cmd/api/main.go <<'GO'
package main

import (
    "log"
    "net/http"
    "time"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok\n"))
    })

    server := &http.Server{
        Addr:              ":8080",
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
    }

    log.Println("Participant Services API listening on :8080")

    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}
GO
```

Run in one terminal:

```bash
go run ./cmd/api
```

In another terminal:

```bash
curl -i http://localhost:8080/healthz
```

**Verified response:**

```http
HTTP/1.1 200 OK
Content-Length: 3
Content-Type: text/plain; charset=utf-8

ok
```

Stop the server with `Ctrl+C`.

**First Git commit (confirmed pushed):**

```bash
git add go.mod cmd/api/main.go
git commit -m "Initial Go API server with health endpoint"
git push -u origin main
```

Commit: `62d916d`.

**Engineering note:** `/healthz` currently confirms that the HTTP process responds. Later, add a distinct readiness endpoint for dependencies, plus graceful shutdown.

</details>

<details>
<summary><strong>4. Install gqlgen and generate the starter GraphQL API</strong></summary>

```bash
go get -tool github.com/99designs/gqlgen
go tool gqlgen version
go tool gqlgen init
```

**Observed version:** `v0.17.98`.

Generated files include:

```text
gqlgen.yml
graph/schema.graphqls
graph/schema.resolvers.go
graph/resolver.go
graph/model/...
graph/generated/...
server.go
```

Some paths and generated filenames can vary with gqlgen configuration. `go.mod` and `go.sum` are updated by dependency installation.

Start the generated GraphQL server:

```bash
go run ./server.go
```

Open <http://localhost:8080/> to use GraphQL Playground. Do not run `go run ./cmd/api` at the same time: both servers initially use port 8080.

**Initial mutation:**

```graphql
mutation {
  createTodo(input: {
    text: "Prepare NDIA middleware platform"
    userId: "shawfire"
  }) {
    id
    text
    done
    user {
      id
      name
    }
  }
}
```

**Observed initial error:**

```json
{
  "errors": [
    {
      "message": "internal system error",
      "path": ["createTodo"]
    }
  ],
  "data": null
}
```

**Root cause:** Generated resolvers in `graph/schema.resolvers.go` still contained placeholder panics:

```go
panic(fmt.Errorf("not implemented: CreateTodo - createTodo"))
panic(fmt.Errorf("not implemented: Todos - todos"))
```

The schema and server were generated, but the application behaviour still needed implementation.

</details>

<details>
<summary><strong>5. Implement Todo mutation and query resolvers</strong></summary>

Replace `graph/schema.resolvers.go` with:

```go
package graph

import (
    "context"
    "strconv"

    "github.com/shawfire/participant-services/graph/model"
)

// CreateTodo handles the GraphQL createTodo mutation.
func (r *mutationResolver) CreateTodo(
    ctx context.Context,
    input model.NewTodo,
) (*model.Todo, error) {
    todo := &model.Todo{
        ID:   strconv.Itoa(len(r.todos) + 1),
        Text: input.Text,
        Done: false,
        User: &model.User{
            ID:   input.UserID,
            Name: "Test Participant",
        },
    }

    r.todos = append(r.todos, todo)
    return todo, nil
}

// Todos handles the GraphQL todos query.
func (r *queryResolver) Todos(
    ctx context.Context,
) ([]*model.Todo, error) {
    return r.todos, nil
}

func (r *Resolver) Mutation() MutationResolver {
    return &mutationResolver{r}
}

func (r *Resolver) Query() QueryResolver {
    return &queryResolver{r}
}

type (
    mutationResolver struct{ *Resolver }
    queryResolver    struct{ *Resolver }
)
```

Replace `graph/resolver.go` with:

```go
package graph

import "github.com/shawfire/participant-services/graph/model"

// Resolver holds dependencies shared by GraphQL resolvers.
type Resolver struct {
    todos []*model.Todo
}
```

Format and start:

```bash
gofmt -w graph/schema.resolvers.go graph/resolver.go
go run ./server.go
```

**Caveat:** This in-memory slice is **not safe for concurrent writes**; it also resets on restart and is not suitable for multiple replicas. It is intentionally temporary. Later, use a synchronised service or database, explicit authorisation and context-aware dependency calls. The unused `ctx` parameter here is acceptable but will matter when services and downstream calls are introduced.

</details>

<details>
<summary><strong>6. Verify GraphQL operations and investigate errors</strong></summary>

Run the mutation:

```graphql
mutation {
  createTodo(input: {
    text: "Prepare NDIA middleware platform"
    userId: "shawfire"
  }) {
    id
    text
    done
    user {
      id
      name
    }
  }
}
```

**Verified result:**

```json
{
  "data": {
    "createTodo": {
      "id": "1",
      "text": "Prepare NDIA middleware platform",
      "done": false,
      "user": {
        "id": "shawfire",
        "name": "Test Participant"
      }
    }
  }
}
```

Then query:

```graphql
query {
  todos {
    id
    text
    done
  }
}
```

**Verified result:**

```json
{
  "data": {
    "todos": [
      {
        "id": "1",
        "text": "Prepare NDIA middleware platform",
        "done": false
      }
    ]
  }
}
```

**Debugging lesson:** An HTTP 200 from a GraphQL endpoint does not imply every resolver succeeded. Inspect the GraphQL `errors` field, resolver code and server logs; isolate whether the failure is in schema validation, resolver logic or a downstream dependency.

</details>

<details>
<summary><strong>7. Test and commit the GraphQL starter implementation (pending confirmation)</strong></summary>

Stop the GraphQL server (`Ctrl+C`) and run:

```bash
go test ./...
git status
```

If tests/build pass:

```bash
git add go.mod go.sum gqlgen.yml graph/ server.go
git commit -m "Add gqlgen GraphQL API with working resolvers"
git push
```

**Status:** Initial `go test ./...` compiled all four packages and reported `[no test files]`. A resolver unit test was subsequently added and both `go test -v ./...` and `go test -race ./...` passed. The Git commit/push for these changes has not been confirmed.

Add this README to the same or a subsequent commit:

```bash
git add README.md
git commit -m "Document setup and GraphQL development walkthrough"
git push
```

</details>

<details>
<summary><strong>8. Next implementation stages (planned, not yet executed)</strong></summary>

1. **Domain and shared service layer:** participant, plan and submission models; dependency-injected Go services; tests.
2. **Two GraphQL schemas:** separate gqlgen configurations and generated packages for `/graphql/portal` and `/graphql/mobile`, with shared domain logic and distinct consumer contracts.
3. **Identity and authorisation:** derive mobile `me` from verified identity; enforce resource-level access, not only authentication.
4. **Performance:** reproduce and fix N+1 resolver calls, add batching/pagination, complexity and depth limits.
5. **Resilience:** context cancellation, deadlines, bounded concurrency, retries/backoff, circuit-breaking considerations and idempotent submissions.
6. **Enterprise integration:** simulated MuleSoft/Salesforce HTTP APIs, including latency, 5xx responses and duplicate submissions.
7. **Operations:** PostgreSQL, structured logs, traces, metrics, Docker, Kubernetes probes, GitHub Actions and eventual AWS EKS deployment.
8. **Interview drills:** slow GraphQL requests, Go memory growth/OOMKilled, failed integration calls and review of insecure AI-generated code.

No commands in this section have been executed yet; document exact commands and verified results as each milestone is completed.

</details>

<details>
<summary><strong>9. Add a resolver unit test and run the race detector (verified)</strong></summary>

Create `graph/schema.resolvers_test.go`:

```go
package graph

import (
    "context"
    "testing"

    "github.com/shawfire/participant-services/graph/model"
)

func TestCreateTodo(t *testing.T) {
    r := &Resolver{}
    todo, err := r.Mutation().CreateTodo(context.Background(), model.NewTodo{
        Text: "Prepare NDIA interview", UserID: "shawfire",
    })
    if err != nil { t.Fatalf("CreateTodo failed: %v", err) }
    if todo.ID != "1" { t.Errorf("expected ID 1, got %s", todo.ID) }
    if todo.Text != "Prepare NDIA interview" { t.Errorf("unexpected text: %s", todo.Text) }
    if todo.Done { t.Error("new todo should not be completed") }

    todos, err := r.Query().Todos(context.Background())
    if err != nil { t.Fatalf("Todos failed: %v", err) }
    if len(todos) != 1 { t.Fatalf("expected 1 todo, got %d", len(todos)) }
}
```

```bash
gofmt -w graph/schema.resolvers_test.go
go test -v ./...
go test -race ./...
```

**Observed:** `TestCreateTodo` passed; race-detector run passed. This does **not** establish concurrency safety: the test only exercises sequential calls, while the current Todo slice has no synchronization.

```bash
git add graph/schema.resolvers_test.go
git commit -m "Add GraphQL resolver unit tests"
git push
```

Commit/push shown for reference; execution not confirmed.

</details>

<details>
<summary><strong>10. Introduce shared participant service (provided; not yet verified)</strong></summary>

```bash
mkdir -p internal/participant
cat > internal/participant/service.go <<'EOF'
package participant

import (
    "context"
    "errors"
    "sync"
)

var ErrNotFound = errors.New("participant not found")

type Participant struct {
    ID        string
    FirstName string
    LastName  string
}

type Service struct {
    mu           sync.RWMutex
    participants map[string]Participant
}

func NewService() *Service {
    return &Service{participants: map[string]Participant{
        "P001": {ID: "P001", FirstName: "Alex", LastName: "Taylor"},
        "P002": {ID: "P002", FirstName: "Jordan", LastName: "Lee"},
    }}
}

func (s *Service) Get(ctx context.Context, id string) (Participant, error) {
    if err := ctx.Err(); err != nil { return Participant{}, err }
    s.mu.RLock()
    defer s.mu.RUnlock()
    p, ok := s.participants[id]
    if !ok { return Participant{}, ErrNotFound }
    return p, nil
}
EOF

gofmt -w internal/participant/service.go
go test ./...
```

**Status:** This implementation was proposed but no execution or test result has been supplied. `RWMutex` guards access to the in-memory map; it does not itself enforce participant-level authorisation. Before production use, add tests, explicit access controls and a persistent store.

</details>

## Current source layout

```text
participant-services/
├── cmd/
│   └── api/
│       └── main.go           # Original /healthz server
├── graph/
│   ├── model/                # gqlgen model types
│   ├── generated/            # gqlgen executable schema (layout may vary)
│   ├── resolver.go           # In-memory Todo state
│   ├── schema.graphqls       # Temporary Todo schema
│   ├── schema.resolvers.go   # Todo mutation/query implementation
│   └── schema.resolvers_test.go # Verified unit test
├── internal/participant/      # Planned/proposed shared service (unverified)
├── DESIGN.md                  # Architecture and design decisions
├── gqlgen.yml
├── server.go                 # Generated GraphQL demo server
├── go.mod
├── go.sum
└── README.md
```

## Security and development guidance

Use **synthetic participant data only**. Never commit credentials, real participant information, tokens, private keys or `.env` files containing secrets. Generated GraphQL schemas and AI-assisted code require human review, particularly around authentication, resource-level authorisation, input validation, timeouts, logging and error disclosure.

## Interview focus

The NDIA supplier briefing emphasises hands-on **Go, GraphQL, AWS/EKS, middleware architecture, production debugging and technical leadership**. This lab is designed to make those concepts tangible, while real interview examples should come from genuine professional experience rather than attributing simulated lab results to production projects.
