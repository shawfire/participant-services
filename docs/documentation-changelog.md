# Documentation change log

[Documentation index](README.md) · [History and maintenance rules](history/README.md) · [Development journal](development-journal.md)

Entries are append-only and ordered from earliest to latest. Dates use Australia/Melbourne local time. This supporting log was introduced during the documentation refactor. The [development journal](development-journal.md) remains the primary continuous audit trail before, during and after that milestone; follow it for the full development sequence. Working-tree entries become exact reproducible revisions only after they are committed.

<details>
<summary>DOC-001 — 28 October 2026 — Step 74 documentation refactor</summary>

**State:** implemented in the working tree; no documentation commit created. The inspected committed baseline was `144ec1d` (structured logging and request correlation). Authentication changes were also uncommitted and were left untouched.

**Reason:** separate current guidance from historical development evidence and resolve stale implementation-status claims.

| Source | Current destination | Treatment |
|---|---|---|
| Original root README, 985 lines | [README snapshot](history/README-before-step-74.md) | Complete byte-identical copy before editing |
| Original root DESIGN, 467 lines | [DESIGN snapshot](history/DESIGN-before-step-74.md) | Complete byte-identical copy before editing |
| README entries 1–23 | [Development journal](development-journal.md) | All original details blocks retained verbatim and in order |
| README overview, status and commands | [Root README](../README.md) | Concise overview and quick start updated for mandatory JWT configuration |
| DESIGN architecture and decisions | [Architecture overview](../DESIGN.md) | Current runtime flow and explicit decision status |
| DESIGN contracts and developer commands | [GraphQL](graphql.md) | Current contracts, generation and authenticated request examples |
| DESIGN integration detail | [Integration](integration.md) | Repository adapter, mock and implemented retries |
| DESIGN lifecycle detail | [Health/lifecycle](health-lifecycle.md) | Probes, timeouts, shutdown and limits |
| DESIGN observability detail | [Observability](observability.md) | Logs, correlation and remaining gaps |
| Earlier verification statements and session handover | [Testing](testing.md) | Centralised coverage and evidence limits; historical records preserved |
| Planned architecture and deployment | [Roadmap](roadmap.md) | Future flows and delivery priorities labelled as planned/proposed |
| Authentication working tree and handover | [Authentication](authentication.md) | New reference; authentication remains distinct from participant authorisation |

Added the [documentation index](README.md) and [history index](history/README.md). The journal gained separately labelled authentication and Step 74 entries; original entries were not rewritten. Historical snapshots retain original relative links and their original root-path context.

**Verification:** current relative paths/anchors and collapsed HTML structure passed; all 23 original journal blocks matched; both snapshots matched the original committed files byte for byte; application diffs and authentication hashes matched the pre-refactor baseline; `git diff --check` passed. Go tests were not rerun for documentation changes. Live authenticated query examples were not exercised.

</details>

<details>
<summary>DOC-002 — 28 October 2026 — Make ongoing audit maintenance explicit</summary>

**State:** implemented in the working tree; no documentation commit created.

**Reason:** ensure readers can follow chronological steps from the original documents through the refactor to subsequent versions.

Added this change log, a source-to-destination map for DOC-001, a chronological reading path and maintenance rules in the [history index](history/README.md). Linked the audit trail from the [documentation index](README.md) and journal. Existing snapshots and journal entries remain unchanged; this audit entry supplements them.

**Verification:** relative Markdown paths/anchors, collapsed HTML structure, original journal preservation, immutable snapshots and unchanged application baseline checked; `git diff --check` passed. No application code or tests changed.

</details>

<details>
<summary>DOC-003 — 28 October 2026 — Clarify the continuous development audit</summary>

**State:** implemented in the working tree; no commits created.

**Reason:** the user clarified that chronological, step-by-step documentation applies to all development now and going forward; Step 74 is one milestone within that process.

Updated the documentation index, history reading path, maintenance rules and introductory wording to make the development journal the primary continuous audit trail. This change log remains a supporting record of documentation changes. Earlier entries and snapshots remain unchanged; the prior documentation-only maintenance guidance is superseded by the broader rules. The corresponding journal entry records this clarification in the main chronology.

**Verification:** current relative paths/anchors and collapsed sections passed; all 23 original journal blocks and both snapshots remain unchanged; the application baseline is unchanged; `git diff --check` passed. No application tests were run for this wording change.

</details>

<details>
<summary>DOC-004 — 28 October 2026 — Record verified authentication commit</summary>

