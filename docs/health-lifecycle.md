# Health and HTTP lifecycle

[Documentation index](README.md) · [Architecture](../DESIGN.md)

<details>
<summary>HTTP lifecycle — implemented</summary>

### Server configuration

The Participant Services API uses Go's `net/http` server with explicit timeouts:

| Setting | Value | Purpose |
|---|---|---|
| ReadHeaderTimeout | 5 seconds | Limits time spent receiving HTTP headers |
| ReadTimeout | 10 seconds | Limits time spent reading the complete request |
| WriteTimeout | 15 seconds | Bounds response writing time |
| IdleTimeout | 60 seconds | Limits idle keep-alive connections |
| Shutdown timeout | 10 seconds | Bounds graceful shutdown |

These settings provide baseline protection against stalled connections and unbounded shutdown.

### Liveness and readiness

The API exposes two separate health endpoints:

- `GET /healthz` returns HTTP 200 while the HTTP server is operating.
- `GET /readyz` returns HTTP 200 when the application is ready and HTTP 503 when readiness is disabled.

Readiness is implemented in `internal/health` using `atomic.Bool`, allowing safe concurrent access from HTTP handlers and the shutdown goroutine.

The current readiness check reflects application lifecycle state. It does not actively test MuleSoft or other external dependencies.

### Graceful shutdown

The API listens for SIGINT and SIGTERM using `signal.NotifyContext`.

On receiving a shutdown signal:

1. The application marks itself not ready.
2. It creates a shutdown context with a 10-second deadline.
3. It calls `http.Server.Shutdown`.
4. The HTTP server stops accepting new connections and waits for active requests to complete.
5. The application exits when shutdown finishes or reports an error if the deadline is exceeded.

The application does not currently include a pre-shutdown delay for Kubernetes endpoint propagation.

### Verification

See [testing](testing.md) for coverage, recorded results and evidence limits. Historical outcomes remain in the [development journal](development-journal.md).

### AWS EKS considerations

A future Kubernetes deployment should define:

- A readiness probe using `/readyz`.
- A liveness probe using `/healthz`.
- A suitable `terminationGracePeriodSeconds`.
- A traffic-draining strategy that accounts for Kubernetes Service and load-balancer propagation.
- Resource requests and limits.

The current implementation demonstrates application-level shutdown behaviour. End-to-end Kubernetes termination and traffic draining remain to be implemented and verified.

</details>

Both probes bypass JWT middleware after startup. Readiness does not establish JWKS availability or participant access. See [testing](testing.md) for verification evidence.
