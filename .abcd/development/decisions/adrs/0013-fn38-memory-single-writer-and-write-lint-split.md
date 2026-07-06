---
id: adr-13
slug: fn38-memory-single-writer-and-write-lint-split
status: accepted
date: 2026-06-09
supersedes: null
superseded_by: null
related_intents: [itd-36]
related_rfcs: []
related_adrs: [adr-8]
---

# ADR-13: fn-38 memory — single-writer crash model + write/lint spec split

## Context

fn-38 (`abcd-memory-unification`) plan review failed to converge: it ran ~23
RepoPrompt re-review rounds (rounds 3–25) and was killed by the 2-hour worker
timeout without ever producing a `ship` verdict. Each round resolved the prior
round's objection but the reviewer surfaced a new, deeper concurrency /
crash-consistency edge case — lost-update races between transactional and
non-transactional registry writers, `txn_id` ownership of `pending_pages`
markers, crash-recovery reconciliation colliding with the SD001 "bare command is
non-mutating" contract.

Two causes, one acute and one chronic:

- **Acute:** the plan had accreted a near-ACID transaction + crash-recovery
  protocol (txn ~955, crash ~1825, pending-reconciliation ~2629 mentions) for a
  memory store built on **markdown files** (`.abcd/memory/index.md`, `log.md`,
  per-page files, `.sources_index.json`). That protocol only needs to exist if
  concurrent writers can race — a premise never validated.
- **Chronic:** fn-38's task bodies are ~41.7k words — more than double fn-37
  (~20k, which itself took 14 rounds). A 44k-word plan exceeds what an
  adversarial line-by-line reviewer can clear inside one 2-hour budget; it will
  always find "one more thing" first.

## Decision

A grill-me interview resolved the design tree:

1. **Single-writer to the memory store.** The memory writer touches only
   `.abcd/memory/**`. Human activities that run concurrently with Ralph
   (facilitator refining `.abcd/development/brief/**`, product thinker amending
   `.abcd/intents/**`, issue capture into `.work/issues.md`) write **disjoint
   file trees** and never mutate the memory store. Ralph itself runs one worker
   at a time. There is no real concurrent-writer scenario.

2. **Atomic-rename only (crash model A).** Each file write is
   write-temp-then-`rename()`. A crash leaves either the complete old or complete
   new file, never a torn file. No transaction journal, no `pending` markers, no
   lock-ordering protocol. (No mid-write corruption has ever been observed in
   the existing memory files — the empirical failure rate of what the transaction
   model defended against is ~zero.)

3. **No explicit crash-recovery.** With atomic rename there is no torn write to
   heal; the only residue is cross-file staleness, healed by the **idempotent
   sibling-reconciliation that already runs before every mutating write**. Bare /
   read-only commands stay strictly SD001-non-mutating and merely *report* drift
   ("index stale; run an ingest"). This dissolves the recovery-vs-SD001 tension
   at the root rather than patching it.

4. **Split along the write/lint seam.** flow-next models specs as flat
   sequential `fn-N` epics (no sub-specs), so the split is **fn-38 + fn-39**:
   - **fn-38 — memory write core:** registry/index (atomic-rename),
     ingest/atomic-promotion (simplified), the `principle-distiller` prompt,
     launch consumer + docs. The part that *stores*.
   - **fn-39 — memory quality/lint:** the MS/ML/MQ/SD constraint validation,
     citations, provenance budgets, the `lint` command. The part that *audits*.
   The retrieval engine was already excised to a separate intent (itd-39); this
   continues that decomposition.

5. **Re-plan via flow-next, not by hand.** Both specs are (re)generated through
   `/flow-next:plan` from a brief encoding decisions 1–4, so flowctl owns all
   state-writing. No hand-authored spec/task JSON (the earlier registration
   repair caused runtime-field drift that the `flow-task-definition-drift` gate
   rejected — see git history around the fn-36/37/38 registration).

6. **Triple-backend plan review, gate left open.** Review escalates
   **Opus in-session (adversarial) → Codex CLI → RP**, extending ADR-8's
   asymmetric-trust posture from dual to triple. The cheapest critic runs first.
   A believed-good plan is **not** recorded as `ship`: `plan_review_status` is
   left open so Ralph runs its own fresh plan review before implementing —
   consistent with "do not fake state flow-next should own."

## Consequences

- The entire class of objection that blocked fn-38 (lost-update, txn ownership,
  recovery-mutation) is removed by construction, not argued down round by round.
- The txn machinery is concentrated in tasks .1/.4/.6; .3/.5/.7 are already
  clean, so the simplification is surgical.
- Two ~20k-word plans, each in fn-37's proven-reviewable range, replace one
  un-reviewable 44k-word plan. The write core (the actually-blocking piece) can
  ship first.
- Risk: re-planning may not perfectly preserve good parts of the existing 44k of
  hardened content. Mitigated by feeding the existing plan in as source material
  to the re-plan, and by the triple-backend review before Ralph's own gate.

## Related Documentation

- [ADR-8: Dual-Backend Review with Asymmetric Trust](0008-dual-backend-review-asymmetric-trust.md)
- Intent: `.abcd/development/roadmap/intents/drafts/itd-36-memory-unification.md`
