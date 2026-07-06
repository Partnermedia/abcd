---
id: itd-22
slug: opencode-portability
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
created: 2026-05-03
updated: 2026-05-03
---

# abcd Runs in OpenCode Too

## Press Release

> **abcd runs in OpenCode as well as Claude Code.** All six commands, all 15 agents, all the dev-sync and lifeboat infrastructure work identically across both harnesses. The harness-specific code lives behind one `harness.py` interface; OpenCode users get the same abcd conventions, intents, and lifeboats that Claude Code users get. Switching harness becomes a config decision, not a tool migration.
>
> "We had a team standardising on OpenCode for cost reasons but couldn't use abcd," said Grace, VP Engineering. "abcd ships an OpenCode harness implementation. Same plugin, same conventions, same lifeboats. The cross-harness portability removed our last objection."

## Why This Matters

abcd ships with a `harness.py` shim from day one (locked decision in the brief) — all Claude Code-specific calls (AskUserQuestion, agent dispatch, MCP, scheduling) go through this interface. The OpenCode port is then a second implementation of `harness.py`, not a rewrite of the plugin.

The portability promise is the entire reason for the harness shim. Delivering on it validates the architectural decision and opens abcd to the OpenCode ecosystem.

## What's In Scope

- Second `harness.py` implementation: `harness_opencode.py`
- OpenCode-specific shims for: agent dispatch, AskUserQuestion-equivalent, MCP calls, background tasks, scheduled wakeups, slash command discovery
- Test corpus: same multi-repo validation set runs through OpenCode harness
- Documentation: how to install + configure abcd under OpenCode
- Minimum-viable parity: every abcd command works in OpenCode (some convenience UX may differ)

## What's Out of Scope

- Third-harness ports (Cursor, Aider, etc.) — same pattern would apply but separate intents
- Cross-harness lifeboats (a Claude Code-produced lifeboat must work in OpenCode and vice versa — this is the test, but lifeboat format itself is harness-agnostic by design)
- Performance parity (OpenCode may be slower or faster; abcd just works)

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** abcd is installed under OpenCode, **when** the user runs `/abcd:ahoy install`, **then** the marker block is installed, the `.abcd/` namespace is created, and the install completion report renders identically to the Claude Code path — the surface text + actions are harness-agnostic.
- **Given** the same project is opened in Claude Code on one machine and OpenCode on another with the same lifeboat, **when** both run `/abcd:disembark to <path>`, **then** the resulting lifeboats are byte-identical (modulo timestamps and oracle backend chat IDs) AND embark of either lifeboat into the other harness completes without error.
- **Given** OpenCode's MCP equivalent (or its absence), **when** abcd's harness shim issues an oracle call, **then** the call routes through `harness_opencode.py` to OpenCode's available primitive AND the resolution chain (RP MCP → Codex CLI → in-session subagent) honours the OpenCode-side capabilities — no harness-specific code leaks into agent prompts.
- **Given** OpenCode's subagent dispatch primitive (per itd-2's wire-protocol contract), **when** abcd dispatches an in-session subagent under OpenCode, **then** the same `<verdict>` tag protocol holds AND the result-return path delivers output back to abcd's loop within the same iteration shape used under Claude Code.
- **Given** the corpus validation suite runs through OpenCode, **when** every command and acceptance gate completes, **then** the corpus passes with the same per-test verdicts as under Claude Code AND any OpenCode-specific divergence is recorded as a documented harness-difference rather than a failure.
- **Given** an OpenCode user reads abcd's installation docs, **when** they reach the harness-specific section, **then** the docs enumerate per-harness configuration (slash command discovery, MCP setup if applicable, subagent invocation differences) — Claude Code is documented as one of two equally-supported harnesses, not as the default.

## Open Questions

- Does OpenCode have an MCP equivalent or a different extension model? Affects oracle backend wiring.
- How does OpenCode handle the agent dispatch primitive — same `Task` tool semantics, or different?
- What's the testing strategy — run the entire acceptance matrix through both harnesses, or sample?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
