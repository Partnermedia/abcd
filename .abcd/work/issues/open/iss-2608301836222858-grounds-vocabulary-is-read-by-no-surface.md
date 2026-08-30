---
schema_version: 1
id: "iss-2608301836222858"
slug: "grounds-vocabulary-is-read-by-no-surface"
severity: "nitpick"
category: "tech-debt"
source: "user-observation"
found_during: "itd-179-delta-close"
found_at: "internal/core/grounds/grounds.go"
---

grounds.Vocabulary claimed to be the one copy every gate and every flag description reads, and no surface reads it

Found by an unused-export sweep while settling the itd-179 delta nits, not by
any of the delta reviews.

`grounds.Vocabulary` carried the comment "It is the ONE copy every gate and
every flag description reads." Nothing outside package `grounds` reads it. The
closed set is rendered for refusals by the package-internal `vocabularyList`,
and every surface spells the three values as literal text instead: four sites
in `internal/surface/cli/cli.go`, one in `internal/core/capture/grounds.go`,
plus `commands/intent.md` and the generated `docs/reference/cli/commands.md`.

The comment is corrected in the change that found it, so the claim now matches
the code. What stays open is the consolidation the claim described.

It is a nitpick because the literals and the closed set agree today. The cost
is what a fourth value would take: five edits in code plus the plugin surface
and a regenerated reference page, which is the arithmetic a closed set exists
to avoid.

`grounds.Heading` has the same shape one step smaller — the `## Grounds`
spelling is a literal in record-lint's remedy string because `core/lint` does
not import `core/grounds`. That one is left alone; its comment is narrowed to
what it establishes rather than an import being added for one prose string.
