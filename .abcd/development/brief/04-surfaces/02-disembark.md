# `/abcd:disembark` — Pack a Lifeboat

> **Status:** design target — builds in Phase 4 (the lifeboat pipeline). Today only the probe / dry-run stubs ship (spc-17); the full flow below is not yet built.

> **Recovery humility.** Disembark packs the highest-fidelity proxy of the project's theory we can leave behind. It is not the theory. The theory of any non-trivial project lives in the people who built it, the conversations where decisions were made, and the alternatives they rejected before this one — what Naur (1985) called the lived activity of building. The lifeboat is the floor we can carry forward across a session, machine, or team boundary; it is not the activity itself. See [`01-product/03-mental-model.md § The Naurian gap`](../01-product/03-mental-model.md#the-naurian-gap) for the framing.

## Sub-verbs

Bare `/abcd:disembark` shows status + help only — never mutates state. Current sub-verbs:

- **`/abcd:disembark to <path>`** — pack a lifeboat to `<path>`. Path is required; use `home` as shorthand for the current repo's `.abcd/lifeboat/`. The flow described in § 1 below is this sub-verb's behaviour. Flag-shaped modifier: `--no-agents` (write the verbatim/deterministic parts, skip LLM synthesis).
- **`/abcd:disembark probe`** — adapter probes only: list sources that would be packed, write nothing. Ultra-light read-only inspection (no LLM dispatch).
- **`/abcd:disembark dry-run`** — full plan: what files would be written, which agents would be dispatched, estimated tokens — write nothing. Heavier than `probe` (runs the inventory + agent budget estimation), lighter than `to <path>` (no actual writes).
- **Later phase: `/abcd:disembark to-spec-kit <path>`** — export shipped intents to GitHub Spec Kit format alongside the lifeboat (per itd-23).

## 1. Architecture (Phase 0 dev-sync + three passes)

```
PHASE 0 — DEV-SYNC (see 05-internals/03-configuration.md § 2)
  abcd dev-sync runs per-source enabled flags (each source is an opt-in adapter over a native default):
    • reviews → render reviews via the wired oracle adapter (host-delegated
                default; native / CLI / API / MCP opt-in)  → .abcd/development/activity/reviews/
    • memory  → distil memory backend  (Claude / OpenCode / ...) → .abcd/memory/
    • work    → curate .work/                            → .abcd/development/activity/{issues,notes}/
    • rp      → RP workspace pull (opt-in RP adapter only, per itd-7) → .abcd/rp/workspace.json
  Idempotent (content-hash dedup).
                           │
                           ▼
PROBE & CONFIRM
  the disembark probe entrypoint (internal/core) runs all adapters' probe() in parallel
  → ActiveSources list (reads curated .abcd/development/activity/, not raw sources)
  → AskUserQuestion (transparent): confirm assets/docs sources
                           │
                           ▼
PASS A — settled artefacts (4 agents in parallel, newest-first, JSON-out)
  • flow-essence            → rescue/spec-essence.json
  • decision-archaeologist  → research/decisions-timeline.json
  • review-collator         → research/reviews-consolidated.json
  • code-rescuer*           → research/code-principles.json
                           │
                           ▼
PASS B — targeted chat fill (chat-distiller, one call per unresolved spine entry)
  • git-blame spec windows  → time-window index
  • filtered transcripts    → research/rationale-fills.json
                              research/unrecorded-decisions.json
                              delta research/pitfalls.json
                           │
                           ▼
PASS C — distil, compose, audit (sequential)
  • principle-distiller     → principles.json
  • artefact-curator        → docs/, assets/_manifest.json
  • brief-composer          → README.json
  • press-release-composer  → press-release.json
       ↳ oracle product audit → audit/press-release-oracle-<ts>.json
         (host-delegated by default; opt-in oracle adapter when wired; findings appended to press-release.md)
  • render                  → .md files from .json sources
  • lifeboat-oracle         → audit/oracle-<ts>.json
                              findings appended to README rendering
  • documentation-auditor (subagent, pre-pack) → audit/documentation-audit-<ts>.json
```

