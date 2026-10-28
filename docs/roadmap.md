# Roadmap and proposed extensions

[Documentation index](README.md) · [Architecture](../DESIGN.md)

**Status — 28 October 2026:** local [Docker/Compose packaging](containers.md) is implemented; see the journal for verification. Cloud deployment is not performed. Delivery sequence:

1. Docker images and local Compose for API, issuer and mock are implemented in uncommitted files. Documentation reconciled and approved repeat verification passed: Go HTTPS JWKS retrieval, both authenticated participant queries and invalid-token rejection, including a fresh API cache. The issuer health-probe bypass remains a known limitation.
2. [Minimal local Kubernetes demonstration](kubernetes.md) implemented and verified on kind: three Deployments/ClusterIP Services, configuration/Secret mounts, resources, security settings, probes and an in-cluster authentication Job. EKS deployment, ingress, autoscaling and verified traffic draining remain planned.
3. [AWS identity/secrets preparation](aws-identity-secrets.md): dedicated unprivileged API service account implemented locally; EKS Pod Identity and exact-secret policy examples prepared. Cloud IAM/CSI integration remains proposed because the API currently has no AWS consumer. Do not add static keys or unused permissions.
4. [CI validation and image builds](ci.md) implemented in an uncommitted workflow with local equivalent checks passing; hosted execution remains unverified. Automated stack tests, schema checks, scanning, publishing and controlled deployment/rollback remain planned.

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
