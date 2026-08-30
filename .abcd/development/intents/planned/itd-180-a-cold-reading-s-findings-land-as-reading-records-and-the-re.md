---
id: itd-180
slug: a-cold-reading-s-findings-land-as-reading-records-and-the-re
spec_id: spc-58
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: [itd-86]
severity: major
impact: additive
---

# A cold reading's findings land as reading records, and the researcher's response is a separate disposition record — two acts, two writes, never collapsed

Typed links: `refines` [itd-86](../drafts/itd-86-cold-reading-surface.md) (the
detection shape); the hold state's axes are shared with the itd-142 hold
register and the iss-2608220750029991 triage-route seed.

## Press Release

> **A reading record is not a capture with a flag, and a disposition is not an
> issue state.** A capture is something a person noticed; a reading record is
> something an instrument returned under a recorded visible world — it
> carries a run-scoped identifier, a run identifier, a manifest reference,
> its position and regime value, and a position-typed body (for the
> detection pass: tension, constraint in play, why it is a tension). The researcher's
> response is a second record, keyed to the item identifier, from a
> five-state vocabulary whose availability varies by position: `accepted`
> (all positions — at the widening position it IS admission), `rejected`
> (a testable purpose is asserted; never at the widening position),
> `declined` (widening only — the proposal was admissible and the
> researcher chose otherwise, asserting nothing testable), `held`
> (directional, with an epistemic exit condition). The record can therefore always show that a
> reading record existed before it was dispositioned.

## What's In Scope

- The reading record type (renamed from "detection", ruled 2026-08-28 —
  the registrative body and the Step-6 instrument keep the detection
  name; the record type is the reading record): a common envelope the
  instrument stamps — identifier (run-scoped, adr-45 mint, never
  content-derived, else a re-raise is indistinguishable from its first
  appearance and the recurrence signal dies), run identifier, manifest
  reference, position and regime value from the definition, and the
  pattern named — plus a position-typed body from the reading.
- The bodies, one per position — the pattern named sits in the envelope,
  never in a body, since a universal core condition must not live in a
  variant part: registrative — tension · constraint in play · why it is a
  tension; generative — configuration · what admits it; explicative —
  claim surfaced · claim type · what implies it; evaluative — candidate
  identifier · criterion · characterisation. One record type: one lint,
  one disposition surface, one identifier scheme; four types are
  rejected. Fallback if the discriminated union proves awkward in build:
  one type, untyped body, and a lint asserting the required fields per
  position — weaker, and adequate.
- The disposition record, written separately (ruled 2026-08-28, R7):
  five states, availability varying by position — `accepted` (all
  positions; at the widening position acceptance IS admission, since a
  state encoding position would duplicate the envelope), `rejected`
  (explicative/evaluative/registrative only: it asserts a purpose the
  closing run tests), `declined` (widening only: the proposal was
  admissible and the researcher chose otherwise — forcing that into
  `rejected` would manufacture a principle never at stake), `held`
  (exit condition required; availability at the widening position still
  open). The grounds field is `disposition_grounds`, required on every
  state except `held`; what it must contain varies by state, enforced by
  lint rather than by four fields. Free text, not enumerations. Nothing
  meaning "already covered" exists in any position; an undispositioned
  item is reported as outstanding, not named as a state. The disposition
  record reads the envelope's position to validate its own state — a
  coupling the schema carries and the lint checks — and the
  admitted-against-declined count at the widening position is the
  ownership evidence, queryable without reading prose.
- The recurrence link, on the warm side: a disposition may cite prior
  item identifiers (`recurs`), and a re-acceptance or re-rejection
  made against evidence of persistence carries that citation — the
  stronger record recurrence-is-signal describes, and the answer to the
  duplicate case (re-disposition with citation, not a fourth state). The
  reading never sees it: the assembler excludes dispositions, so the link
  lives entirely on the ledger side. **Adopted (2026-08-28):** the citation is the recorded form of the
  researcher's warm recognition — new material relative to the governing
  design, adopted, and routed back as an amendment to it.
- On-disk shape: reading records land under `issues/readings/<run-id>/`
  and dispositions under `issues/dispositions/`, keyed by the item
  identifier. The status signal is the presence of the keyed disposition —
  never folder membership — and RS001–RS003, `capture resolve`, and the
  changelog derivation are taught in the same change to scope to
  open/resolved/wontfix and ignore both new types.