\* code-rescuer reads source via a wired oracle adapter if available, else spec-driven git-window file selection. Outputs principles only — never copies code into the lifeboat (`--with-code` comes in a later phase as itd-8).

## 2. Recency rule

Hard rule: later structural artefacts supersede earlier ones — spec numbers, ADR `Superseded-By` headers, file mtimes, git log. **Never resolve recency semantically.**

When structure and content disagree (e.g., later spec mentions earlier decision approvingly without restating it), **route to chat-distiller in Pass B**: receives the relevant time-windowed transcripts, emits a finding for the unrecorded-decisions report. (Risk: Pass B becomes load-bearing for correctness; transcript signal density resolved in Phase 0 (`research/phase/0/transcript-sampling.md`); Pass B viability gates encoded in `06-delivery/01-build-sequence.md § 7`.)

## 3. Agent context budget

Each agent estimates input tokens before dispatch:

- Under `disembark.maxAgentTokens` (default 100,000) → one shot
- Over → stream + summarise (map-reduce: per-input summary, then merge pass)

Pass B's chat-distiller is exempt (already streams by per-spine-entry queries). Phase 0 sampling measures actual sizes on the corpus before locking the strategy.

## 4. Backgrounded execution

- `disembark to <path>` launches as a background task via `harness.run_background()`.
- After each agent run, writes checkpoint to `.abcd/logbook/disembark/<timestamp>/_state.json` (forensic record only — see resume note below).
- `harness.schedule()` schedules a wake-up every ~5 minutes to surface a one-line summary ("Pass B: 12/30 spine entries resolved").
- Ctrl-c interrupts cleanly; re-running `disembark to <path>` starts fresh. **Note:** abcd does not ship a resume verb. The checkpoint exists for forensic purposes (post-mortem on a failed run); a future intent may add a `resume` sub-verb if real-world friction emerges.
- Final report at completion.

## 5. Output shape

```
.abcd/lifeboat/
├── README.md                           # rendered from README.json (brief-composer + oracle findings)
├── README.json
├── press-release.md                    # rendered from press-release.json (+ product-audit findings appended)
├── press-release.json                  # the embark interview contract
├── principles.md                       # rendered from principles.json
├── principles.json
├── rescue/
│   ├── extraction.md                   # only if --with-code (itd-8, a later phase)
│   ├── spec-plan.md                    # rendered (rebuild plan)
│   ├── spec-essence.json               # spine
│   ├── spec-essence.md                 # rendered
│   └── specs/                          # verbatim copies of the native spec store's spec files
├── research/
│   ├── decisions-timeline.{json,md}
│   ├── pitfalls.{json,md}              # source memory + Pass B deltas
│   ├── reviews-consolidated.{json,md}
│   ├── rationale-fills.{json,md}
│   ├── unrecorded-decisions.{json,md}
│   └── code-principles.{json,md}       # from code-rescuer
├── docs/
│   ├── adrs/                           # verbatim ADR copies
│   ├── terminology.md                  # rendered from .abcd/development/foundation/terminology/<context>/<term>.md (per itd-27 / spc-3)
│   ├── claude-md-snapshot.md
│   └── tutorials/ guides/ reference/ explanation/   # verbatim from source docs/
├── assets/
│   ├── logos/ charts/ screenshots/
│   └── _manifest.json                  # source path → lifeboat path, caption, provenance, classification (keep/adapt/drop)
├── audit/
│   ├── oracle-<timestamp>.{json,md}                # lifeboat-oracle (content fidelity)
│   ├── press-release-oracle-<timestamp>.{json,md}  # press-release-composer's product audit
│   └── documentation-audit-<timestamp>.{json,md}   # documentation-auditor (subagent pre-pack)
├── activity/                           # curated issue ledger snapshot (per itd-4)
│   └── issues/{open,resolved,wontfix}/
└── _provenance.json                    # current snapshot's provenance (sources probed, adapter runs, agent runs, oracle backend used, schema_version). Hash matches the manifest_sha256 in voyage/disembark/history.jsonl for this run; cumulative history lives in voyage/, not here.
```

