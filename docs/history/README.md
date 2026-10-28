# Historical snapshots

[Documentation index](../README.md) · [Development journal](../development-journal.md) · [Documentation change log](../documentation-changelog.md)

These immutable copies preserve the complete documents as they stood before Step 74 on 28 October 2026. They are historical evidence, not current technical guidance. Original relative links are retained verbatim and resolve from the original repository-root location, not this archive directory.

- [Original README](README-before-step-74.md): all 985 lines, including setup, source layout, journal, commands, problems and verification notes.
- [Original DESIGN](DESIGN-before-step-74.md): all 467 lines, including architectural proposals and earlier status claims.

Maintain current explanations in the focused references. Do not update these snapshots as implementation evolves.

## Follow the chronology

1. Follow the [development journal](../development-journal.md) from its original entries through every subsequent development step. It is the primary chronological audit trail for implementation, documentation, testing and deployment work.
2. Consult the immutable snapshots above when an entry refers to earlier documentation.
3. Use the [documentation change log](../documentation-changelog.md) for supporting detail about content moves and documentation corrections, then return to the journal to continue the chronology.
4. Use the [current documentation index](../README.md) for the maintained technical explanation of each topic.

## Maintenance rules

For every substantive development step, append a dated, collapsed entry to the development journal. This applies continuously to application code, documentation, configuration, tests, infrastructure and deployments. Record what changed and why, files affected, problems encountered, decisions and trade-offs, checks actually performed and their results, remaining work and commit references when available. Use Australia/Melbourne dates and keep same-day entries in execution order. Retain established step identifiers; do not invent missing steps, dates, checks or commit IDs.

The documentation refactor labelled Step 74 is one milestone in that ongoing chronology. It does not start or reset the audit. Documentation-specific identifiers such as DOC-001 are supporting references, not a replacement development sequence. For substantive documentation changes, also append supporting detail to the documentation change log and link it from the corresponding journal entry.

Keep existing journal entries, change-log entries and snapshots unchanged. Record corrections in a new entry identifying the earlier claim and linking to the current explanation. Maintain overview text and technical references as the system evolves. When moving or removing unique content, preserve it first and record its destination. Create another snapshot for a major restructuring when Git history and the journal alone would make reconstruction difficult; routine edits need a journal entry and Git history, not a full copy.

Keep the authentication implementation and documentation refactor in separate commits, as requested for this milestone. Record the commit containing an entry in a subsequent entry when useful; a commit cannot include its own final hash. Until committed, label changes as working-tree only. Git records exact committed revisions; the journal explains the continuous development process.

From the repository root, inspect the chronology and any recorded commit with:

```bash
git log --reverse --date=iso --format='%h %ad %s'
# Limit to documentation when investigating a documentation-specific change:
git log --reverse --date=iso --format='%h %ad %s' -- README.md DESIGN.md docs/
git log --follow -- docs/development-journal.md
git show <commit> -- README.md DESIGN.md docs/
```
