---
id: itd-188
slug: machine-assistance-in-maintaining-the-ledger-without-any-con
spec_id: spc-66
kind: standalone
suggested_kind: standalone
reclassification_history: []
builds_on: []
severity: minor
impact: additive
---

# Machine assistance in maintaining the ledger, without any context that holds both ledger content and a reading

## Press Release

> **The scribe sees the ledger and nothing else.** Its access rule is the
> assembler's inverse: it reads ledger content, transcribes reading outputs
> and researcher dispositions, and authors nothing — and it never receives
> the shipped repository as an object of judgement. No session holds both a
> reading and the ledger; each is a distinct retained session, and session
> retention can show it.

## What's In Scope

- The scribe definition with the inverse access rule, and (when the verb
  lands) its ingest path.
- The fidelity-flag permission: the scribe may flag an internal
  inconsistency in what it is transcribing ("this disposition contradicts
  the ruling recorded earlier in the session") — that is transcription
  fidelity, not judgement, and it does not breach authors-nothing. What it
  may never do is propose a resolution.
- The contribution stamp: anything the scribe is explicitly asked to
  produce beyond formatting opens with a stamped attribution that travels
  with the material if adopted — the hand-run precursor of the `origin` /
  `scribe-transcribed` keys, in force until those keys ship. An unstamped
  contribution is never delivered.
- The protocol until then, documented and followed: entries are transcribed
  when the reading returns, not later — a protocol invented under time
  pressure is a protocol that gets skipped.

## Acceptance Criteria

- **Given** a scribe invocation, **when** its context is assembled,
  **then** it contains ledger content and no shipped-tree material.
- **Given** a reading run and a scribe run, **when** session retention is
  inspected, **then** each is a distinct retained session and no session
  holds both.