- The output contract that "validated" refers to is owned by the
  cold-reading output-contract intent; this intent consumes it and adds no
  second validation path.
- The surprise entry (per the recording-obligations ruling): a distinct record shape, schema reserved in
  this family now and populated in Iteration 2 — the reading's output, the
  researcher's disposition, and the surprise that occasions abduction are
  three acts, three records. The admission-side records (grounds on
  admission, declined-proposal dispositions) live in the
  step2-admission-records draft.
- The reserved two-axis hold field (frame-location × MoSCoW), present in
  the schema and unpopulated — deferred by decision; reserving costs
  nothing, retrofitting is expensive. Defined-and-dormant: the value
  grammars are stated (frame-location: free text naming the frame element;
  MoSCoW: must / should / could / wont) and a populated value is refused
  until activation is ruled — refusal, not silent acceptance, so the
  reservation is a behaviour rather than a comment.
- The lint: every reading record in a run either carries a disposition or is
  reported as outstanding — report-only, on the capture status board and
  the lint summary; it never gates preflight or CI, so a reading can never
  block unrelated pushes. Open holds render on the same board with their
  exit conditions; a hold exits only through a superseding disposition
  that cites it — never by expiry, and never silently.
- Routing on acceptance: action is a separate admission joined by the
  item identifier stamped forward on `promoted_to` and back in
  `origin`. Item-to-intent without a disposition is the collapse this
  record family exists to prevent: `capture promote` refuses an item
  identifier that carries no disposition, and a circumvention is a
  lapse-log entry.

## Ruled

- **Ruled (maintainer, 2026-08-28; decision log):** the reading record is a distinct record type in
  the existing issue tier, with a separate disposition record and its own
  disposition vocabulary — the five-state, availability-by-position slate
  (R7) this draft implements. Ground: the
  surprise entry and the disposition are different acts and must be
  distinguishable records; reusing the issue states collapses them and
  misdescribes all three.
- **Recurrence matching is warm work (per the closing-run ruling):** run-scoped identifiers
  join nothing mechanically; the researcher recognises a recurrence
  against the ledger, and the recognition is itself a disposition
  judgement — the `recurs` citation in scope is that recognition's
  recorded form.
- **Where an accepted item goes (per the acceptance-routing ruling):** acceptance is one
  record; the action is a separate admission and build, joined by the
  item identifier (forward on `promoted_to`, back in `origin` with
  the run identifier). The landings are enumerated — artefact via the
  intent lifecycle, cross-cutting rule via a discipline, redecision via a
  superseding ADR, the brief's description via the delivering change, the
  construal via a section rewrite with the prior construal passing to
  ledger content; the frame-level record is iteration 2, the fourth
  verdict deferred.

## Open (maintainer readings design, 2026-08-28)

- Whether `held` is available at the widening position: a configuration
  held with an exit condition is a candidate deferred, which is what
  `deferred` already means in the selection-grounds vocabulary
  (pursued / deferred / declined) — two words for one act is the drift
  the design avoids elsewhere. The alternative routes a deferred
  configuration through the selection surface instead of the disposition
  surface. Reaches the selection vocabulary as well; does not gate the
  build.

## What's Out of Scope

- Reusing open/resolved/wontfix: `resolved` means fixed, but an accepted
  detection may be deliberately not acted on; `wontfix` means will-not-act,
  whereas rejection asserts an intentional constraint; `open` is a parking
  space, whereas held is directional.
- Passing dispositions to a reading — warm by definition; the assembler
  excludes them and the read-block eval asserts it.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a validated reading output, **when** it is ingested, **then**
  one reading record is written per item, each with a run-scoped
  identifier.
- **Given** two runs producing the same tension, **when** both are
  ingested, **then** the two records carry different identifiers.
- **Given** any disposition with empty `disposition_grounds` (or a hold
  with an empty exit condition), **when** the command runs, **then** it
  refuses.
- **Given** a disposition whose state is unavailable at the item's
  position (a `declined` on a detection, a `rejected` on a widening
  configuration), **when** the command runs, **then** it refuses and
  names the availability rule.
- **Given** a run's reading records, **when** the lint runs, **then** every
  record either carries a disposition or is reported as outstanding.

