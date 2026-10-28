# Docker and Docker Compose

[Documentation index](README.md) · [Development journal](development-journal.md)

**Implemented in the uncommitted working tree — 28 October 2026:** API, local HTTPS issuer and mock MuleSoft images. At handover all three services were healthy, but Go JWT/JWKS authentication was still awaiting successful verification. The approved repeat verification subsequently passed against rebuilt images, including a fresh API JWKS cache. See [testing](testing.md#docker-authentication-evidence) for evidence and limits.

## First-time local setup

Follow steps 1–6 in order. Run commands from the repository root in a macOS/Linux Bash or Zsh terminal. Docker runs the Go toolchain, so host Go is optional unless you run Go tests directly. No external identity provider, AWS account, host DNS entry or system-wide CA installation is needed for this local stack.

### 1. Check prerequisites and checkout

Install Git, Docker Desktop (or Docker Engine plus Compose on Linux), OpenSSL 3 and curl. Start Docker. On macOS, OpenSSL 3 is available with `brew install openssl@3`; if `openssl version` reports LibreSSL, run `export PATH="$(brew --prefix openssl@3)/bin:$PATH"` in this terminal. Linux users can install these tools through their distribution's package manager.

```bash
git clone https://github.com/shawfire/participant-services.git
cd participant-services
git status --short
docker info
docker compose version
openssl version
curl --version
```

If already in this checkout, skip cloning. A private repository requires GitHub access. The Docker/issuer files are currently uncommitted, so a remote clone will not include them yet: obtain the reviewed working files or use a revision containing them before continuing. Confirm `Dockerfile`, `compose.yaml`, `compose.local.yaml` and `cmd/localissuer/main.go` exist. Docker/Compose must support `up --wait` and health-conditioned dependencies. Port 8080 must be free.

### 2. Create local configuration without overwriting existing work

```bash
mkdir -p config
if [ ! -e config/local.env ]; then
  (umask 077; cat > config/local.env <<'ENV'
APP_ENV=local
MULESOFT_URL=http://mockmulesoft:8081
JWT_JWKS_URL=https://local-issuer:8443/.well-known/jwks.json
JWT_ISSUER=https://local-issuer:8443
JWT_AUDIENCE=participant-services
LOG_LEVEL=debug
ENV
  )
else
  echo 'Preserving existing config/local.env; check it against the values above.'
fi
```

These are local settings, not secrets. Keep existing files intact and edit discrepancies deliberately. Use `config/local.env` for this three-service walkthrough. Root `.env.example` describes an external provider/upstream and now includes every required API setting. Dev/test/prod files contain placeholder URLs. `APP_ENV` and `LOG_LEVEL` currently have no effect in Go code.

### 3. Create the local CA and issuer certificate

This block refuses to replace any existing `config/certs` path. For an existing setup, skip generation and run step 4. Incomplete or expired setups need the renewal procedure below.

```bash
(
  set -eu
  if [ -e config/certs ]; then
    echo 'Preserving config/certs. Skip generation and validate it in step 4.'
    exit 0
  fi
  umask 077
  cert_stage=$(mktemp -d /tmp/participant-certs.XXXXXX)
  echo "Generating local certificates in $cert_stage"
  cat > "$cert_stage/ca.cnf" <<'CONF'
[req]
distinguished_name = dn
x509_extensions = ca
prompt = no
[dn]
CN = Participant Services Local Development CA
[ca]
basicConstraints = critical,CA:TRUE
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
CONF
  cat > "$cert_stage/local-issuer.ext" <<'CONF'
basicConstraints = critical,CA:FALSE
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = DNS:local-issuer
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid,issuer
CONF
  openssl genrsa -out "$cert_stage/local-ca.key" 3072
  openssl req -new -x509 -sha256 -days 365 \
    -key "$cert_stage/local-ca.key" -out "$cert_stage/local-ca.crt" \
    -config "$cert_stage/ca.cnf"
  openssl genrsa -out "$cert_stage/local-issuer.key" 2048
  openssl req -new -sha256 -subj '/CN=local-issuer' \
    -key "$cert_stage/local-issuer.key" -out "$cert_stage/local-issuer.csr"
  openssl x509 -req -sha256 -days 30 \
    -in "$cert_stage/local-issuer.csr" \
    -CA "$cert_stage/local-ca.crt" -CAkey "$cert_stage/local-ca.key" \
    -CAcreateserial -extfile "$cert_stage/local-issuer.ext" \
    -out "$cert_stage/local-issuer.crt"
  cp "$cert_stage/local-ca.crt" "$cert_stage/ca-bundle.crt"
  chmod 700 "$cert_stage"
  chmod 600 "$cert_stage/local-ca.key"
  chmod 644 "$cert_stage/local-issuer.key" "$cert_stage/local-issuer.crt" \
    "$cert_stage/local-ca.crt" "$cert_stage/ca-bundle.crt"
  mv "$cert_stage" config/certs
)
```

The CA private key stays mode 600 and is never mounted. The local issuer key is mode 644 **inside a mode-700 host directory** so container UID 65532 can read the bind-mounted file; do not copy it into a shared directory. This is a local-development compromise. Certificate mounts are read-only. Neither keys nor certificates need to be printed or committed. The bundle contains this local CA; the walkthrough uses only local HTTPS.

These commands use OpenSSL's [request configuration](https://docs.openssl.org/3.0/man1/openssl-req/) and [certificate signing extensions](https://docs.openssl.org/3.0/man1/openssl-x509/). A failed generation leaves the temporary directory for diagnosis; do not install a partial certificate set.

### 4. Validate certificates and configure Compose

```bash
openssl verify -CAfile config/certs/ca-bundle.crt \
  -purpose sslserver -verify_hostname local-issuer config/certs/local-issuer.crt
openssl x509 -in config/certs/local-issuer.crt -noout -dates -ext subjectAltName
openssl x509 -in config/certs/local-issuer.crt -noout -checkend 86400
git check-ignore config/certs/local-ca.key config/certs/local-issuer.key

dc() {
  env -u JWT_JWKS_URL -u JWT_ISSUER -u JWT_AUDIENCE -u MULESOFT_URL -u APP_ENV -u LOG_LEVEL \
    docker compose --project-name participant-services --env-file config/local.env \
    -f compose.yaml -f compose.local.yaml "$@"
}
dc config --quiet
```

Expected: certificate verification reports `OK`, the SAN contains `DNS:local-issuer`, the expiry check succeeds, both key paths are ignored, and Compose validation exits successfully. The `dc` function fixes the project/network name and prevents inherited shell values overriding the local file. Define it again in each new terminal. It does not print environment contents.

### 5. Start and verify the stack

```bash
dc up -d --build --wait
dc ps
curl --fail-with-body -i http://127.0.0.1:8080/healthz
curl --fail-with-body -i http://127.0.0.1:8080/readyz
dc exec -T api sh -c 'test "$SSL_CERT_FILE" = /certs/ca-bundle.crt && test -r "$SSL_CERT_FILE"'
dc exec -T localissuer sh -c 'test -r /certs/local-issuer.crt && test -r /certs/local-issuer.key'
```

Expected: three healthy services, two HTTP 200 probe responses, and successful readability checks. Next run the [authentication smoke check](testing.md#local-authentication-smoke-check): it obtains a token without displaying it, checks both participant queries through mock MuleSoft and verifies invalid-token rejection. **Healthy containers are not proof of successful authentication.**

To inspect failures, run `dc logs --tail=50 api localissuer mockmulesoft` without enabling token/body logging. Stop an existing host API if Docker reports port 8080 is occupied. For bind-mount errors, confirm all certificate paths are regular files and Docker Desktop allows access to the checkout. If issuer startup reports permission denied, review step 3's directory/key permissions. Do not use `sudo` or disable TLS to conceal a trust error.

### 6. Run checks and stop

```bash
# Optional host checks: requires Go 1.27.3 as declared in go.mod.
go test -count=1 -race ./...
go vet ./...
# Shut down only this project's containers and network; local files remain.
dc down
```

Without host Go, run tests using the builder image (on ARM64/AMD64 Linux containers):

```bash
docker run --rm --mount "type=bind,src=$(pwd),dst=/src,readonly" -w /src \
  golang:1.27.3-bookworm sh -c 'go test -count=1 -race ./... && go vet ./...'
```

This separate Debian builder has the C compiler needed by the race detector. Internet access is required for initial images/modules. Tests use synthetic data; no external provider or running Compose stack is required.

<details>
<summary>Certificate renewal and reuse</summary>

Issuer certificates last 30 days; the generated local CA lasts 365 days. Check expiry with step 4. To renew this local fixture without overwriting keys: stop the stack with `dc down`, move the existing certificate directory aside, then rerun steps 3–5 and the authentication test:

```bash
# With dc defined and run from the repository root:
dc down
(
  set -eu
  cert_archive=$(mktemp -d /tmp/participant-certs-backup.XXXXXX)
  chmod 700 "$cert_archive"
  mv config/certs "$cert_archive/certs"
  echo "Previous certificates preserved at $cert_archive/certs"
)
```

The backup is outside Git; keep it private. Step 3 generates a new CA and signing key. Recreate containers, obtain fresh tokens and rerun checks; old tokens no longer match the new JWKS. Do not install this local CA into your operating system's trust store. If generation fails, preserve the failed staging directory and restore the backup deliberately before starting. This renews the complete local fixture; it is not production key rotation.

</details>

<details>
<summary>Image and Compose architecture</summary>

[Dockerfile](../Dockerfile) builds three static executables. Runtime images use Alpine, CA certificates and UID/GID 65532. [compose.yaml](../compose.yaml) is API-only; [compose.local.yaml](../compose.local.yaml) adds issuer/mock, dependency health checks and API CA mounts. Base tags are versioned, not digest-pinned; scanning and multi-platform publication remain unverified. Local provisioning does not change API JWT or TLS validation.

</details>

<details>
<summary>DNS, CA trust and security settings</summary>

The API is published only at `127.0.0.1:8080`. The issuer and mock expose no host ports. The API uses `http://mockmulesoft:8081`; its configured HTTPS JWKS URL is `https://local-issuer:8443/.well-known/jwks.json`. The service is named `localissuer`, so the local overlay supplies network alias `local-issuer` to match the URL, token issuer and certificate SAN. Container `localhost` means that container, not another service or the Mac.

The local overlay mounts the bundle read-only into the API at `/certs/ca-bundle.crt` and sets `SSL_CERT_FILE=/certs/ca-bundle.crt`. Changing Compose does not prove an existing container received the new setting: inspect/recreate it during approved verification. A readable file and successful DNS are prerequisites, not proof that Go has accepted the certificate. BusyBox `wget` and Go can use different TLS implementations and trust configuration.

All three services use read-only root filesystems, dropped capabilities and no-new-privileges. Issuer certificate and private key mounts are read-only. `.dockerignore` excludes `config/`, backup files, private-key patterns and `.env` patterns. Runtime configuration and local certificate material stay outside the build context; read-only bind mounts still supply certificates at runtime. No secrets are reproduced here.

The issuer is a local test fixture with an unauthenticated token endpoint; it must not become a production identity service. Participant-level authorisation is absent. API health/readiness routes remain public.

</details>

<details>
<summary>Health checks, shutdown and troubleshooting</summary>

The API checks `/readyz`; mock health checks synthetic participant P001. The overlay requires both dependencies to satisfy `condition: service_healthy`. In the preserved source configuration, issuer health uses BusyBox `wget` with `--no-check-certificate` against loopback. This existing bypass applies to the health probe, not the API Go client, and cannot prove TLS trust. Replacing it with a certificate-validating probe remains proposed; successful Go/API checks do not remove this health-probe limitation.

The API has 15 seconds to stop against its 10-second shutdown deadline. Mock/issuer have simple HTTP lifecycles; Kubernetes termination and traffic draining remain unverified. See [health/lifecycle](health-lifecycle.md).

| Symptom | Diagnosis and correction recorded in the handover |
|---|---|
| First authenticated Portal request returns 401 | Generic rejection alone does not identify a signature, claim, DNS or TLS cause; follow the chain below |
| `--ca-certificate` unsupported | BusyBox feature limitation; use a Go TLS client for authoritative verification |
| `local-issuer` cannot resolve | Service name was `localissuer`; add the matching Docker network alias |
| Certificate verification failure after DNS works | Check SAN, validity, CA bundle and running API environment; do not disable API TLS validation |
| API missing `SSL_CERT_FILE` | Set it in the local overlay, mount the bundle, and verify the running container receives both |
| Missing `mockmulesoft` service definition | Restore the service in the local overlay and validate the merged files |
| Missing issuer health check with `service_healthy` | Restore issuer health check and hardening before retrying startup |

The full chronological sequence, evidence sources and remaining uncertainty are preserved in the [journal](development-journal.md). A historical Compose backup is retained as user work; use the current two-file configuration for maintained instructions. Test commands and outcomes belong in [testing](testing.md).

</details>
