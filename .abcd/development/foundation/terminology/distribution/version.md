---
term: version
bounded_context: distribution
definition: A strict-SemVer string carried in the public plugin's SELECTED version location (recorded by .abcd/config/version-location.json — plugin.json.version under the ACCEPT outcome) and in git tags, identifying a published snapshot of abcd for install and update. It is an OUTPUT of publishing, distinct from the internal sequencing unit (phase).
aliases: ["semver", "plugin version"]
forbidden_synonyms: []
status: stable
introduced_in: itd-67
starts_when: null
ends_when: null
not_to_be_confused_with: core/phase
versions: null
---

# version (distribution)

A **version** in the distribution context is the semantic-version string that
identifies a published snapshot of the abcd plugin — carried in the SELECTED
manifest location recorded by the fn-77.1 decision artifact
([`.abcd/config/version-location.json`](../../../../config/version-location.json))
and as a git tag, so Claude Code can compare what is installed against what is
available and users can update. The decision artifact records where the version
lives: today the outcome is **ACCEPT**, so the version is `plugin.json.version`
(`manifest_path: .claude-plugin/plugin.json`, `json_pointer: /version`). Were the
outcome ever **BLOCKED**, there would be no schema-valid distribution version
location and Tier A would be escalated — no location string is then canonical.

This is deliberately distinct from [phase](../core/phase.md), which forbids
"version" as a synonym *in the core context*. That prohibition is correct there:
abcd organises its internal development work by phase, not by version number. But
when abcd is PUBLISHED as a plugin, a semantic version is the precise, correct
term — it is what "falls out" when a phase (or a routine snapshot) is promoted to
the public repo. The two coexist: `phase` sequences the work; `version` labels
the published result.

## When to use

Use "version" for the semver string of a published abcd plugin snapshot —
in `plugin.json`, git tags, the marketplace entry, and the changelog.

## When NOT to use

Do not use "version" for the internal sequencing of development work — that is a
[phase](../core/phase.md). A version is the output of publishing, never the unit
that organises what ships together.

## Examples

- "`launch ship` bumps `plugin.json.version` from `0.1.0` to `0.2.0` (strict SemVer, no leading `v`) and tags the public repo `v0.2.0`."
- "Users update to the latest version with `/plugin update abcd`."

## Related terms

- [phase](../core/phase.md) — the internal sequencing unit; a version is an output of completing one
- [release](release.md) — the published act that carries a version
