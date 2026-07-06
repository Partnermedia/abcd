# ADR-01: Harness Interface Design

## Status

Accepted (Phase 0 lock)

---

## Context

abcd runs as a Claude Code plugin.  A later phase (itd-22) targets OpenCode.
Every command (``/abcd:ahoy``, ``/abcd:disembark``, ``/abcd:embark``,
``/abcd:launch``, ``/abcd:intent``) and every agent must communicate with its
host runtime — asking the user questions, dispatching sub-agents, calling MCP
tools, spawning background processes, scheduling wakeups, and resolving file
paths relative to the plugin root.

Without a stable, host-agnostic interface for these surfaces, the OpenCode port requires
command-layer rewrites.  With one, the OpenCode port is a second concrete implementation of
the same Protocol and the commands need no changes.

This ADR locks the interface contract at Phase 0 so that Phase 1+ command
implementations can be written against it without re-litigation.

**Architectural lock from itd-6:** abcd's only integration with RepoPrompt
(RP) is the MCP API.  abcd never picks a model, never reads RP's preset
selection, never spawns ``claude -p``, and makes no direct vendor LLM API
calls.  This lock shapes the ``mcp_call`` method semantics and the explicit
"what is NOT in the harness" list.

**Predecessor study (Task 1):** abcdSubZero's closest predecessor to this
module is ``abcd/orchestrator/cli_adapter_base.py`` — an ``abc.ABC`` with
``_build_command()`` as the single override point.  The shape is right; the
delivery mechanism (``abc.ABC`` inheritance, ``PipelineContext`` tight coupling,
88-file orchestrator) is the anti-pattern (see Drop D2 in predecessor-notes.md).

**idelphi rescue study (Task 3) anti-patterns informing this ADR:**

- **AP3 God-object engine accumulation** (``DelphiEngine.swift`` 589 → 1766
  lines post-rescue): harness must remain a *thin shim*, not grow into an
  orchestrator.  The scope-violation signal is **non-docstring lines** (``...``
  stubs, imports, dataclass fields, concrete logic): if those exceed ~200 lines,
  treat it as over-growth.  Docstring and comment weight is expected and
  acceptable in an interface-only file.
- **AP2 Routing state machine in the wrong layer** (``AppShell`` in
  ContentView): commands own their own decision logic; the harness must not
  embed routing state machines.  Its methods are pure request/response.
- **AP1 Wizard as a single file** (SetupWizardView 29KB): multi-step flows
  belong in commands, not in harness methods.  ``ask_user`` is a single-turn
  Q&A, not a wizard sequencer.

---

## Decision

### 1. ``typing.Protocol`` over ``abc.ABC``

**Chosen:** ``typing.Protocol`` (structural subtyping, PEP 544).

**Rationale:**

- Structural subtyping means the second harness (OpenCode) does not need to
  inherit from anything in the abcd package.  It just implements the same
  method signatures and passes the ``isinstance(impl, Harness)`` check at
  runtime (``@runtime_checkable``).
- No metaclass machinery at runtime; no ``ABCMeta`` registration.
- Full mypy/pyright support: ``--strict`` passes with zero errors on the
  interface file.
- abcdSubZero used ``abc.ABC`` with ``PipelineContext`` tight coupling (Drop
  D5 in predecessor-notes.md).  Structural typing cuts that coupling at source.

### 2. Six narrow Protocols + aggregate ``Harness``

**Chosen:** Six single-responsibility Protocols composed into one aggregate.

The six capabilities:

| Protocol | Method(s) | Rationale |
|----------|-----------|-----------|
| ``UserIO`` | ``ask_user`` | User interaction is one concern |
| ``AgentDispatch`` | ``dispatch_agent`` | Sub-agent dispatch is one concern |
| ``MCPBridge`` | ``mcp_call`` | MCP tool invocation is one concern |
| ``BackgroundExec`` | ``run_background`` | Background execution is one concern |
| ``Scheduling`` | ``schedule`` | Future wakeup is one concern |
| ``PluginContext`` | ``plugin_root`` | Path resolution is one concern |

**Rationale (Interface Segregation Principle):**

- A module that only asks the user questions depends only on ``UserIO``.  A
  test stub for that module implements only ``ask_user``.  Without narrow
  Protocols the stub must implement six methods to satisfy the type checker.
- An agent that only does oracle calls depends on ``MCPBridge`` +
  ``AgentDispatch``.  It does not import scheduling or background execution.
- Partial backends (e.g. a minimal in-test harness, a CI-only harness that
  cannot spawn background processes) can satisfy narrow Protocols without
  implementing the full ``Harness``.
- The aggregate ``Harness`` is available for callers that genuinely need
  everything (e.g. the top-level command dispatcher).

