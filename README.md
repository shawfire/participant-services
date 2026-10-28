# Participant Services

A synthetic Go/GraphQL middleware lab for the NDIA EL2 Lead/Principal Developer interview (RFQ LH-07533). This independent learning project is not an NDIA system or a claim of NDIA production experience.

Portal and Mobile expose separate GraphQL contracts over a shared participant service and mock MuleSoft HTTP adapter. The working tree implements JWT authentication, bounded retries, JSON logging, request correlation, health/readiness probes and graceful shutdown.

**Status — 28 October 2026:** authentication is implemented and committed as `4a4bec9`. Participant-level authorisation is not implemented. Docker/Compose packaging is implemented; see the journal for verification. A local [Kubernetes demonstration](docs/kubernetes.md) is implemented and verified. [CI validation and image builds](docs/ci.md) are configured but have not run on GitHub yet. AWS EKS deployment, IAM/secrets integration and CD remain planned. See [testing](docs/testing.md) for verification evidence and limits.

## Quick start

**New developers:** follow the [first-time local setup](docs/containers.md#first-time-local-setup) for prerequisites, certificates, configuration, startup, authentication tests and shutdown. Host Go is optional for that path.

The commands below are an alternative for running Go directly with an existing identity provider.

Use Go 1.27.3 as declared in `go.mod`. Run the mock upstream in one terminal:

```bash
go run ./cmd/mockmulesoft
```

In a second terminal, configure your trusted identity provider and start the API. Replace the example values with an actual HTTPS JWKS endpoint, issuer and audience:

```bash
export JWT_JWKS_URL='https://your-identity-provider.example/.well-known/jwks.json'
export JWT_ISSUER='https://your-identity-provider.example/'
export JWT_AUDIENCE='participant-services'
export MULESOFT_URL='http://localhost:8081'
go run ./cmd/api
```

The example identity-provider URL is a placeholder. The local Compose overlay now provides a development issuer; the direct-Go example above still requires a reachable trusted provider. See the [authentication setup](docs/authentication.md) for requirements. Missing JWT configuration prevents API startup. Once running:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

See [authenticated GraphQL queries](docs/graphql.md) for Portal and Mobile requests. Stop with `Ctrl+C`. For an automated local check requiring no external identity provider:

```bash
go test -count=1 -race ./...
go vet ./...
```

## Documentation

Start with the [documentation index](docs/README.md) or [architecture overview](DESIGN.md). The [development journal](docs/development-journal.md) is the continuous audit trail for every development step, preserving commands, failures, decisions and verification results from initial setup onward. Follow the [maintenance rules](docs/history/README.md#maintenance-rules) as work continues. Complete pre-refactor documents are in [historical snapshots](docs/history/README.md). Next priorities are recorded in the [roadmap](docs/roadmap.md).

Use synthetic participant data only. Never commit credentials, tokens, private keys, real participant information or secret-bearing `.env` files. Keep this prototype local while resource authorisation remains absent. Interview examples should distinguish lab exercises from professional production experience.
