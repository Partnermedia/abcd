---
id: itd-181
slug: a-shipped-intent-s-scope-conditions-are-dispositioned-by-the
spec_id: spc-59
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# A shipped intent's scope conditions are dispositioned by the fidelity verdict — what was assumed ex ante and what survived are recorded as different things

## Press Release

> **The difference between what an intent assumed and what held is itself a
> finding.** When the fidelity verdict is ingested for a shipped intent,
> every scope condition — by its persistent identity, not its wording —
> receives one of four values: `survived`, `narrowed`, `falsified`,
> `untested`. A narrowing is stated, never implied by silently changed
> text. Later work inherits only what held.

## What's In Scope

- The disposition surface keyed to scope-condition identity, populated at
  verdict ingest; exercised by the fidelity verdict only for now.

## What's Out of Scope

- Passing dispositions to a reading — warm; excluded by the assembler.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a shipped intent with scope conditions, **when** the fidelity
  verdict is ingested, **then** every condition carries one of the four
  values and none is left absent.
- **Given** a narrowed condition, **when** it is recorded, **then** the
  narrowing is stated rather than implied by a changed text.

## Grounds

- pursued: Pursued now because the difference between what an intent assumed and what held is itself a finding, and it is currently unrecorded: the audit can compare promise to delivery but has nowhere to say that a condition the design leaned on turned out false. It rides with spc-55 because the identity it keys on has no other consumer, and a mechanism with no consumer is scaffolding.