**Alternative rejected: God-Protocol.**  One 6-method ``Harness`` Protocol.
Rejected because it forces full implementation for any partial consumer, makes
test stubs heavier, and mirrors the ``DelphiEngine`` god-object anti-pattern
(AP3) from the idelphi rescue study.

### 3. Synchronous (``def``) interface — async comes in a later phase

**Chosen:** All six methods are synchronous (``def``).

**Rationale:**

- Claude Code's tool-use loop is synchronous.  An async interface now would
  require wrapping every call in ``asyncio.run()`` or ``asyncio.to_thread()``
  with no throughput benefit.
- abcd's harness calls are I/O-bound but short (user prompt round-trip, MCP
  call, subprocess spawn).  The complexity cost of mandatory async exceeds
  the throughput benefit at current scale.
- If the OpenCode harness implementation needs async semantics, it wraps the
  sync harness methods in ``asyncio.to_thread()`` at call sites — a
  well-trodden adapter pattern that does not require changing this interface.
- practice-scout: "single hardest thing to retrofit" — this is the explicit
  choice: sync now, async wrapping in the OpenCode port if needed.

**Alternative rejected: async-first interface.**  Every method declared
``async def``.  Rejected because: (a) mandatory event-loop coupling in a
CLI tool that is invoked synchronously from a bash wrapper; (b) all callers
require ``await`` even for operations that complete in < 1 ms; (c) testing is
significantly harder (``asyncio.run`` in every test or ``pytest-asyncio``
dependency).

### 4. ``mcp_call`` semantics — RP-only via direct MCP client (itd-6 lock)

**Chosen:** ``mcp_call(server, tool, args) -> McpResult`` is a synchronous
wrapper around the official ``mcp`` Python SDK using stdio transport to the
named MCP server.

**Rationale:**

- abcd CLI runs from the bash wrapper (``scripts/abcd-cli``) outside any agent
  tool-use loop.  The host's tool-use loop is not available there; abcd cannot
  emit MCP intents and wait for the host to route them.  Direct in-process MCP
  client is the only path that works.
- itd-6 locks RP as MCP-only: abcd calls RP via ``mcp__RepoPrompt__*`` tools
  exclusively.  ``mcp_call`` is the harness surface for that path.
- ``McpResult.chat_id`` is mandatory for RP oracle calls.  The audit-fix loop
  in ``oracle.py`` must thread the chat ID back on re-review (same-chat
  semantics per itd-6): never ``--new-chat``, never a fresh ``rp builder``.
- Current servers: ``"RepoPrompt"`` only.  No other MCP servers in scope.

**Codex CLI is a peer transport (not via mcp_call):**  Codex CLI subprocess
is invoked via ``dispatch_agent(agent_name="codex", ...)`` when RP is
unavailable (non-Mac / no-RP users).  This is NOT a rejection of Codex; it is
a routing decision.  Codex calls go through ``AgentDispatch``, not
``MCPBridge``.

**Alternative rejected: emit-MCP-intent dispatch via host's tool-use loop.**
abcd would signal "call this MCP tool" to the host (Claude Code / OpenCode)
and wait for the host to route the call.  Rejected because: abcd CLI runs
outside the host's tool-use loop (invoked from bash wrapper, not from inside
an agent session).  The host loop is not available at CLI invocation time.

**Alternative rejected (variant): model auto-detection / preset-switching for
RP.**  abcd reads RP's active preset and selects a model accordingly.
Rejected per itd-6: RP handles model selection internally; abcd never reads or
sets RP presets.

### 5. Result type design — frozen dataclasses, no vendor SDK leakage

**Chosen:** Five ``@dataclass(frozen=True)`` result types:  ``UserAnswer``,
``AgentResult``, ``McpResult``, ``BackgroundHandle``, ``ScheduleHandle``.

**Rationale:**

- Neutral shapes: no Anthropic ``Message``, no Codex ``CompletionResult``, no
  MCP SDK internal type appears in a return signature.  Callers can be written
  without any vendor SDK import.
- Shallow-frozen by default: ``@dataclass(frozen=True)`` prevents field
  reassignment, which is the most common mutation error.  Container fields
  (``list[str]``, ``list[Path]``, ``list[dict]``) remain internally mutable —
  callers must not mutate them.  This is a deliberate trade-off: using
  ``tuple`` and ``Mapping`` for all containers would be the fully-immutable
  alternative, but adds overhead and makes construction verbose at Phase 0 call
  sites.  If mutation bugs appear in practice, tighten to immutable containers
  at that point.
- practice-scout: "don't bake Anthropic in" — if ``AgentResult`` returned an
  ``anthropic.types.Message``, the OpenCode harness would be forced to
  construct ``anthropic.types.Message`` objects, creating a dependency on the
  Anthropic SDK from a non-Anthropic host.

