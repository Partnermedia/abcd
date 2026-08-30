---
id: itd-182
slug: the-record-s-discipline-failures-are-themselves-recorded-a-l
spec_id: spc-60
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# The record's discipline failures are themselves recorded — a lapse is a capture category, timestamped at the lapse

## Press Release

> **When the recording discipline is suspended, deferred, or evaded, that
> is captured too.** A `lapse` entry carries the point in the process at
> which the discipline gave way and a timestamp at the lapse rather than at
> write-up. The log is not merely a disclosure obligation: the working
> claim is that recording at the point of commitment prevents retrospective
> reconstruction, and the lapse log is the evidence bearing on that claim.

## What's In Scope

- The `lapse` value in capture's validated category list.
- The first three entries, written at the outset rather than discovered:
  the pre-tooling window (which entries were hand-authored before their
  surfaces existed); anticipation (those populating the record know what
  the readings will look for, and the instrument is specified alongside the
  record it will read); any commitment made outside the tooling during the
  build.

## Acceptance Criteria

- **Given** a lapse, **when** it is captured, **then** the entry carries
  the category, the point in the process at which the discipline was
  suspended, deferred or evaded, and a timestamp at the lapse rather than
  at write-up.

