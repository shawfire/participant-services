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

The API fails startup when these values are absent. It does not issue tokens or bundle an identity provider. Obtain an RS256 access token from your configured provider with matching issuer/audience, an expiry, a nonempty subject and a `kid` matching its JWKS. The test suite generates temporary keys and tokens locally; see [testing](testing.md).

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