**Alternative rejected: return SDK response objects directly.**  Return
``anthropic.types.Message`` from ``dispatch_agent``.  Rejected because it
leaks the Anthropic SDK type into the Protocol and forces the OpenCode
implementation to depend on the Anthropic SDK.

### 6. Protocol grouping — capability-based over lifecycle or layer

**Alternative groupings considered:**

- **Lifecycle-based** (``SessionStart``, ``SessionMid``, ``SessionEnd``):
  rejected because capability consumers do not align with lifecycle phases.
  An oracle agent needs ``MCPBridge`` at any lifecycle point; grouping by phase
  would produce artificial partitions.
- **Agent-vs-not-agent** (``AgentSurfaces``, ``HostSurfaces``): rejected
  because ``ask_user`` and ``plugin_root`` are host surfaces but unrelated in
  consumer scope.  A module reading the plugin root has no reason to depend
  on ``ask_user``.
- **Coarse two-group** (``Interaction``, ``Execution``): rejected because it
  merges ``mcp_call`` (synchronous, returns content) with ``run_background``
  (fire-and-forget) without ISP benefit.

Capability-based grouping wins: each consumer declares exactly the capabilities
it needs, no more.

---

## Alternatives Considered (summary)

**pluggy-based plugin registry.** pluggy is designed for large plugin
ecosystems where third-party plugins need to register hooks without modifying
the host.  abcd's portability target is exactly two harness implementations
(Claude Code now, OpenCode later).  pluggy adds a runtime dependency, a hook
specification ceremony, and a plugin registration layer that serves no purpose
at this scale.  Rejected: overkill for a 2-backend shim.

**LiteLLM-style provider normalisation.** LiteLLM normalises *LLM API call
shapes* across OpenAI, Anthropic, Google, and others.  abcd's harness
normalises *host runtime interactions* — asking the user a question, dispatching
a sub-agent, calling an MCP tool, managing background processes.  These are
completely different concern levels.  Rejected: wrong abstraction layer; also
conflicts with itd-6 (abcd makes no direct LLM API calls).

**abc.ABC (abstract base class).** Using ``ABCMeta`` forces the OpenCode port to
import and inherit from abcd's ``abc.ABC`` class.  That creates a hard
dependency from the OpenCode harness implementation on abcd's package — exactly the
coupling this shim is designed to eliminate.  ``typing.Protocol`` (PEP 544)
gives structural subtyping: the OpenCode port satisfies the Protocol by implementing the right
methods, without any import from abcd.  Rejected: forced inheritance blocks
structural typing.

**async-first interface (async def everywhere).** Claude Code's tool-use loop
is synchronous.  Making every harness method ``async def`` would require
``asyncio.run()`` or ``asyncio.to_thread()`` wrappers at every call site, with
no throughput benefit at current scale (harness calls are short I/O-bound
round-trips).  If the OpenCode port needs async semantics, the sync methods can be
wrapped in ``asyncio.to_thread()`` at the call site — a well-trodden adapter
pattern.  Rejected: complexity cost exceeds benefit at current scale.

**emit-MCP-intent dispatch via host's tool-use loop.** abcd CLI is invoked
from the bash wrapper (``scripts/abcd-cli``) outside any agent session.  The
host's tool-use loop is not available there; abcd cannot emit MCP intent signals
and wait for the host to route them.  A direct in-process MCP client (official
``mcp`` Python SDK, stdio transport) is the only viable path for RP calls.
Rejected: unavailable at CLI invocation time.

**God-Protocol (one 6-method Protocol).** A single ``Harness`` Protocol
listing all methods forces any partial consumer — a module that only asks the
user questions — to depend on all six methods and forces any test stub to
implement all six.  Narrow Protocols per capability (``UserIO``,
``AgentDispatch``, etc.) apply the Interface Segregation Principle and make
partial implementations and lightweight stubs straightforward.  Rejected:
violates ISP; mirrors the ``DelphiEngine`` god-object anti-pattern.

**Codex-direct subprocess via mcp_call.** Codex CLI is a subprocess transport,
not an MCP server.  Routing it through ``mcp_call(server="codex", ...)`` would
conflate two distinct transport mechanisms and mislead implementers into thinking
Codex exposes an MCP interface.  Codex is routed through
``dispatch_agent(agent_name="codex")`` — the subprocess transport path.
Rejected: wrong transport abstraction.

**Model auto-detection / preset switching for RP.** Rejected per itd-6: RP
handles model selection internally based on the user's configured presets.
abcd issues MCP calls and accepts whatever model RP routes to.  Any
model-selection logic in abcd would re-introduce the "which preset?" complexity
that itd-6 explicitly eliminates.

---

## What Is NOT in the Harness

The harness is a thin shim for host-specific surfaces only.  Anything
achievable with the standard library is done directly in the caller — not
routed through the harness.

