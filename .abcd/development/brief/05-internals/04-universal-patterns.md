# Universal Patterns

All abcd commands share these patterns. Implement once in shared helpers, not per-command.

## 1. Transparent prompts

Every `AskUserQuestion` (or equivalent harness call) shows:

1. **Current state** ("Current: private")
2. **Consequence of each option** ("Switching to public removes .specstory/, .flow/ ... from tracking")
3. **The question + how to change later** ("Keep private? — change later with `abcd config set repo.visibility public`")

No silent defaults. No surprises.

## 2. MCP-preferred, structural-fallback

Every external-tool call follows this pattern:

```
configured backend (e.g., RepoPrompt MCP, Codex MCP)
    └─ if unavailable → structural fallback (in-session subagent, git/file scan)
```

Examples:
- **Oracle audit** (lifeboat-oracle, press-release-composer, intent-fidelity-reviewer): RP MCP → Codex CLI → in-session subagent (per itd-6 + itd-2)
- **Code-rescuer**: RP MCP for codemap → spec-driven git-window file selection
- Future agents follow the same pattern by default.

## 3. Plugin-preferred + internal-fallback

When abcd needs a capability that another installed plugin already provides (flow-next, RepoPrompt, etc.), prefer calling the other plugin's agent/skill over reimplementing. Fall back to abcd's own implementation only when the preferred provider isn't installed.

```
preferred plugin's agent/skill (e.g., flow-next:github-scout)
    └─ if plugin not installed → abcd's internal implementation
```

ahoy probes for known plugins on install and records detection in `.abcd/config.json` → `plugins.<name>.detected = true|false`. Detection is re-run on every ahoy (consistent with the transparent re-confirm rule). Examples:

- **Issue scout**: `flow-next:github-scout` → abcd's internal `issue-scout` agent
- **Intent → spec creation**: `/abcd:intent plan <itd-N>` calls `/flow-next:plan` to scaffold the bidirectionally-linked spec (the canonical create `/abcd:intent "<text>"` only writes a draft — no flow-next call). flow-next is the upstream spec system; abcd never reimplements that surface.
- **Future**: any abcd capability with a strong external counterpart should follow this pattern, not duplicate functionality

## 4. JSON internal, MD render

All inter-agent data is JSON; markdown is a render step at the end of each pass.

- Each JSON artefact has a JSON Schema in `scripts/abcd/schemas/`
- `render.py` provides deterministic JSON → MD renderers
- Easier to validate, agents are unit-testable against schemas
- Re-rendering with different templates is cheap

## 5. Reports as JSON + MD pairs

Every command emits `<command>-report.json` (full structured detail) and `<command>-report.md` (human skim summary, rendered from JSON). Both stored in `.abcd/logbook/<command>/<timestamp>/`.

## 6. `.abcd/logbook/` layout

