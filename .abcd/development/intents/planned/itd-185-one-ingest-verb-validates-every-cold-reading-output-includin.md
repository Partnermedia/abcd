---
id: itd-185
slug: one-ingest-verb-validates-every-cold-reading-output-includin
spec_id: spc-63
kind: bundle-member
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

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a malformed reading output, **when** it is ingested, **then** ingest
  refuses and names the offending field, and no durable record exists for that
  run in the reading-record family, in the readings tree, or in the stage.
- **Given** a fault injected after the reading records are staged and before the
  run-metadata commit marker is written, **when** ingest runs, **then** no
  reading records and no run metadata are durable for that run, and the next
  invocation names the orphaned stage and clears it.
- **Given** a run, **when** its metadata is read, **then** the manifest
  reference resolves to that run's manifest — the stored hash equalling the
  content hash of the manifest itself — and a reference that resolves to
  nothing, or to a manifest whose hash disagrees, refuses the run.
- **Given** a `registrative` output whose item carries a reserved name
  (`resolution`, `fix`, `remedy`), **when** it is ingested, **then** ingest
  refuses and names the item's ordinal, the field, and the licence breached.
- **Given** a `registrative` output whose item body matches a registered
  fix-proposal signature, **when** it is ingested, **then** ingest refuses and
  names the item and the signature id.
- **Given** an `evaluative` output carrying a `rank`, `score`, `order` or
  `recommended` field, **when** it is ingested, **then** ingest refuses and
  names the field.
- **Given** an `evaluative` output whose items are merely arranged in an order
  and carry no reserved field, **when** it is ingested, **then** ingest accepts
  it: arrangement order is never inspected and never refused.
- **Given** an `explicative` output in which a surfaced claim carries a
  disposition-bearing field — `disposition`, `status`, or any field outside the
  explicative body schema — **when** it is ingested, **then** ingest refuses and
  names the field.
- **Given** an `explicative` output whose claim body matches a registered
  disposition signature, **when** it is ingested, **then** ingest refuses and
  names the item and the signature id.
- **Given** a run refused at list level, **when** the refusal completes, **then**
  a refusal record exists carrying the run metadata and the named reason, and no
  reading records exist for that run.
- **Given** an item at any of the four regimes whose `pattern_named` envelope
  field is empty or absent, **when** it is ingested, **then** ingest refuses that
  item, without exception at any regime.
- **Given** an output whose self-declared regime disagrees with the regime stated
  by the definition its position resolves to, **when** it is ingested, **then**
  ingest refuses the run at list level.
- **Given** an accepted output, **when** its reading records are read, **then**
  every item carries an identifier the verb minted, and a payload supplying its
  own item identifier is refused as an unknown field.

**Disclosed residue (ac-5 and ac-9).** The two semantic criteria are enforced
over the signature registry, not over the space of things a sentence can do, and
the residue has two parts.

The first is the one this intent has always disclosed: a fix proposal or a
disposition **phrased** outside the registry's signatures is not caught. The
registry sits in the calibration band, and whether the signatures lint cleanly
in practice is this intent's recorded open question.

The second is narrower and is named here rather than left to be discovered. The
detectors read a folded copy of the body, so an invisible or
compatibility-equivalent rune cannot decide whether a signature fires: every
Unicode space folds to ASCII, every invisible rune is dropped, and NFKC folds
the compatibility forms. A script-**confusable** substitution is not folded — a
Cyrillic that is not the Latin one, and no normalisation equates them — so a
signature's own phrasing written in a confusable script is not caught. Closing
that needs a confusables table, which is a new dependency and a maintainer's
decision; until it is taken, this class is open.

The distinction between the two parts is the same test throughout: the
registry's phrasing with a byte substituted is a defect in the gate, and
phrasing outside the registry is a limit of it. Their structural halves — ac-4,
ac-6 and ac-8 — carry no residue of either kind, because a field is present or
it is not.


## Grounds

- pursued: This conjecture is pursued now because the failure it catches is silent everywhere else: a reading that quietly proposes, ranks, or arrives already dispositioned passes every structural test while violating the one property its position is defined by, and the read-block eval covers only what a reading saw, never what it was licensed to produce. Building the gate later would mean trusting outputs produced before anything checked them.