The lifeboat is gitignored unless visibility=private (then it's committed alongside the rest of `.abcd/`).

## 6. Per-phase acceptance

Each phase passes when **both gates** succeed:

1. **Oracle gate**: `lifeboat-oracle` audit on phase outputs returns a "sufficient" verdict with specific findings (not vague approval). The oracle is **host-delegated by default** (per [adr-25](../../decisions/adrs/0025-host-delegated-llm-default.md)); an opt-in oracle adapter (native / CLI / API / MCP) runs the audit when wired — never blocks.
2. **Round-trip gate**: phase outputs feed cleanly into the next phase's expected inputs (e.g., Pass A's `spec-essence.json` validates against schema and is consumed by Pass B's chat-distiller without parse errors).

Acceptance is checked across the validation corpus (`.abcd/corpus.json`), with documented per-repo exemptions where a feature genuinely doesn't apply.

## 7. Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:disembark`, **then** the dispatcher shows when the project last disembarked, where the lifeboat lives, the available sub-verbs (`to <path>`, `probe`, `dry-run`; later phase: `to-spec-kit`), and suggested next actions — bare invocation never mutates state.
- **Given** a corpus repo with a native spec store, ADRs, and a memory backend present, **when** `/abcd:disembark to home` runs to completion, **then** `.abcd/lifeboat/` contains all sections in [§ 5](#5-output-shape) and the oracle audit returns a "sufficient" verdict with specific findings.
- **Given** a corpus repo where one source is sparse (e.g., a repo with no spec store), **when** `disembark to home` runs, **then** the affected agent emits a documented exemption in its report (e.g., empty `spec-essence.json` with a `reason: "no spec store in source"` note) and the oracle gate accepts the exemption.
- **Given** the user runs `/abcd:disembark probe`, **when** the command completes, **then** all adapters' `probe()` runs in parallel, the ActiveSources list is reported to stdout, no `.abcd/lifeboat/` mutation occurs, and the run takes a small fraction of the time `to home` would take (no LLM dispatches).
- **Given** the user runs `/abcd:disembark dry-run`, **when** the command completes, **then** the inventory + agent budget estimation runs end-to-end, the would-be writes are listed (file paths, agent dispatches, estimated tokens), no `.abcd/lifeboat/` mutation occurs, and the report explicitly notes which Pass A/B/C agents would have been dispatched.
- **Given** an existing lifeboat at the destination, **when** `disembark to <path>` is re-run, **then** the user is asked transparently whether to overwrite (with `.bak` safety net) or skip.
- **Given** an agent's input estimate exceeds `disembark.maxAgentTokens`, **when** that agent dispatches, **then** the stream + summarise (map-reduce) path is taken instead of one-shot dispatch.
- **Given** Phase 0 dev-sync fails on any source, **when** Pass C runs, **then** the affected Pass C agent runs in degraded mode and the disembark report flags the degradation explicitly.
- **Given** any `disembark to <path>` run completes, **when** the lifeboat is written, **then** a new line is appended to `.abcd/development/voyage/disembark/history.jsonl` with the run's `manifest_sha256`, file list, oracle backend used, and verdict (per [`03-embark.md § 7`](03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history)). The hash matches `_provenance.json` in the lifeboat for this run.
- **Given** `disembark to <path>` is re-run after a previous run, **when** the new snapshot lands at `<path>`, **then** the previous snapshot is overwritten (with `.bak` safety net) AND the previous snapshot's manifest remains in `voyage/disembark/history.jsonl` — there is never a `lifeboat-v1/` / `lifeboat-v2/` directory; history is preserved in the manifest log, not in stale snapshots.
- **Given** `<path>` is `home`, **when** any sub-verb resolves the path, **then** it expands to the current repo's `.abcd/lifeboat/` — providing an ergonomic default while keeping path explicit.
