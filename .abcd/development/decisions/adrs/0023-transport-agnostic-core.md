---
id: adr-23
slug: transport-agnostic-core
status: accepted
date: 2026-07-06
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-21, adr-22, adr-25, adr-26]
---

# ADR-23: A transport-agnostic Go core behind thin front doors

## Context

abcd's old surface was welded to one host: behaviour lived in slash-commands,
hooks, and in-session Python that assumed Claude Code was driving. Adding a
second entry point (an MCP server, a scriptable CLI, a markdown plugin surface)
meant re-implementing logic against a different caller, because there was no
neutral place the logic lived independent of how it was invoked.

The rebuild ([ADR-21](0021-rebuild-in-go.md)) needs several front doors over
time — a Cobra CLI first, an MCP server and a markdown plugin surface later —
and cannot afford to re-express its decisions per surface.

## Decision

abcd's logic lives in a **transport-agnostic core** (`internal/core`) that does
deterministic work and returns **structured results**. It knows nothing about
who called it — no terminal formatting, no MCP framing, no slash-command
parsing. Front doors are thin: each parses its own input, calls the same core,
and renders the structured result in its own idiom.

- **CLI (Cobra)** is the first front door and ships in the MVP.
- **MCP server** and a **markdown plugin surface that shells to the binary**
  are later front doors — each is a new adapter over the unchanged core, not a
  reimplementation.

Adding a transport is therefore an additive change: write a thin door, map its
inputs to core calls, render the results. "Add MCP later" is trivial precisely
because the core never assumed a transport in the first place.

## Alternatives Considered

- **Logic in the CLI layer, MCP added by refactor later.** The obvious path.
  Rejected: it defers the seam to the moment it is most expensive to cut, and
  invites CLI-shaped assumptions (exit codes, stdout formatting) to leak into
  logic that MCP then has to unpick.
- **A shared library that still emits presentation.** A core that returns
  rendered strings. Rejected: presentation is transport-specific; a core that
  formats for a terminal cannot serve MCP's structured responses without each
  caller re-parsing text. Structured results in, structured results out.
- **Chosen: deterministic core returning structured results, thin front doors
  over it.** One place for behaviour; transports are cheap and additive.

## Consequences

- The core is directly testable without a transport — table tests over
  structured inputs and outputs, no terminal or protocol harness.
- The oracle and adapter seams ([ADR-22](0022-bundled-deps-as-pluggable-adapters.md),
  [ADR-25](0025-host-delegated-llm-default.md)) plug into the core, so every
  front door inherits the same adapters for free.
- Each front door owns its own input validation at the system boundary; the
  core trusts the structured values it receives from them.
- A behaviour is "wired" only when a front door reaches it — core code no front
  door calls is dead scaffolding, and the CLI is the surface that proves the
  MVP core is live.
