# Authentication

[Documentation index](README.md) · [Architecture](../DESIGN.md)

**Implemented and committed as `4a4bec9`.** It was uncommitted at the initial documentation inspection. Participant-level authorisation remains unimplemented.

<details>
<summary>Configuration and local use</summary>

| Variable | Requirement |
|---|---|
| `JWT_JWKS_URL` | Trusted HTTPS JWKS endpoint; required |
| `JWT_ISSUER` | Expected issuer; required |
| `JWT_AUDIENCE` | Expected audience; required |

The API fails startup when these values are absent. The API does not issue tokens. The uncommitted local Compose overlay now includes a development issuer; external providers remain supported. Obtain an RS256 access token from your configured provider with matching issuer/audience, an expiry, a nonempty subject and a `kid` matching its JWKS. The test suite generates temporary keys and tokens locally; see [testing](testing.md).

Send exactly one `Authorization: Bearer <token>` header to either GraphQL POST endpoint. Playground pages remain public development tools; set their Headers panel to JSON containing the Authorization header. Probe requests require no token. Use the [quick start](../README.md#quick-start) and [query examples](graphql.md).

The runtime requires HTTPS. The fetcher allows HTTP only when explicitly constructed with the test-only flag; `cmd/api` passes false. There is no runtime environment switch to disable this requirement.

</details>

<details>
<summary>Verification and middleware behaviour</summary>

`RequireJWT` checks the Bearer header, delegates cryptographic verification and places trusted claims in `context.Context`; `ClaimsFromContext` retrieves them. Rejection returns HTTP 401, `WWW-Authenticate: Bearer`, and a generic `unauthorized` body before GraphQL executes.

The cached verifier accepts only RS256, requires a nonempty key ID and subject, verifies the RSA signature, checks configured issuer/audience, requires expiration, and validates issued-at when present. Standard registered-claim validation also checks not-before when supplied. Unverified token headers are used only to select a trusted public key, never to establish identity. Tokens must be nonempty and at most 8 KiB.

The package also contains earlier Bearer-only and fixed-key verifier building blocks. The runtime uses `RequireJWT` with `CachedJWKSVerifier`; presence of a Bearer header alone does not grant access.

</details>

<details>
<summary>JWKS retrieval, caching and rotation</summary>

JWKS fetching uses a three-second client timeout, rejects redirects and limits the response to 1 MiB. Endpoint configuration rejects missing hosts, embedded credentials and fragments. Keys are validated before entering the cache, including nonempty unique key IDs and valid RSA key material.

The API configures a five-minute cache TTL. Access is mutex-protected, with independent copies returned to callers. Expired-cache fetch failures reject authentication rather than serving stale keys. Invalid refreshes do not replace stored keys.

A genuinely unknown key ID may trigger an early refresh, limited to one attempt per 30 seconds. Concurrent requests can reuse an already-refreshed key set. An invalid signature using a known key does not trigger early refresh. This supports rotation while limiting unknown-key refresh pressure; it is not a guarantee of immediate acceptance for every rotation sequence.

</details>

<details>
<summary>Security boundary and remaining work</summary>

A valid token identifies the caller but does not restrict which participant ID they can query. Shared-service ownership/scope policy and negative cross-participant tests remain required. GraphQL depth/complexity limits, production identity-provider operations, metrics and audit policy are future work.

Never log tokens, private keys or participant personal information. Use synthetic records and keep this prototype local. Deployment planning is recorded in the [roadmap](roadmap.md).

</details>

Verification coverage and its limits are authoritative in [testing](testing.md).

<details>
<summary>Local HTTPS issuer — implemented fixture, limited assurance</summary>

[cmd/localissuer/main.go](../cmd/localissuer/main.go) loads an RSA TLS private key and uses the same key to sign development JWTs. `POST /dev/token` issues RS256 tokens with `kid=local-dev-key`, subject `participant-123`, audience `participant-services`, issuer `https://local-issuer:8443` and a 15-minute expiry. `GET /.well-known/jwks.json` publishes the matching RSA public key.

Using one key for TLS and signing and exposing an unauthenticated token endpoint are local-fixture shortcuts, not production identity design. No scopes or participant-ownership policy are enforced. A token subject is not proof of entitlement to P001.

The [first-time setup](containers.md#first-time-local-setup) owns CA/key/certificate creation, safe reuse, renewal, configuration, DNS alias, mounts and health-check details. Go's HTTPS client must validate certificate trust, validity and the configured hostname before the JWKS can be accepted. Successful BusyBox diagnostics, a mounted bundle or a healthy issuer alone are insufficient.

At handover, successful end-to-end JWT validation remained pending. A subsequent session check observed valid participant results and invalid-token rejection before the user requested documentation-first approval; see [testing evidence](testing.md#docker-authentication-evidence). The approved repeat verification now passed against rebuilt images and a recreated API with an empty JWKS cache. This confirms the Go/API trust and JWT path for the current local setup, while preserving the historical handover uncertainty. It does not verify production identity operations or resource authorisation.

</details>