```
.abcd/logbook/
├── ahoy/<timestamp>/
│   ├── ahoy-report.json
│   ├── ahoy-report.md
│   └── prompts.json              # what was asked, what was answered
├── disembark/<timestamp>/
│   ├── disembark-report.{json,md}
│   ├── _state.json               # forensic checkpoint state (no resume; for post-mortem only)
│   ├── progress.log              # streaming progress (tail -f friendly)
│   └── agents/<agent>/<run>.{json,md}
├── embark/<timestamp>/
│   └── embark-report.{json,md}
├── launch/<timestamp>/
│   ├── launch-report.{json,md}
│   └── preflight.{json,md}       # PII/secret scan output
├── intent/<timestamp>/
│   └── intent-report.{json,md}   # one per /abcd:intent invocation
├── capture/<timestamp>/
│   └── capture-report.{json,md}  # one per /abcd:capture invocation (per itd-4)
├── grill/<utc-ts>-<intent-id>/   # one per /abcd:intent grill session (per itd-27)
│   └── grill-report.{json,md}    # glossary terms are written inline to terminology/, not batched here
├── audit/<sub-tier>-<ts>/        # review/audit reports across six sub-tiers land here
│   └── report.{json,md}          # sub-tier ∈ {review, spec-mg, consistency, shape, chain, lifeboat}:
│                                 #   audit/review-<ts>/      (Role 1 itd-1 pass / /abcd:intent review,        itd-1)
│                                 #   audit/spec-mg-<ts>/     (Role 1 MG004 pass / itd-37 boilerplate receipt, itd-37)
│                                 #   audit/consistency-<ts>/ (Role 2 / /abcd:intent consistency,   itd-48 — superseded itd-31; live as of fn-29)
│                                 #   audit/shape-<ts>/       (Role 3 / /abcd:intent shape,         itd-34)
│                                 #   audit/chain-<ts>/       (default app of /abcd:audit chain,    itd-16, later phase)
│                                 #   audit/lifeboat-<ts>/    (sibling app of /abcd:audit lifeboat, itd-35, later phase)
│                                 # Directory name (audit/) reflects "this is the on-disk audit trail"
│                                 # regardless of which verb produced it; sub-tier prefix names the verb.
│                                 # `chain` and `lifeboat` are sub-verbs of /abcd:audit umbrella;
│                                 # `review`, `consistency`, `shape` are sub-verbs of /abcd:intent.
│                                 # Bare /abcd:audit and bare /abcd:intent are status+help only.
├── sota-audits/<date>.{json,md}  # periodic prompt SOTA audit findings (option D)
└── phase/<phase-id>/             # validation cadence outputs per phase (Phase 0 study, Phase 1 acceptance, etc.)
    └── <test-name>.{json,md}
```

**Note: `.abcd/logbook/` is for reports only.** Coordination state (file locks like `shape.lock`, multi-agent claims per itd-33) lives at `.abcd/coordination/` — a *sibling* of `logbook/` under `.abcd/`, not a subdirectory. See `04-surfaces/05-intent.md § 6` for the canonical lock-path contract (`.abcd/coordination/shape.lock`).

**Later-phase additions to `.abcd/logbook/`** (appear when their parent intent ships):
- `dredge/<timestamp>/` — cross-corpus synthesis output (itd-25)
- `frontier/<timestamp>/` — per-run frontier-mapping events (Frontier Awareness; idea-4)
- `doc-fidelity/<input_fingerprint>/` — the doc-fidelity anti-drift pass (itd-60). **An explicit EXCEPTION to the `<command>/<timestamp>/` convention above:** this tier is **content-addressed**, keyed by the run's `input_fingerprint` (a sha256 over the deterministic trust+reality inputs — receipts, target manifest, bundle manifest, prompts) rather than a timestamp, so an identical re-run reuses the same `report.json` + bound `decision.json` instead of accreting a fresh ts dir. A `decision.json` (approve/defer) binds to a fingerprint dir; `deferred.jsonl` sits directly under `doc-fidelity/` (not inside a fingerprint dir) so an open obligation stays discoverable after the gate clears. The **pre-fingerprint-failures/`<timestamp>/`** sibling is the one ts-keyed slice (a failure that occurs *before* a well-formed fingerprint can be computed — invalid config/manifest, intent-resolution conflict — has no reusable content-addressed report, so its diagnostics are ts-keyed and never reused). This layout contract is the single source of truth for the tier's on-disk shape.

**Forward-source provenance marker (the itd-61/fn-75 dedup contract).** When the doc-fidelity pass (itd-60) drafts a brief delta to repair forward drift, the staged lines are wrapped in a **paired** HTML-comment stamp so the covered region is unambiguous:

```
<!-- abcd:forward-doc-sync:begin origin=itd-60 spec=fn-N input_fingerprint=<hex> consumed_receipts_sha=<hex> -->
...the drafted delta lines...
<!-- abcd:forward-doc-sync:end -->
```

The pairing is load-bearing: a single self-closing comment cannot delimit a multi-line block, so itd-61/fn-75's derivation dedup needs a **matched** begin/end pair to exclude exactly the freshly-stamped lines and nothing else. `consumed_receipts_sha` is the sha256 over the sorted per-receipt **stable trust-and-reality digests** — the same `{spec_id, parse_error, rollup_agreement, criteria:[{criterion, verdict, detail_key}]}` digest the report's `input_fingerprint` uses (deterministic trust+reality fields only; no LLM-authored `detail` *value*, no timestamp). So the marker is reproducible across reviewer re-runs, does not churn on a forensic-prose rewording, but **does** change when a trust field changes. The grammar pins `origin=itd-60` and requires full lowercase sha256 widths (a short or foreign-origin marker is not valid coverage). **fn-75 fails closed on any unmatched or legacy single-line marker.** The grammar is owned by `scripts/abcd/doc_fidelity/_marker.py` (first-created with the stamping in fn-74.3); the CI gate, the pre-commit advisory wrapper, and the spec-close preflight all reference it but none re-implement it.

