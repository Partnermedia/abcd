<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: phase
bounded_context: core
definition: An ordered stretch of development work that bundles a set of intents and brief plumbing-phases and ends in a milestone; abcd's sequencing layer, recorded as a document in roadmap/phases/.
aliases: ["build phase", "roadmap phase"]
forbidden_synonyms: ["version", "release", "milestone", "sprint", "iteration"]
status: stable
introduced_in: adr-9
starts_when: null
ends_when: null
not_to_be_confused_with: core/spec
versions: null
---

# phase

A **phase** is abcd's sequencing layer — an ordered stretch of work that ends in a
**milestone** (a concrete, checkable end condition). Phases replace plugin-version language
(`v1`, `v2`) as the way the project organises what ships together and in what order. Each
phase is a document in `.abcd/development/roadmap/phases/` that opens with a product
`## Expectation`: a working-backwards re-statement, at phase granularity, of what the phase is
expected to make true for the user.

A phase sits between the [brief](../../brief/README.md) (the whole project) and the
[intent](intent.md) (one user-facing capability) on the question it answers — see the
four-layer mental model in `brief/01-product/03-mental-model.md` and
[adr-9](../../decisions/adrs/adr-9-phase-as-product-layer.md).

## When to use

Use "phase" for an ordered stretch of work that bundles intents and plumbing into a coherent
milestone-ending unit. Use it when discussing *what order* work happens in, or *what
expectation* a stretch of work is measured against.

## When NOT to use

Do not use "phase" for an individual work block (that is a [spec](spec.md)), for a plugin
release number (a release version is an *output* of completing a phase, never the organising
unit), or interchangeably with "milestone" — the milestone is the *end condition* of a phase,
not the phase itself.

## Examples

- "Phase 1 — Substrate ends when the oracle backend ships and project state is reconciled."
- "Spec `fn-5` carries `phase: phase-1-substrate` in its frontmatter." (The `phase:` anchor's tooling shipped in fn-66 — the phase-audit reviewer that reads it and the `PA001` verify-exists lint both exist; what stays deferred is the *corpus backfill* that would make the anchor a standing convention. Phase membership today is still editorial, via each phase doc's `## Scope`.)
- "The phase audit reviews delivered reality against the phase's structured `## Phase Acceptance`." (Per the adr-9 amendment: the reviewable cut is the structured `## Phase Acceptance` bullets, NOT the prose `## Expectation` — prose is the narrative re-statement, not the audit target.)

## Related terms

- [spec](spec.md) — the implementation unit; many specs belong to one phase
- [intent](intent.md) — the user-facing capability; a phase bundles a set of intents
- [voyage](voyage.md) — a full lifecycle arc