**Reason:** authentication was verified and committed separately as `4a4bec9`; current guidance must reflect that state while preserving earlier observations.

Updated root README, DESIGN, documentation index, authentication reference and testing evidence. Added the corresponding verification/commit entry to the continuous development journal. Prior audit entries and snapshots remain unchanged.

**Verification:** fresh Go race tests and vet passed before the authentication commit. Current documentation links/anchors, collapsed sections, original journal blocks, immutable snapshots and `git diff --check` passed before the documentation commit. Staged documentation contains only root README, DESIGN and `docs/`. The documentation commit hash can be found in Git history; no hash is invented here.

</details>

<details>
<summary>DOC-005 — 28 October 2026 — Document local container packaging</summary>

Added [containers](containers.md) as the authoritative Docker/Compose guide; linked it from the root README and documentation index. Updated roadmap status and testing evidence. The corresponding Docker/Compose journal entry records implementation, decisions, environment problems, commands/check outcomes and remaining validation. Documentation refactor commit `a1a622e` contains the previous entries.

**Verification:** both images built; local stack health, missing-token rejection, request IDs, non-root execution, upstream connectivity and graceful shutdown were checked. Valid-token Docker requests remain unverified. Current Markdown links/anchors, collapsed sections, immutable snapshots and original journal blocks passed preservation checks; `git diff --check` passed. Changes remain uncommitted at this entry.

</details>

<details>
<summary>DOC-006 — 28 October 2026 — Reconcile three-service Docker handover and continuous audit</summary>

**Objective:** reconcile documentation with the supplied 16-event Docker troubleshooting sequence and current repository evidence before further development. Preserve all previous journal and change-log text.

**Revisions:** container reference now describes the three Docker targets, base/local Compose split, DNS alias, CA trust, dependency health checks, hardening, startup and troubleshooting. Authentication reference adds the development issuer, RS256 issuance and its local-only security limits. Testing adds reproducible commands, a token-safe temporary Go client and separate evidence rows for handover, inspection, observed checks and source/runtime drift. Roadmap gates minimal Kubernetes work on approved Docker verification. Root README corrects its obsolete no-local-issuer claim; documentation index refines container navigation. DESIGN was inspected and remains the architecture overview without duplicating this troubleshooting detail.

**Audit additions:** appended the full ordered handover sequence and a later session/scope-correction entry to the development journal. Earlier two-service statements remain intact as historical evidence. Exact historical error messages/commands were not invented where the handover provided only symptoms. The initial 401 is not conclusively attributed to TLS. The handover's pending Go verification is preserved separately from the later observed Go/API success.

**Validation:** original working-tree journal and change-log are unchanged prefixes; all original implementation details blocks and immutable snapshots remain preserved; relative paths/anchors, collapsed section structure and `git diff --check` checked. The temporary non-documentation edits from the interrupted development attempt were restored to recorded baselines; pre-existing user work remains. No new application checks were run during documentation reconciliation. No commit or push was made; approval is required before further JWT verification.

</details>

<details>
<summary>DOC-007 — 28 October 2026 — Complete first-time local onboarding</summary>

**Reason:** the local guide assumed existing certificates/config and did not offer a complete new-developer sequence.

