# abcd Phases

The **phase** is abcd's sequencing layer — an ordered stretch of work that ends
in a **milestone**. Phases replace plugin-version language (`v1`, `v2`) as the
way the project organises *what ships together, in what order*. See
[adr-9](../../decisions/adrs/0009-phase-as-product-layer.md) for why this layer
exists and [`01-product/03-mental-model.md`](../../brief/01-product/03-mental-model.md)
for where it sits among brief / phase / intent / spec.

## What a phase document is

Each phase is a **product-reflection point**. A phase doc opens with the product
thinker's re-statement, in user terms, of what the phase is expected to make
true — coarser than a single intent's press release, finer than the
whole-project brief. Every phase doc carries:

| Section | What it holds |
|---|---|
| `## Expectation` | Working-backwards prose: what is true for the user at phase end. Press-release voice, at phase granularity. The reflection point. |
| `## Milestone` | The concrete engineering done-cut — what must pass for the work to be "finished". |
| `## Phase Acceptance` | Given/When/Then bullets — the *user-truth* cut and the phase-audit target. Roll-up only: asserts emergent, cross-intent truths or phase-spanning journeys; never copies an intent's own acceptance. Same format as intent `## Acceptance Criteria`, one grain up. |
| `## Scope` | Which intents and which brief plumbing-phases this phase bundles. |
| `## Maps against` | Traceability — which brief sections this phase realises; which intents deliver the expectation. |
| `## Dependency rationale` | Why this phase sits where it does; what it must run after. |
| `## Open questions` | Anything not yet decided. |

`## Expectation` (prose) and `## Phase Acceptance` (structured) mirror an
intent's press-release-then-`## Acceptance Criteria` shape, one grain up. The
product thinker authors both. `## Milestone` answers "is the work finished?";
`## Phase Acceptance` answers "is the expectation met?" — see
[adr-9](../../decisions/adrs/0009-phase-as-product-layer.md) and its amendment.

A phase doc carries **no status** — per [adr-5](../../decisions/adrs/0005-brief-is-current-state.md),
status is never stored in design docs. The [roadmap dashboard](../README.md)
renders phase progress by reading `.flow/`.

## How phases connect to the rest of the model

- **Intents → phase.** The intent → phase mapping is editorial and lives here,
  in each phase doc's `## Scope`. Intent files themselves carry no phase field
  (per adr-9 — phase-grain mapping is editorial, not per-item linkage).
- **Specs → phase.** Phase membership is reconstructed **editorially from each
  phase doc's `## Scope`** — this is the live mechanism today. The `phase:`
  frontmatter anchor described in adr-9 has its **tooling DELIVERED by fn-66**:
  the phase-audit reviewer that reads it (`scripts/abcd/phase_audit_reviewer/`)
  and the `PA001` verify-exists lint (a `phase:` naming a non-existent phase is
  an error) both exist. What stays **deferred** is making `phase:` a *standing
  convention* — the corpus anchor backfill onto the unanchored specs is a
  separate planning act (fn-49.2's deferral), now **unblocked** by fn-66 but not
  performed by it. Until that backfill lands, no spec is expected to carry the
  anchor and its absence on the corpus is not drift; a spec listed in no phase
  doc's `## Scope` is implicitly unscheduled and correctly carries no anchor
  (the spec analogue of adr-9's unscheduled-intent rule). `PA001` only validates
  an anchor that IS present — a missing anchor is legal.

## Phase index

| Phase | Milestone | Document |
|---|---|---|
| Phase 0 — Substrate & disciplines | Oracle backend shipped; project state honest; three disciplines in force | [phase-0-substrate.md](phase-0-substrate.md) |
| Phase 1 — ahoy | `/abcd:ahoy` installs cleanly on any folder | [phase-1-ahoy.md](phase-1-ahoy.md) |
| Phase 2 — capture | `/abcd:capture` writes a structured issue ledger | [phase-2-capture.md](phase-2-capture.md) |
| Phase 3 — intent | `/abcd:intent` + `grill` harden an intent into a spec-ready PRD | [phase-3-intent.md](phase-3-intent.md) |
| Phase 4 — The lifeboat pipeline | `/abcd:disembark` packs a faithful lifeboat | [phase-4-lifeboat.md](phase-4-lifeboat.md) |
| Phase 5 — Round-trip and ship | `/abcd:embark` + `/abcd:launch`; abcd ships itself | [phase-5-roundtrip.md](phase-5-roundtrip.md) |

Phases are organised by **user-capability moment** — each one ends in a
milestone a contributor can demo. Phase 0 is the exception: it has no
user-facing command, and is numbered 0 to say so honestly — it is the floor the
five capability phases stand on. Phases are sequenced but their *contents* may
run in any dependency-respecting order — see each phase's `## Dependency
rationale`.

## Beyond Phase 5

Phase 5 closes abcd's loop and cuts the first public release; it is not the end
of the work. Most of the intent corpus is not yet phased — an intent listed in
no phase doc's `## Scope` is implicitly unscheduled (per adr-9), and that is a
valid state. Further phases will be authored as that work is committed to.

This section is deliberately a placeholder, not a forecast: naming which intents
land in which future phase before the phase is planned would only go stale. See
[intents/README.md](../../intents/README.md) for the full corpus; an intent moves
into a phase's `## Scope` when the phase that bundles it is written.

## Related Documentation

- [Roadmap](../README.md) — status dashboard
- [Intents](../../intents/README.md) — the intent registry
- [Build sequence](../../brief/06-delivery/01-build-sequence.md) — the brief's plumbing-phase DAG, which phases bundle from
- [adr-9](../../decisions/adrs/0009-phase-as-product-layer.md) — the phase layer decision
