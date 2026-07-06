---
id: itd-22
slug: opencode-portability
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
---

# abcd Runs in OpenCode Too

## Press Release

> **abcd reaches OpenCode as well as Claude Code.** abcd's core is a transport-agnostic Go engine, and hosts reach it through thin front doors. All the dev-sync and lifeboat infrastructure works identically whichever host is driving: OpenCode connects through the MCP surface and gets the same abcd conventions, intents, and lifeboats a Claude Code user gets through the CLI front door. Reaching a new host is adding a front door onto the same core, not porting the plugin.
>
> "We had a team standardising on OpenCode for cost reasons but couldn't use abcd," said Grace, VP Engineering. "OpenCode connects to abcd through its MCP surface. Same core, same conventions, same lifeboats. The cross-host reach removed our last objection."

## Why This Matters

abcd's core logic is transport-agnostic by design (ADR-23): it never calls a host's primitives directly. Host-specific concerns — prompting the user, dispatching an agent, oracle calls, scheduling — sit behind thin front doors (the CLI today; the MCP surface and a markdown plugin later). Reaching OpenCode is then a matter of the host speaking to abcd through the MCP surface, not a second copy of the plugin.

The portability promise is the entire reason for the transport-agnostic core. Delivering on it validates the architectural decision and opens abcd to the OpenCode ecosystem.

## What's In Scope

- The MCP front door exercised from OpenCode — the transport-agnostic Go core is reached identically to the CLI path
- Host-agnostic handling of: agent dispatch, user prompts, oracle calls, background tasks, scheduled wakeups, command discovery — expressed through the front-door interface, never host-specific plugin code
- Test corpus: same multi-repo validation set runs with OpenCode as the driving host
- Documentation: how to install + configure abcd under OpenCode
- Minimum-viable parity: every abcd command works under OpenCode (some convenience UX may differ)

## What's Out of Scope

- Third-host front doors (Cursor, Aider, etc.) — same pattern would apply but separate intents
- Cross-host lifeboats (a Claude Code-produced lifeboat must work in OpenCode and vice versa — this is the test, but lifeboat format itself is host-agnostic by design)
- Performance parity (OpenCode may be slower or faster; abcd just works)

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** abcd is driven from OpenCode, **when** the user runs `/abcd:ahoy install`, **then** the marker block is installed, the `.abcd/` namespace is created, and the install completion report renders identically to the Claude Code path — the surface text + actions are host-agnostic.
- **Given** the same project is opened in Claude Code on one machine and OpenCode on another with the same lifeboat, **when** both run `/abcd:disembark to <path>`, **then** the resulting lifeboats are byte-identical (modulo timestamps and oracle adapter chat IDs) AND embark of either lifeboat into the other host completes without error.
- **Given** OpenCode's MCP surface, **when** abcd's core issues an oracle call, **then** the call routes through the front door to OpenCode's available primitive AND the configured oracle adapters and host-delegated default honour the OpenCode-side capabilities — no host-specific code leaks into agent prompts.
- **Given** OpenCode's subagent dispatch primitive (per itd-2's wire-protocol contract), **when** abcd dispatches an in-session subagent under OpenCode, **then** the same `<verdict>` tag protocol holds AND the result-return path delivers output back to abcd's loop within the same iteration shape used under Claude Code.
- **Given** the corpus validation suite runs with OpenCode driving, **when** every command and acceptance gate completes, **then** the corpus passes with the same per-test verdicts as under Claude Code AND any OpenCode-specific divergence is recorded as a documented host-difference rather than a failure.
- **Given** an OpenCode user reads abcd's installation docs, **when** they reach the host-specific section, **then** the docs enumerate per-host configuration (command discovery, MCP setup, subagent invocation differences) — Claude Code is documented as one of two equally-supported hosts, not as the default.

## Open Questions

- Does OpenCode's MCP surface cover every abcd front-door call, or are there gaps? Affects oracle-call wiring.
- How does OpenCode handle the agent dispatch primitive — same `Task` tool semantics, or different?
- What's the testing strategy — run the entire acceptance matrix through both hosts, or sample?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
