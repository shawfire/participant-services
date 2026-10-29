# Participant authorisation

[Documentation index](README.md) · [Authentication](authentication.md) · [Testing](testing.md)

**Implemented and verified locally on the review branch — 29 October 2026:** participant self-access in the shared service. Go race tests/vet, Docker and the in-cluster Kubernetes smoke Job passed; implementation commit `e5c748f` is prepared for review; hosted CI results belong to the pull request. Authentication and authorisation are separate: verified identity alone does not grant arbitrary participant access. Staff/delegate access, dynamic entitlements and production identity provisioning remain unimplemented.

<details>
<summary>Policy and trusted mapping</summary>

The trusted issuer supplies a signed string claim `participant_id`. JWT `sub` remains the caller identity and is not interpreted as a record ID. After successful RS256/signature/registered-claim checks, JWT middleware copies the subject and mapping into an `internal/identity.Principal` value in request context. Incoming headers, GraphQL arguments and unverified JWT contents cannot establish that principal.

`participant.Service.Get` requires a nonempty subject and mapping, then compares the requested ID exactly with the trusted mapping before reading memory or calling the repository. Missing identity/mapping, empty requested IDs and mismatched IDs are denied. The service also rejects a repository response containing a different participant ID, preventing a faulty upstream from exposing another record. Request cancellation remains honoured.

The issuer must own and protect the identity-to-participant mapping. A valid signature proves the issuer signed the mapping, not that the organisation provisioned it correctly. Revocation/freshness depends on token lifetime and issuer controls; this prototype has no entitlement database or revocation service. Internal adapters may construct identity context after establishing trust; it is not an externally writable session mechanism.

</details>

<details>
<summary>HTTP and GraphQL outcomes</summary>

| Case | Outcome | Repository access |
|---|---|---|
| Missing/invalid JWT | HTTP 401 before GraphQL | None |
| Valid JWT, missing `participant_id` | Participant field `null`, generic GraphQL `FORBIDDEN` | None |
| Mapped participant requests own existing record | HTTP 200 with expected data | Allowed |
| Mapped participant requests another or unknown other ID | HTTP 200, participant field `null`, error message `forbidden`, code `FORBIDDEN` | None |
| Caller requests own mapped but nonexistent record | Nullable `null`, preserving existing not-found behaviour | Allowed |
| Upstream returns a different ID | Generic GraphQL `FORBIDDEN`; no mismatched data returned | Called, then response rejected |

The `FORBIDDEN` mapping is identical in Portal and Mobile. Requests can receive HTTP 200 with GraphQL errors, so clients must inspect `errors`. `__typename` and other operations that do not read participants do not exercise this service policy. The same rule applies to every participant field in a multi-field operation.

</details>

<details>
<summary>Local fixture and compatibility</summary>

The development issuer signs `sub=participant-123` and `participant_id=P001`; clients cannot select a mapping at its token endpoint. It remains an unauthenticated local fixture, not a production IdP. Rebuild both API and issuer and obtain a fresh token when adopting this change; previously issued tokens without the mapping still authenticate but cannot read participant data.

External providers must supply the trusted claim (or a future explicitly designed server-side mapping adapter). Do not substitute a request header or derive entitlement from arbitrary caller input. This policy deliberately excludes staff, guardians/delegates, multiple-participant access and field-level permissions until separately designed and tested.

Use the [container setup](containers.md) and [token-safe smoke check](testing.md#local-authentication-smoke-check). Run `go test -count=1 -race ./...` and `go vet ./...` for unit/HTTP coverage. Current verification evidence belongs in [testing](testing.md); historical pre-policy results remain in the chronological journal.

</details>
