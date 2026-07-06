---
term: release
bounded_context: distribution
definition: A published, version-tagged snapshot of the abcd plugin in the public repo — the act and the artefact of promoting abcdDev to the public sibling, carrying a version and a changelog entry.
aliases: ["published snapshot", "plugin release", "snapshot"]
forbidden_synonyms: []
status: stable
introduced_in: itd-67
starts_when: null
ends_when: null
not_to_be_confused_with: core/phase
versions: null
---

# release (distribution)

A **release** is a published, version-tagged snapshot of the abcd plugin in the
public repo — both the act of promoting abcdDev to its public sibling and the
resulting artefact, carrying a [version](version.md), a changelog entry, and a
git tag.

As with [version](version.md), [phase](../core/phase.md) forbids "release" as a
synonym *in the core context* — abcd does not organise development by releases.
But in the distribution context, "release" is the correct term for the published
output: a phase completing may TRIGGER a release (a minor version bump), but the
release is the publish event, not the sequencing unit.

## When to use

Use "release" for a published, version-tagged snapshot promoted to the public
repo, and for the act of publishing one via `launch ship`.

## When NOT to use

Do not use "release" for an internal stretch of development work (that is a
[phase](../core/phase.md)) or as a loose synonym for "milestone" (the phase's end
condition).

## Examples

- "The v0.2.0 release publishes the completed launch phase to the public repo."
- "`launch ship` refuses a no-change release unless `--force` is passed."

## Related terms

- [version](version.md) — the semver string a release carries
- [phase](../core/phase.md) — completing one may trigger a minor-version release
