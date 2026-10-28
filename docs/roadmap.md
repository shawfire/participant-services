# Roadmap and proposed extensions

[Documentation index](README.md) · [Architecture](../DESIGN.md)

**Planned, not deployed — 28 October 2026.** Next priorities after documentation:

1. Docker image and Docker Compose for the API and mock upstream, including explicit identity-provider configuration.
2. Kubernetes manifests and EKS exercises: Deployment, Service, ingress, requests/limits, probes, rolling updates, autoscaling and verified traffic draining.
3. AWS workload IAM and managed secrets; avoid static keys, restrict networking and database access.
4. CI/CD: automated tests, race checks, vet, schema checks, scanning, image build and controlled deployment/rollback.

Participant-level authorisation remains unimplemented and must be resolved before exposing participant data beyond this local synthetic lab. PostgreSQL, query limits, N+1/DataLoader exercises, circuit breaking, metrics, distributed tracing and production log aggregation remain future work.

<details>
<summary>Proposed persistence and submission flow</summary>

Candidate PostgreSQL entities: `participants`, `plans`, `submissions`, `submission_events`; columns and constraints remain to be designed. Use migrations, indexed access, parameterised queries and explicit transactions.

A future read flow adds trusted scope/ownership checks in the shared service, complexity and pagination limits, request-scoped batching and field policy. A future submission flow validates identity/input, persists an idempotency key with a unique constraint, calls the adapter with a timeout/correlation ID, retains an unambiguous pending state on uncertain failures and reconciles outcomes without duplicates. Return a stable submission ID/status. These flows are proposals; the current schemas expose reads only.

A fuller mock integration could simulate Salesforce, latency, 429/500/503, connection failures and duplicate requests. Keep integration DTOs separate and use synthetic records.

</details>

<details>
<summary>Proposed AWS topology and operational exercises</summary>

```text
Client -> controlled edge / ingress -> EKS Go API pods
                                     |-> managed PostgreSQL (RDS)
                                     |-> mock integration service
                                     |-> CloudWatch / telemetry backend
```

ECR/EKS deployment is not performed. Cloudflare Zero Trust/Access/Tunnels is an optional study topic, not an existing component. Budget and cost review should guide EKS exercises. Define latency/error SLOs, incident diagnosis and rollback drills as the platform develops.

</details>
