# Development journal

[Documentation index](README.md) · [Current architecture](../DESIGN.md) · [Audit maintenance rules](history/README.md#maintenance-rules)

This is the primary continuous audit trail for the project, from initial setup through current and future development. Append each substantive step here, covering changes, reasons, problems, decisions, verification, remaining work and available commit references. Technical references describe the current system; the documentation change log provides supporting detail.

Entries 1–23 below are preserved verbatim from the pre-refactor README, in their original order. Their pending/planned statements describe knowledge at that point in development, not current status. Commands can overwrite files and are historical reproduction instructions; use the current [quick start](../README.md#quick-start) for normal operation. The original entry numbering is separate from the interview preparation workflow's Step 74; no missing intermediate workflow steps have been invented.

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


## Enterprise middleware integration — 27 October 2026

The following milestones were implemented and verified locally using synthetic participant data.

<details>
<summary><strong>16. Repository dependency injection (verified)</strong></summary>

Created `internal/participant/repository.go` defining the `Repository` interface with `GetParticipant(ctx, id)`.

Updated `internal/participant/service.go` to support:

- `NewService()` for in-memory data.
- `NewServiceWithRepository(repository)` for injected data access.

Added `internal/participant/repository_test.go` covering delegation, error propagation and context cancellation.

**Verification:** All seven Participant Service tests passed, including the race-detector suite.

</details>

<details>
<summary><strong>17. Mock MuleSoft HTTP API (verified)</strong></summary>

Created `internal/mockmulesoft/handler.go`, `handler_test.go` and `cmd/mockmulesoft/main.go`.

The mock API listens on port 8081 and exposes `GET /api/participants/{id}`.

It returns synthetic participants P001 (Alex Taylor) and P002 (Jordan Lee). Unknown IDs return HTTP 404.

**Verification:** Both HTTP handler tests passed.

</details>

<details>
<summary><strong>18. MuleSoft HTTP client (verified)</strong></summary>

Created `internal/integration/mulesoft/client.go` and `client_test.go`.

The client implements `participant.Repository` and supports configurable HTTP timeouts, context propagation, 404 mapping, upstream 5xx errors, JSON decoding and network error handling.

**Verification:** Six HTTP client tests passed, covering success, not found, upstream errors, timeout, cancellation and network failure.

</details>

<details>
<summary><strong>19. Connect GraphQL APIs to MuleSoft (verified)</strong></summary>

Updated `cmd/api/main.go` to construct a MuleSoft HTTP client and inject it into the shared Participant Service.

The upstream URL is configured using `MULESOFT_URL`, defaulting to `http://localhost:8081`.

Both Portal and Mobile GraphQL schemas use the same service.

**Verification:** Direct MuleSoft, Portal GraphQL and Mobile GraphQL HTTP requests returned the expected synthetic data.

</details>

<details>
<summary><strong>20. Automated end-to-end integration tests (verified)</strong></summary>

Created `internal/integration/mulesoft/endtoend_test.go`.

The test uses `httptest.Server` to exercise both GraphQL APIs through the Participant Service, MuleSoft HTTP client and mock upstream.

**Verification:** Both GraphQL end-to-end subtests passed. The full `go test -race ./...` suite passed.

</details>

### Remaining implementation work

Bounded retries, circuit breaking, authentication, authorisation, observability, PostgreSQL persistence, Docker packaging and AWS EKS deployment remain planned.

<details>
<summary><strong>21. MuleSoft HTTP resilience — bounded retries (verified 27 October 2026)</strong></summary>

Implemented three total HTTP attempts with exponential backoff (100 ms, then 200 ms). HTTP 502, 503 and 504 are retryable; HTTP 404 and client timeouts are not. Context cancellation stops retry waiting.

Added retry tests covering transient recovery, maximum attempts, HTTP 404 and cancellation during backoff.

Verification: go test -count=1 -race ./... passed. git diff --check passed.

Future improvements: jitter, Retry-After, overall retry budget and metrics.

</details>

<details>
<summary>22. Production-grade Go HTTP server lifecycle</summary>

### Implementation

- Configured HTTP server timeouts:
  - Read header: 5 seconds
  - Read: 10 seconds
  - Write: 15 seconds
  - Idle: 60 seconds
- Added SIGINT/SIGTERM handling using `signal.NotifyContext`.
- Added graceful HTTP shutdown with a 10-second deadline.
- Added `/healthz` liveness endpoint.
- Added `/readyz` readiness endpoint backed by `atomic.Bool`.
- Moved readiness handling into `internal/health`.
- Marked the application not ready when shutdown begins.

### Verification

- Verified `/healthz` and `/readyz` return HTTP 200 during normal operation.
- Verified the API exits cleanly after SIGTERM.
- Added automated readiness tests for HTTP 200, HTTP 503, and concurrent access.
- Added an HTTP integration test proving `http.Server.Shutdown` waits for an active request to complete.
- Passed `go test -count=1 -race ./...`, `go vet ./...`, and `git diff --check`.

### Production considerations

- The readiness state is currently in-memory and does not verify upstream MuleSoft availability.
- Kubernetes readiness changes and load-balancer endpoint updates are not instantaneous.
- A production EKS deployment needs coordinated traffic draining, appropriate termination grace periods, and lifecycle configuration.
- The shutdown integration test verifies Go HTTP server behaviour, not the complete Kubernetes termination sequence.

</details>

<details>
<summary>23. Structured logging and request correlation</summary>

### Implementation

- Added structured JSON logging using Go's standard `log/slog` package.
- Configured the API to emit JSON startup and shutdown logs.
- Added HTTP middleware that generates a random 128-bit request ID.
- Exposed the request ID through the `X-Request-ID` response header.
- Stored the request ID in `context.Context` for downstream handlers.
- Added structured HTTP access logs containing:
  - Request ID
  - HTTP method
  - Request path
  - Response status
  - Duration in milliseconds
- Propagated request IDs into outgoing MuleSoft HTTP requests.
- Preserved request IDs across retry attempts.

### Verification

- Verified JSON startup and shutdown logs from the running API.
- Verified unique request IDs in `/healthz` and `/readyz` responses.
- Verified that the access-log request ID matches the HTTP response header.
- Added automated tests for HTTP 200, 404 and 500 access logs.
- Added an integration test verifying request ID propagation to mock MuleSoft.
- Passed the full Go test suite with race detection.
- Passed `go vet` and `git diff --check`.

### Production considerations

- Request IDs are generated by the API rather than accepted from untrusted clients.
- Request correlation is implemented, but distributed tracing is not yet implemented.
- The access-logging response wrapper does not yet preserve optional HTTP interfaces such as `http.Flusher` and `http.Hijacker`.
- Metrics, tracing, log aggregation, and production dashboards remain future work.
- Avoid logging participant personal information, authentication tokens, or sensitive GraphQL request bodies.

</details>


<details>
<summary>Authentication baseline — 28 October 2026</summary>

The continuing-session handover reports RS256 JWT authentication with HTTPS JWKS, caching, rotation and cryptographic validation implemented. Tests cover valid, missing, expired, malformed and wrong-audience tokens for both GraphQL endpoints, and public health probes. The reported checks passed: `go test -count=1 -race ./...`, `go vet ./...`, and `git diff --check`. These are handover results, not newly executed checks in Step 74.

Repository inspection found authentication changes uncommitted: `cmd/api/main.go`, `go.mod`, `go.sum`, and 20 new files in `internal/auth/`. Participant-level authorisation is still absent. Current details belong in the [authentication reference](authentication.md).

</details>

<details>
<summary>Step 74 — Documentation refactor, 28 October 2026</summary>

Preserved both original documents before editing; see the [history index](history/README.md). Extracted journal entries unchanged, separated current technical references, and replaced stale overview claims with explicit implementation and verification status. Application code was outside this refactor's scope. Authentication and documentation must remain separate commits.

Step 74 checks passed: current relative Markdown paths and anchors, balanced collapsed details/summary sections, all 23 original journal blocks unchanged, both historical snapshots byte-identical to their originals, `git diff --check`, and unchanged application diffs/authentication file hashes. Original snapshot links retain their historical root context. Go tests were not rerun for this documentation-only change.

The next delivery priorities are Docker/Compose, Kubernetes/EKS, AWS IAM/secrets and CI/CD. Participant-level authorisation remains an explicit security gap.

</details>

<details>
<summary>Step 74 follow-up — Documentation audit maintenance, 28 October 2026</summary>

Added the [documentation change log](documentation-changelog.md) with stable chronological identifiers, a refactor source/destination map and explicit working-tree status. The [history index](history/README.md) now defines the reading sequence and append-only maintenance rules. Original journal entries and immutable snapshots remain unchanged. See DOC-002 for verification and scope.

</details>

<details>
<summary>Continuous audit clarification — 28 October 2026</summary>

**Reason and decision:** the user clarified that the audit is a continuous, step-by-step development approach, not a separate process starting at Step 74. The development journal remains the primary chronology for application, documentation, tests, infrastructure and deployment work, now and going forward.

**Changes:** updated the documentation index, history navigation, maintenance rules and change-log introduction. The documentation change log supports the journal with content-movement detail; it does not replace or reset the development sequence. See DOC-003 in the [supporting change log](documentation-changelog.md).

**Verification:** current relative Markdown paths/anchors and collapsed HTML structure passed; original journal blocks and snapshots remain intact; application files match the pre-refactor baseline; `git diff --check` passed. No Go tests were rerun for this documentation wording change.

**Remaining work:** authentication and documentation changes are still uncommitted. Commit work remains on hold pending approval. Subsequent development steps must continue this journal using the same maintenance rules.

</details>

<details>
<summary>Authentication verification and separate commit — 28 October 2026</summary>

**Changes and decision:** committed only the authentication implementation, API wiring and JWT dependency files as `4a4bec9` (`Add RS256 JWT authentication with cached HTTPS JWKS`). Documentation remains a separate change. No application code was edited during the documentation work.

**Verification:** fresh `go test -count=1 -race ./...` and `go vet ./...` passed. `git diff --check` passed. The initial sandboxed check could not access the Go build cache; the approved rerun succeeded. Git staging also required approval to write the repository index.

**Documentation:** current status references and the testing evidence now identify the authentication commit. Earlier uncommitted-status observations remain preserved as historical entries. See DOC-004 in the [documentation change log](documentation-changelog.md).

**Remaining work:** create the separate documentation commit after link, structure and preservation checks. No push is authorised in this sequence. Participant authorisation remains absent; delivery priorities remain Docker/Compose, Kubernetes/EKS, IAM/secrets and CI/CD.

</details>

<details>
<summary>Docker and Compose packaging — 28 October 2026</summary>

**Baseline:** authentication is committed as `4a4bec9`; the documentation refactor and continuous audit rules are committed as `a1a622e`. This entry continues the development chronology without assigning an unconfirmed workflow step number.

**Changes:** added a multi-stage Dockerfile with separate API and mock targets, `.dockerignore`, `compose.yaml` and `.env.example`. Added the [container reference](containers.md), updated README/index/roadmap and testing evidence. Application Go source is unchanged.

**Decisions:** retain Go 1.27.3, static binaries, HTTPS CA trust, non-root UID/GID 65532, read-only filesystems and dropped capabilities. Bind the API only to localhost; keep the mock internal. Require JWT configuration without providing an authentication bypass. Wait for mock health at startup and allow 15 seconds for API termination. No bundled identity provider or signing keys were added.

**Problems and resolution:** Docker access initially failed under the sandbox; an approved daemon check showed Docker Desktop was stopped. Started it with approval, then built and ran the stack. No application changes were needed.

**Verification:** Compose configuration passed with placeholder identity settings and rejected missing JWT configuration. Both images built successfully on the local Docker engine and both service health checks passed. Host probes returned 200; missing-token requests to both GraphQL endpoints returned 401. Every checked response included a request ID. `id` inside the API reported UID/GID 65532. An API-container request to `http://mockmulesoft:8081/api/participants/P001` returned the expected synthetic participant. Compose stop produced API shutdown initiation/completion logs; the test containers and network were removed. Documentation checks and `git diff --check` passed. No Go source changed, so the previously passing Go checks were not repeated.

**Limits and remaining work:** placeholder identity settings do not prove HTTPS JWKS retrieval or valid-token GraphQL requests inside Docker. A real provider/container integration exercise remains required. Multi-platform publishing, vulnerability scanning and digest pinning are unverified. Container files and accompanying documentation remain uncommitted at this entry. No images were published and no Git push occurred. Next priorities are Kubernetes/EKS, IAM/secrets and CI/CD; participant-level authorisation remains absent.

</details>

<details>
<summary>Local issuer and Docker authentication troubleshooting — 28 October 2026, reconstructed from handover</summary>

**Objective and context:** extend the earlier two-service Docker stage with a local HTTPS RS256 issuer, then authenticate Portal/Mobile queries through the shared service and mock MuleSoft. This missing stage is appended after the earlier Docker entry; no historical text or numbering is replaced.

**Evidence provenance:** the numbered sequence below is supplied by the user's handover, in that order. The repository corroborates the final configuration, not every intermediate failure. No exact timestamps or unprovided command transcripts/error wording are inferred. Where only a symptom was supplied, it is described rather than presented as a verbatim log.

1. **Build — handover:** Docker images for API, local issuer and mock MuleSoft built successfully. The Dockerfile now contains all three targets.
2. **Initial startup — handover:** all three containers initially reached healthy status. Health alone did not establish token validation or Go JWKS trust.
3. **Token issuance — handover:** a locally generated RS256 JWT was obtained successfully. The issuer implementation corroborates a local `POST /dev/token` route; its output and key material are not reproduced.
4. **First authenticated request — handover:** the Portal GraphQL request returned HTTP 401. This confirms rejection, not its exact internal cause; the middleware intentionally returns a generic response.
5. **Diagnostic tooling — handover:** BusyBox `wget` diagnostics encountered unsupported options, including `--ca-certificate`. This is a tool limitation, not proof of an invalid certificate. No precise earlier command string is available.
6. **DNS failure — handover:** the API initially could not resolve `local-issuer`; the Compose service was named `localissuer`. The handover identifies the naming mismatch as the DNS cause; earlier raw resolver errors are unavailable.
7. **DNS correction — handover and current source:** added Docker network alias `local-issuer`, matching the HTTPS URL and certificate hostname while retaining the service name `localissuer`.
8. **Next diagnostic result — handover:** DNS resolution succeeded, then BusyBox `wget` reported certificate verification failure. This was a distinct failure stage. It did not establish how Go would load trust roots.
9. **Trust investigation — handover:** `SSL_CERT_FILE` was absent from the running API container. This is a confirmed configuration omission in the handover, but not enough to attribute every earlier 401 conclusively to TLS; Go can also use default trust roots.
10. **Trust correction — handover and current source:** mounted the local CA bundle at `/certs/ca-bundle.crt` and set `SSL_CERT_FILE=/certs/ca-bundle.crt` in the local API overlay. Rationale: supply the local CA to Go without disabling certificate validation.
11. **Compose validation failure — handover:** a missing `mockmulesoft` service definition was exposed and restored. The local overlay now contains it; the exact historical Compose error text is unavailable.
12. **Dependency startup failure — handover:** `localissuer` lacked a health check required by `depends_on: condition: service_healthy`. This is a Compose dependency configuration issue, separate from API token validation.
13. **Dependency correction — handover and current source:** restored the issuer health check and hardening. The inspected source uses a loopback `wget` probe with `--no-check-certificate`; record this existing limitation explicitly. It does not bypass the API's Go TLS validation and cannot prove certificate trust.
14. **Final startup — handover:** Compose validation/startup succeeded and all three containers were healthy. Repository-session inspection later independently observed a valid merged configuration and all three services healthy.
15. **Final trust prerequisites — handover:** the running API was verified to have `SSL_CERT_FILE=/certs/ca-bundle.crt`, with the mounted bundle readable. Current overlay source corroborates those settings. No broader trust-resolution claim follows from readability alone.
16. **Handover boundary:** successful end-to-end JWT authentication was still awaiting verification. Preserve that boundary; do not retroactively replace it with later observations.

**Implementation details confirmed by inspection:** `compose.yaml` is API-only; `compose.local.yaml` adds the issuer/mock and CA mount. `config/local.env` selects HTTPS JWKS and matching issuer/audience. Issuer certificate SAN contains `DNS:local-issuer`. The Go fetcher still requires HTTPS, rejects redirects, and performs normal TLS validation. The JWT verifier validates RS256 signatures and configured claims. `APP_ENV`/`LOG_LEVEL` are configured in Compose but not consumed by API code. No participant-level authorisation exists.

**Commands/evidence limits:** the original handover does not provide full historical commands for steps 1–16. Current maintained commands are documented in [containers](containers.md) and [testing](testing.md); they are not fabricated historical transcripts. Remaining work at this boundary: verify a valid token and invalid rejection through the Go/API path and query actual participant data before moving to Kubernetes.

</details>

<details>
<summary>Repository reconciliation and scope correction — 28 October 2026</summary>

**Initial inspection:** authentication commit `4a4bec9` and documentation commit `a1a622e` are present. Existing Docker, issuer, config and documentation work is uncommitted. Inspected the requested nine documentation files, Git history/diffs, Compose files, issuer implementation, JWT middleware/verifier and JWKS fetch/cache code. Earlier two-service statements are historical; current references now describe the three-service overlay.

**Observed checks before documentation-only steering:** merged Compose `config --quiet` passed and `ps` showed three healthy containers. The temporary Python/wget token attempt failed with `wget: unrecognized option: ca-certificate=/certs/ca-bundle.crt`; it did not send a verified token request. A temporary Go client then ran on the Compose network, obtained a token using explicit CA/hostname validation, and sent actual P001 participant queries. Portal and Mobile each returned 200 with expected data and no GraphQL errors; tampered-signature, malformed and missing-token requests each returned 401 on both routes. Request IDs were present. No tokens or private keys were printed. `go test -count=1 -race ./...` and `go vet ./...` completed successfully before the temporary source edit. These are later observations, not claims about the handover state.

**Premature development and correction:** before the user requested documentation-first approval, the assistant temporarily added a Go-based issuer TLS health probe, a CA mount for that probe, broader Docker context exclusions and a reusable smoke source, then launched `up -d --build --wait`. On the scope correction, restored issuer source, local Compose and Docker ignore contents to captured pre-turn hashes and removed only the newly introduced smoke source. All pre-existing uncommitted work was preserved. The already-started rebuild subsequently completed and reported three healthy services with the temporary issuer image/configuration. Source restoration did not recreate containers; the live issuer may therefore differ from the restored source. This drift requires explicit reconciliation in the next approved verification stage. No further Docker mutation or JWT checks were performed after the documentation-only instruction.

**Documentation decisions:** append the missing handover chronology without rewriting existing entries; separate handover, inspected-source and observed-run evidence. Update container/authentication/testing references, roadmap and navigation. A proposed repeatable Go client is documented, not installed into the application. See DOC-006 in the [documentation change log](documentation-changelog.md).

**Verification for this documentation stage:** checked Markdown paths/anchors and collapsed sections, unchanged journal/change-log prefixes, immutable snapshots, preserved original journal blocks and `git diff --check`. This stage does not add a new application-test or live-JWT result.

**Remaining work and approval boundary:** present the documentation proposal for user approval before continuing JWT verification or resolving live/source drift. Do not commit or push. After approved Docker verification, the next implementation priority is a minimal Kubernetes demonstration suitable for explaining EKS concepts. Participant-level authorisation remains missing.

</details>

<details>
<summary>First-time developer setup documentation — 28 October 2026</summary>

**Objective:** close onboarding gaps so a developer can prepare and run the local three-service prototype without pre-existing certificates or an external identity provider. The user requested documentation changes only; no further live JWT or Kubernetes work was undertaken.

**Problem and cause:** the earlier guide assumed provisioned certificate files and local environment configuration. Root `.env.example` still targets an external issuer and lacks `MULESOFT_URL` required by the current base Compose file. The authentication helper required manual creation of a temporary Go file. Those assumptions prevented a complete first-time walkthrough.

**Documentation corrections:** added an ordered path for prerequisites/checkout, preserved-or-created local configuration, local CA/RSA issuer certificate generation with `DNS:local-issuer`, bundle creation, certificate verification, startup, authentication checks, tests and shutdown. Added renewal with preserved backups. Defined a Compose function with fixed project/network name and explicit environment-file precedence. Documented private-directory/key permissions for UID 65532, the existing health-probe limitation, optional host Go and a Debian Go image alternative for race tests. Converted the documented smoke client into a single shell block using a new temporary directory. Updated README/index/authentication navigation; commands remain authoritative in the container/testing references.

**Decisions and limits:** certificate creation refuses to overwrite an existing certificate directory; existing config is retained. The local CA is not installed into OS trust. The local issuer key is readable via its bind mount for the non-root container, while its host parent directory is mode 700 and CA key mode 600. This remains a synthetic local-only identity fixture without participant-level authorisation. Docker/issuer files are still uncommitted, so the guide explicitly states that cloning the current remote alone does not supply the working-tree additions.

**Observed documentation validation:** all Bash blocks in container/testing references passed `bash -n`. Executed the exact certificate-generation block in an isolated temporary checkout directory: RSA keys/certificates were created; OpenSSL CA-chain, server-purpose and `local-issuer` hostname validation passed. Checked documented permissions and reran the block; all generated files retained identical hashes. Existing project certificates and running containers were untouched. Relative paths/anchors, collapsed sections, unchanged prior journal/change-log prefixes, original journal blocks, immutable snapshots and `git diff --check` passed.

**Remaining work:** review/approve documentation, then reconcile source/runtime state and repeat the full live authentication check. The complete new-developer Docker walkthrough has not been rerun end to end during this documentation stage. No application code, Compose configuration or dependencies changed; no commit or push occurred. See DOC-007 in the [documentation change log](documentation-changelog.md).

</details>

<details>
<summary>Approved Docker JWT verification against preserved source — 28 October 2026</summary>

**Objective and approval:** after the documentation-first reconciliation and onboarding instructions, the user asked to continue. Resumed the planned end-to-end checks without committing, publishing images or starting Kubernetes development.

**Source/runtime reconciliation:** validated the two-file local Compose configuration, then ran `docker compose --project-name participant-services --env-file config/local.env -f compose.yaml -f compose.local.yaml up -d --build --wait`. All three images built and services became healthy. The issuer was recreated from the restored original source, replacing the earlier temporary health-probe image. Checked all running image IDs against the rebuilt tagged images; they matched. No source edits were needed.

**Actual authentication test:** extracted the exact shell block from `docs/testing.md` and ran it against the local stack. Its temporary Go helper fetched an RS256 token over CA/hostname-validated HTTPS and kept it in memory. Both Portal and Mobile P001 participant queries returned HTTP 200 and the exact expected synthetic data with no GraphQL errors. Tampered-signature, malformed and missing-token requests returned HTTP 401 on each endpoint. Request IDs were present. No token, private key or secret environment output was displayed.

**Fresh-cache verification:** recreated only the API using the same Compose options with `up -d --force-recreate --no-deps --wait api`, then reran the exact documented helper. All eight checks passed again. Because the new process had an empty JWKS cache, valid participant results establish a fresh Go HTTPS JWKS/signature-validation path for this setup, not just cached trust or a BusyBox diagnostic. The handover's unresolved state remains preserved historically.

**Other observed checks:** `go test -count=1 -race ./...` and `go vet ./...` completed successfully. The API environment contained the expected `SSL_CERT_FILE`; the mounted file was readable. Public `/healthz` and `/readyz` returned 200 with request IDs. All three containers were healthy. Existing certificates and user configuration were preserved. Markdown links/anchors, collapsed section structure, prior journal/change-log prefixes and `git diff --check` passed.

**Documentation:** updated current authentication, container, testing and roadmap status with the approved verification outcome; see DOC-008 in the [documentation change log](documentation-changelog.md).

**Limits and next steps:** participant-level authorisation is still absent. The existing issuer health check still bypasses certificate validation even though the API and test helper do not. Local token issuance is an unauthenticated development fixture and reuses the TLS signing key. No production security assurance, Docker rotation exercise, clean-checkout end-to-end run or Kubernetes deployment is claimed. The stack is left running for local development. All existing uncommitted work remains; no commit or push was made. Next priority is a minimal demonstrable Kubernetes deployment explaining EKS concepts.

</details>

<details>
<summary>Minimal local Kubernetes demonstration — 28 October 2026</summary>

**Objective and context:** after verified Docker/JWT integration, prepare a minimal demonstrable Kubernetes deployment for explaining EKS concepts. Preserve the Compose setup and existing uncommitted work. No AWS provisioning, image publication, commit or push was authorised or performed.

**Initial assessment:** kubectl v1.36.1 was installed, but no kubeconfig contexts or kind/minikube executables were available. Installed kind v0.33.0 into `/tmp/participant-k8s-tools` with approval, then created an isolated Docker-based cluster named `participant-services` with `/tmp/participant-services-kubeconfig`. The node runs Kubernetes v1.37.0. The normal kubeconfig was not modified.

**Implementation:** added `deploy/kubernetes/local/` with a restricted namespace, API configuration, three single-replica Deployments/ClusterIP Services and a separate authentication Job. Added `docs/kubernetes.md` with prerequisites, image build/load commands, Secret/CA setup, rollout checks, authentication testing, port-forward access, cleanup and EKS mapping. Updated current overview/navigation, roadmap and testing evidence. No application or Compose source changed.

**Decisions:** use the existing TLS/JWT behaviour and certificate hostname `local-issuer`; keep issuer/mock internal. Generate the local certificate Secret and CA ConfigMap from existing ignored files without persisting Secret YAML. The CA private key remains on the host. Apply non-root UID/GID 65532, fsGroup for group-readable Secret mounts, disabled service-account token mounting, dropped capabilities, read-only root filesystems, RuntimeDefault seccomp and resource requests/limits. Set imagePullPolicy Never for loaded kind images; ECR publication remains future work. API probes use its health/readiness routes, mock probes read synthetic P001 and issuer probes check TCP reachability without pretending to verify certificate trust. End-to-end verification belongs to the Job.

**Failure and confirmed root cause:** the first generated CA ConfigMap apply was rejected: `metadata.annotations: Too long: may not be more than 262144 bytes`. Client-side apply tried to duplicate the existing large CA bundle in its last-applied annotation. Switched generated Secret/ConfigMap commands to server-side apply with a dedicated field manager. The retry succeeded without changing the certificate bundle. The runbook records the corrected commands.

**Commands and observed results:** `kubectl kustomize deploy/kubernetes/local` rendered successfully. Loaded the already-built API/issuer/mock images plus the Go test image into kind. Namespace/Secret/CA setup and `kubectl apply --dry-run=server -k deploy/kubernetes/local` passed; applied manifests and waited for all three rollouts, which became ready with zero initial restarts. The test Job also passed server dry run, then completed successfully. It used the existing documented Go client with CA/hostname validation: Portal and Mobile participant queries returned 200 and expected P001 data; tampered-signature, malformed and missing tokens returned 401 on each endpoint. Tokens and keys were not printed. The API began as a new process with an empty JWKS cache. Brief localhost port-forwarding on 18080 returned 200 with request IDs from `/healthz` and `/readyz`; the forward was stopped.

**Documentation verification:** checked current Markdown links/anchors, collapsed sections, shell examples, unchanged prior journal/change-log prefixes and snapshots, and `git diff --check`. Application Go was unchanged, so the previously passing race tests and vet were not repeated for manifest-only work. See DOC-009 in the [documentation change log](documentation-changelog.md).

**Limitations and next steps:** the cluster remains running for demonstration; cleanup commands remove only the named kind cluster/namespace. Compose and host certificates remain intact. One replica and illustrative resource values are not production sizing or high availability. No network policy, ingress, autoscaling, AWS IAM/managed secrets, ECR or EKS deployment is implemented. Kubernetes traffic draining is not verified. Participant-level authorisation remains absent; the local issuer remains an unauthenticated synthetic fixture reusing its TLS key for signing. Next priorities are workload IAM/secrets and a controlled EKS deployment design, then CI/CD; do not expose this lab publicly. All files remain uncommitted pending review.

</details>

<details>
<summary>Workload identity and managed-secrets preparation — 28 October 2026</summary>

**Objective:** follow the local Kubernetes demonstration with a minimal AWS identity/secrets design, implementing only what can be justified and validated locally. Preserve existing uncommitted work and avoid AWS provisioning or credential introduction.

**Assessment and decision:** the API has no AWS SDK calls or secret-backed upstream credential consumer. Its issuer URL/audience/JWKS and CA are configuration, not application AWS credentials. Granting IAM or Secret-read permissions now would add unused access. Added a dedicated `participant-api` Kubernetes service account with automatic token mounting disabled, selected it in the API Deployment and left RBAC grants absent. The issuer receives its existing Secret as a volume; the API needs public JWKS, not issuer-key access.

**Prepared design:** added `docs/aws-identity-secrets.md` covering end-user/Kubernetes/AWS identity boundaries, least-privilege policy scope, future Pod Identity association, managed-secret delivery/reload decisions and cloud acceptance criteria. Added two non-deployable JSON examples under `deploy/aws/examples/`: a Pod Identity trust policy constrained by cluster/namespace/service-account tags and an exact-secret read/describe policy. Placeholders remain explicit; no role, policy, secret or EKS association was created. Current AWS primary documentation informed the examples; references are linked in the guide. AWS CLI is installed but no AWS account/API calls were needed or performed.

**Verification:** Kustomize rendering and API-server dry run passed, manifests applied to the isolated kind cluster, and API rollout completed. Checked the selected service account and disabled automatic token flag; the default Kubernetes API-token file was absent. `kubectl auth can-i get secrets --as=system:serviceaccount:participant-services:participant-api` returned `no`/exit 1 as expected. Recreated only the disposable authentication test Job; it completed with valid Portal/Mobile participant queries returning expected data and tampered, malformed and missing tokens rejected on both routes. No tokens or secrets were printed. Policy examples parsed as JSON and runbook shell syntax passed; this is not AWS authorization validation. Documentation links/anchors, collapsed sections, previous journal/change-log prefixes, snapshots and `git diff --check` passed. No application Go changes required repeating the already-passing Go suite.

**Problems and limits:** no deployment failure occurred in this stage. The main boundary is lack of a real AWS consumer/account-specific deployment, so cloud permission tests cannot be claimed from kind. Pod Identity/Secrets Manager/CSI/KMS integration is proposed, and named service accounts do not add network isolation or participant authorisation. The local cluster and Compose stack remain available. No commit or push occurred.

**Next steps:** add credential-free CI validation/build automation, or prepare a reviewed cloud exercise only after selecting account/region, supported cluster compute, cost limit, operator identity, one synthetic secret consumer and cleanup owner. Preserve participant-level authorisation as an outstanding security requirement before public exposure. See DOC-010 in the [documentation change log](documentation-changelog.md).

</details>

<details>
<summary>Credential-free CI validation and image builds — 28 October 2026</summary>

**Objective:** add repeatable validation/build automation after local Docker/Kubernetes verification, without cloud credentials, registry publication or deployment authority. Preserve existing uncommitted work and the continuous audit.

**Assessment and implementation:** no `.github` workflow existed. Added `.github/workflows/ci.yaml` with pull-request, main-push and manual triggers; read-only contents permission; timeouts/concurrency cancellation; Go formatting, race tests, vet and module-file checks; documentation/JSON/Compose/Kustomize checks; and all three image target builds after validation. Both official actions are pinned to commit hashes resolved from their v6 tags, and checkout credentials are not persisted. Kubectl is versioned and checksum-verified. No AWS permissions, registry login, secret values, cluster access or publishing/deployment steps were introduced.

**Documentation tooling:** added a small Python-standard-library checker for current local Markdown paths/anchors and collapsed/balanced details sections. Immutable snapshots retain their original link context and are excluded from path/anchor checks while still checked structurally. The checker does not request external URLs or rewrite content. Added `docs/ci.md` with workflow scope, local equivalent commands, activation conditions and evidence limits; updated README/index/roadmap/testing.

**Observed verification:** fresh `go test -count=1 -race ./...`, `go vet ./...`, formatting and unchanged module-file checks passed. Ran local merged Compose validation with non-secret explicit values, Kustomize rendering and all three Docker target builds with `:ci` tags; they passed without changing running stacks. IAM example JSON parsing and documentation checks passed. Actionlint v1.7.7 accepted the workflow; optional ShellCheck was disabled, with multiline shell/runbook Bash syntax checked separately. A disposable negative Markdown fixture confirmed missing-link, open-section and unbalanced-section failures and was removed. No application module dependencies were changed by installing the temporary lint tool. Previous journal/change-log prefixes, snapshots, original journal entries and `git diff --check` passed preservation checks.

**Problems and evidence boundaries:** no runtime failure was observed in this stage. The workflow is uncommitted and has never run on GitHub, so hosted runner/tool-install behaviour, Actions settings and AMD64 builds are not claimed as verified. Local Docker builds used the development engine and cached layers. Offline manifest/JSON checks are not API-server/IAM validation. Automated Compose/kind authentication runs, image/dependency scanning, schema compatibility, release publication and CD remain future work.

**Remaining work:** review the accumulated implementation and documentation before any commit or push. Once explicitly authorised and published, observe the first hosted run before calling CI verified on GitHub. Future CD needs a concrete OIDC/ECR/EKS permission and rollback design; participant-level authorisation remains missing. Local Compose/kind services and existing certificates remain untouched. See DOC-011 in the [documentation change log](documentation-changelog.md).

</details>

<details>
<summary>Accumulated-change review and commit preparation — 28 October 2026</summary>

**Objective:** review the completed local deployment/CI preparation before requesting permission to commit. Preserve all existing uncommitted work and avoid inventing a further cloud feature or treating earlier continue instructions as commit authorisation.

**Findings and corrections:** the external-provider `.env.example` lacked `MULESOFT_URL`, which the current base Compose file requires; added a placeholder upstream URL and clarified that the bundled local stack uses `config/local.env`. The Docker ignore rules excluded private-key patterns but still admitted the local configuration directory and historical backup into the context; excluded `config/` and `*.bak`, then updated the container guide. Runtime bind mounts are unaffected. The user's `compose.before-local-split.yaml.bak` remains intact and is excluded from the proposed commit set. No application logic changed.

**Observed validation:** the example now supplies all four required API variables and base Compose validation with that file passed. Existing CA/issuer key paths remain ignored by Git. A temporary Docker context-check build first failed before reading its Dockerfile with a `/private/tmp` metadata/xattr permission error. Passing the same Dockerfile through stdin avoided that host-file metadata issue; the build then proved config and backup absence and required source/module-file presence without printing contents. Running services were unchanged. Current Markdown links/anchors, collapsed sections, previous journal/change-log prefixes and `git diff --check` passed. Broader Go/runtime checks from the preceding stages remain their recorded evidence; they were not repeated for these example/ignore changes.

**Proposed commit boundaries:** Docker/issuer/environment packaging; local Kubernetes and workload-identity/AWS examples; reconciled modular documentation and complete chronological audit; credential-free CI workflow and reusable documentation checker. The final commit series must be checked together before any push. Do not include host certificates, keys, token material or the user's backup. Git staging/commit/push remain pending explicit approval; no commit or push occurred in this review. See DOC-012 in the [documentation change log](documentation-changelog.md).

</details>

<details>
<summary>Approved commits, push and first hosted CI verification — 28 October 2026</summary>

**Authorisation and scope:** the user explicitly approved committing, pushing and continuing. Rechecked Git state and staged only the agreed files. Reviewed environment files for expected local settings/placeholder URLs; certificate/private-key paths remain ignored. Preserved the user's untracked `compose.before-local-split.yaml.bak` outside all commits.

**Created commits:** `9a81351` Docker Compose/local HTTPS issuer and non-secret environment configuration; `ef72e82` local Kubernetes, service-account and proposed AWS-policy examples; `9885c1a` modular documentation and chronological audit; `bd91def` credential-free CI and documentation checker. Pushed to `origin/main`; the remote advanced from `1797011` to `bd91def`, also publishing the existing lifecycle/logging/authentication/documentation commits already present locally.

**Hosted results observed:** initial run listing was empty immediately after push, then GitHub queued run 40651138381. Monitored it to successful completion. Documentation/configuration passed in 9 seconds, Go tests/vet in 1 minute 36 seconds and all three service image builds in 41 seconds. This includes hosted race tests, vet, formatting/module checks, Markdown/JSON/Compose checks, checksum-verified kubectl installation, Kustomize rendering and Ubuntu/AMD64 builds. No failures required workflow fixes. [Run evidence](https://github.com/shawfire/participant-services/actions/runs/40651138381).

**Current documentation:** corrected current uncommitted/unpushed status statements, recorded commit IDs and hosted evidence, and preserved all previous entries as historical records. Checked current documentation links/anchors, collapsed sections, unchanged journal/change-log prefixes, immutable snapshots and `git diff --check` before the follow-up documentation commit. This follow-up cannot record its own final hash; Git history provides it.

**Remaining scope:** no hosted runtime-stack test, AWS IAM/EKS deployment, image publication or participant-level authorisation is claimed. Existing Compose/kind services and local certificates are unchanged. The only intentionally untracked artefact after the follow-up is the user's preserved Compose backup. Next development should address remaining explicitly documented gaps or a reviewed cloud exercise, not treat passing CI as production readiness.

</details>

<details>
<summary>Shared participant self-access authorisation — 29 October 2026</summary>

**Objective and initial state:** the user approved continuing the suggested self-access stage and requested ongoing explanations and maintained Markdown documentation. Initial Git state contained only the user's untracked Compose backup. Authentication and hosted CI were previously verified, but service `Get` accepted any participant ID and the local issuer subject `participant-123` did not match domain IDs such as P001. No commits or pushes were requested for this new stage.

**Decision and rationale:** choose participant self-access only, with an explicit trusted signed `participant_id` claim rather than treating `sub` as a record ID or accepting a caller-supplied header. The issuer owns that mapping. Add a small internal identity-context package with value-copy principals; JWT middleware creates it only after existing cryptographic/registered-claim verification. The participant service depends on identity, not the JWT/HTTP implementation, so both consumers share the policy.

**Implementation:** JWT claims now include optional `participant_id`; the local fixture signs P001 for its fixed subject. Shared `Service.Get` checks nonempty subject/mapping and exact requested-ID equality before memory lookup or repository access. Missing identity/mapping and cross-participant IDs return `ErrForbidden`. A repository response containing another ID is rejected too. Portal/Mobile translate denial to a null participant field and generic `forbidden`/`FORBIDDEN` GraphQL error without exposing target existence. Invalid JWTs still produce HTTP 401 before GraphQL; valid-but-forbidden GraphQL fields use HTTP 200 with errors. Cancellation and authorised not-found behaviour are retained. There is no service bypass constructor or production test-mode switch.

**Tests and observed results:** focused race tests passed, then the full `go test -count=1 -race ./...` suite and `go vet ./...` passed. Existing direct service/handler fixtures now provide explicit trusted context. Service tests cover absent/empty identity, missing mapping, cross/unknown/empty IDs and no repository calls on denial, plus wrong-record upstream responses. The JWT HTTP integration fixture now uses real mock MuleSoft participant queries in addition to existing authentication cases: P001 and P002 mappings can read their own records; cross-participant, unknown-other and missing-mapping requests return generic denials without upstream access. A tampered mapping without re-signing is rejected as unauthenticated; identity headers do not override signed claims. No implementation/test failure was encountered in this stage.

**Working demonstrations:** rebuilt the local Compose stack without replacing certificate files, obtained a fresh token in memory over CA/hostname-validated HTTPS, and ran the updated documented smoke client. Both own-P001 queries returned exact expected data; missing/malformed/tampered tokens returned 401; valid P001 requests for P002/P999 returned `FORBIDDEN` and no data on both routes, despite a spoofed `X-Participant-ID`. All twelve cases passed. Loaded the rebuilt API/issuer/mock images into the existing isolated kind cluster, restarted/waited for all three Deployments, refreshed the smoke-client ConfigMap and ran a new disposable authentication Job; all twelve cases passed again. Tokens/private keys were not printed. Local stacks remain running with this working-tree change; old tokens without a mapping must be replaced before reading participants.

**Documentation:** added the focused authorisation reference; updated current authentication/GraphQL/architecture/README/navigation and remaining-work statements. Extended the single copy-and-run test client in testing rather than duplicating another helper. Updated Docker/Kubernetes evidence and kept older absent-authorisation statements in this journal and historical snapshots untouched. Current Markdown links/anchors, collapsed sections, shell examples, original historical blocks/snapshots and previous journal/change-log prefixes passed checks; `git diff --check` passed. See DOC-014 in the [documentation change log](documentation-changelog.md).

**Risks and next steps:** issuer provisioning accuracy is a trust assumption; there is no dynamic entitlement/revocation service, staff/delegate/multiple-participant/field-level policy or production identity assurance. The local issuer remains unauthenticated and reuses its TLS key for signing. Its existing health-probe TLS bypass is unchanged; the Go API/client continue full TLS/signature validation. No new dependencies or AWS resources were introduced. The prior hosted CI results describe earlier commits; this implementation is uncommitted and locally verified only. Review before a separate implementation/documentation commit; do not push without approval. The user's Compose backup remains intact.

</details>

<details>
<summary>Participant self-access pull request preparation — 29 October 2026</summary>

**Authorisation and scope:** the user requested all pending implementation/documentation changes in a pull request for review. Refreshed `origin/main` and confirmed it matched local HEAD, then created `feat/participant-self-access`. Preserved the user's untracked Compose backup outside the review changes.

**Verification and commits:** repeated the full `go test -count=1 -race ./...` suite and `go vet ./...`; both passed. Current Markdown checks and `git diff --check` passed. Committed implementation/tests as `e5c748f` (`Enforce shared participant self-access from verified JWT mapping`). Documentation follows separately with the complete policy reference, extended test commands, current status and append-only audit. Earlier uncommitted-state entries remain historical observations.

**Review focus:** explicit signed mapping rather than subject-as-ID; shared enforcement before repository access; generic GraphQL denial; missing-mapping/old-token compatibility; no staff/delegate policy or production entitlement lifecycle. Prior Docker and Kubernetes twelve-case results remain the runtime evidence. Push only the review branch, open against `main`, and do not merge. Hosted checks are to be inspected after publication. See DOC-015 in the [documentation change log](documentation-changelog.md).

</details>
