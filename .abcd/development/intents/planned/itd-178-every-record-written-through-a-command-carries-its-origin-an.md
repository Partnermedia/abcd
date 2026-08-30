---
id: itd-178
slug: every-record-written-through-a-command-carries-its-origin-an
spec_id: spc-56
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Every record written through a command carries its origin and its production mode — stamped by the command, never typed by hand

## Press Release

> **A record now says where each of its items came from and how its text was
> produced.** `origin` names the arrival path — `researcher-authored`,
> `contributed-by-reading` (carrying the run and item identifiers that
> resolve to a reading record), or `extracted-from-record`. The production
> mode distinguishes `hand-written` from `dictated-and-formatted` from
> `scribe-transcribed`. Both are frontmatter keys written only by commands:
> no flag carries them as free text, hand editing is caught by the lint, and
> neither key touches authorship — disclosure at field granularity, on the
> same footing as the Assisted-by trailer at commit granularity.

## Alignment

- The three-term production-mode vocabulary is the ruled design's (the ruled authorship account); this intent ships the mechanism that stamps it, not
  the vocabulary decision.
- The attribution config seam to consult is itd-91's
  (`.abcd/config/identity.json`) — extend, never duplicate.
- Population is forward-only, per the ruled population properties: existing records are untouched, sparseness is information, and an
  absent stamp is never backfilled.

## What's In Scope

- The two keys in the record schemas, resolver support, and the
  command-side stamping on every write path that mints or mutates a record.
- The lint that reports a hand-edited record carrying either key.

## What's Out of Scope

- Passing either key to a reading — both are excluded by the input
  assembler's field projection, and the read-block eval asserts it.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a record written through a command, **when** it is committed,
  **then** both keys are present and neither was supplied as free text by
  the operator.
- **Given** a record with `origin: contributed-by-reading`, **when** it is
  read, **then** the item identifier and run identifier resolve to a
  reading record.
- **Given** a hand-edited record carrying either key, **when** the lint
  runs, **then** the lint reports it.