Expanded [containers](containers.md#first-time-local-setup) with prerequisites, checkout availability, non-overwriting local configuration/certificate creation, SAN/CA verification, permissions, Compose startup, diagnostics, renewal, tests and shutdown. Updated [testing](testing.md#local-authentication-smoke-check) with a copy-and-run token-safe helper, linked from README, documentation index and authentication reference. Technical details remain in their authoritative references. Appended the corresponding development journal entry without changing previous entries.

**Validation:** Bash syntax checked; exact certificate block executed in an isolated temporary directory; CA/hostname/server-purpose checks, permissions and non-overwrite rerun passed. Markdown links/anchors, collapsed sections, original content preservation and diff checks passed. No live Docker/JWT checks were rerun, no existing certificates were replaced and no application/config files were changed. Changes remain uncommitted pending review.

</details>

<details>
<summary>DOC-008 — 28 October 2026 — Record approved fresh-cache Docker authentication verification</summary>

Updated current container, authentication, testing and roadmap references after approved verification. Added the cold-cache command and actual results to testing; appended the complete development stage to the journal. Earlier handover uncertainty, failed diagnostics and scope-correction entries remain unchanged.

**Observed evidence:** all three rebuilt services healthy with matching running image IDs; exact documented smoke client passed both before and after API recreation. Both valid participant queries returned expected data; tampered, malformed and missing tokens were rejected on both routes. Fresh Go JWKS retrieval, API CA configuration/readability, public probes, Go race tests and vet passed. Documentation structure, links, preservation and diff checks passed. No code/config edits, commits or pushes were required. Security limits remain explicit and Kubernetes is still planned.

</details>

<details>
<summary>DOC-009 — 28 October 2026 — Document verified local Kubernetes demonstration</summary>

Added the Kubernetes runbook as the authoritative guide for kind installation/isolated kubeconfig, local image loading, certificate Secret and CA ConfigMap provisioning, manifests, rollouts, token-safe test Job, localhost access, troubleshooting, EKS concept mapping and cleanup. Linked it from root README, DESIGN and documentation index; updated testing evidence and roadmap. Appended the full development stage, including the confirmed CA ConfigMap annotation failure and server-side apply correction, without rewriting history.

**Observed validation:** Kustomize rendering and server dry runs passed, all three Deployments became ready, authentication Job passed eight valid/invalid-token cases with exact participant data, and localhost probes passed. Prior journal/change-log prefixes and snapshots, shell syntax, current links/anchors, collapsed sections and diff checks passed. No Go/Compose code changes, AWS resources, image publication, commit or push occurred. Runtime scope and unimplemented participant authorisation remain explicit.

</details>

<details>
<summary>DOC-010 — 28 October 2026 — Separate local workload identity from proposed AWS integration</summary>

Added the AWS identity/secrets guide and linked it from index, architecture, Kubernetes guide and roadmap. It distinguishes implemented local service-account controls from proposed EKS Pod Identity, least-privilege policies and CSI delivery. Policy examples are explicit non-deployable templates with no actual AWS ARNs or credentials. Testing records observed local identity/RBAC checks and the successful authentication regression Job. Appended the development stage without changing previous history.

**Validation:** local rendering/server dry run/rollout, account/token/RBAC assertions and eight authentication cases passed. IAM examples passed JSON syntax only; shell syntax, current Markdown links/anchors, collapsed sections, preservation and diff checks passed. No AWS resources or permissions were provisioned and no commit/push occurred.

</details>

<details>
<summary>DOC-011 — 28 October 2026 — Document credential-free CI and evidence limits</summary>

Added the CI guide as the authoritative description of workflow jobs, permission boundaries, local commands, activation and unverified hosted execution. Linked it from README/index/roadmap and appended testing results plus the corresponding development journal stage. Added a reusable local Markdown/HTML checker used by CI; original content and immutable snapshots are not rewritten.

**Observed validation:** race tests, vet, formatting/module checks, merged Compose configuration, Kustomize render, IAM JSON syntax, documentation checks and all three local image builds passed. Workflow lint and Bash syntax passed; a temporary negative fixture confirmed checker failure detection. Prior journal/change-log content, original journal blocks and snapshots remain preserved, and diff checks passed. No hosted CI run, commit, push, publication or deployment occurred.

</details>

<details>
<summary>DOC-012 — 28 October 2026 — Reconcile environment example and build-context guidance</summary>

Updated the container guide after review fixes: the external-provider example now contains the required upstream URL variable; Docker context exclusions now cover local config/certificates and backup files. Existing first-time setup instructions still use `config/local.env`. Appended the review findings, temporary build-tool failure/correction, observed checks and proposed commit boundaries to the development journal; previous history is unchanged.

**Validation:** required-variable coverage, example-based Compose validation, ignored key paths, backup preservation and explicit Docker context assertions passed. Documentation links/anchors, collapsed sections, history preservation and diff checks passed. No application/runtime changes, commit or push occurred.

</details>

<details>
<summary>DOC-013 — 28 October 2026 — Record pushed commits and first successful hosted CI</summary>

Updated current README, container, authentication, CI, roadmap and testing status after authorised commit/push. Recorded the four commit IDs and first successful GitHub-hosted run in the journal, keeping historical uncommitted/unverified observations intact. Remote checkout instructions now reflect published files.

**Observed evidence:** run 40651138381 passed all three jobs for `bd91def`. Hosted test/build evidence replaces current uncertainty without claiming runtime/AWS verification. Prior audit content, snapshots, Markdown structure/links and diff checks passed. Certificates/keys and the user's Compose backup were not committed. The follow-up documentation update is within the user's commit/push authorisation.

</details>
