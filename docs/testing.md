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
| Participant service | Lookup, missing ID, cancelled context, 100 concurrent reads, repository delegation/errors | Self-access implemented; no staff/delegate policy or write path |
| Portal and Mobile | Resolver/HTTP participant queries | Isolated handler tests do not prove runtime security wiring |
| MuleSoft | Mock handlers, success, missing ID, upstream errors, timeout, cancellation, network failure; retries and backoff cancellation | Simulated backend, no actual enterprise system |
| End-to-end | Both GraphQL APIs through service and mock MuleSoft | Separate from JWT integration fixture |
| JWT | Signature/claim rejection, JWK parsing/validation, HTTPS fetch rules, cache expiry/concurrency, rotation and refresh limiting, middleware | No resource-level policy |
| JWT HTTP integration | Valid, missing, expired, malformed and wrong-audience tokens on both GraphQL paths; public `/healthz` and `/readyz` | Also tests participant queries, mapping tampering and forbidden/no-upstream behaviour; builds its own router |
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

## Container verification

On 28 October 2026, both Docker targets built and Compose health checks passed. Host probes returned 200; both GraphQL routes rejected missing tokens with 401 and returned request IDs. Verified UID/GID 65532, API-container connectivity to mock P001, and graceful API shutdown logs. Missing JWT settings were rejected by Compose. Test containers/network were removed afterward. This earlier two-service stage used placeholder identity settings; valid-token requests and HTTPS JWKS retrieval were unverified at that stage. See the [container guide](containers.md) and latest [journal entry](development-journal.md).

## Docker authentication evidence

| Stage / source | Evidence | Limit |
|---|---|---|
| Supplied 28 October handover | Three images built; all services healthy; token obtained; initial Portal request 401; DNS alias, CA mount/environment and Compose dependencies corrected | Successful Go JWT/JWKS verification still pending at handover; earlier raw command output is unavailable |
| Repository inspection in this session | Local overlay contains alias, API CA setting/mount, mock and issuer health checks; certificate SAN is `local-issuer`; merged Compose validation and `ps` succeeded | Files and health status alone do not prove Go TLS trust |
| Session check before documentation-only steering | Temporary Go client on the Compose network obtained a token over CA/hostname-validated HTTPS; Portal and Mobile P001 queries returned 200 and exact expected data; tampered-signature, malformed and missing tokens returned 401 on each route; request IDs present | One observed run, before subsequent issuer rebuild; not a fresh-run/cache-isolation or rotation exercise |
| Session Go checks before steering | `go test -count=1 -race ./...` and `go vet ./...` completed successfully | Ran before the temporary health-probe edit; no new development tests during documentation-only work |
| Scope correction | Temporary source changes were restored to captured pre-turn hashes; new smoke source removed | Already-started rebuild completed with a temporary verified issuer probe; running image may differ from restored source until an approved reconciliation |

**Approved repeat verification — 28 October 2026:** rebuilt all three images from the preserved source; local Compose validation/startup passed. The exact documented authentication shell block passed twice, including after force-recreating the API with an empty JWKS cache. Both valid participant queries returned 200 with expected P001 data and no GraphQL errors; each route rejected tampered-signature, malformed and missing tokens with 401. The helper verified issuer CA/hostname without displaying tokens. Fresh API success establishes Go HTTPS JWKS validation for this configuration, rather than a BusyBox trust inference. All containers were healthy and their image IDs matched the rebuilt images; earlier source/runtime drift is resolved. API CA environment/readability and public probes (200 with request IDs) passed. `go test -count=1 -race ./...` and `go vet ./...` passed. Existing certificates were preserved; the entire clean-checkout walkthrough and alternate Debian test-image command were not rerun in this stage.

To repeat a cold-cache check after local startup:

```bash
# dc is defined in the container setup guide. Recreating API interrupts local requests.
dc up -d --force-recreate --no-deps --wait api
# Then run the authentication shell block below.
```

