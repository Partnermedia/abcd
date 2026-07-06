---
id: itd-59
slug: autonomous-worker-transcript-capture
spec_id: null
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: []
routed_from: []
created: 2026-06-25
updated: 2026-06-25
prd_path: null
---

# Every Autonomous Ralph Loop Pass Leaves the Same Durable, Queryable Transcript an Interactive Session Does

## Press Release

> **abcd captures the instruction-set and reasoning of every autonomous Ralph worker loop pass into the same durable history store it already keeps for interactive sessions — so the work done while no human is watching is exactly as auditable as the work done with one.** Today there is an asymmetry: an operator's interactive Claude session runs under `specstory run` and its transcript is redirected into `~/.abcd/history/<root_sha>/specstory/` (a durable, queryable store). But the autonomous Ralph worker — the thing that runs unattended for hours, plans, implements, and self-reviews — is spawned by `ralph.sh` as a bare `claude -p …` piped to a per-run `iter-NNN.log` under `scripts/ralph/runs/`, which is **gitignored and pruned**. The agent doing the most consequential, least-supervised work leaves the least durable record. This intent closes that gap: every autonomous loop-pass's instructions + reasoning is captured to the same history store, with the same queryability, as an interactive session.
>
> "abcd's whole promise is that you can trust what the agent did while you were away," said Alex, framework author. "But when I went to audit a Ralph run after the fact, the interactive sessions were all there in the history store and the autonomous worker — the one I most needed to check — had left only an ephemeral log that the run-dir cleanup had already eaten. The unattended work is exactly the work that most needs a durable trail."

## Why This Matters

abcd is an autonomous-development framework whose differentiator is **auditable, provenance-bearing automation** — the operator can leave Ralph running and later reconstruct what it decided and why. That promise has a hole precisely where it matters most:

- **Interactive sessions ARE captured** (verified 2026-06-25): `scripts/abcd/specstory_shim.py` writes `<repo>/.specstory/cli/config.toml` redirecting SpecStory's `output_dir` to `~/.abcd/history/<root_sha>/specstory/`, and operator sessions run under `specstory run claude` (observed live: `specstory run claude -c` processes parented to interactive shells).
- **Autonomous Ralph workers are NOT** (verified 2026-06-25): `ralph.sh:1266-1291` spawns the worker as `$CLAUDE_BIN "${claude_args[@]}" "$prompt" | tee "$iter_log" | watch-filter.py`, where `CLAUDE_BIN` defaults to bare `claude` (`:527`) — never `specstory run`. The only record is `scripts/ralph/runs/<run>/iter-NNN.log`, and `scripts/ralph/.gitignore` ignores `runs/` (and `runs-archive/`), so it is neither committed nor durably stored; run-dir hygiene prunes it.

The consequence: the worker's full chain of reasoning (the instructions it received, the `--append-system-prompt` autonomous rules, every tool call and decision across a loop pass) exists only transiently. After a run is pruned — or after the loop-death-and-restart churn this very session experienced — the autonomous reasoning is gone, while every interactive session that touched the repo is permanently queryable. For a framework that markets unattended trustworthiness, the unattended path is the one missing its black box.

This is also a *symmetry* failure, not just a missing feature: the capture machinery already exists and already works for one caller class. The autonomous caller was simply never routed through it (or through an equivalent durable sink).

## What's In Scope

- **The capture gap itself:** every autonomous Ralph worker loop-pass's instructions + reasoning lands in a durable, queryable store under `~/.abcd/history/<root_sha>/` — parity with interactive-session capture.
- **All autonomous entry points:** the `ralph.sh` worker spawn (plan / work / completion-review phases), and any other unattended `claude` invocation Ralph makes, are covered — not just the work phase.
- **Queryability + provenance:** captured transcripts are associated with the run id, spec/task id, loop pass, and phase, so a later audit can answer "what did the worker reason for fn-NN.M, loop pass K?".
- **Honor abcd's dependency discipline:** whatever the mechanism, it wraps/configures the unmodified dependency (SpecStory and/or the existing history store) rather than forking it, and survives a Ralph re-vendor (overlay-managed if it touches `ralph.sh`, which is harness surface).

