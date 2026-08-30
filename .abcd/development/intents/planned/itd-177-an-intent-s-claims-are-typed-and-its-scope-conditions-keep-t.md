---
id: itd-177
slug: an-intent-s-claims-are-typed-and-its-scope-conditions-keep-t
spec_id: spc-55
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: major
impact: additive
---

# An intent's claims are typed and its scope conditions keep their identity across edits — the readiness gate refuses an intent that leaves a context claim unrecorded

Typed links: consumes [adr-51](../../decisions/adrs/0051-intents-declare-mechanism-and-scope-conditions.md)
(the sections exist as a format; this intent is the anticipated enforcement);
gated by the `claim-recording-gradient` discipline; identities minted on the
adr-45 mint.

## Press Release

> **An intent now says what kind of claim each of its sections carries, and
> its scope conditions survive their own rewording.** Criteria were always
> mandatory; the mechanism claim is prompted and may be declined with a
> recorded nullity; scope conditions are required — or their absence is
> declared, explicitly, as "none stated". Each scope condition carries a
> persistent identity, so when its text is edited the condition is still the
> same condition, and the disposition that later attaches to it (survived,
> narrowed, falsified, untested) attaches to the claim rather than to a
> sentence that no longer exists.

## What's In Scope

- Schema first: the three claim types and the per-condition identity, in the
  intent record shape, before any command enforces them.
- Command second: the readiness-gate refusals per the gradient.
- The nullity forms: one token, `None stated.`, alone on its line under the
  section heading — the same grammar for scope conditions and for a
  declined mechanism. An absent section, an empty section, and a recorded
  nullity are three distinct byte states with three semantics: a claim not
  carried; a gate fault; a claim considered and declined.
- Condition identity: each scope-condition bullet closes with a stamped
  identity marker minted on the adr-45 mint, written by `intent plan` (the
  lifecycle's write-capable verb) — never hand-typed. `abcd intent --json`
  renders conditions with their identities, which is the observable surface
  the identity criteria assert against; the gate refuses a duplicated or
  missing marker by name.
- Identity lifecycle: an edit keeps the marker; a split keeps the marker on
  the first part and mints for the second; a merge keeps the surviving
  marker and retires the other (surfaced as `narrowed` at the next fidelity
  verdict); a deletion retires the marker, reported at the next gate pass.
- "Prompted" has a named surface: the create-path scaffold comment and the
  planning interview prompt for the mechanism claim; the readiness gate
  only checks and reports.

## What's Out of Scope

- The gradient's rationale and staging — the discipline record owns it.
- Scope-condition dispositions — a separate intent (scope-condition disposition); this one only makes
  them attachable.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** an intent without context conditions or an explicit nullity,
  **when** the readiness gate runs, **then** the gate exits non-zero and
  names the missing field.
- **Given** a scope condition whose text is edited, **when** the intent is
  re-read, **then** the condition's identity is unchanged.
- **Given** an intent whose `## Mechanism` section carries the nullity
  token, **when** the readiness gate runs, **then** the gate passes and
  reports the recorded nullity.
- **Given** an intent whose `## Mechanism` section is present but empty,
  **when** the readiness gate runs, **then** the gate exits non-zero and
  names the section: write the claim or the nullity token.
- **Given** a scope condition with a duplicated or missing identity marker,
  **when** the readiness gate runs, **then** the gate exits non-zero and
  names the fault.

