# Phase 6 — Lifeboat round-trip

## Expectation

By the end of this phase — the last one — abcd closes its own loop. A user can
run `/abcd:disembark` in a project and get a faithful lifeboat: a structured
snapshot that carries the project's verbatim artefacts (specs, ADRs, memory,
reviews) *and* the synthesised why — distilled principles, a composed brief, a
press release. They can run `/abcd:embark` to unpack that lifeboat into a fresh
target, and the new repo's brief, memory, and ADRs are a faithful subset of the
source. The lifeboat is high-fidelity enough that someone with no prior context
can read it and understand not just *what* the project is but *why* it was built
the way it was. This is the heart of abcd — the surface that makes a project
survivable — and it lands last because it depends on every prior substrate being
native.

## Milestone

- `/abcd:disembark to <path>` runs end-to-end on each corpus repo, per the
  acceptance in `04-surfaces/02-disembark.md`.
- The three preview shapes work: `probe` (adapter probes only), `dry-run`
  (full plan, no writes), `to <path> --no-agents` (verbatim parts only).
- Pass A/B/C all pass their round-trip gates; Pass B retrieval reads the
  **native transcript store** (per
  [adr-29](../../decisions/adrs/0029-native-transcript-corpus.md)) rather than an
  external recorder; each pass's oracle audit is opt-in (per
  [adr-25](../../decisions/adrs/0025-host-delegated-llm-default.md)).
- `/abcd:embark from <path>` unpacks a lifeboat into an empty target per the
  acceptance in `04-surfaces/03-embark.md`; the round-trip test passes
  (disembark a corpus repo → embark into an empty target → faithful subset).
- Pass C consumes the curated memory substrate built in Phase 2 (itd-36); the
  disembark output is a curated single-repo artifact, not a promotion into a
  second repository (per
  [adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)).

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/0009-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** a corpus repo with no abcd lifeboat, **when** a user runs
  `/abcd:disembark to <path>`, **then** a lifeboat is produced that carries both
  the verbatim artefacts AND the synthesised why (distilled principles, composed
  brief, press release) — a journey across adapters, Pass A, Pass B, Pass C, and
  Phase 2's memory substrate that no single stage delivers alone.
- **Given** a project disembarked in this phase, **when** a user runs
  `/abcd:embark` into an empty target and then inspects it, **then** the new
  repo's brief, memory, and ADRs are a faithful subset of the source — the
  round-trip property that only exists because disembark and embark agree on one
  lifeboat schema.
- **Given** a produced lifeboat, **when** a reader with no prior context reads
  it, **then** they can reconstruct not just *what* the project is but *why* it
  was built that way — the emergent fidelity property that is the whole point of
  the pipeline, and that no individual pass can guarantee on its own.
- **Given** a `dev-sync` source fails, **when** disembark runs, **then** the
  pipeline degrades gracefully with the failure named in the report rather than
  aborting — an end-to-end resilience property spanning every pass.

## Scope

**Intents:** itd-7 (RepoPrompt workspace pull — carries
`.abcd/rp/workspace.json` in the lifeboat for user-account migration, the opt-in
RepoPrompt-adapter case only; presets and routing are deferred to a later phase).

**Disembark and embark are one round-trip.** Disembark packs a lifeboat; embark
unpacks it into a fresh target. They share one lifeboat schema, and the phase's
milestone is the round-trip proving the schema faithful. The output is a curated
single-repo artifact (per adr-28) — there is no dev→public mirror and no sibling
repository to promote into.

**The memory substrate is Phase 2's.** Pass C's `principle-distiller` and
`press-release-composer` consume the curated substrate `/abcd:memory` built in
Phase 2 (itd-36); this phase draws on it rather than building it. Pass B's
retrieval reads the native transcript store (adr-29).

This is the most plumbing-heavy phase — settled-artefact adapters, Pass A spine
agents, Pass B targeted retrieval over the native transcript store, and Pass C
principles/compose/audit — plus the embark unpack half of the round-trip.

## Maps against

- **Brief:** `04-surfaces/02-disembark.md` (disembark);
  `04-surfaces/03-embark.md` (embark); `05-internals/02-adapters.md` (the
  adapters); `05-internals/01-agents.md` (the Pass A/B/C agents);
  `05-internals/07-memory.md` (Phase 2's itd-36 substrate this phase consumes);
  `06-delivery/01-build-sequence.md`.
- **Intents deliver the expectation:** itd-7 delivers the RepoPrompt-workspace
  portability that makes a lifeboat carry an opt-in RepoPrompt setup across
  machines; Phase 2's itd-36 supplies the memory the composed lifeboat draws its
  synthesised content from.
- **ADRs realised:** adr-4 (lifeboat as regenerable output — the disembark/embark
  model); adr-28 (single-repo curated release — the disembark output is a
  curated artifact, not a promotion); adr-29 (native transcript corpus — Pass B's
  source); adr-1's "phase audit" feedback loop becomes exercisable here once the
  pipeline produces auditable output.

## Dependency rationale

- **Lands last** — the lifeboat round-trip depends on every prior substrate being
  native: the history and memory stores (Phase 2), the review artefacts
  (Phase 3), the native spec/task engine (Phase 4), and the run seam (Phase 5)
  all feed the verbatim-and-synthesised snapshot disembark packs.
- **Runs after Phase 1** — disembark's `dev-sync` foundation and adapter dispatch
  build on the ahoy install flow and the rules loader.
- **Adapters → Pass A → Pass B → Pass C → embark** is a hard internal chain:
  each pass consumes the previous pass's validated output through a round-trip
  gate, and embark unpacks what Pass C composes.
- **itd-7 attaches at the RepoPrompt adapter** — it carries `workspace.json` only
  when the opt-in RepoPrompt adapter is in use; it is otherwise independent of
  the disembark/embark chain and can run in parallel.

## Open questions

- Pass B's signal-density gating threshold (15%) was set from early transcript
  sampling — confirm it holds against the full validation corpus, read from the
  native transcript store, before Pass B is considered done.
- Degraded-mode behaviour when `dev-sync` fails on a source: the brief specifies
  Pass C runs degraded with notes — verify this is exercised by a test, not just
  specified.
- itd-7 ships `workspace.json` only; presets and routing are explicitly
  deferred — confirm which later phase picks them up when this phase plans.
