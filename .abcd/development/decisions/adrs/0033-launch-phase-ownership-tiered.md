---
id: adr-33
slug: launch-phase-ownership-tiered
status: accepted
date: 2026-07-08
supersedes: null
superseded_by: null
related_intents: [itd-65, itd-66, itd-70, itd-72, itd-73]
related_rfcs: []
related_adrs: [adr-9, adr-28, adr-31]
---

# ADR-33: Launch phase ownership is tiered; the phase index is the sole ownership source

## Context

The record disagreed about which phase owns `/abcd:launch`. The roadmap
(`phases/phase-1-ahoy.md`, `phases/README.md`) and the brief's build sequence
(`06-delivery/01-build-sequence.md`) make **install + launch the first
milestone** — Phase 1 cuts a curated single-repo release. But
`brief/04-surfaces/04-launch.md` opened with "builds in Phase 5 (round-trip and
ship)", and `05-internals/01-agents.md` anchored launch agents to "Phase 5".

"Phase 5 (round-trip and ship)" is the **pre-rebuild six-phase numbering**, in
which the final phase bundled the lifeboat round-trip *and* the release flow.
The current roadmap has seven phases (0–6): the round-trip is Phase 6, and the
release cut moved forward to Phase 1 as the MVP proof of the packaging path. The
stale banner survived the copy of the design record; an implementer reading only
the brief would schedule launch five phases late. (Captured as `iss-1`; the
sibling six-vs-seven-phase and disembark/embark-banner findings are `iss-4` and
`iss-5`.)

There is also a real distinction the stale banner was gesturing at: the *full*
launch flow (pre-flight gate suite, render parity, tier-b publishing, retention,
derived versioning) is deliberately not Phase 1 work.

## Decision

1. **The roadmap phase index is the sole source of phase numbering and phase
   ownership.** `roadmap/phases/README.md` (Phases 0–6) and each phase doc's
   `## Scope` own the mapping (per adr-9). A brief file states phase ownership
   only by pointing at the owning phase doc — never with its own phase number
   that can drift. The six-phase "Phase 0–5, round-trip-and-ship last"
   numbering is superseded.

2. **Launch ownership is tiered.** **Phase 1 owns the curated-release cut**:
   packaging from the one repo with `.abcd/**` excluded, the secret/PII scan
   gate, and the bootstrap release path (adr-28). The **launch deepenings are
   separately scheduled intents** — the full pre-flight gate suite (itd-65),
   payload render parity (itd-66), release retention (itd-70), tier-b
   publishing (itd-72), and derived versioning (itd-73) — each entering a
   phase's `## Scope` when sequenced, unscheduled until then (adr-9's
   unscheduled-intent rule).

3. **Deepened behaviour is attributed to its intent, not a phase number.**
   Where `04-surfaces/04-launch.md` describes behaviour beyond the Phase 1 cut,
   it names the owning intent (e.g. "the gate suite is itd-65's"), so the doc
   stays correct however those intents are sequenced.

## Alternatives Considered

- **Keep launch whole and late (the banner's reading).** Rejected: it
  contradicts the first-milestone MVP logic — install + launch is what proves
  the Go core, adapter seams, and packaging path end to end (build sequence,
  phase-1 doc), and the packaging boundary is the highest-risk privacy surface,
  so it ships first, not last.
- **Make Phase 1 own the full launch flow.** Rejected: the gate suite and
  versioning automation depend on stores and machinery from later phases (the
  native spec store for intent-impact derivation, adr-31), and pinning them to
  Phase 1 would either stall the milestone or ship them hollow.
- **Renumber the brief's banners to "Phase 6".** Rejected: it repairs the
  number but keeps duplicated phase ownership in the brief, which is the drift
  mechanism itself. Ownership statements live once, in the phase index.

## Consequences

- `04-surfaces/04-launch.md` loses its "builds in Phase 5" banner in favour of
  a phase-ownership pointer (Phase 1 cut + intent-attributed deepenings);
  in-body "Phase-5 `ship`" references become intent attributions.
- `05-internals/01-agents.md` anchors `launch-gatekeeper` to itd-65 rather than
  a phase number.
- The six-vs-seven phase count and the disembark/embark banners (`iss-4`,
  `iss-5`) now have their canonical basis: seven phases, ownership per the
  phase index.
- Future phase renumbering touches exactly one file, the phase index.
