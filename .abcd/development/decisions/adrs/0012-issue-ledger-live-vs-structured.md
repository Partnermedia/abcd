---
id: adr-12
slug: issue-ledger-live-vs-structured
status: accepted
date: 2026-06-06
supersedes: null
superseded_by: null
related_intents: [itd-4, itd-46]
related_rfcs: []
related_adrs: [adr-26]
---

# ADR-12: `.work/issues.md` stays the live operational ledger; the structured per-issue store is deferred until the native spec layer schedules the migration

## Context

abcd records issues discovered during any work in `.work/issues.md` — the
operational working ledger, append-as-you-find, per the workspace's
issue-recording discipline. `.work/` is git-ignored, so this ledger is a
local-only operational surface, not a committed durable record.

Two things pull against that:

1. **A structured per-issue store is designed.** A migration spec plans the move
   of the flat `.work/issues.md` ledger into structured per-issue files under
   `.abcd/development/activity/issues/{open,resolved,wontfix}/iss-*.md`, with
   `iss-N` IDs and YAML frontmatter — the substrate the `/abcd:capture` promote
   bridge (itd-46) and the `PR000`/`PR001`/`PR002` promote-backlink lint family
   already reference.

2. **AC evidence routed only to `.work/` is invisible to PR reviewers** (the
   gitignore makes the ledger unreviewable from a diff). That tension is what
   makes "is `.work/issues.md` the durable record, or a staging surface?" a real
   decision rather than a content edit.

The spc-33 Phase-3→4 cleanup sweep (I13) forced the question to a head: either
run the spc-22 migration now, or record a durable decision that the flat ledger
stays the live operational surface — with the durable record being a committed
ADR, *not* `.work/issues.md` itself (which, being git-ignored, can never be the
checker-assertable artefact).

**Current-state evidence:**

- The structured-ledger root **does not exist**:
  `.abcd/development/activity/issues/` is absent (no `activity/` tree at all).
  Nothing under the structured path is populated — the migration has not run.
- The migration spec **is designed but unrun**: no migration has executed and
  no `iss-*.md` files have been produced.
- The flat ledger **is live and in active use**:
  `.work/issues.md` carries current-session operational entries.

So the structured path is *designed but empty*, and the flat path is *the only
ledger actually carrying issues today*.

## Decision

We will keep **`.work/issues.md` as the live operational issue ledger**, and
**defer the structured per-issue migration until the native spec layer schedules
it** ([adr-26](0026-native-spec-layer-ccpm-backend.md)).

The deferral trigger is concrete, not vague: the structured-ledger cutover
(`/abcd:capture → iss-N` plus the `.abcd/development/activity/issues/` store)
happens **when, and only when, a future native spec explicitly schedules the
`.work/issues.md` → `iss-N` migration** into a phase and runs it. Absent that
named trigger, the flat ledger remains canonical for operational issue capture.

This ADR is the durable, checker-assertable record of that choice. It does
**not** legitimise `.work/` as a durable routing target: `.work/issues.md`
remains an *operational* working surface only. Every future **routed**
follow-up — anything that must survive in the committed record — still needs a
committed home (`itd-*` intent, `iss-*` structured issue once the migration
runs, or an ADR). The decision here is narrow: the *operational live-capture
ledger* stays flat until the migration is scheduled; durable routing is
unchanged.

## Alternatives Considered

1. **Run the structured migration now.** Build the structured
   `.abcd/development/activity/issues/` store, migrate every `.work/issues.md`
   entry to `iss-N` files, and wire the live ledger to the structured path.
   Rejected: it is a heavyweight, multi-task migration (legacy parse + `iss-N`
   assignment + promote-bridge integration + lint-corpus wiring) that warrants
   its own plan-review rather than being smuggled into a cleanup sweep. The
   structured store being *empty today* (evidence above) means nothing is
   currently broken by deferring — there is no half-migrated state to rescue.

2. **Record the decision as a free-form `.work/` note.** Rejected: `.work/` is
   git-ignored, so a note there cannot be the durable, checker-assertable
   record. The whole point of this decision is that the *durable* artefact must
   be committed. A `.work/` note would re-commit the exact failure mode the
   decision is resolving (treating a git-ignored surface as the system of
   record).

3. **Vague deferral — "structured store deferred, TBD."** Rejected: a deferral
   with no named trigger is indistinguishable from indefinite drift. The
   discipline here requires a concrete re-entry condition; "until spc-22 is
   re-planned" is that condition.

## Consequences

**Gains:**
- The live issue-capture workflow is unchanged and unblocked — contributors and
  agents keep appending to `.work/issues.md` with no migration friction.
- The decision is now committed and checker-assertable: this ADR (not a
  git-ignored file) is the artefact a reviewer or lint can point to for "why is
  the issue ledger still flat?"
- No half-migrated structured store exists to maintain — deferral keeps the
  empty `activity/issues/` path from becoming a second, partial ledger
  (the split-ledger failure mode).

**Costs / obligations:**
- AC evidence and routed follow-ups recorded *only* in `.work/issues.md` remain
  invisible to PR reviewers (the gitignore). The obligation stands: any
  follow-up that must be durable still earns a committed home (`itd-*` / ADR /
  `iss-*` once spc-22 runs) — this ADR does not relax that.
- The promote-bridge lint family (`PR000`/`PR001`/`PR002`, a Go implementation)
  and the `/abcd:capture promote` path reference the structured `iss-*` store,
  which stays empty until the migration runs; those surfaces are designed-ahead
  but not yet exercised against a populated ledger.

**Downstream decisions enabled:**
- The native spec layer scheduling the migration is the single, named re-entry
  point for the flat-→-structured cutover. When a future spec schedules it, this
  ADR is superseded by the record of that migration decision.
