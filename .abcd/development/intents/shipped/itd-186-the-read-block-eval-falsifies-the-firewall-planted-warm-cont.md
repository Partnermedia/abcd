---
id: itd-186
slug: the-read-block-eval-falsifies-the-firewall-planted-warm-cont
spec_id: spc-64
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: major
impact: additive
---

# The read-block eval falsifies the firewall — planted warm content that reaches a reading fails the build loudly

## Press Release

> **The only component capable of falsifying the blindfold rather than
> asserting it.** The eval plants sentinel warm content across every warm
> location class in a fixture repository state — a decision record, a
> wontfix reason, a framing trace, transcript-class text, an `origin`
> stamp on an included record type — and asserts its absence from the
> assembler's output. It asserts on fields, not paths: a warm field
> landing in a new place is exactly the failure a path assertion misses.
> Its oracle is independent of the assembler's include table — an eval
> that read the same table could only assert the table, never falsify it.

## What's In Scope

- The sentinel fixture state and the planted warm content, one plant per
  warm location class, maintained in the eval and never derived from the
  assembler's configuration.
- The field-level assertion over the assembler's output, wired into CI so
  the firewall is checked on every change, not per case run.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** the fixture state, **when** the eval runs, **then** it passes only if
  the assembler output contains no planted warm content and no field on the
  exclusion list.
- **Given** a ledger path moved to a new location holding a plant, **when** the
  eval runs, **then** it fails with a non-zero exit and a message naming the
  leaked sentinel's class token and the position at which it leaked.
- **Given** a warm field introduced on a record type already on the include
  list, **when** the eval runs, **then** it fails.
- **Given** a repository state containing manifests and reading records from
  prior runs, **when** the eval runs, **then** none of them appears in the
  assembler's output — the instrument's own exhaust is tested against its own
  read-block (added 2026-08-28; nothing else tests it).
- **Given** the eval package, **when** every Go file under `evals/` is parsed,
  **then** no import path names the assembler's package or its include list, so
  the oracle's exclusion table is transcribed rather than derived.
- **Given** the materialised fixture, **when** the eval runs, **then** every
  declared sentinel class is present the declared number of times, so a corpus
  that lost a plant fails rather than passing silently.

**Disclosed residue (ac-5).** The import guard holds the oracle structurally
independent of the assembler, but a hand-transcribed table can still fall behind
the exclusion list it mirrors. That staleness is bounded by the holed variant
and by ac-6, never by the import check. Prose-borne warmth inside an included
chapter carries no structural signal and stays the residue itd-183 records.


## Grounds

- pursued: Every other component in the workstream asserts the blindfold; this one is the only component capable of falsifying it, and an eval that read the assembler's own include table could only ever confirm the table rather than test the property. It is pursued now because the assembler exists to hold a firewall whose failure is otherwise silent: warm content that reaches a reading contaminates the reading invisibly, and nothing else in the cycle would notice.

## Audit Notes

<!-- abcd-review: OWED receipt=rcp-e3f354adbbee -->
Fidelity review OWED (receipt rcp-e3f354adbbee).
