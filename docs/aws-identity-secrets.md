# AWS workload identity and secrets

[Documentation index](README.md) · [Local Kubernetes](kubernetes.md) · [Roadmap](roadmap.md)

**Implemented locally:** API Pods use the dedicated `participant-api` Kubernetes service account, with automatic API token mounting disabled and no RoleBinding or ClusterRoleBinding added. **Proposed only:** EKS Pod Identity, IAM roles and Secrets Manager integration. No AWS account has been inspected or changed.

<details>
<summary>Current needs and identity boundaries</summary>

| Concern | Current implementation | Future design |
|---|---|---|
| End-user identity | RS256 Bearer JWT from configured issuer | Production identity provider and participant policy |
| Kubernetes workload identity | Named API service account; no application RBAC grants | Same namespace/account names can identify an EKS workload |
| AWS identity | No AWS SDK calls or credentials in API | EKS Pod Identity only when an AWS consumer exists |
| API runtime configuration | Upstream URL, issuer/audience/JWKS URL and CA | Keep non-secret settings in ConfigMaps |
| API secrets | No upstream password, API key or database credential is consumed | Add a consumer and lifecycle tests before wiring managed secrets |
| Local issuer key | Ignored host file, injected through a Kubernetes Secret | Replace local fixture with a managed/trusted provider; do not promote its development signing key |

A Kubernetes service account does not grant AWS access. End-user JWT validation does not grant Kubernetes or IAM permissions. [Self-access](authorisation.md) now additionally checks the issuer-managed participant mapping. Staff/delegate entitlement remains unimplemented.

The API does not need permission to read its local issuer's private key. Kubernetes mounts that Secret into the issuer Pod; the API verifies public JWKS keys. Reading a mounted file does not require application `get secrets` RBAC permission. Named service accounts and disabled token mounts do not provide network isolation.

</details>

<details>
<summary>Proposed EKS approach and reviewable policy examples</summary>

For a future Linux EC2-backed EKS exercise, associate an IAM role with the namespace/service-account pair through EKS Pod Identity. It requires cluster support and the agent unless supplied by the selected EKS mode. Kind cannot validate this integration. A future SDK consumer should use supported default credential-chain behaviour, not static AWS keys. See [AWS Pod Identity guidance](https://docs.aws.amazon.com/eks/latest/userguide/pod-identities.html).

Two **non-deployable examples**, outside the local Kustomization, capture the proposed policy boundaries:

- [Trust policy example](../deploy/aws/examples/pod-identity-trust.json.example): trust `pods.eks.amazonaws.com` for `sts:AssumeRole`/`sts:TagSession`, constrained to the intended cluster ARN, namespace and service account. Replace `<CLUSTER_ARN>` only after selecting the target cluster; preserve session tags for those conditions. Based on [AWS trust-policy guidance](https://docs.aws.amazon.com/eks/latest/userguide/pod-id-role.html).
- [Secret-read policy example](../deploy/aws/examples/secret-read.json.example): read/describe one exact secret ARN, including its generated suffix. It grants no create, update, delete, list-all or wildcard resource access. Replace the placeholder with an approved synthetic secret; this example is not attached to any role.

If a customer-managed KMS key is selected, review the exact key policy and narrowly scoped decrypt permission separately. No KMS grant is included because no key or consumer exists. Restrict who can create Pod Identity associations and deploy Pods using a privileged service account; an IAM role boundary alone is not workload admission control.

</details>

<details>
<summary>Proposed secret delivery and application changes still needed</summary>

For a file-based consumer, consider AWS Secrets and Configuration Provider with the Secrets Store CSI driver and `usePodIdentity: "true"`. Install compatible components and bind a reviewed `SecretProviderClass` to the workload only during an approved EKS exercise. This lets the provider mount selected Secrets Manager values; it does not automatically teach the Go API to consume or reload them. See [AWS integration instructions](https://docs.aws.amazon.com/secretsmanager/latest/userguide/ascp-pod-identity-integration.html).

For this prototype, first define a real requirement such as an upstream credential. Then add a small file-based credential reader with safe failure behaviour and tests before adding a secret mount. Decide whether a restart or explicit reload handles changes; a refreshed file does not guarantee an application reload. Avoid syncing into a Kubernetes Secret unless an application integration requires it. File ownership/permissions for UID/GID 65532 must be verified with the selected CSI implementation.

Public issuer URLs, JWKS keys and a CA bundle do not become secrets simply because they configure authentication. Keep confidential values out of images, environment examples, logs and Git. Kubernetes Secret objects are not an AWS managed-secret integration; base64 encoding alone does not establish confidentiality.

</details>

<details>
<summary>Local checks and verification limits</summary>

With the isolated kubeconfig from the Kubernetes guide:

```bash
export KUBECONFIG=/tmp/participant-services-kubeconfig
kubectl config current-context
kubectl apply --dry-run=server -k deploy/kubernetes/local
kubectl -n participant-services get serviceaccount participant-api
kubectl -n participant-services get deployment api -o jsonpath='{.spec.template.spec.serviceAccountName}{"\n"}'
kubectl -n participant-services exec deployment/api -- sh -c \
  'test ! -e /var/run/secrets/kubernetes.io/serviceaccount/token'
kubectl -n participant-services auth can-i get secrets \
  --as=system:serviceaccount:participant-services:participant-api
```

Expected: service account is `participant-api`, no default token file exists, and the final command prints `no` and exits with status 1. That nonzero exit is the expected denial; do not chain it with `&&` expecting success. The impersonation check needs the administrator rights provided by the isolated kind kubeconfig. It tests Kubernetes RBAC only, not AWS IAM. Repeat the [in-cluster authentication Job](kubernetes.md) after the API rollout to ensure end-user authentication and upstream queries still work.

The JSON examples can be parsed offline, but JSON syntax is not IAM authorization validation. Cloud access, trust conditions, cross-secret denial, KMS, rotation and CSI mounts remain unverified until exercised on a reviewed EKS setup. Do not present local RBAC checks as proof of AWS policy behaviour.

</details>

<details>
<summary>Next approved cloud exercise</summary>

Before any AWS mutation, select the account/region, cost limit, supported EKS compute/version, operator identity and cleanup owner. Prepare a concrete deployment plan: ECR images, cluster/network access, external identity provider, workload identity components and one synthetic secret consumer. Production participant data and the local issuer are out of scope.

The acceptance exercise should verify intended-role identity without printing credentials, approved-secret access without printing values, denial for an unrelated secret, application behaviour on rotation and outage, and cleanup of every created resource. Record actual results and costs in the journal. Account-specific IAM/CSI configuration is deliberately not installed into kind or represented as an implemented EKS deployment.

CI validation/build automation can be added locally before cloud credentials or deployment are authorised.

</details>
