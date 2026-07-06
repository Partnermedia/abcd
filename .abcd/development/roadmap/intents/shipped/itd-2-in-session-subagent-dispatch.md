---
id: itd-2
slug: in-session-subagent-dispatch
spec_id: fn-10-in-session-subagent-dispatch-oracle
kind: standalone
suggested_kind: null
reclassification_history: []
created: 2026-05-04
updated: 2026-05-07
---

# In-Session Subagent Dispatch — Always-Available Oracle Fallback

## Press Release

> **abcd ships a fallback oracle backend that needs zero external setup.** When neither RepoPrompt nor Codex CLI is available — fresh laptop, restricted network, locked-down corporate environment — abcd's oracle calls now route to an in-session Claude subagent via the host's Task tool. No subscriptions to configure, no MCP servers to install, no CLIs to find. The persona gets Carmack-style reviews from the same `oracle` already running their session. Oracle audits, plan-reviews, and impl-reviews all work out of the box.
>
> "I was on a borrowed laptop with no RP and no Codex setup," said Henry, junior developer. "Without the in-session fallback abcd would have errored out. abcd just used the in-session `oracle`. The review wasn't as cross-vendor as RP-routed, but it was instant and free."

## Why This Matters

itd-6 ships the RP MCP and Codex CLI transports — both need external setup (RP desktop app on macOS; Codex CLI + OpenAI subscription elsewhere). Most users will have at least one configured, but the friction of a "no oracle backend available" failure mode is real. The fix is a fallback that uses the abcd-running Claude session itself.

The challenge: a Python plugin running from `scripts/abcd_cli.py` (invoked by the bash wrapper) cannot directly call the host's `Task` tool — that's a model-side capability, only Claude (running in the harness loop) can invoke it. Crossing that boundary requires a wire protocol.

## What's In Scope

- **Sentinel-output protocol**: when `AgentDispatch.dispatch_agent(agent_name="in_session", ...)` is invoked, the harness emits a structured payload (JSON) that the orchestrating Claude session reads as a tool result and translates into a `Task` tool invocation. The slash-command markdown (`commands/abcd/intent.md`, `disembark.md`, etc.) embeds instructions telling the host: "if the script returned a `task_dispatch` payload, invoke Task with these args."
- **Subagent context isolation**: each in-session dispatch spawns a fresh subagent context (no continuity by design — the no-state-between-calls property is documented as an in-session-specific limitation, distinct from RP/Codex's chat_id-based continuity).
- **Result return path**: the Task subagent's output flows back via the host session into a file or stdout that abcd's loop reads on next iteration.
- **No-continuity caveat**: same-chat re-review semantics from itd-6 do NOT apply to in-session — each iteration is a fresh Task call. The audit-fix loop must re-prime context every iteration. This is acceptable per § 12 (in-session is the always-available fallback, not the preferred path).
- **Verdict parser interop**: in-session subagent outputs use the same `<verdict>...</verdict>` tag protocol as RP and Codex backends.

## What's Out of Scope

- **Autonomously orchestrating multiple in-session subagents** (e.g. for parallel review of different facets) — single subagent per call; multi-agent orchestration is a future intent or part of a broader brief revision.
- **Routing to specific Anthropic models** (Opus vs Sonnet vs Haiku) — uses whatever model the user's Claude Code session is running. abcd doesn't pick.
- **Cost optimisation** (in-session calls bill against the user's Claude subscription; abcd doesn't budget) — covered by itd-17 model effectiveness tracking if relevant.
- **Capability-aware backend selection** (added 2026-05-08 per idea-4 jagged-frontier review). The cascade is and stays availability-driven (per `04-universal-patterns.md § 7` "fixed cascade"). Any future capability-aware routing — when Frontier Awareness ships — is a *pre-cascade selector* layer ABOVE the cascade, NOT a modification to the cascade contract. Selector picks which backend the cascade *starts from* based on `(task_class, agent, model_id) → preferred_backend_ranking`; the fixed cascade per this intent and itd-6 begins from that backend without contract change. Thin seam between the layers.

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** a fresh laptop with neither RP nor Codex CLI installed and `oracle.backend = "auto"` in `.abcd/config.json`, **when** any abcd command invokes the oracle (e.g. `/abcd:disembark`'s lifeboat-oracle Phase C audit), **then** the audit completes successfully via in-session subagent dispatch and writes a finding JSON to `.abcd/lifeboat/audit/oracle-<ts>.json` without ever attempting an external network call or shell-out to `claude -p`.
- **Given** `oracle.backend = "in-session"` is explicitly pinned, **when** abcd's host script returns its sentinel-output payload, **then** the orchestrating Claude session reads the payload, invokes the host's `Task` tool with the agent name and prompt declared in the payload, and returns the subagent's output back to abcd via the documented result-return path (file or stdout) for the next iteration to consume.
- **Given** an in-session-dispatched subagent emits a Carmack-style review, **when** abcd parses the result, **then** the verdict tag matches the same `<verdict>SHIP|NEEDS_WORK|MAJOR_RETHINK</verdict>` protocol used by RP MCP and Codex CLI backends — the parser is backend-agnostic.
- **Given** an audit-fix loop runs over multiple iterations under `oracle.backend = "in-session"`, **when** the second and subsequent iterations dispatch, **then** each iteration is a fresh `Task` call with no inherited `chat_id` continuity AND abcd's loop re-primes the subagent context (artefact + prior verdict + applied fixes) every iteration — the no-state property is documented in `04-surfaces/02-disembark.md` and surfaced in CLI help text.
- **Given** a user with RP MCP available runs an oracle call and `oracle.backend = "auto"`, **when** the resolution chain runs, **then** RP MCP is selected (preferred) and in-session is *not* invoked — the in-session backend is the final fallback, never the default when alternatives exist.
- **Given** OpenCode is the host harness (per itd-22), **when** an in-session dispatch is requested, **then** the harness's `AgentDispatch.dispatch_agent("in_session", ...)` call routes through OpenCode's subagent invocation primitive AND the same wire-protocol contract holds — the per-harness branch is invisible to the calling agent.

## Open Questions

- What's the canonical wire protocol shape? JSON-on-stdout with a `next_action: "task_dispatch"` field, or some other sentinel? Slash-command markdown needs to know what to look for.
- How does the orchestrating session know to read the wire-protocol payload? The slash-command markdown instructs Claude — but commands like `/abcd:disembark` invoke `scripts/abcd` which spawns a long-running process; how does the post-process payload get back to Claude's view?
- Does this also work in OpenCode (per itd-22)? OpenCode has its own subagent invocation primitive; the harness's `AgentDispatch.dispatch_agent("in_session")` impl needs a per-harness branch.
- What's the right way to test this without spinning up a real subagent? Mock the host's Task tool? Stub the wire-protocol consumer side?
- Does itd-2 need to ship before itd-22 (OpenCode), or can both be designed in parallel since they share the AgentDispatch Protocol surface?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
