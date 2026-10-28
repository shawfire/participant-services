# Testing and verification

[Documentation index](README.md) · [Development journal](development-journal.md)

## Verification evidence

The 28 October 2026 session handover reports these checks passed on the authentication working tree:

```bash
go test -count=1 -race ./...
go vet ./...
git diff --check
```

These initial results were reported evidence. A subsequent verification run on 28 October 2026 passed `go test -count=1 -race ./...` and `go vet ./...` before authentication commit `4a4bec9`; `git diff --check` also passed. The first sandboxed attempt could not access the Go build cache; the approved rerun completed successfully. Earlier locally verified results and problems are preserved chronologically in the journal. Documentation validation does not establish production readiness.

<details>
<summary>Coverage and boundaries</summary>

| Area | Existing coverage | Limit |
|---|---|---|
| Todo | Sequential create/query resolver test | Does not prove concurrent slice writes safe |
| Participant service | Lookup, missing ID, cancelled context, 100 concurrent reads, repository delegation/errors | No participant authorisation or write path |
| Portal and Mobile | Resolver/HTTP participant queries | Isolated handler tests do not prove runtime security wiring |
| MuleSoft | Mock handlers, success, missing ID, upstream errors, timeout, cancellation, network failure; retries and backoff cancellation | Simulated backend, no actual enterprise system |
| End-to-end | Both GraphQL APIs through service and mock MuleSoft | Separate from JWT integration fixture |
| JWT | Signature/claim rejection, JWK parsing/validation, HTTPS fetch rules, cache expiry/concurrency, rotation and refresh limiting, middleware | No resource-level policy |
| JWT HTTP integration | Valid, missing, expired, malformed and wrong-audience tokens on both GraphQL paths; public `/healthz` and `/readyz` | Uses `__typename`; does not exercise participant resolvers or MuleSoft, and builds its own router |
| Health/lifecycle | Readiness 200/503 and concurrency, active request completes during shutdown | No Kubernetes termination/traffic-draining verification |
| Observability | Request ID middleware, access logs for 200/404/500, upstream correlation | No distributed tracing or log backend |

</details>

<details>
<summary>Focused commands</summary>

```bash
go test -v ./internal/auth
go test -v ./graph/portal ./graph/mobile
go test -v ./internal/participant ./internal/integration/mulesoft ./internal/mockmulesoft
go test -v ./internal/health ./internal/observability ./cmd/api
```

Tests use local fixtures, generated test signing keys and `httptest` servers; an external identity provider is not required. Full race-enabled checks remain the release baseline. Regeneration commands are in [GraphQL](graphql.md).

</details>

## Step 74 documentation checks

Check local Markdown paths/anchors, balanced `<details>`/`<summary>` sections with no `open` attribute, preserved original journal blocks and byte-identical snapshots. Run `git diff --check`, and compare application diffs and authentication file hashes against the pre-refactor baseline. Historical snapshots retain original links verbatim; their index explains the original path context.
