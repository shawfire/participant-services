# Observability

[Documentation index](README.md) · [Architecture](../DESIGN.md)

<details>
<summary>Logs and correlation — implemented</summary>

### Structured logging

The API uses Go's standard `log/slog` package with a JSON handler.

Structured lifecycle logs include startup, SIGTERM receipt, shutdown initiation and shutdown completion.

JSON logs are written to standard output, making them suitable for collection by a container logging agent in a future EKS deployment.

### HTTP request correlation

The `internal/observability` package provides two HTTP middleware components:

- `RequestIDMiddleware` generates a random 128-bit request ID.
- `AccessLogMiddleware` records the request ID, HTTP method, path, response status and duration in milliseconds.

The middleware is registered in this order:

1. Request ID middleware
2. Access logging middleware
3. HTTP router
4. JWT middleware, then Portal or Mobile GraphQL handler

The request ID is returned in the `X-Request-ID` response header and stored in `context.Context`.

### MuleSoft propagation

The MuleSoft HTTP client reads the request ID from the Go context and adds it to outgoing requests using the `X-Request-ID` header.

This allows requests to be correlated across the API and mock MuleSoft boundary.

Retry attempts use the same request context and therefore retain the same request ID.

### Verification

See [testing](testing.md) for coverage, recorded results and evidence limits. Historical outcomes remain in the [development journal](development-journal.md).

### Production considerations

The following capabilities are not yet implemented or verified:

- Distributed tracing using OpenTelemetry.
- Prometheus-compatible application metrics.
- Centralised log aggregation and alerting.
- Kubernetes deployment and production monitoring.
- Full GraphQL request-to-MuleSoft integration-level correlation tests.

The current access-log response wrapper does not preserve optional HTTP interfaces such as `http.Flusher` and `http.Hijacker`. It must be improved before enabling streaming or WebSocket transports.

Logging must exclude participant personal information, credentials, tokens and sensitive GraphQL request bodies.

</details>

See [testing](testing.md) for verification evidence.
