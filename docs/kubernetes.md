# Local Kubernetes demonstration

[Documentation index](README.md) · [Containers and certificates](containers.md) · [Development journal](development-journal.md)

This local kind exercise demonstrates Kubernetes concepts relevant to EKS. It does not provision AWS. Use synthetic data only: participant-level authorisation is still missing. **Verified locally — 28 October 2026:** kind v0.33.0 / Kubernetes v1.37.0, all three rollouts, in-cluster TLS/JWT participant checks and localhost probes passed. See [testing](testing.md#kubernetes-verification) and the journal for evidence and limits.

<details>
<summary>1. Prerequisites and isolated cluster</summary>

Complete [local certificate setup](containers.md#first-time-local-setup) first. Keep Docker running; the Compose stack can remain up because this demo uses local port 18080. Install `kubectl` and [kind](https://kind.sigs.k8s.io/docs/user/quick-start/) (the verification uses kind v0.33.0). Python 3 extracts the documented smoke client; the test itself runs Go inside Kubernetes, so host Go is not required.

On macOS with Homebrew:

```bash
brew install kind kubectl python
```

Run from the repository root. Use a separate kubeconfig so the demo cannot select your normal cluster:

```bash
export KUBECONFIG=/tmp/participant-services-kubeconfig
kind create cluster --name participant-services --kubeconfig "$KUBECONFIG" --wait 120s
kubectl config current-context
kubectl get nodes
```

Expected context: `kind-participant-services`. If the named cluster already exists, reuse its kubeconfig with `kind export kubeconfig --name participant-services --kubeconfig "$KUBECONFIG"`; do not delete an existing cluster as an installation step. All following kubectl commands assume this isolated kubeconfig. Network access is needed for initial image downloads. Allow Docker enough memory for a node plus workloads (4 GiB is a practical starting allocation, not a measured minimum).

</details>

<details>
<summary>2. Build and load local images</summary>

```bash
docker build --target api -t participant-services-api:local .
docker build --target localissuer -t participant-services-localissuer:local .
docker build --target mockmulesoft -t participant-services-mockmulesoft:local .
docker pull golang:1.27.3-alpine3.23
kind load docker-image --name participant-services \
  participant-services-api:local participant-services-localissuer:local \
  participant-services-mockmulesoft:local golang:1.27.3-alpine3.23
```

Manifests use `imagePullPolicy: Never` so local images must be loaded on every kind node. For rebuilds, reload images and run `kubectl -n participant-services rollout restart deployment/<name>`; changing a local image under the same tag does not automatically restart Pods. Production/EKS would use a registry such as ECR and immutable image references.

</details>

<details>
<summary>3. Supply configuration and deploy</summary>

The CA is public configuration; only the issuer's certificate and private key enter a Kubernetes Secret. Never store rendered Secret YAML or print the private key. The CA private key stays on the host. Generated objects use server-side apply to avoid duplicating a potentially large CA bundle (or secret data) in a last-applied annotation.

```bash
kubectl apply -f deploy/kubernetes/local/namespace.yaml
kubectl -n participant-services create secret generic local-issuer-tls \
  --from-file=local-issuer.crt=config/certs/local-issuer.crt \
  --from-file=local-issuer.key=config/certs/local-issuer.key \
  --dry-run=client -o yaml | kubectl apply --server-side --field-manager=participant-local -f -
kubectl -n participant-services create configmap local-ca \
  --from-file=ca-bundle.crt=config/certs/ca-bundle.crt \
  --dry-run=client -o yaml | kubectl apply --server-side --field-manager=participant-local -f -
kubectl apply --dry-run=server -k deploy/kubernetes/local
kubectl apply -k deploy/kubernetes/local
kubectl -n participant-services rollout status deployment/local-issuer --timeout=120s
kubectl -n participant-services rollout status deployment/mockmulesoft --timeout=120s
kubectl -n participant-services rollout status deployment/api --timeout=120s
kubectl -n participant-services get pods,services
```

The `local-issuer` Service preserves the certificate SAN/token issuer hostname; the Deployment is also named `local-issuer`. API configuration comes from `api-config`; CA trust is mounted at the same path used in Compose. The Secret's group-readable mount and Pod `fsGroup` allow UID 65532 to read the key without making it world-readable in the Pod.

Deployments reconcile independently; Kubernetes has no Compose `depends_on`. API readiness currently reflects lifecycle only, so it can be ready before its dependencies. Wait for all rollouts and perform the authentication check before declaring the demo functional. After rotating the fixture certificates, update the Secret/ConfigMap and restart API and issuer; obtain a fresh token. This is not a zero-downtime rotation procedure.

</details>

<details>
<summary>4. Check access and end-to-end authentication</summary>

In one terminal, with the isolated kubeconfig:

```bash
kubectl -n participant-services port-forward --address 127.0.0.1 service/api 18080:8080
```

In another:

```bash
curl --fail-with-body -i http://127.0.0.1:18080/healthz
curl --fail-with-body -i http://127.0.0.1:18080/readyz
```

For token-safe authentication checks, create a ConfigMap from the existing [documented Go smoke client](testing.md#local-authentication-smoke-check), then run the Job. This extracts the same client without running the Docker smoke block or displaying a token:

```bash
smoke_dir=$(mktemp -d /tmp/participant-k8s-smoke.XXXXXX)
python3 - "$smoke_dir/main.go" <<'PY'
from pathlib import Path
import sys
text = Path('docs/testing.md').read_text()
start = text.index("<<'GO'\n") + len("<<'GO'\n")
end = text.index('\nGO\n', start)
Path(sys.argv[1]).write_text(text[start:end] + '\n')
PY
kubectl -n participant-services create configmap smoke-client \
  --from-file="main.go=$smoke_dir/main.go" --dry-run=client -o yaml | kubectl apply --server-side --field-manager=participant-local -f -
# Delete only an earlier disposable test Job when repeating the check.
kubectl -n participant-services delete job auth-smoke --ignore-not-found
kubectl apply -f deploy/kubernetes/local/smoke-job.yaml
kubectl -n participant-services wait --for=condition=complete job/auth-smoke --timeout=180s
kubectl -n participant-services logs job/auth-smoke
```

Expected: valid RS256 participant queries return 200 with matching Portal/Mobile P001 data; tampered, malformed and missing tokens return 401 on each endpoint. The helper verifies issuer CA/hostname and compares GraphQL data without printing tokens. To repeat with an empty JWKS cache, restart API and wait for its rollout before rerunning the Job.

For failures use `kubectl -n participant-services get pods`, `describe pod <name>`, `logs deployment/<name>` or `logs job/auth-smoke`. `ErrImageNeverPull` means an image was not loaded. `CreateContainerConfigError`/`FailedMount` often points to a missing Secret or ConfigMap. A certificate error requires checking CA, SAN, expiry and mounted settings; do not disable API TLS verification.

</details>

<details>
<summary>5. Explain the design and its EKS equivalents</summary>

| Local object / behaviour | Interview explanation / EKS mapping |
|---|---|
| Namespace | Logical grouping with restricted Pod Security admission |
| Three Deployments, one replica each | Controllers maintain Pods; EKS runs the same workload API on managed Kubernetes |
| ClusterIP Services | Stable in-cluster DNS/routing, independent of Pod IP changes |
| Dedicated `participant-api` service account | No application RBAC grants or automatic API token; future EKS identity association is separate |
| API ConfigMap and CA ConfigMap | Non-secret runtime settings; changes do not automatically reload process environment |
| Issuer Secret | Local certificate/key injection; Kubernetes Secrets are not a managed secret store or automatically encrypted simply because they are base64-encoded |
| Non-root UID, read-only filesystem, dropped capabilities, seccomp | Least-privilege runtime settings; service-account token mounting disabled |
| CPU/memory requests and limits | Scheduling reservations and runtime bounds; illustrative values, not load-tested sizing |
| Startup/readiness/liveness probes | Startup gate, endpoint eligibility and restart detection; API readiness does not check upstream availability |
| API SIGTERM and 20-second termination grace | Application gets its 10-second shutdown budget; Kubernetes/load-balancer traffic draining still needs a verified strategy |
| kind image loading | Replace with ECR image distribution and controlled rollout in EKS |
| localhost port-forward | Local demonstration access; production ingress/load balancer, TLS and network policy are separate work |

The issuer uses TCP probes: they detect a listening port, not certificate trust or JWKS correctness. The end-to-end Job establishes TLS/JWT behaviour. [Probe semantics](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/) and [security context settings](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/) follow Kubernetes documentation.

See [AWS identity and secrets](aws-identity-secrets.md) for the proposed design and verified local service-account checks. No IAM workload roles, AWS Secrets Manager integration, ECR publication, ingress, network policy, autoscaling, multi-AZ resilience or production RBAC design is implemented. The local issuer is an unauthenticated synthetic fixture, not a production IdP. Participant-level authorisation remains missing. One replica is a demonstration, not high availability; no Kubernetes traffic-draining or load test is claimed.

</details>

<details>
<summary>6. Stop access and remove only the demo cluster</summary>

Stop port-forward with `Ctrl+C`. To remove the demo and its in-cluster secrets:

```bash
kind delete cluster --name participant-services
unset KUBECONFIG
```

This deletes the named local cluster only. It preserves host certificate files, Docker images and the Compose stack. Do not run it against a cluster containing other work. You can instead delete only the demo namespace with `kubectl delete namespace participant-services`, which removes its workloads, test Job, ConfigMaps and Secrets while preserving the cluster.

</details>
