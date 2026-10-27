# Participant Services — Go, GraphQL and Cloud-Native Middleware Lab

A hands-on, **synthetic** participant-services platform built to practise the engineering skills relevant to the NDIA EL2 Lead Developer interview (RFQ LH-07533). This is an independent learning project, **not an NDIA system** and not a claim of NDIA production experience.

> **Status (27 October 2026):** The Go health endpoint, starter Todo GraphQL example, shared Participant Service, and **separate Portal and Mobile GraphQL APIs** have been implemented and verified. Both HTTP GraphQL endpoints returned the expected synthetic participant data. Participant Service unit tests (including 100 concurrent reads), the Portal HTTP integration test, and the full `go test -race ./...` suite passed before the Mobile HTTP route was added. The Mobile resolver compiled and both endpoints responded successfully; the **Mobile HTTP integration test and final race suite after wiring both routes remain unconfirmed**. GitHub push of the Mobile changes is also unconfirmed.

## Project roadmap

- [x] Initialise private GitHub repository and Go module
- [x] Implement `/healthz` and starter gqlgen Todo operations
- [x] Test Todo resolver with `go test` and `-race`
- [x] Implement `internal/participant` with `context.Context`, `sync.RWMutex`, and `ErrNotFound`
- [x] Test participant lookup, missing ID, cancelled context and 100 concurrent reads
- [x] Generate separate Portal and Mobile gqlgen packages and schemas
- [x] Implement Portal and Mobile resolvers using the shared Participant Service
- [x] Serve `/graphql/portal` and `/graphql/mobile` from one Go HTTP server
- [x] Verify both endpoints with `curl` using participant P001
- [x] Pass Portal GraphQL HTTP integration test
- [ ] Confirm Mobile HTTP integration test and final full race-enabled suite
- [ ] Commit and push Mobile milestone; verify clean Git status
- [ ] Implement authenticated identity and participant-level authorisation
- [ ] Add mock MuleSoft/Salesforce, timeouts, retry/idempotency and N+1 exercises
- [ ] Add PostgreSQL, observability, Docker, Kubernetes, AWS EKS and CI/CD

## Current architecture (verified locally)

```text
Portal client -> POST /graphql/portal -> Portal gqlgen resolver --+
                                                              +--> shared participant.Service -> in-memory synthetic data
Mobile client -> POST /graphql/mobile -> Mobile gqlgen resolver --+

GET /healthz; GET /playground/portal; GET /playground/mobile
```

**Planned, not implemented:** authentication/authorisation, PostgreSQL, MuleSoft/Salesforce adapters, production hardening, Docker/Kubernetes and EKS. Both consumers currently accept arbitrary participant IDs: **do not deploy this API publicly**.

## Design documentation

See **[DESIGN.md](DESIGN.md)** for the as-built architecture, request contracts, data flow, verified test coverage, design decisions, security gaps and future architecture.

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
<summary><strong>8. Remaining implementation stages (planned)</strong></summary>

1. Mobile HTTP integration test and final race suite; commit/push changes.
2. Identity-based access control for participant records (Portal and Mobile), GraphQL query limits and secure errors.
3. Mock MuleSoft/Salesforce HTTP integration with context propagation, timeouts, safe retries and idempotency.
4. N+1 demonstration, DataLoader batching, metrics and traces.
5. PostgreSQL, Docker, local Kubernetes, AWS EKS and GitHub Actions.

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
<summary><strong>10. Introduce shared participant service (verified)</strong></summary>

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

**Verified:** `gofmt`, `go test ./...` and `go test -race ./...` passed. Tests `TestGetParticipant`, `TestGetParticipantNotFound`, `TestGetParticipantCancelled`, and `TestConcurrentParticipantReads` (100 goroutines) all passed. `RWMutex` guards map access; it does not implement participant-level authorisation. The current implementation reads a fixed in-memory map; concurrent writes are not exercised.

</details>

<details>
<summary><strong>11. Generate the separate Portal GraphQL API (verified)</strong></summary>

