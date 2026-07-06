# Phase 4 — The lifeboat pipeline

## Expectation

By the end of this phase, a user can run `/abcd:disembark` in a project and get
a faithful lifeboat: a structured snapshot that carries the project's verbatim
artefacts (specs, ADRs, memory, reviews) *and* the synthesised why — distilled
principles, a composed brief, a press release. The lifeboat is high-fidelity
enough that someone with no prior context can read it and understand not just
*what* the project is but *why* it was built the way it was. This is the heart
of abcd: the surface that makes a project survivable.

## Milestone

- `/abcd:disembark to <path>` runs end-to-end on each corpus repo, per the
  acceptance in `04-surfaces/02-disembark.md`.
- The three preview shapes work: `probe` (adapter probes only), `dry-run`
  (full plan, no writes), `to <path> --no-agents` (verbatim parts only).
- Pass A/B/C all pass their round-trip gates and oracle audits.
- `/abcd:memory` ingests external sources into the curated substrate that
  Pass C consumes.

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/0009-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** a corpus repo with no abcd lifeboat, **when** a user runs
  `/abcd:disembark to <path>`, **then** a lifeboat is produced that carries both
  the verbatim artefacts AND the synthesised why (distilled principles,
  composed brief, press release) — a journey across adapters, Pass A, Pass B,
  Pass C, and itd-36's memory substrate that no single stage delivers alone.
- **Given** a produced lifeboat, **when** a reader with no prior context reads
  it, **then** they can reconstruct not just *what* the project is but *why* it
  was built that way — the emergent fidelity property that is the whole point
  of the pipeline, and that no individual pass can guarantee on its own.
- **Given** a `dev-sync` source fails, **when** disembark runs, **then** the
  pipeline degrades gracefully with the failure named in the report rather than
  aborting — an end-to-end resilience property spanning every pass.

## Scope

**Intents:** itd-36 (`/abcd:memory` unification — multi-upstream curated
knowledge substrate; Pass C's `principle-distiller` and `press-release-composer`
consume `.abcd/memory/`).

**Brief plumbing-phases:** Phase 2 — settled-artefact adapters; Phase 3 — Pass A
spine agents; Phase 4 — Pass B targeted chat retrieval; Phase 5 — Pass C
principles/compose/audit + `/abcd:disembark to <path>` end-to-end.

This is the most plumbing-heavy phase — four brief phases of agent and adapter
work — with a single intent (itd-36) supporting it.

## Maps against

- **Brief:** `04-surfaces/02-disembark.md` (the command);
  `05-internals/02-adapters.md` (the 11 adapters); `05-internals/01-agents.md`
  (the Pass A/B/C agents); `06-delivery/01-build-sequence.md` Phases 2–5;
  `05-internals/07-memory.md` (itd-36's component spec).
- **Intents deliver the expectation:** itd-36 delivers the memory substrate the
  composed lifeboat draws its synthesised content from.
- **ADRs realised:** adr-4 (lifeboat as regenerable output — the disembark
  output model); adr-1's "phase audit" feedback loop becomes exercisable here
  once the pipeline produces auditable output.

## Dependency rationale

- **Runs after Phase 1** — disembark's `dev-sync` foundation and adapter
  dispatch build on the ahoy install flow and the rules loader. (Phases 2 and
  3 — capture and intent — are independent of the lifeboat pipeline; this
  phase need not wait on them, only on Phase 1's install surface.)
- **Adapters → Pass A → Pass B → Pass C** is a hard internal chain: each pass
  consumes the previous pass's validated output through a round-trip gate.
- **itd-36 before Pass C** — Pass C agents consume `.abcd/memory/`; the memory
  substrate must exist before the compose step runs. itd-36 is otherwise
  independent and can run in parallel with the adapter and Pass A/B work.
- **The oracle cascade (Phases 0–1) must be whole** — every pass ends with an
  oracle audit; a missing cascade leg would block the pipeline.

## Open questions

- Pass B's signal-density gating threshold (15%) was set from Phase 0
  transcript sampling — confirm it holds against the full validation corpus
  before Pass B is considered done.
- Degraded-mode behaviour when `dev-sync` fails on a source: the brief
  specifies Pass C runs degraded with notes — verify this is exercised by a
  test, not just specified.
