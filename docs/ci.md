# Continuous integration

[Documentation index](README.md) · [Testing](testing.md) · [Development journal](development-journal.md)

**Implemented:** [.github/workflows/ci.yaml](../.github/workflows/ci.yaml). **Verified locally:** workflow syntax and equivalent Go/configuration/image-build checks. **Verified on GitHub:** [run 40651138381](https://github.com/shawfire/participant-services/actions/runs/40651138381) passed all three jobs for pushed commit `bd91def` on 28 October 2026. Publishing and deployment remain planned.

<details>
<summary>Triggers, jobs and permissions</summary>

The workflow runs for pull requests, pushes to `main` and manual dispatch. Concurrent runs for the same ref replace older runs. It uses GitHub-hosted Ubuntu 24.04 runners with job timeouts.

| Job | Checks |
|---|---|
| Go tests and vet | Go version from `go.mod`, formatting, uncached race suite, vet and unchanged module files |
| Documentation and configuration | Local Markdown paths/anchors and collapsed sections, IAM example JSON syntax, merged Compose configuration, Kubernetes Kustomize rendering and whitespace |
| Build service images | Builds `api`, `localissuer` and `mockmulesoft` targets after both validation jobs pass |

Repository permissions are `contents: read`; checkout does not persist Git credentials. Both actions are pinned to immutable commit hashes resolved from their official v6 tags. There is no AWS/OIDC permission, registry login, image push, cluster credential, secret value or deploy step. Setup-go uses the module version and dependency cache. See the official [workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax), [checkout](https://github.com/actions/checkout) and [setup-go](https://github.com/actions/setup-go) guidance.

A pinned kubectl v1.36.1 Linux binary is downloaded with its SHA-256 checksum using the [Kubernetes installation procedure](https://kubernetes.io/docs/tasks/tools/install-kubectl-linux/). CI renders manifests without accessing a cluster. Compose uses non-secret local example values and does not start containers or require certificate files. The build job needs public image/module downloads but no deployment credentials.

</details>

<details>
<summary>Run equivalent checks locally</summary>

Use Go 1.27.3, Python 3, Docker/Compose and kubectl. From the repository root:

```bash
# Expected: no file names from the formatting command.
gofmt -l .
go test -count=1 -race ./...
go vet ./...
git diff --exit-code -- go.mod go.sum
python3 scripts/check_docs.py
for policy in deploy/aws/examples/*.json.example; do
  python3 -m json.tool "$policy" > /dev/null
done
MULESOFT_URL=http://mockmulesoft:8081 \\
JWT_JWKS_URL=https://local-issuer:8443/.well-known/jwks.json \\
JWT_ISSUER=https://local-issuer:8443 \\
JWT_AUDIENCE=participant-services \\
  docker compose -f compose.yaml -f compose.local.yaml config --quiet
kubectl kustomize deploy/kubernetes/local > /dev/null
for target in api localissuer mockmulesoft; do
  docker build --target "$target" --tag "participant-services-${target}:ci" .
done
git diff --check
```

CI fails when formatting returns file names; `gofmt -l` itself exits successfully for unformatted files, so inspect its output locally. The workflow includes the explicit failure check. These builds add local `:ci` image tags; they do not change running Compose/kind containers.

The standard-library documentation checker checks current inline Markdown links and heading anchors, plus balanced/collapsed details sections. It makes no external requests and is not a complete Markdown renderer. Historical snapshot links retain their original root context and are excluded from path/anchor checking; their HTML structure is still checked. Historical journal text is preserved, not automatically corrected by the checker.

Optional workflow linting used during development (host Go, network access; tool installs only under the selected temporary directory):

```bash
mkdir -p /tmp/participant-ci-tools
GOBIN=/tmp/participant-ci-tools go install github.com/rhysd/actionlint/cmd/actionlint@v1.7.7
/tmp/participant-ci-tools/actionlint -shellcheck= .github/workflows/ci.yaml
```

The empty shellcheck option disables that optional external tool; Bash syntax was checked separately. Actionlint is a local development tool, not a new Go application dependency.

</details>

<details>
<summary>Activation and limits</summary>

The initial workflow was committed and pushed with explicit user authorisation; its first hosted run passed. For subsequent changes, obtain the required authorisation before committing/pushing, then inspect the repository's Actions tab. Workflow files must exist on the relevant GitHub branch; manual dispatch normally requires the workflow on the default branch. Private repository Actions settings, fork approvals and runner quotas may affect execution. No branch-protection or repository settings were changed.

Offline rendering is not Kubernetes API validation, JSON parsing is not IAM permission validation, and a successful image build is not a runtime authentication test. The local [Compose](containers.md) and [Kubernetes](kubernetes.md) authentication exercises remain separate. Hosted CI does not yet run those stacks, scan images/dependencies, check schema compatibility or produce release artifacts. The first hosted Ubuntu/AMD64 run and all image builds passed; this does not establish multi-platform runtime behaviour.

No CD is implemented. A future publishing/deployment workflow needs reviewed OIDC trust, exact ECR/EKS permissions, environment approval, immutable image promotion and rollback procedures. Participant-level authorisation remains missing, so a passing pipeline does not authorise public exposure.

</details>
