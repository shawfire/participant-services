# Project TODO

A concise delivery checklist. Checked milestones reflect recorded evidence; unchecked items are planned work, not commitments or production-readiness claims.

## Completed milestones

- [x] Separate Portal and Mobile GraphQL APIs over a shared participant service.
- [x] RS256 JWT validation with HTTPS JWKS, caching and rotation handling.
- [x] Signed participant self-access with cross-participant denial in the shared service.
- [x] Mock MuleSoft integration with timeouts, bounded retries and cancellation.
- [x] Structured logging, request correlation, health/readiness probes and graceful shutdown.
- [x] Local Docker Compose and kind Kubernetes demonstrations with authenticated API checks.
- [x] Hosted CI passing Go race tests, vet, documentation/configuration checks and service image builds at `6ca994e`.
- [x] Modular documentation, chronological development audit and AI-assisted engineering guide.
- [x] Public source release with MIT/third-party notices and enabled private vulnerability reporting.

Implementation and verification evidence is recorded in [testing](docs/testing.md) and the [development journal](docs/development-journal.md). Public access and reporting enablement were verified on 30 October 2026. Local Kubernetes verification does not imply AWS deployment.

## Current priorities

- [ ] Record and publish Video 1: fresh checkout and local setup.
- [ ] Record and publish Video 2: features and security behaviour.
- [ ] Record and publish Video 3: architecture, documentation and AI-assisted engineering.
- [ ] Introduce PostgreSQL persistence.
- [ ] Demonstrate AWS RDS integration.
- [ ] Build a React Native mobile application using the protected Mobile GraphQL API.
- [ ] Build a React web portal using the protected Portal GraphQL API.

No videos are recorded or published. Video 1's local fresh-checkout rehearsal passed; this does not claim a pristine-machine setup. The recording pack and rehearsal evidence remain separate local drafts.

## Future enhancements

- [ ] Add staff/delegate authorisation and entitlement provisioning/revocation.
- [ ] Add GraphQL query/rate controls, pagination and request-scoped batching where needed.
- [ ] Add submission workflows with idempotency and reconciliation.
- [ ] Add circuit breaking, metrics, distributed tracing and production log aggregation.
- [ ] Add automated runtime-stack CI checks, dependency/image vulnerability scanning and controlled image publication.
- [ ] Demonstrate EKS deployment, managed identity/secrets, ingress, autoscaling and verified traffic draining/rollback.

## Supporting documentation

- [Documentation index](docs/README.md) and [local setup](docs/containers.md#first-time-local-setup).
- [Detailed architectural roadmap](docs/roadmap.md) and [architecture](DESIGN.md).
- [Authentication](docs/authentication.md), [authorisation](docs/authorisation.md) and [security guidance](SECURITY.md).
- [Testing evidence](docs/testing.md), [development journal](docs/development-journal.md) and [documentation changelog](docs/documentation-changelog.md).
- [AI-assisted engineering](docs/ai-assisted-engineering.md).