## What's Out of Scope

- **Re-architecting the history store.** `~/.abcd/history/<root_sha>/` and its single-owner `register_repo` provisioning (fn-15) stay as-is; this intent FEEDS that store, it does not redesign it.
- **Capturing the operator's interactive sessions** — already solved (specstory redirect). This is only the autonomous-worker half.
- **The watch-filter / live `--watch` UI** — that is a live console view, orthogonal to durable capture; unchanged.
- **Retroactive recovery of already-pruned runs** — this is forward-looking capture, not archaeology of lost logs.

## Approach (deliberately OPEN — for plan/grill to decide)

The mechanism is intentionally left undecided; the planning/grill stage chooses between (at least) these candidates, weighing the wrap-not-fork rule, re-vendor survival, and the autonomous worker's non-interactive (`claude -p`) mode:

- **(a) Wrap the worker in SpecStory** — route `ralph.sh`'s worker spawn through `specstory run` (or point `CLAUDE_BIN` at a specstory wrapper) so autonomous transcripts land in the SAME store and format as interactive ones (maximal symmetry; but verify SpecStory captures a non-interactive `-p` run, and that wrapping survives re-vendor as an overlay).
- **(b) Ingest the existing `iter-NNN.log`** — keep the bare-`claude` spawn but durably ingest each loop-pass's log into the history store on loop pass end (less invasive to the spawn path; but not SpecStory-format, and must capture the instructions too, not just stdout).
- **(c) A hybrid / structured sink** — capture the worker's stream-json (`--output-format stream-json` is already used) into a structured per-loop pass record in the store.

Open questions for grill: does SpecStory record a headless `claude -p` session at all? Does wrapping interfere with the `tee | watch-filter.py` pipe or the timeout/cancel-storm controls? Should the capture be best-effort (never block a loop pass) or gated? What is the retention/size policy for autonomous transcripts vs interactive?

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** an autonomous Ralph worker loop pass (plan, work, or completion-review phase), **when** it completes (or is terminated), **then** its instructions + reasoning is present in a durable store under `~/.abcd/history/<root_sha>/`, queryable after the run dir is pruned, and associated with run id + spec/task id + loop-pass + phase.
- **Given** the existing interactive-session capture, **when** the autonomous capture lands, **then** both reach the same history store (parity), and a single audit query can enumerate autonomous AND interactive activity for the repo.
- **Given** abcd's dependency discipline, **when** the mechanism is implemented, **then** it wraps/configures unmodified SpecStory / history-store machinery (no fork), and if it touches `ralph.sh` it is overlay-managed so it survives `/abcd:ralph-up` re-vendor.
- **Given** the live controls, **when** capture is active, **then** it does not break the `tee | watch-filter.py` live view, the worker timeout, or the cancel-storm/budget controls, and a capture failure never blocks or fails a loop pass (best-effort, fail-open on the capture path only).

## References

- Research (2026-06-25, this session): verified interactive = captured, autonomous = not.
- `scripts/ralph/ralph.sh:527` (`CLAUDE_BIN="${CLAUDE_BIN:-claude}"`), `:1266-1291` (worker spawn → `tee iter_log | watch-filter.py`, no specstory)
- `scripts/ralph/.gitignore` (`runs/`, `runs-archive/` ignored → iter logs ephemeral)
- `scripts/abcd/specstory_shim.py` (the interactive-session redirect to `~/.abcd/history/<root_sha>/specstory/`), `scripts/abcd/history_store.py` (`register_repo` single-owner store provisioning, fn-15)
- `.abcd/development/research/notes/ahoy-history-store-manual-scaffolding.md` (the history-store design this feeds)
- Related: the abcd transparency/provenance promise in the brief (`01-product/`), itd-58 (a sibling session-provenance closure from the same monitoring run)
