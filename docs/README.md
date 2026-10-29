# Documentation index

[Project and quick start](../README.md) · [Architecture overview](../DESIGN.md)

**Start here:** [First-time local setup](containers.md#first-time-local-setup), then the [authentication smoke check](testing.md#local-authentication-smoke-check).

| Document | Authoritative scope |
|---|---|
| [Authentication](authentication.md) | JWT configuration, verification, JWKS caching and security boundaries |
| [Participant authorisation](authorisation.md) | Self-access policy, trusted mapping, error behaviour and remaining scope |
| [GraphQL](graphql.md) | Consumer contracts, queries, generation and resolver behaviour |
| [Integration](integration.md) | Repository adapter, mock MuleSoft, failure mapping and retries |
| [Observability](observability.md) | JSON logs, request IDs and propagation |
| [Health and lifecycle](health-lifecycle.md) | Probes, HTTP timeouts and graceful shutdown |
| [Testing](testing.md) | Verification commands, coverage and evidence limits |
| [Containers](containers.md) | Three-service Compose setup, DNS/CA trust, health checks and troubleshooting |
| [Kubernetes demonstration](kubernetes.md) | Isolated kind setup, manifests, authentication Job and EKS mapping |
| [AWS identity and secrets](aws-identity-secrets.md) | Verified local service account and proposed EKS/managed-secret design |
| [Continuous integration](ci.md) | Validation/image-build workflow, local checks and hosted-execution limits |
| [Roadmap](roadmap.md) | Deployment priorities and proposed extensions |
| [Development journal](development-journal.md) | Primary continuous audit trail for every development step |
| [Documentation change log](documentation-changelog.md) | Supporting documentation changes, content movement and verification |
| [Historical snapshots](history/README.md) | Complete original README and DESIGN |

## Status conventions

**Implemented** means code exists in the working tree; it does not imply a commit, deployment or production assurance. **Verified** names a completed check and its evidence. **Proposed** describes a design awaiting adoption. **Planned** describes future work with no implementation claim.

Current references describe the working tree inspected on 28 October 2026, including authentication committed as `4a4bec9`. The journal and snapshots preserve historical statements even where later work supersedes them. Keep new journal entries chronological, and update the authoritative reference when behaviour changes. Keep detailed sections collapsed using `<details>` without an `open` attribute.

For every substantive development step, follow the [audit maintenance rules](history/README.md#maintenance-rules) and append a dated entry to the [development journal](development-journal.md). Record documentation-specific movement and corrections in the supporting [change log](documentation-changelog.md), linked from the journal. This process continues across all future milestones.