Create `graph/portal/schema.graphqls`:

```graphql
type Participant {
  id: ID!
  firstName: String!
  lastName: String!
}

type Query {
  participant(id: ID!): Participant
}
```

Create `gqlgen.portal.yml`:

```yaml
schema:
  - graph/portal/schema.graphqls
exec:
  filename: graph/portal/generated/generated.go
  package: generated
model:
  filename: graph/portal/model/models_gen.go
  package: model
resolver:
  layout: follow-schema
  dir: graph/portal
  package: portal
  filename_template: "{name}.resolvers.go"
```

```bash
go tool gqlgen generate --config gqlgen.portal.yml
go test ./...
```

The Portal resolver injects `*participant.Service` and calls `Participants.Get(ctx, id)`. It converts `ErrNotFound` to a nullable GraphQL result and returns a `model.Participant`. The handwritten resolver implements `generated.QueryResolver` from `graph/portal/generated`, not a local `QueryResolver` interface.

**Verified:** generation and compilation; `TestPortalParticipantQuery` using `httptest`; successful HTTP query for `P001`.

</details>

<details>
<summary><strong>12. Generate the separate Mobile GraphQL API (verified)</strong></summary>

Create `graph/mobile/schema.graphqls`:

```graphql
type ParticipantProfile {
  participantId: ID!
  displayName: String!
}

type Query {
  profile(participantId: ID!): ParticipantProfile
}
```

Create `gqlgen.mobile.yml`:

```yaml
schema:
  - graph/mobile/schema.graphqls
exec:
  filename: graph/mobile/generated/generated.go
  package: generated
model:
  filename: graph/mobile/model/models_gen.go
  package: model
resolver:
  layout: follow-schema
  dir: graph/mobile
  package: mobile
  filename_template: "{name}.resolvers.go"
```

```bash
go tool gqlgen generate --config gqlgen.mobile.yml
go test ./...
go test -race ./...
```

The Mobile resolver calls the **same service type** and maps `FirstName + " " + LastName` to `displayName`. It uses its own generated `model.ParticipantProfile` and `generated.QueryResolver` interface. Compilation and race-enabled tests passed before the final HTTP wiring.

</details>

<details>
<summary><strong>13. Serve Portal and Mobile from one HTTP server (verified)</strong></summary>

In `cmd/api/main.go`, use distinct aliases for the two packages named `generated`:

```go
import (
    "github.com/shawfire/participant-services/graph/mobile"
    mobilegenerated "github.com/shawfire/participant-services/graph/mobile/generated"
    "github.com/shawfire/participant-services/graph/portal"
    portalgenerated "github.com/shawfire/participant-services/graph/portal/generated"
    "github.com/shawfire/participant-services/internal/participant"
)
```

The handler wiring (within `main`, with `mux`, `handler`, `transport`, and `playground` imports) is:

```go
participantService := participant.NewService()
portalAPI := handler.New(portalgenerated.NewExecutableSchema(portalgenerated.Config{
    Resolvers: &portal.Resolver{Participants: participantService},
}))
portalAPI.AddTransport(transport.Options{})
portalAPI.AddTransport(transport.POST{})
mux.Handle("POST /graphql/portal", portalAPI)
mux.Handle("GET /playground/portal", playground.Handler("Portal GraphQL", "/graphql/portal"))

mobileAPI := handler.New(mobilegenerated.NewExecutableSchema(mobilegenerated.Config{
    Resolvers: &mobile.Resolver{Participants: participantService},
}))
mobileAPI.AddTransport(transport.Options{})
mobileAPI.AddTransport(transport.POST{})
mux.Handle("POST /graphql/mobile", mobileAPI)
mux.Handle("GET /playground/mobile", playground.Handler("Mobile GraphQL", "/graphql/mobile"))
```

Both resolvers receive the **same `participantService` instance**. A previous build failed because both generated packages were imported under the same name; distinct aliases fixed that. Another build failed due to a malformed import block; replacing `cmd/api/main.go` with a complete file resolved the issue.