| Capability | How to use | Why NOT in harness |
|-----------|------------|-------------------|
| File reads/writes | ``pathlib`` directly | stdlib; no host-specific surface |
| HTTP calls | ``urllib`` / ``httpx`` directly | stdlib; no host coupling |
| Git commands | ``subprocess.run(["git", ...])`` directly | stdlib; no host coupling |
| JSON parsing | ``json`` stdlib | stdlib; no host coupling |
| **Direct LLM API calls** (any vendor) | **Forbidden per itd-6** | abcd never calls Anthropic/OpenAI/Google APIs; use ``mcp_call`` to RP instead |
| **``claude -p`` subprocess spawning** | **Forbidden per itd-6** | abcd is the consumer, not the spawner; ``claude -p`` bypasses the RP MCP path |
| **Subprocess spawning of other vendor CLIs** (Gemini, etc.) | **Forbidden per itd-6** | same rationale as ``claude -p`` |
| **Codex CLI subprocess** | ``dispatch_agent(agent_name="codex", ...)`` | Codex CLI IS used as the fallback oracle transport when RP is unavailable; it is routed via ``AgentDispatch``, not spawned directly from commands |
| **Model picking / preset detection within RP** | **Forbidden per itd-6** | RP handles model selection internally; abcd issues MCP calls and accepts whatever model RP routes to |

**Thin-shim test:** "If you can't draw the request/response on a napkin
without referring to Anthropic SDK types, the abstraction line is in the wrong
place." (practice-scout)

**Anti-patterns from Task 3 (idelphi rescue study) explicitly avoided:**

- ``AppState`` god-object (AP3) → harness is a thin shim, not an orchestrator.
  Scope-violation signal: non-docstring lines (stubs, fields, concrete logic)
  exceeding ~200.  Docstring weight is expected and exempt from this count.
- ``AppShell`` routing state machine in the wrong layer (AP2) → harness methods
  are pure request/response; no routing logic inside them.
- Wizard multi-step state machines in one file (AP1) → ``ask_user`` is a
  single-turn Q&A; multi-step flows belong in commands.

---

## Consequences

### Enables

- **Phase 1+ commands** can be written against the ``Harness`` Protocol without
  knowing whether the host is Claude Code or OpenCode.
- **The OpenCode port (itd-22, a later phase)** requires only a new concrete ``Harness``
  implementation — no command-layer changes.
- **Test stubs** are lightweight: implement only the narrow Protocol(s) the
  module under test depends on.
- **mypy --strict** passes on the interface file with zero errors (verified at
  Phase 0 with Python 3.9 + mypy 1.18.2).

### Forecloses

- **Async-first harness**: the sync-first choice means any async wrapper
  for the OpenCode port must be added at the caller level (``asyncio.to_thread``), not inside
  the harness.  This is the accepted trade-off (revisit at the OpenCode port if OpenCode
  benchmarks show contention).
- **Per-vendor LLM adapters** (abcdSubZero's ``codex_adapter.py``,
  ``gemini_adapter.py`` pattern): itd-6 explicitly prohibits per-vendor adapters.
  RP is the only LLM surface; Codex is a CLI subprocess transport, not an LLM
  adapter.

### Risk

- **Async mismatch at the OpenCode port**: if OpenCode's plugin model requires async dispatch,
  the sync harness will need adapter wrappers.  This is explicitly accepted as
  the OpenCode port's problem to solve.  The alternative (async-first now) was judged more
  expensive now than the wrapper cost at the OpenCode port.

---

## Related Documentation

- ``scripts/abcd/harness.py`` — the interface file this ADR governs
- ``.abcd/development/intents/planned/itd-6-rp-mcp-only-integration.md``
  — architectural lock: RP is MCP-only
- ``.abcd/development/intents/drafts/itd-22-opencode-portability.md``
  — OpenCode portability target that harness.py is designed to serve
- ``.abcd/development/intents/drafts/itd-2-in-session-subagent-dispatch.md``
  — in-session subagent dispatch (one of the three oracle transports)
- ``.abcd/development/research/phase/0/predecessor-notes.md`` — Task 1 output;
  abcdSubZero's ``cli_adapter_base.py`` is the closest predecessor
- ``.abcd/development/research/phase/0/idelphi-rescue-study.md`` — Task 3
  output; anti-patterns AP1, AP2, AP3 inform the "what is NOT in the harness"
  section

---

## Related ADRs

- ADR-02: MCPBridge Implementation Contract
  — closes the implementation-level gaps deferred by § 4 (server discovery, session lifecycle,
  timeout, cancellation, RP-unavailable typed exception). Does NOT supersede this ADR.
- ADR-03: MCPBridge Host-Reuse Extension
  (`adr-3-mcpbridge-host-reuse-extension`, fn-5) — extends § 4's `MCPBridge` to a
  second caller shape (a host harness injected at construction) without changing
  the `mcp_call` signature. Additive only; does NOT supersede this ADR.
