# Scope

The scope this brief covers is bounded by what is bundled into the six planned phases (Phase 0 — Substrate & disciplines through Phase 5 — round-trip; see [`roadmap/phases/README.md`](../../roadmap/phases/README.md)) plus what is explicitly slated for a later phase.

## What the phases deliver

**User-facing commands (design target across the planned phases — not all shipped yet):**

- `/abcd:ahoy` — install/update — see [`04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)
- `/abcd:disembark` — pack a lifeboat — see [`04-surfaces/02-disembark.md`](../04-surfaces/02-disembark.md)
- `/abcd:embark` — unpack a lifeboat — see [`04-surfaces/03-embark.md`](../04-surfaces/03-embark.md)
- `/abcd:launch` — public promotion — see [`04-surfaces/04-launch.md`](../04-surfaces/04-launch.md)
- `/abcd:intent` — press-release intent capture — see [`04-surfaces/05-intent.md`](../04-surfaces/05-intent.md)
- `/abcd:capture` — issue ledger — see [`04-surfaces/06-capture.md`](../04-surfaces/06-capture.md)
- `/abcd:memory` — multi-upstream curated knowledge substrate (per itd-36) — see [`05-internals/07-memory.md`](../05-internals/07-memory.md)

**Operator-internal commands** (wiring under `commands/abcd/`, NOT part of the user-facing surface above): `deps-check`, `ralph-up`, `session`, and `/abcd:run` — the itd-29 autonomous-run operator surface (`status`/`pause`/`resume`/`preflight`; read-mostly over an autonomous Ralph run, v1 never starts or kills the loop). See [`04-surfaces/README.md`](../04-surfaces/README.md) for the user-facing-vs-operator-internal boundary.

**Thirteen phased intents — ten standalone plus three disciplines:**
- **Standalone (10):** itd-2, itd-3, itd-4, itd-6, itd-7, itd-27, itd-28, itd-34, itd-36, itd-40.
- **Disciplines (3):** itd-1 (acceptance gates), itd-5 (prompt-quality + capability_scope per idea-4), itd-37 (modification grammar per ideas 2+3). Disciplines have no user moment; they impose acceptance gates on every other spec per the three-kinds taxonomy in [`01-product/03-mental-model.md`](03-mental-model.md) and itd-34.

See [`phases/README.md`](../../roadmap/phases/README.md) for the phase plan and each phase's intent scope, and [`intents/README.md`](../../intents/README.md) for the intent index. itd-27 (`/abcd:intent grill` sub-verb), itd-28 (spec-tied RP reviews), and itd-34 (three intent kinds) were captured post-brief on 2026-05-07. itd-36 (memory unification) and itd-37 (modification grammar) were captured on 2026-05-08 following adversarial RP review of four candidate ideas (LLM Wiki, Naur theory-building, systems thinking, jagged frontier); itd-40 (folder classification) was captured 2026-05-16.

**Plumbing infrastructure** (15 agents, 11 adapters, harness shim, prompt-quality stack, hooks): see [`05-internals/`](../05-internals).

## What comes in a later phase

**All later-phase items live as press-release intents** in `.abcd/development/roadmap/intents/drafts/`. The canonical out-of-scope list is at [`06-delivery/03-out-of-scope.md`](../06-delivery/03-out-of-scope.md).

The later-phase set is the live `drafts/` corpus minus the intents already scoped into a planned phase; it is enumerated — and kept non-drifting via a filesystem-derived command rather than a hand-count — in the canonical [`06-delivery/03-out-of-scope.md`](../06-delivery/03-out-of-scope.md). itd-31 and itd-32 are superseded (preserved as historical record in `intents/superseded/`).

Each intent captures the press-release-shaped scope and acceptance criteria. A later-phase intent enters work by being scoped into a phase, then promoted to `planned/` via `/abcd:intent plan <itd-N>` and to `shipped/` via `/abcd:intent ship <itd-N>` (or automatically when the linked spec closes).

The brief does not get re-versioned. What has shipped is defined by which phases are complete and which intents are in `shipped/`; this brief stays the canonical current-state design record.
