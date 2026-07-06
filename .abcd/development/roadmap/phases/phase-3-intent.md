# Phase 3 — intent

## Expectation

By the end of this phase, a user can take a vague intent and harden it into a
spec-ready one. `/abcd:intent` manages the intent lifecycle; `/abcd:intent
grill` runs a two-phase Socratic challenge — interrogating the intent for
vagueness, then silently synthesising a Pocock-shaped PRD frozen at promotion —
and emits an emergent glossary. Spec-tied review artefacts land in `.flow/` so
the plan-review step has somewhere to put its output. This is the phase where
an idea stops being a sentence and becomes something `/flow-next:plan` can turn
into a spec.

This phase is the **end of intent authoring**, not the start of implementation.
abcd does not own the build step — `/flow-next:plan` and `/flow-next:work` do.
Grill's job is to make the hand-off to flow-next clean: a grilled intent is one
a planner can consume without re-interrogating it.

## Milestone

- `/abcd:intent grill <itd-N>` runs both phases per `04-surfaces/05-intent.md`:
  Phase 1 Socratic vagueness interrogation, Phase 2 silent PRD synthesis. The
  PRD is frozen at `.abcd/intents/<itd-N>/prd.md` on promotion.
- The emergent glossary is written under
  `.abcd/development/foundation/terminology/`; grill's lint codes
  (GL001–GL005, GR001–GR005) are live.
- The coherence-aware grill tiers work: Tier 2 (brief-coherence) and Tier 3
  (sibling-intent index) run; light vs. full grill is selected by lifecycle
  position.
- Spec-tied RP reviews land at `.flow/reviews/<spec-id>/<NNNN>-<slug>-<ref>/`
  with a `review.json` sidecar and two-stage redaction, per
  `04-surfaces/05-intent.md` and the fn-2 spec.

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/adr-9-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** a vague draft intent, **when** a user runs `/abcd:intent grill`
  and then `/abcd:intent plan`, **then** the intent reaches `planned/` carrying
  a frozen PRD, a populated glossary, and acceptance criteria that survive the
  Phase 0 itd-1 gate — a journey across itd-27, itd-42, and the Phase 0
  disciplines that no single intent delivers alone.
- **Given** a grilled intent whose PRD is frozen, **when** `/flow-next:plan`
  consumes it, **then** the planner has enough specificity to produce a spec
  without re-interrogating the user — the emergent "clean hand-off to
  flow-next" property that is the whole point of the phase.
- **Given** a plan-review or impl-review runs on a spec, **when** it produces
  review output, **then** that output lands spec-tied in `.flow/reviews/` with
  redaction applied — itd-28 making the review trail a durable, pinned
  artefact rather than transient chat scrollback.

## Scope

**Intents:** itd-27 (`/abcd:intent grill` sub-verb + emergent glossary —
two-phase Socratic challenger producing a frozen Pocock PRD), itd-42
(coherence-aware grill — Tier 2 brief-coherence and Tier 3 sibling-intent
index, light vs. full grill by lifecycle), itd-28 (spec-tied RP reviews land in
`.flow/reviews/` — the review artefacts the plan-review step of this phase
produces need a home).

**Why grill is here and not in an "implementation" phase.** Grill's Phase 2
output is a PRD *frozen at intent promotion* — it is the last step of intent
authoring, the thing that turns a vague intent into a spec-ready one. The
actual build is `/flow-next:plan` → `/flow-next:work`, which abcd consumes
rather than builds. There is no abcd "implementation phase"; there is this
intent-hardening phase, and grill is its centrepiece.

**Why capture (Phase 2) and intent are separate phases.** Both surfaces let a
user "record something", but capture is a fast, low-stakes ledger entry and
intent authoring is a high-stakes, Socratic act producing a frozen PRD. They
have different audiences, different risk, and different demoable milestones —
bundling them would blur both. They are sequenced, not merged.

**Brief plumbing-phases:** none of its own — the `/abcd:intent` surface flow is
covered by `04-surfaces/05-intent.md`. The probe-only bare `/abcd:intent`
render stub from brief-Phase 1 is joined here by the real `grill` sub-verb.

## Maps against

- **Brief:** `04-surfaces/05-intent.md` (the intent surface, grill, the
  intent-fidelity-reviewer's three roles); `05-internals/06-lint.md` (the
  GL/GR lint families); `05-internals/08-skills.md` (the abcdGrill skill).
- **Intents deliver the expectation:** itd-27 delivers the grill sub-verb and
  glossary; itd-42 delivers the coherence-aware tiers; itd-28 delivers the
  spec-tied review trail the phase's plan-reviews write into.
- **ADRs realised:** adr-9 (phase-as-product-layer — grill's PRD and the
  phase's `## Expectation` mirror at intent grain); adr-8 (dual-backend
  review — the review artefacts itd-28 lands).

## Dependency rationale

- **Runs after Phase 0** — every spec planned in this phase inherits the Phase
  0 disciplines, and grill's own acceptance criteria must pass the itd-1 gate.
  Grill is in part an *application* of the acceptance-gate discipline at
  authoring time, so the discipline must already be in force.
- **Runs after Phase 1** — `/abcd:intent` is a command, and a command needs
  the ahoy install flow and the rules loader to be live before its sub-verbs
  are wired.
- **itd-27 depends on itd-42** — the coherence-aware tiers are the grill
  machinery itd-27's sub-verb invokes; itd-42 lands the tier system, itd-27
  lands the user-facing sub-verb and glossary on top of it. itd-27 also
  depends on itd-1 (its own acceptance criteria pass the gate).
- **itd-28 depends on Phase 0's itd-6** — spec-tied reviews come from the RP
  MCP integration; itd-28 is the post-processor that lands what the oracle
  emits. It is grouped here because the plan-reviews this phase runs are the
  first heavy producers of review artefacts that need a durable home.

## Open questions

- itd-27 and itd-28 currently sit in `intents/planned/` with plan-reviewed
  specs (`fn-3`, `fn-2`). Phase 0's reconciliation work moves them to
  `shipped/` once the flowctl `spec close` close-hook (fn-36) → `intent_lifecycle.reconcile`
  (fn-28) runs on their close. Confirm: do they ship as
  part of Phase 0's reconciliation (because their specs are already complete
  on disk) or are they re-opened as Phase 3 scope? If their specs are genuinely
  done, Phase 3's scope for itd-27/itd-28 is *verification that the shipped
  reality matches the phase expectation*, not fresh build — state this when
  the phase is worked.
- itd-42 is currently an unplanned draft — it must be planned (and likely
  grilled) before itd-27's spec can depend on it. Confirm the itd-42 → itd-27
  ordering holds once both have specs.

## Addendum (2026-05-30)

Duplicate `itd-45` resolved by fn-31 on 2026-05-30: the drift-detector intent
was renumbered to `itd-49` (`itd-49-flow-state-drift-detector.md`); the kept
`itd-45` is the phase-1-cleanup intent (`itd-45-phase-1-cleanup-before-phase-2.md`).
Any historical mentions of `itd-45` in this doc refer to the kept
phase-1-cleanup intent.
