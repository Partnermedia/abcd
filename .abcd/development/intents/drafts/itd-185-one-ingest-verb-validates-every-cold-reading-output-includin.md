---
id: itd-185
slug: one-ingest-verb-validates-every-cold-reading-output-includin
spec_id: null
kind: null
suggested_kind: bundle-member
reclassification_history: []
builds_on: []
severity: major
impact: additive
---

# One ingest verb validates every cold-reading output — including what the reading was licensed to produce, not only what it saw

## Press Release

> **A reading that quietly exceeds its licence is refused at ingest.** The
> output contract carries, per item, an identifier the ingest verb mints
> (adr-45 mint, run-scoped — the reading itself holds no mint, so ids are
> assigned at validation, never self-supplied) and, per run, the
> instrument name, manifest reference, target state, and regime value. Ingest validates before any
> durable record is written — malformed output is rejected without partial
> writes — and then checks the supply regime: a detection-position output
> attaching a proposed resolution, a comparative output containing an
> ordering, a score or a single named recommendation, an entailment output
> in which a surfaced claim arrives already dispositioned — each is refused
> and the offending item named. The failure this catches is silent
> everywhere else: such an output would pass every structural test while
> violating the one property the position is defined by.

## Ruled

- **Ruled (maintainer, 2026-08-28; decision log):** build the supply-regime check now, as
  a regime field on the output contract validated at ingest — the form
  drafted here. Ground: a reading that quietly proposes or aggregates
  passes every other planned test; the read-block eval covers what a
  reading saw, never what it was licensed to produce, and the failure is
  silent. The ruling matches the shape implemented below — structural
  signatures at `evaluative` (aggregation) and `registrative`
  (fix-proposing), `explicative` checked as record shape rather than
  prose, and the record-and-flag degradation path for noisy signatures.
  Still open (recorded open question): whether the signatures lint cleanly in practice.

## What's In Scope

- The ingest verb: JSON in, validation, durable reading records out; the
  read-block and contract written in one place.
- The regime table, checked structurally at ingest against a strict schema
  (unknown fields are refused, so every violation is a named field, never
  a guess): `generative` (no regime-specific refusal — the constraint
  falls on admission at the dispositioning end); `explicative` (the input
  schema carries no disposition field, so a dispositioned claim is refused
  as an unknown field — the violation is impossible to express, not merely
  caught); `evaluative` (refuses an explicit rank or score field, or an
  item marked recommended; arrangement order is never refused — items
  arrive in document order by mandate); `registrative` (refuses a
  resolution field attached to a detection). Semantic violations — prose
  that ranks, settles, or proposes without the fields — are checked too;
  any signature degrades only on observed noise, per the rung bullet
  below, never pre-emptively.
- Item shapes are position-typed and validated per position: generative —
  configuration / what-admits-it; explicative — claim / type /
  what-implies-it; evaluative — candidate / criterion / characterisation;
  registrative — tension / constraint-in-play / why. The pattern named is
  an envelope field, validated once for every position; the strict schema
  is per-position and unknown fields still refuse.
- The regime's source of truth is the definition, resolved through the
  run's position: an output's self-declared regime that mismatches it is
  itself a refusal; no operator input can set a regime.
- Named provenance is enforced here, for every regime: each item carries a
  non-empty pattern-basis field or ingest refuses. The definitions
  instruct it; this contract enforces it; nothing else does.
- Ingest is staged: outputs validate into a write-aside area, records move
  together, and the run-metadata record lands last as the commit marker —
  an orphaned stage found on the next invocation is reported and cleared,
  so a crash mid-ingest leaves evidence, never half a run.
- The manifest reference is the content hash of the assembler's manifest —
  the one unforgeable form; a reference that resolves to nothing, or to a
  manifest whose hash disagrees, refuses the run.
- Instrument identity in run metadata (ruled 2026-08-28): "instrument
  name" comprises the model identity, the definition's content hash, and
  the assembler version — two runs claiming the same instrument are
  thereby provably the same, which the closing-run comparison requires.
- Refusal granularity: an item-level violation refuses the item and lands
  the rest; a list-level violation refuses the run. A refused run still
  writes a refusal record — run metadata and the named reason, no
  items — so the event is durable and a rerun is a new run.
- Enforcement per the ruling: every check ships enforced —
  refusal is the default at birth. The degradation path is reserved, not
  pre-taken: only where a signature proves noisy in practice does that
  check degrade to record-and-flag, and the degradation is itself a
  recorded decision that weakens the claimed property from enforced to
  observed, said out loud per `loud-staging`. (2026-08-28 review: the earlier
  observed-first stance for semantic signatures pre-empted the recorded open question, and is corrected here. Standing tension with
  the repo's widen-options promotion clause — "calibrated before it
  gates" — recorded as a standing tension; the ruled design governs the instrument
  meanwhile.)

## What's Out of Scope

- The definitions that state each regime — a sibling intent.
- Whether the regime signatures lint cleanly — untested; the degradation
  path exists precisely because of it.

## Acceptance Criteria

- **Given** a reading output, **when** it is ingested, **then** the verb
  validates it before any durable record is written, and rejects malformed
  output without partial writes.
- **Given** a run, **when** its metadata is read, **then** the manifest
  reference resolves to the manifest for that run.
- **Given** a `registrative` output containing a proposed resolution,
  **when** it is ingested, **then** ingest refuses and names the item.
- **Given** an `evaluative` output carrying a rank or score field or a
  recommended marker, **when** it is ingested, **then** ingest refuses and
  names the field — document order alone is never refused.
- **Given** a run refused at list level, **when** the refusal completes,
  **then** a refusal record exists carrying the run metadata and the named
  reason, and no reading records exist for that run.
- **Given** an `explicative` output in which a surfaced claim carries a
  disposition, **when** it is ingested, **then** ingest refuses.