No API TLS bypass or JWT verifier change was required. At that 28 October stage, the issuer health-probe bypass and missing participant authorisation remained limitations. The later self-access implementation is described in [authorisation](authorisation.md).

## Local authentication smoke check

<details>
<summary>Copy-and-run check after starting the local stack</summary>

These are instructions, not additional results from this documentation stage. Complete the [first-time local setup](containers.md#first-time-local-setup) first; it fixes the project name and creates certificates/configuration. Inspect only the specific non-secret setting and bundle readability:

```bash
docker compose --env-file config/local.env -f compose.yaml -f compose.local.yaml exec -T api sh -c 'test "$SSL_CERT_FILE" = /certs/ca-bundle.crt && test -r "$SSL_CERT_FILE"'
go test -count=1 -race ./...
go vet ./...
docker compose --env-file config/local.env -f compose.yaml -f compose.local.yaml config --quiet
```

For the end-to-end check, copy this complete shell block into a terminal at the repository root. It writes a standard-library Go client into a new private temporary directory. It keeps the token in memory and prints only route, test label and status. It validates issuer CA/hostname, then compares both participant responses and checks tampered, malformed and missing tokens. Do not use shell tracing or print token responses. Panics contain generic diagnostics, not JWTs or keys.

```bash
smoke_dir=$(mktemp -d /tmp/participant-smoke.XXXXXX)
cat > "$smoke_dir/main.go" <<'GO'
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func must(ok bool, msg string) {
	if !ok {
		panic(msg)
	}
}
func main() {
	ca, err := os.ReadFile("/certs/ca-bundle.crt")
	must(err == nil, "read CA")
	roots := x509.NewCertPool()
	must(roots.AppendCertsFromPEM(ca), "parse CA")
	client := &http.Client{Timeout: 10 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}}}
	resp, err := client.Post("https://local-issuer:8443/dev/token", "application/json", nil)
	must(err == nil, "verified TLS issuer request failed")
	defer resp.Body.Close()
	must(resp.StatusCode == 200, "issuer response")
	var token struct {
		AccessToken string `json:"access_token"`
	}
	must(json.NewDecoder(resp.Body).Decode(&token) == nil, "token decode")
	var parts []string
	// Split without exposing token in diagnostics.
	start := 0
	for i, c := range token.AccessToken {
		if c == '.' {
			parts = append(parts, token.AccessToken[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token.AccessToken[start:])
	must(len(parts) == 3, "JWT format")
	header, err := base64.RawURLEncoding.DecodeString(parts[0])
	must(err == nil, "header decode")
	var h map[string]any
	json.Unmarshal(header, &h)
	must(h["alg"] == "RS256", "RS256 required")
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	must(err == nil && len(sig) > 0, "signature decode")
	sig[0] ^= 1
	bad := parts[0] + "." + parts[1] + "." + base64.RawURLEncoding.EncodeToString(sig)
	cases := []struct {
		path, query, field string
		expected           string
	}{
		{"/graphql/portal", `{ participant(id: "P001") { id firstName lastName } }`, "participant", `{"participant":{"id":"P001","firstName":"Alex","lastName":"Taylor"}}`},
		{"/graphql/mobile", `{ profile(participantId: "P001") { participantId displayName } }`, "profile", `{"profile":{"participantId":"P001","displayName":"Alex Taylor"}}`}}
	for _, c := range cases {
		for _, tc := range []struct {
			name, token string
			status      int
		}{{"valid RS256", token.AccessToken, 200}, {"tampered signature", bad, 401}, {"malformed", "not-a-jwt", 401}, {"missing", "", 401}} {
			body, _ := json.Marshal(map[string]string{"query": c.query})
			req, _ := http.NewRequest("POST", "http://api:8080"+c.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			r, e := client.Do(req)
			must(e == nil, "API request failed")
			b, _ := io.ReadAll(r.Body)
			r.Body.Close()
			must(r.StatusCode == tc.status, c.path+" "+tc.name+" unexpected status")
			must(r.Header.Get("X-Request-ID") != "", "missing request ID")
			if tc.status == 200 {
				var actual struct {
					Data   json.RawMessage `json:"data"`
					Errors []any           `json:"errors"`
				}
				must(json.Unmarshal(b, &actual) == nil && len(actual.Errors) == 0, "GraphQL errors")
				var a, w any
				json.Unmarshal(actual.Data, &a)
				json.Unmarshal([]byte(c.expected), &w)
				aa, _ := json.Marshal(a)
				ww, _ := json.Marshal(w)
				must(bytes.Equal(aa, ww), "participant data mismatch")
			}
			fmt.Println(c.path, tc.name, r.StatusCode)
		}
		for _, id := range []string{"P002", "P999"} {
			body, _ := json.Marshal(map[string]string{"query": strings.ReplaceAll(c.query, "P001", id)})
			req, _ := http.NewRequest("POST", "http://api:8080"+c.path, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
			req.Header.Set("X-Participant-ID", id)
			r, err := client.Do(req)
			must(err == nil, "denied query request failed")
			var denied struct {
				Data   map[string]json.RawMessage `json:"data"`
				Errors []struct {
					Message    string         `json:"message"`
					Extensions map[string]any `json:"extensions"`
				} `json:"errors"`
			}
			err = json.NewDecoder(r.Body).Decode(&denied)
			r.Body.Close()
			must(err == nil && r.StatusCode == 200, "expected GraphQL forbidden response")
			must(len(denied.Errors) == 1 && denied.Errors[0].Message == "forbidden" && denied.Errors[0].Extensions["code"] == "FORBIDDEN", "expected generic forbidden error")
			must(string(denied.Data[c.field]) == "null", "denied participant data exposed")
			must(r.Header.Get("X-Request-ID") != "", "missing denied-query request ID")
			fmt.Println(c.path, "cross-participant", id, "FORBIDDEN (no data)")
		}

	}
	fmt.Println("Verified TLS issuer retrieval and authenticated participant data passed; tokens never printed.")
}

GO

docker run --rm --network participant-services_default \
  --mount "type=bind,src=$smoke_dir/main.go,dst=/smoke.go,readonly" \
  --mount "type=bind,src=$(pwd)/config/certs/ca-bundle.crt,dst=/certs/ca-bundle.crt,readonly" \
  golang:1.27.3-alpine3.23 go run /smoke.go
```

Expected: valid RS256 200 on both routes with P001/Alex Taylor data and no GraphQL errors; tampered, malformed and missing tokens 401 on each route; every response has a request ID. The client intentionally uses trusted CA verification without `InsecureSkipVerify`. It additionally verifies that valid P001 tokens requesting P002/P999 return generic GraphQL `FORBIDDEN` and null participant data despite a spoofed identity header. It does not test expired/wrong-audience tokens or rotation; the Go suite covers claim rejection, missing identity/mapping, and no repository calls on denial. The original block passed the 28 October authentication run; the extended self-access block has separate 29 October results below.

</details>

## Kubernetes verification

On 28 October 2026, the [local Kubernetes runbook](kubernetes.md) was exercised using kind v0.33.0, Kubernetes v1.37.0 and kubectl v1.36.1 with a separate kubeconfig. Kustomize rendering and API-server dry runs passed, including restricted Pod Security admission. All three Deployments rolled out ready with zero initial restarts. The authentication Job reused the documented Go client: both valid participant queries returned 200 and expected data; tampered-signature, malformed and missing tokens returned 401 on each route. The Job's Go client verified the issuer CA/hostname; the newly deployed API began with an empty JWKS cache. Host port-forward probes returned 200 with request IDs; the temporary forward was stopped.

The first generated CA ConfigMap apply failed because client-side apply's annotation exceeded 262144 bytes with the existing bundle. Server-side apply resolved it without certificate changes; the runbook uses that approach. No application Go changes were made and the previously passing Go suite was not repeated for Kubernetes manifests. EKS, AWS IAM, ingress, network isolation, load/draining tests and resource authorisation remain unverified/unimplemented. The demo cluster remains running; cleanup instructions are in the runbook.

## Workload identity preparation

On 28 October 2026, local Kustomize rendering and server dry run passed with a dedicated `participant-api` service account; the API rollout succeeded. Verified the Deployment selects that account, default API-token mounting remains disabled, the default token file is absent, and an impersonated Kubernetes `get secrets` request returns `no` with exit status 1. The in-cluster authentication Job then passed both valid participant queries and all six invalid/missing-token cases. This verifies local Kubernetes identity isolation without proving AWS IAM behaviour. Example IAM policies passed JSON parsing only; EKS Pod Identity, Secrets Manager, KMS and CSI delivery remain untested. See [identity/secrets](aws-identity-secrets.md) for commands and proposed scope.

## CI preparation evidence

On 28 October 2026, fresh `go test -count=1 -race ./...`, `go vet ./...`, formatting and module-file checks passed. Local equivalent Compose configuration, Kustomize rendering, policy JSON syntax, documentation checks and all three Docker target builds passed. Actionlint v1.7.7 accepted the new workflow; its optional ShellCheck integration was disabled and shell blocks were syntax-checked separately. A disposable negative Markdown fixture confirmed the checker rejects missing targets, open details and unbalanced sections, then was removed. See [CI](ci.md) for commands and limits. No workflow was committed/pushed or run on GitHub; local macOS/engine results do not prove hosted Ubuntu/AMD64 execution.

## First hosted CI result

On 28 October 2026, GitHub Actions [run 40651138381](https://github.com/shawfire/participant-services/actions/runs/40651138381) for commit `bd91def` completed successfully: documentation/configuration (9 seconds), Go tests/vet (1 minute 36 seconds), and all service image builds (41 seconds). Observed race tests, vet, module-file checks, checksum-verified kubectl installation, Compose validation and Kustomize rendering all passed on hosted Ubuntu/AMD64. Earlier local-only/hosted-unverified statements above describe their historical stage; this run supplies the later hosted evidence. It does not add hosted runtime-stack or AWS verification.

## Participant self-access verification — 29 October 2026

**Observed:** focused race tests passed, followed by the complete `go test -count=1 -race ./...` suite and `go vet ./...`. New service tests deny absent/empty identity, missing mapping, cross-participant/unknown/empty requests before repository invocation, on both memory and injected-repository paths. A mismatched upstream record ID is rejected. The HTTP JWT test now traverses real participant resolvers and mock MuleSoft: signed mappings for P001 and P002 access their own records; cross-participant, unknown-other and missing-mapping requests return generic `FORBIDDEN` without upstream calls. A modified mapping without a new signature returns 401; spoofed identity headers do not grant access. Existing cancellation, not-found and concurrent-read tests pass with explicit trusted fixtures.

The local issuer now signs `participant_id=P001`. Rebuilt Compose passed validation/startup with existing certificates, and the exact updated smoke block passed twelve cases: own reads 200, six invalid/missing-token rejections in total, and P002/P999 forbidden responses on both routes with null participant data. Loaded the same images into kind, restarted/waited for all three Deployments, updated the smoke-client ConfigMap and ran a fresh disposable Job; all twelve cases passed there too. Tokens/keys were not displayed. Generic GraphQL denials use HTTP 200 plus `errors[].extensions.code=FORBIDDEN`; they are not HTTP 401 authentication failures.

[Authorisation](authorisation.md) owns the policy and claim-provisioning contract. New code remains uncommitted and has not run on hosted CI. Staff/delegate rules, entitlement revocation/provisioning, production identity assurance and the existing issuer health-probe TLS bypass remain outside this implementation. No AWS resources changed.