**Later-phase sibling additions under `.abcd/`** (NOT under logbook — operational state, not run reports):
- `.abcd/coordination/audit/<YYYY-MM-DD>.jsonl` — multi-agent coordination append-log (itd-33; JSONL, daily UTC rotation, committed). Sibling local-only state (gitignored): `.abcd/coordination/active-work.json` and `.abcd/coordination/*.lock`.

Tracked alongside the rest of `.abcd/` per the visibility rule ([`03-configuration.md § 1`](03-configuration.md#1-visibility-driven-gitignore-policy)) — committed in private repos, gitignored in public. No special exception. Sensitivity is handled at launch time: the launch payload manifest ([`../04-surfaces/04-launch.md § 2`](../04-surfaces/04-launch.md#2-payload-manifest-default-deny)) excludes the entire `.abcd/` namespace from what ships publicly.

**`logbook/` vs `voyage/` distinction:** `logbook/<command>/<timestamp>/` holds *per-run* command output (reports, prompts asked, forensic checkpoint state — abcd ships no resume sub-verb, so checkpoints are post-mortem only) — ephemeral relative to a single invocation. `.abcd/development/voyage/` (see [`../04-surfaces/03-embark.md § 7`](../04-surfaces/03-embark.md#7-voyage-layout-embarkdisembark-provenance-and-history)) holds *cross-run* embark/disembark provenance and history that the project carries forward. Both are tracked under the visibility rule; they answer different questions ("what happened in this run?" vs "what is the lifeboat history of this repo?").

## 7. Vendor-agnostic adapters with environment branching

abcd consumes data from vendor-specific stores (Claude Code's `~/.claude/projects/...`, RepoPrompt's Application Support, OpenCode's TBD path, etc.). To stay portable across LLM harnesses and external tools, adapters are named by **semantic role** (what content they handle), not by vendor:

- `memory.py` — handles agent memory, regardless of whether the source is Claude Code, OpenCode, or a future harness
- `reviews.py` — handles oracle/Carmack-style reviews, regardless of whether the source is RepoPrompt, Codex, or a future tool

Each semantic adapter is a **thin dispatcher** (~30 lines: detection logic + delegation). Concrete per-vendor implementations live in a sibling subdirectory (`memory_backends/`, `reviews_backends/`), one file per vendor. Architecture:

```
adapters/
├── memory.py                    # dispatcher (semantic name)
├── memory_backends/
│   ├── claude.py                # Claude Code-specific
│   ├── opencode.py              # OpenCode-specific (stub for now; itd-22)
│   └── ...                      # one file per vendor
├── reviews.py                   # dispatcher
└── reviews_backends/
    ├── repoprompt.py
    ├── codex.py
    └── ...
```

**Backend resolution: detect by default, config to override.**

- `.abcd/config.json` → `<adapter>.backend = "auto" | "<vendor-name>" | "<custom-path>"`
- `"auto"` (default): dispatcher runs environment detection, picks the matching backend
- Explicit value: bypasses detection, forces the named backend (useful for unusual setups, testing, or when detection is ambiguous)
- Adding a new backend = drop a new `<adapter>_backends/<name>.py` + add to dispatcher's known list. No edits to consumers (`review-collator`, `principle-distiller`, etc.)

**Why semantic naming, not vendor naming:** the consumers (agents) shouldn't care where data came from. They consume "memory" or "reviews", not "Claude Code memory" or "RepoPrompt reviews". The vendor layer is implementation detail.

**Pattern application:**
- `memory.py` (currently: Claude Code backend active, OpenCode backend stubbed; later phase: OpenCode backend activated as part of itd-22 OpenCode portability)
- `reviews.py` (currently: RepoPrompt backend; future: Codex CLI, generic markdown imports, other tools)
- `oracle.py` follows a related pattern but with a **fixed cascade** rather than vendor auto-detection: `oracle.backend` config names the *preferred starting point* (`"rp"`, `"codex"`, or `"in-session"`); on failure, abcd cascades down the chain until something succeeds. Default `"auto"` cascades RP → Codex → in-session per availability. (Post-fn-5 footnote: the RP leg's transport is the concrete `MCPBridge` in `scripts/abcd/mcp_bridge.py`; an unreachable RP surfaces as the typed `RPUnavailable` exception declared by fn-5, which is the signal the cascade catches to fall through to Codex.)
- `rp_state_backend.py` (currently: RP workspace.json pull only per itd-7; later phase: presets, mcp-routing scoping, etc.)
- Future agents/adapters with vendor-specific sources should follow this pattern by default

## 8. Artefact-lifecycle taxonomy

abcd produces three classes of durable artefact, each with distinct curation rules. **Lifecycle class is declared in the parent README of each artefact namespace.** Lint blocks if a namespace's curation behaviour disagrees with its declared class (lint code reserved at `06-lint.md`).

**Three classes, three behaviours:**

| Class | Behaviour | Examples |
|---|---|---|
| **Regenerable** | Overwritten in place; regenerated from authoritative inputs on next run. Single canonical version at any time; history preserved separately if at all. | `.abcd/lifeboat/` (latest disembark snapshot only), `.abcd/development/voyage/` cards, sota-audit findings, intent-fidelity audit reports |
| **Append-only** | Never modified after creation; new entries accrete; old entries preserved verbatim. | `.abcd/logbook/<command>/<timestamp>/` per-run reports, `.abcd/development/voyage/disembark/history.jsonl`, capture issue-ledger entries (immutable post-create) |
| **Compounding-curated** | Accumulates across sessions/runs; pages added, modified, contradicted, deprecated by curator. Carries provenance per entry; lint surfaces drift between curated form and source-of-truth. | `.abcd/memory/` (multi-upstream knowledge substrate per itd-36), `.abcd/development/activity/notes/`, `.abcd/development/activity/reviews/`, the brief itself |

**Why the taxonomy is load-bearing:** without it, "regenerable" and "compounding" get conflated, the curator agent (e.g., `principle-distiller` post-itd-36) loses its contract with consumers, and a pattern like itd-36's memory-unification looks like ceremony when it's actually a different lifecycle class than the lifeboat. Naming the three classes lets each artefact namespace declare its rules explicitly and lets cross-document fidelity audit (Role 2) catch drift.

**Recomputation discipline for regenerable artefacts.** Regenerable artefacts use **full-crawl-on-demand** recomputation, not incremental delta-application. Three failure modes that justify the discipline:

1. **Drift-without-detection.** Delta-application accumulates small bookkeeping errors silently; full crawl is stateless and idempotent.
2. **Schema fragility.** Schema bumps require re-interpreting every past delta; full crawl re-parses with the current schema each run, no museum of past schemas.
3. **False O(1).** Deltas that reference cross-cutting state ("this spec newly depends on spec-N's output") read back into the corpus, collapsing the O(1) claim where it matters most.

Cadence for regenerable artefacts: **on-demand + at phase-milestone boundaries**, NOT every state-change event. At current corpus sizes (10-30 specs × ~4 sub-bullets each), full crawl runs in seconds. Re-evaluate cadence if the corpus grows past ~50 entries; not before.

**Why compounding-curated is NOT regenerable.** A compounding artefact's value is the curated synthesis across upstream sources — it cannot be reconstructed from sources alone without the curator agent's accumulated decisions (which contradictions to surface, which sources to weight, which entries to deprecate). Regenerable artefacts are stateless functions of inputs; compounding-curated artefacts carry curator state.

## 9. The abcd abstraction boundary — surface classification

abcd's guarantees live at the `/abcd:*` surface; the bundled dependencies (flow-next/Ralph, RepoPrompt, codex) are implementation detail. The boundary rule: **sanctioned** = an `/abcd:*` verb (or the fn-37.3 sentinel shim) is the caller; **reach-past** = a *person* is the caller. Reach-past is warned-and-redirected, never blocked (adr-15; itd-52).

Two distinct boundaries, with different detection ceilings:

- **Live-call boundary** — decided at exec time by the process-scoped `--abcd-driven` argv sentinel (shipped by fn-37.3; emission in `scripts/abcd/session/ralph_mirror.py`, consumption in `scripts/ralph/flowctl` / `scripts/abcd/tools/_dispatch.py` — see [`10-in-session-dispatch.md`](10-in-session-dispatch.md) for the in-session dispatch wire protocol this rides alongside). A doctor probe runs after the call and **cannot** observe the sentinel; live detection of non-dispatch reach-past is the deferred PreToolUse hook.
- **Static-artifact boundary** — `ahoy doctor` detects persistent bypass *artifacts* on disk, and only those. The table below records per-surface detectability honestly; rows without an artifact are `no`, not pretended-yes.

This table is the **maintained classification** of bundled-dep surfaces (itd-52 Open-Q#2: a maintained map, governing documentation + future heads-up keywords — not a probe input). abcd's own `<!-- BEGIN ABCD -->` fence (`scripts/abcd/markers.py`) is sanctioned and is not a member of the reach-past set.

| Surface | Classification | Statically detectable? | Detector/artifact | Remediation |
|---|---|---|---|---|
| `/flow-next:ralph-init` before `/abcd:ralph-up` | sanctioned | n/a (sanctioned) | — | — |
| `/abcd:intent plan` → `/flow-next:plan` (+ `/flow-next:plan-review` for bundles) | sanctioned | n/a (sanctioned) | — | — |
| `/abcd:intent ship` → `/flow-next:work` | sanctioned | n/a (sanctioned) | — | — |
| In-session Ralph driving `flowctl.py` via the `--abcd-driven` sentinel shim | sanctioned | n/a (sanctioned) | — | — |
| `/flow-next:interview` redirect for the spec/task layer | sanctioned | n/a (sanctioned) | — | — |
| `/flow-next:setup` — unmarked `.flow/bin/flowctl` bypass artifact | reach-past | yes | `flow_bin_flowctl_probe` (`scripts/abcd/tools/doctor.py`, fn-33; surfaced by `ahoy doctor`) | run `/abcd:ralph-up` to re-assert the dispatcher |
| `/flow-next:setup` — `<!-- BEGIN FLOW-NEXT -->` marker block in CLAUDE.md/AGENTS.md | reach-past | yes | `flow_next_marker_block_probe` (`scripts/abcd/tools/doctor.py`, fn-42.2; surfaced by `ahoy doctor`; re-confirmed against the installed flow-next 1.13.0 setup templates) | remove the block; follow the dispatcher mandate in `CLAUDE.md` |
| Bare `flowctl` / `.flow/bin/flowctl` invoked outside Ralph | reach-past | no | — (no persistent artifact; the live sentinel is process-scoped) | use the `/abcd:*` verb, or the `scripts/ralph/flowctl` dispatcher when scripting |
| Direct `/flow-next:*` skill invocation (e.g. `/flow-next:plan`) | reach-past | no | — (indistinguishable on disk from the sanctioned wrapped call) | use the wrapping `/abcd:*` verb (`/abcd:intent plan` / `ship`) |
| Direct RepoPrompt / codex use | reach-past | no | — | use the `/abcd:*` review surfaces |

The detectable rows each earn a doctor probe (the fn-33 probe is the template; fn-42.2 shipped the second — `flow_next_marker_block_probe`). The always-loaded principle statement lives in `CLAUDE.md` § "The abcd abstraction boundary"; the decision record is [`adr-15`](../../decisions/adrs/0015-abstraction-boundary-warn-not-block.md).

## 10. Guard receipt classification — review-backend receipts the Stop gate recognises

The abcd sibling guard (`scripts/abcd/hooks/abcd_ralph_guard.py`) gates a Ralph worker's Stop / SubagentStop on a completed review: a review backend writes a **receipt** when it reaches a verdict, and the guard's Stop handler refuses to stop until a valid receipt exists. This section is the single source of truth for the receipt *classes* the guard recognises — their schema, their on-disk store path + discovery rule, and their replay rule. It does not restate the Stop-gate block semantics (those mirror upstream `ralph-guard.py` and live in `_handle_stop_event`'s docstring).

Two receipt classes exist, one per review-transport family. They are mutually exclusive per review mode (a given review is either a flowctl-backend review or an oracle_send review, never both); the guard checks the oracle path first so an oracle-backed review is never silently passed by the flowctl-receipt fall-through.

**(a) flowctl-backend receipt** — the `rp` / `codex` / `copilot` / `cursor` review backends.
- **Shape:** `{type, id, verdict}` (`type` ∈ plan_review/impl_review/completion_review; `verdict` ∈ SHIP/NEEDS_WORK/MAJOR_RETHINK). Path-derived type/id binding when the filename is parseable.
- **Validator:** `scripts/abcd/hooks/receipt_validate.py` — the abcd-owned SSOT that MIRRORS the upstream contract independently (never imports the vendored guard it backs up), pinned by `tests/abcd/test_receipt_validate_parity.py`.
- **Store path + discovery:** the explicit `REVIEW_RECEIPT_PATH` env var, set by `ralph.sh` (unset for `review=none`). A single deterministic path — no globbing over user-writable trees.
- **Replay:** none intrinsic to the shape; the two-factor task-seal model (`is_owner_sealed`, receipt timestamp ≥ cutoff) governs freshness for the SHIP-seal path.

**(b) oracle_send receipt** — an MCP `oracle_send` review (fn-82.7 / R8).
- **Schema:** `scripts/abcd/schemas/oracle_send_receipt.schema.json` — inline-only (no `$ref`), `additionalProperties:false`. Fields `{type, id, mode: "oracle_send", verdict, chat_id, timestamp}` (`type`/`verdict` enums as in (a); `mode` is the const discriminator distinguishing this from the flowctl-backend receipts; `chat_id` ties the verdict to a real oracle_send conversation; `timestamp` is ISO-8601 UTC).
- **Validator:** `scripts/abcd/hooks/oracle_receipt_validate.py` — a fail-closed, stdlib-only, full-schema validator (single authoritative path: exact key set, `mode` const, both enums, non-empty `id`/`chat_id`, timestamp shape, and the path-derived type/id binding). Parallel to `receipt_validate.py`; it re-uses that module's `parse_receipt_path` for the filename grammar but does not fold into it (folding would either widen the upstream-parity contract or leave the new fields unchecked).
- **Store path + discovery:** the explicit `ORACLE_REVIEW_RECEIPT_PATH` env var, set by the session to the RUN_DIR-keyed path `$RUN_DIR/oracle-review-receipt.json` (`oracle_receipt_validate.store_path_for_run_dir`). A single deterministic path mirroring the `REVIEW_RECEIPT_PATH` convention — RUN_DIR-keyed, never a glob over user-writable trees.
- **Replay:** two defenses, both in `validate_and_consume`. (1) A **single-consumption marker** — on a fresh valid receipt the validator writes a `<store>.consumed` sidecar; a second validation of the same bytes is rejected as replayed (a marker-write failure fails closed, since an un-markable receipt could be replayed). (2) A **freshness window** — a receipt whose `timestamp` is older than `ABCD_ORACLE_RECEIPT_MAX_AGE_S` (default 24h; a non-positive/unparseable override disables the window, leaving the marker as the primary defense) is rejected so a leaked earlier-run receipt cannot seal a later run.

**ID binding (both classes):** the receipt `id` must match the spec/task the gated write claims. Where the store filename encodes a `plan-`/`impl-`/`completion-` prefix + id, the path-derived `(kind, id)` is enforced against the receipt; an unparseable filename (e.g. the RUN_DIR-keyed `oracle-review-receipt.json`) skips the path-match and validates shape/mode/verdict only.

The hermetic end-to-end test for (b) (`tests/abcd/test_oracle_send_stop_gate.py`) consumes a **committed fixture receipt** (`tests/abcd/fixtures/oracle_send_receipt_valid.json`) at the store path — no live `oracle_send` MCP call (fn-67 hermeticity).
