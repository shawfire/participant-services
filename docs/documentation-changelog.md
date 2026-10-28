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
