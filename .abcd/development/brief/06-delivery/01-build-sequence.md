# Build Sequence

The build follows the delivery order **MVP → the companion harness → Claude Code**, with
**install + launch first**. It is a dependency DAG, not a linear list: the
autonomous `run` seam picks ready work up in dependency order, with parallelism
where dependencies allow. The canonical intent set and its bundling into product
phases live in the phase docs and the intent index — see
[`roadmap/phases/README.md`](../../roadmap/phases/README.md) and
[`intents/README.md`](../../intents/README.md); this file is the
canonical **build-milestone** detail (what each milestone stands up in the Go
core, the adapters, and the front doors).

abcd ships as a Go binary ([adr-21](../../decisions/adrs/0021-rebuild-in-go.md))
with a transport-agnostic core
([adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)) behind thin front
doors, and **no external tool as a hard dependency**
([adr-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md)). The
first milestone is **install + launch**; every capability after it is a native
default behind an already-wired interface, with an optional external adapter.

## 0. Go scaffold + core/adapter skeleton

Stand up the module layout before any command logic:
`cmd/abcd/main.go`; `internal/core/` (one package per capability);
`internal/adapter/` (the five seams — oracle | history | spec | run | scanner —
each an interface plus a stub native default); `internal/registry`;
`internal/surface/cli` (Cobra). The plugin manifest
(`.claude-plugin/plugin.json`, `marketplace.json`) and the markdown surfaces
(`commands/`, `agents/`, `skills/`) that shell to the binary load cleanly. This
locks the interface seams on day one, so every later milestone fills a native
default behind an interface the core already depends on.

## 1. Install (`ahoy`) + launch (curated release)

The first user-visible milestone, proving the CLI front door reaches the core and
the packaging boundary holds.

- **Install** — `/abcd:ahoy` end-to-end: `abcd init`, `abcd config get|set`, the
  visibility-driven gitignore policy, the CLAUDE.md/AGENTS.md marker block + rules
  loader (itd-3), and the user-scope history store + workspace registry bootstrap.
- **Launch** — `/abcd:launch` cuts a **curated GitHub Release** from the single
  repo ([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md));
  packaging **excludes `.abcd/**` from the release artifact** (a dry-run proves
  nothing under `.abcd/` leaks). There is **no dev→public mirror** — the repo is
  the marketplace.

## 2. Native history, capture, memory

- **history seam** — the native local redacted transcript store
  ([adr-29](../../decisions/adrs/0029-native-transcript-corpus.md)): root-SHA-keyed,
  gitignored, redacted on capture (reusing the two-stage redaction model of
  adr-6).
  specstory is an opt-in import over the same store. This is the research and
  benchmark corpus abcd studies its own flows against.
- **capture** — `/abcd:capture` issue ledger (itd-4) into
  `.abcd/development/activity/issues/`.
- **memory** — the `.abcd/memory/` curated substrate (itd-36); vendor memory
  harvest is an opt-in, read-only source over it.

## 3. Intent + brief + review via host-delegated oracle (+ MCP front door)

- **intent** — `/abcd:intent` create / plan / ship / grill (itd-1, itd-27,
  itd-34), with brief and press-release composition.
- **review** — the oracle seam, **host-delegated by default**
  ([adr-25](../../decisions/adrs/0025-host-delegated-llm-default.md)): abcd emits a
  prompt, the host's subagent dispatch runs it, abcd consumes the structured
  result. Native / CLI / API / MCP oracle adapters are opt-in for an operator who
  wants abcd to reach a model directly; the default install needs no API keys.
- **MCP front door** — enable `internal/surface/mcp` as a second thin door over
  the unchanged core
  ([adr-23](../../decisions/adrs/0023-transport-agnostic-core.md)).

## 4. Native minimal spec engine + ccpm

- **spec seam** — the native minimal store
  ([adr-26](../../decisions/adrs/0026-native-spec-layer-ccpm-backend.md)):
  directory-as-truth ([adr-3](../../decisions/adrs/0003-directory-as-truth-for-lifecycle.md))
  plus a dependency graph over specs and tasks — enough to plan, sequence, and
  track work with no external tool.
- **ccpm backend** — the companion harness `ccpm` as the primary deeper backend, read and
  written at the **convention level**
  ([adr-24](../../decisions/adrs/0024-the companion harness-peer-via-conventions-and-mcp.md)) —
  a peer over conventions + MCP, never a code dependency. **flow-next is not
  built.**

## 5. Autonomous run seam

The `run` seam
([adr-27](../../decisions/adrs/0027-autonomous-run-pluggable-seam.md)): iterate
ready work, gate each step on a **receipt**, enforce a **safety guard**. The thin
native Go loop is the always-available fallback; Claude Workflows and the the companion harness
agent loop are opt-in adapter loops behind the same seam contract. It is **not a
Ralph port** — the receipt-gated, report-not-block iteration boundary is the seam
contract every adapter loop inherits.

## 6. Lifeboat round-trip

- **disembark** — the lifeboat pipeline (Pass A/B/C) reads the repo's own settled
  artefacts through the source readers
  ([`../05-internals/02-adapters.md`](../05-internals/02-adapters.md)) over the
  native spec / history / memory stores, synthesises the lifeboat, and runs the
  host-delegated oracle audit.
- **embark** — scaffold a target repo from a lifeboat.
- The **round-trip** (disembark on a corpus repo → embark into an empty target)
  is the integration milestone that exercises every seam end-to-end.

## Validation cadence

After **every milestone**, run `/abcd:disembark to home` (or the relevant preview
sub-verb — `probe` for adapter-only inspection, `dry-run` for the full plan
without writes) against the validation corpus. Catch regressions early. Acceptance
recorded in `.abcd/logbook/phase/<phase-id>/` (per the logbook layout in
[`../05-internals/04-universal-patterns.md § 6`](../05-internals/04-universal-patterns.md#6-abcdlogbook-layout)).
