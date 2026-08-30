---
id: itd-179
slug: the-reasoning-behind-what-was-pursued-no-longer-evaporates-a
spec_id: spc-57
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# The reasoning behind what was pursued no longer evaporates at the gate — readiness and triage record grounds for the conjecture, not only the decision

## Press Release

> **abcd records why a thing was pursued, at the moment it is pursued.**
> Grounds were recorded for deliberate non-action only — a wontfix carries a
> note, an ADR carries its alternatives — while the reasoning behind what
> went forward evaporated at the gate. Now the readiness gate and capture's
> triage routes take a grounds argument with a small vocabulary — `pursued`,
> `deferred`, `declined` — and the grounds name the conjecture being acted
> on, not merely the route taken. A triage without grounds is refused.

## What's In Scope

- A grounds argument on `intent ready` and on the capture triage routes
  (promote / resolve / wontfix), refusal on absence.
- The three-value vocabulary, with the grounds text free-form and naming the
  conjecture.

## What's Out of Scope

- Rewriting the ADR family's grounds — decision-granularity grounds stay
  where they are; this intent adds the finer grain beside them.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a capture routed to an intent draft, **when** triage runs
  without grounds, **then** the command refuses.
- **Given** a gate decision, **when** it is recorded, **then** the grounds
  name the conjecture, not only the decision.

