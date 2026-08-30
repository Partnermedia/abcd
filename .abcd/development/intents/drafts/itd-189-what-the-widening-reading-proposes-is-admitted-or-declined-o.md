---
id: itd-189
slug: what-the-widening-reading-proposes-is-admitted-or-declined-o
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# What the widening reading proposes is admitted or declined on the record — grounds on admission, dispositions on declines, and surprises as their own entries

## Press Release

> **Declining a proposal costs nothing epistemically; admitting one is
> where the frame is engaged.** When the widening reading proposes
> candidate framings, every admission into the candidate set carries
> recorded grounds, and every declined proposal carries a disposition — so
> the record can show that not every contributed proposal was taken, since
> uniform adoption is equally consistent with abdication. Separately, a
> surprise entry records what was unexpected, distinct from the
> disposition record, because the reading's output and the researcher's
> response are different acts — and the surprise that occasions abduction
> is a third thing again.

## What's In Scope

- The admission-grounds record: keyed to the proposal, grounds free text,
  written at admission (the analogy of the `disposition_grounds` required on
  rejection).
- The declined-proposal disposition: keyed to the proposal, on the ledger
  side.
- The surprise entry: a distinct record shape, keyed to whatever
  occasioned it (a detection, a consequence), never folded into a
  disposition.
- Schema now; **[HAND]** in Iteration 1 (no reading runs, so nothing to
  record yet); enforced at the command in Iteration 2.

## Acceptance Criteria

- **Given** a widening proposal admitted to the candidate set, **when**
  the admission is recorded, **then** it carries grounds, and a blank is
  refused once command enforcement lands.
- **Given** a declined widening proposal, **when** the session ends,
  **then** its disposition exists on the ledger side.
- **Given** a surprise, **when** it is recorded, **then** the surprise
  entry is a distinct record from any disposition it relates to.