```bash
gofmt -w cmd/api/main.go
go test ./...
go run ./cmd/api
```

**Verified:** both routes responded correctly via `curl` after the file replacement. The final `go test` output after that replacement was not supplied.

</details>

<details>
<summary><strong>14. Query both endpoints and validate results (verified)</strong></summary>

```bash
curl -sS http://localhost:8080/graphql/portal \
  -H "Content-Type: application/json" \
  -d '{"query":"{ participant(id: \"P001\") { id firstName lastName } }"}'
```

Expected and observed (compact formatting):

```json
{"data":{"participant":{"id":"P001","firstName":"Alex","lastName":"Taylor"}}}
```

```bash
curl -sS http://localhost:8080/graphql/mobile \
  -H "Content-Type: application/json" \
  -d '{"query":"{ profile(participantId: \"P001\") { participantId displayName } }"}'
```

Expected and observed:

```json
{"data":{"profile":{"participantId":"P001","displayName":"Alex Taylor"}}}
```

GraphQL Playground URLs: `http://localhost:8080/playground/portal` and `http://localhost:8080/playground/mobile`. Query text belongs in the main query editor; the Variables and Headers panels expect JSON.

</details>

<details>
<summary><strong>15. Mobile integration test and Git milestone (pending verification)</strong></summary>

The proposed `graph/mobile/integration_test.go` uses `httptest.NewRequest`, gqlgen `handler.New`, and `json.Unmarshal` to check that `profile(participantId: "P001")` returns `participantId: "P001"` and `displayName: "Alex Taylor"` without GraphQL errors. The user has not yet supplied execution output for this test.

```bash
gofmt -w graph/mobile/integration_test.go
go test -v ./graph/mobile
go test -race ./...
# Only after passing:
git add cmd/api/main.go gqlgen.mobile.yml graph/mobile/
git commit -m "Add Mobile GraphQL API with shared participant service"
git push
git status
```

The most recent Git status showed `cmd/api/main.go` modified, and `gqlgen.mobile.yml` plus `graph/mobile/` untracked. Do not mark the Mobile commit/push complete until confirmed.

</details>

## Current source layout

```text
participant-services/
├── cmd/api/main.go                 # Health, Portal and Mobile HTTP routes
├── graph/
│   ├── schema.graphqls             # Legacy Todo scaffold
│   ├── schema.resolvers.go         # Legacy Todo resolvers
│   ├── schema.resolvers_test.go    # Verified Todo test
│   ├── portal/
│   │   ├── schema.graphqls         # Portal-specific contract
│   │   ├── resolver.go
│   │   ├── schema.resolvers.go
│   │   ├── integration_test.go     # Verified Portal HTTP test
│   │   ├── generated/
│   │   └── model/
│   └── mobile/
│       ├── schema.graphqls         # Mobile-specific contract
│       ├── resolver.go
│       ├── schema.resolvers.go
│       ├── generated/
│       ├── model/
│       └── integration_test.go     # Proposed; execution unconfirmed
├── internal/participant/
│   ├── service.go
│   └── service_test.go             # Four passing service tests
├── gqlgen.yml                     # Legacy Todo config
├── gqlgen.portal.yml
├── gqlgen.mobile.yml
├── server.go                      # Legacy Todo demo server
├── DESIGN.md
├── README.md
├── go.mod
└── go.sum
```

## Security and development guidance

Use **synthetic participant data only**. Never commit credentials, real participant information, tokens, private keys or `.env` files containing secrets. Generated GraphQL schemas and AI-assisted code require human review, particularly around authentication, resource-level authorisation, input validation, timeouts, logging and error disclosure.

## Interview focus

The NDIA supplier briefing emphasises hands-on **Go, GraphQL, AWS/EKS, middleware architecture, production debugging and technical leadership**. This lab is designed to make those concepts tangible, while real interview examples should come from genuine professional experience rather than attributing simulated lab results to production projects.
