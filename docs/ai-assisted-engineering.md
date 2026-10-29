# AI-assisted engineering

[Documentation index](README.md) · [Development journal](development-journal.md) · [Architecture](../DESIGN.md)

This project uses a small reference implementation to make engineering choices understandable and testable. The initial slice is two GraphQL contracts reading the same synthetic participant through a shared service and mock integration. Features are added only when they clarify a concrete requirement; this is not an attempt to reproduce an enterprise platform.

## Working approach

1. **Bound the task.** State the objective, acceptance criteria, exclusions, relevant files and current evidence. Ask the assistant to inspect the repository rather than trust a handover summary.
2. **Keep context relevant.** Supply contracts, architecture boundaries and a small reproduction. Limited context reduces ambiguity and irrelevant changes, but can hide dependencies and does not eliminate AI errors.
3. **Make small changes.** Separate implementation, documentation and verification where practical. Review the diff against the stated task before accepting another stage.
4. **Verify behaviour.** Use unit/integration tests, race checks, vet and reproducible runtime commands as appropriate. Negative cases and trust boundaries matter as much as successful requests. Report observed failures and checks not performed.
5. **Preserve the investigation.** Append the problem, hypotheses, confirmed cause, decision, correction, commands/results and remaining limits to the journal. Label reconstructed handover evidence separately from newly observed results.
6. **Require human acceptance.** AI output is a draft, not authority. The developer owns requirements, reviews sensitive logic and dependencies, and approves commits, pushes, merges and publication at the agreed boundaries.

## Controls demonstrated here

Synthetic fixtures, ignored local keys, bounded task instructions, explicit repository inspection, incremental commits, separate review branches and pull requests, and read-only credential-free CI make changes easier to inspect. The tests exercise JWT rejection, participant denial before repository access and upstream failure behaviour. These controls support review; none proves all generated code correct or production-ready.

The journal includes a scope mistake: development began before a documentation-first correction was incorporated. The subsequent entry records source restoration and later runtime reconciliation. Preserve that evidence. A claim that AI instructions or sandboxing prevented every mistake would contradict the actual history.

## How to evaluate or extend the reference

Start at [onboarding](containers.md#first-time-local-setup), reproduce the documented checks and follow one request through HTTP middleware, resolver, service and adapter. Read [authentication](authentication.md) and [authorisation](authorisation.md) before changing access controls. Keep the original observations intact and append corrections rather than rewriting prior claims.

Do not treat a model explanation as test evidence, ask it to infer unseen confidential systems, or provide real credentials/participant data as context. Record reviewable decisions and results; publishing private conversations or exhaustive model reasoning is unnecessary. The [verification record](testing.md) distinguishes local demonstrations, hosted checks and work still unverified.
