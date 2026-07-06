# ADR-02: MCPBridge Implementation Contract

## Status

Accepted

References ADR-01 § 4. Does NOT supersede ADR-01.

ADR-01's Status field remains "Accepted (Phase 0 lock)". This ADR closes the implementation-level
gaps that ADR-01 § 4 explicitly deferred, without reopening the Phase 0 locked decisions (§§ 1–3)
or the result-type design (§ 5).

---

## Context

ADR-01 § 4 locked `mcp_call` as a synchronous wrapper around the official `mcp` Python SDK via
stdio transport, and noted that the following implementation-level decisions were deferred to Phase 1:

> "Implementation details — server discovery, subprocess management, session lifetime, per-call
> timeout, cancellation — are deferred to Phase 1 when the first concrete harness implementation
> is written."

That deferral point has been reached. Task `.3` (fn-4-phase-0-p1-patch-viability-framing-mcp.3)
executed an in-process MCP spike against the RepoPrompt binary (`rp-cli (repoprompt-mcp) 2.1.23`)
and produced `spike-mcp-evidence.md` with results against seven criteria. This ADR derives its
decisions from that evidence.

**Evidence file:** `.abcd/development/research/phase/0/spike-mcp-evidence.md`

The five gap areas from ADR-01 § 4 that this ADR closes, plus one new:

1. Server discovery — config file resolution order, environment variable overrides
2. Server startup — cold-start vs reuse semantics, subprocess management
3. Session-lifecycle scope — bound to spike Criterion 3b (cross-session `chat_id` outcome)
4. Timeout — dimensions, defaults, mechanism
5. Cancellation — `anyio.fail_after` semantics, `KeyboardInterrupt` handling
6. RP-unavailable typed exception — reserved as `RPUnavailable`, declaration deferred to itd-6

**Pre-existing vocabulary** (ADR-01 / `scripts/abcd/harness.py`):
- `MCPBridge` — the Protocol declared in harness.py
- `mcp_call(server, tool, args) -> McpResult` — the single method on `MCPBridge`
- `chat_id` — field on `McpResult`; must be threaded back on re-review
- `dispatch_agent(agent_name="codex", ...)` — the RP-unavailable fallback path

**Newly reserved by this ADR** (not pre-existing):
- `RPUnavailable` — exception name reserved here; concrete `class RPUnavailable(OSError)`
  declaration deferred to itd-6. Downstream epics (fn-2, fn-3) CANNOT import this symbol
  until itd-6 ships.

---

## Decision

### 1. Server Discovery — Env Override Wins, First-Match-Wins

The `MCPBridge` concrete implementation MUST resolve the RP MCP server configuration in this
order:

1. `ABCD_MCP_CONFIG` environment variable — absolute path to a JSON config file
2. Project-local `.mcp.json` — `<cwd>/.mcp.json`
3. User-global `~/.claude/mcp.json`

**First-match-wins semantics.** The first location that exists and contains a valid `"RepoPrompt"`
server entry is used. Remaining locations are not consulted. There is no layered merge.

**Expected JSON shape** (each level, key `"RepoPrompt"`):

```json
{
  "mcpServers": {
    "RepoPrompt": {
      "command": "/path/to/repoprompt-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

The `ABCD_MCP_CONFIG` path MUST be an absolute path to a file with this shape (entire config, not
just the server entry). Relative paths are not supported via the env override.

**Rationale:** Env override enables CI and test harnesses to redirect to a stub server without
editing project or user files. First-match-wins is predictable and debuggable; layered merge
introduces non-obvious precedence questions (which key wins when both project and user files
define `"args"`?). abcd does not currently have a use case for layered merge.

**Rejected alternatives:**
- Hard-coded binary path — brittle; breaks on non-standard installs and multi-version setups.
- Layered merge — deferred to itd-6 if a real use case emerges. First-match-wins is the current
  commitment; changing to merge requires a new ADR.

### 2. Server Startup — Lazy Spawn, Warm Through Lifecycle Scope

The RP MCP subprocess is spawned on the first `mcp_call` invocation in a command's lifecycle
(lazy / spawn-on-first-call). The subprocess is held warm for the duration of the lifecycle
scope chosen in § 3. The subprocess is always reaped when the `AsyncExitStack` exits.

**Rejected alternatives:**
- Pre-warm on import — the import of `harness.py` has no command context; no `ABCD_MCP_CONFIG`
  or project root is known at import time. Pre-warming would require the server to always be
  reachable, turning an optional dependency into a hard one.
- Per-call spawn — see § 3 (per-call lifecycle).

### 3. Session-Lifecycle Scope — Per-Command, FINAL

**Evidence from spike Criterion 3b (`spike-mcp-evidence.md § 2 / § 3`):**

> "Result: NO — cross-session `chat_id` reuse is IMPOSSIBLE with RepoPrompt"
>
> "Definitive answer: NO (version 2.1.23). Cross-session chat_id reuse is structurally
> impossible because: 1. Each new RP stdio session requires GUI-level connection approval.
> 2. Scripts cannot bypass this approval gate (no token, flag, or socket path override).
> 3. Therefore, no second session can be established from a CLI command invocation."

**Decision: per-command lifecycle, FINAL.**

One stdio session per `abcd-cli` invocation that uses MCP. The `AsyncExitStack` opens lazily
on the first `mcp_call` (not at command start — see § 2) and closes at command exit. No
persistent session manager, no session cache.

**Same-chat semantics narrowing (itd-6):** "same chat" means "within one `abcd-cli` command
invocation's stdio session." Cross-invocation chat continuation is not supported and is not
achievable without GUI re-approval. Any abcd audit-fix loop that spans multiple `abcd-cli`
invocations starts a fresh chat context on each invocation.

Within a single invocation, the implementation MUST thread `chat_id`: the first tool call
returns `chat_id` in `McpResult`; subsequent tool calls in the same established stdio session
MUST pass that `chat_id` in `arguments`; RP is expected to maintain chat continuity within an
approved session. The harness `mcp_call` MUST return `chat_id` in `McpResult` (per ADR-01 § 4),
and the audit-fix loop in `oracle.py` MUST thread that ID back on re-calls within the same
command invocation.

**Evidence note:** SDK `chat_id` threading plumbing was verified by stub (spike Criterion 3a).
Real RP within-session chat continuity is architecturally expected (RP documentation confirms
same-session threading) but was not directly verified by the spike — the approval barrier
prevented real-RP tool calls from a standalone script. Itd-6 implementation must confirm
real-RP within-session continuity as part of the first concrete `MCPBridge` integration test.

**Evidence cited (at least three findings):**

1. Criterion 3b (§ 2): "cross-session `chat_id` reuse is IMPOSSIBLE with RepoPrompt" —
   the approval gate (RP macOS app GUI) fires on each new subprocess; scripts cannot obtain
   an approval token.
2. § 4 Subprocess-Collision Observations: "New processes that connect without prior GUI
   approval are denied within ~1s" — the denial comes as `McpError: Connection closed` during
   `session.initialize()`.
3. § 6 Recommendations: "Per-command lifecycle: One stdio session per `abcd-cli` invocation.
   Session opens at command start, closes at command exit (via AsyncExitStack). No persistent
   session manager; no session cache." (Note: this ADR refines the spike's "command start"
   phrasing to lazy/spawn-on-first-call per § 2; the lifecycle boundary — one session per
   invocation, closed on exit — is unchanged.)

**Rejected alternatives:**
- Per-call lifecycle (separate subprocess per `mcp_call` invocation) — `chat_id` continuity
  would be lost between calls (no cross-session reuse per Criterion 3b evidence), and subprocess
  thrash would be observable on the RP macOS approval queue. Rejected: load-bearing failure mode.
- Per-process lifecycle (one session for the whole Python process lifetime) — would require a
  dedicated-thread session manager and a persistent approval token. Not supported by RP 2.1.23;
  Criterion 3b eliminates the technical path. Rejected: no spike evidence supporting it;
  requires a new ADR if RP adds session persistence in a future version.

### 4. Timeout — Two Dimensions, Combined Startup Budget

The spike confirmed that **two independent timeout dimensions are required**, and that the
`anyio` cancel-scope constraint (python-sdk #521) collapses the startup budget to a single
value covering both subprocess spawn and `session.initialize()`. There is no separate
`initialize_timeout_s` parameter; one is not feasible.

#### 4a. `startup_timeout_s` (combined — subprocess spawn + MCP handshake)

**Default: 10.0 s.**

This single budget covers:
- `stack.enter_async_context(stdio_client(params))` — subprocess spawn and stdio connection
- `await session.initialize()` — MCP handshake (the python-sdk #1452 hang path)

**Mechanism:** `anyio.fail_after(startup_timeout_s)` wraps only the startup phase (spawn +
initialize). The `AsyncExitStack` itself stays open for the lifetime of the command so the
session is held warm for subsequent `call_tool` calls (see § 2 and § 4b).

```python
# Correct lifecycle: AsyncExitStack opened once, session held warm through command lifetime.
async with AsyncExitStack() as stack:
    # Phase A: startup — single combined timeout covering spawn + initialize
    with anyio.fail_after(startup_timeout_s):    # TimeoutError → RPUnavailable
        read, write = await stack.enter_async_context(stdio_client(params))
        session = await stack.enter_async_context(ClientSession(read, write))
        try:
            await session.initialize()           # sdk #1452 hang path
        except McpError as e:
            raise RPUnavailable("RP connection denied or closed") from e

    # Phase B: tool calls — separate per-call timeout within the live session
    with anyio.fail_after(call_timeout_s):       # TimeoutError → RPUnavailable
        result = await session.call_tool(tool, arguments=args)

# AsyncExitStack.__aexit__ runs here at command exit, reaping subprocess
```

**Why `startup_timeout_s` must NOT wrap the full `AsyncExitStack`:** doing so would cancel
the entire stack (including any in-flight `call_tool`) after the startup budget expires,
making long oracle calls (600 s `context_builder`) impossible. The two cancel scopes are
independent: `startup_timeout_s` guards only the startup sequence; `call_timeout_s` guards
each tool call independently.

**Evidence for default (spike § 5 § 2 Criterion 7b):**

| Phase | Observed wall-clock | Included in budget |
|-------|---------------------|-------------------|
| subprocess spawn | < 0.5 s | yes |
| socket connection + RP retries | 0.5–1.5 s | yes |
| `session.initialize()` (approved) | < 1 s | yes |
| Total observed | < 3 s | — |
| Recommended budget | 10.0 s (3.3× headroom) | — |

**Why a separate `initialize_timeout_s` is infeasible:**

Spike Criterion 7b (§ 2) confirmed that placing a second `anyio.fail_after` around
`session.initialize()` nested inside the outer `AsyncExitStack` block causes:

```
RuntimeError: Attempted to exit a cancel scope that isn't the current task's
              current cancel scope
```

This is the exact python-sdk #521 failure mode. `anyio` 4.x requires all cancel scopes to
exit in strict LIFO order within the same task; nesting `fail_after` inside an async context
manager that itself manages cancel scopes violates this constraint. A single combined budget
is not a simplification — it is the only pattern that does not trigger this RuntimeError.

**Mechanism rule (original):** `anyio.fail_after` only. `asyncio.wait_for` is explicitly
rejected (python-sdk #521: `asyncio.wait_for` wraps futures with `asyncio`-level cancel
that conflicts with `anyio`'s cancel-scope bookkeeping).

**Erratum (fn-5.2 implementation):** The ADR originally specified `anyio.fail_after` for
Phase A startup. In practice this causes the same RuntimeError it was meant to avoid:
`stdio_client` internally opens an anyio TaskGroup (its own cancel-scope tree), so wrapping
the entire `stdio_client` + `ClientSession.initialize()` sequence with an outer
`anyio.fail_after` creates a cancel scope that anyio refuses to exit while `stdio_client`'s
inner scopes are still active. The implementation therefore uses `asyncio.timeout()` (Python
3.11+ stdlib) for the startup phase only — it is a native asyncio mechanism that does not
participate in anyio's cancel-scope bookkeeping and does not trigger the nesting conflict.

Per-call `anyio.fail_after` around `session.call_tool(...)` is intentionally kept: the
`call_tool` coroutine does not open additional anyio task groups, so the single-level
`fail_after` remains safe there.

**Subprocess reaping on expiry:** when `TimeoutError` fires, `AsyncExitStack.__aexit__`
runs and tears down `stdio_client`, which reaps the subprocess. Spike Criteria 7a and 7b
(§ 2) confirmed zero orphan processes in both the slow-spawn and hung-initialize cases.

#### 4b. `call_timeout_s` (per-tool call budget)

**Default: 30.0 s for unknown tools.**

Per-tool overrides (derived from task spec; spike § 5 recommended narrower values — these
overrides extend spike § 5 to cover RP-specific heavy-oracle patterns):

| Tool pattern | Timeout |
|---|---|
| `tools/list`, read-only listing tools | 2 s |
| `oracle_send`, `context_builder` | 600 s (heavy sub-agent, long LLM runs) |
| `chat-send` (RP chat continuation) | 120 s |
| All other tools | 30 s (default) |

**Mechanism:** `anyio.fail_after(call_timeout_s)` wraps `await session.call_tool(...)`. This
is a separate cancel scope from the startup scope and is applied within the already-established
session context.

**Subprocess reaping on expiry:** Spike Criterion 4 (§ 2) confirmed that `anyio.fail_after`
expiry during `session.call_tool` tears down the `AsyncExitStack`, reaping the subprocess with
no orphans observed.

### 5. Cancellation — `anyio.fail_after`, Clean `AsyncExitStack` Teardown

`anyio.fail_after` raises `TimeoutError` on expiry; `KeyboardInterrupt` (SIGINT) propagates
through `AsyncExitStack.__aexit__` cleanly.

**Evidence from spike (§ 2):**

- Criterion 4: `TimeoutError` from `anyio.fail_after` around `call_tool` — no orphan process.
- Criterion 5: `KeyboardInterrupt` (`os.kill(os.getpid(), signal.SIGINT)`) fired while tool
  call in flight — `asyncio.CancelledError` / `KeyboardInterrupt` caught inside the
  `AsyncExitStack`; transport and subprocess cleaned up; no zombie observed.
- Criterion 6: Normal `return` from inside `async with AsyncExitStack()` — clean `__aexit__`;
  subprocess reaped; no orphan.

**Rejected alternative:** Silent timeout-then-retry (suppress `TimeoutError`, re-issue the
same `mcp_call`). Rejected because: (a) the subprocess has been reaped and the stdio session
is closed after timeout; a silent retry would spawn a new subprocess, incurring another startup
cost and RP approval gate; (b) callers must be able to distinguish "RP timed out → fall through
to Codex" from "RP returned an error result". Typed exceptions (`RPUnavailable`) are the
correct signal.

### 6. RP-Unavailable Typed Exception — Reserved, NOT Declared

**`RPUnavailable` is a reserved exception name only.** The concrete declaration:

```python
class RPUnavailable(OSError): ...
```

is **deferred to itd-6**, where the first concrete `MCPBridge` implementation will be written.
Until itd-6 ships, callers expecting RP-unavailable signalling MUST catch `OSError` per
ADR-01 L334-338.

**Sub-class hint:** `RPUnavailable` is a sub-class of `OSError`. This is compatible with
ADR-01's specification that `mcp_call` raises `OSError or equivalent` on infrastructure
failure.

**Detection mechanism — all reachable failure paths map to `RPUnavailable`:**

The `oracle.py` cascade needs a single exception type to catch and route to
`dispatch_agent(agent_name="codex", ...)`. All five reachable failure paths converge to
`RPUnavailable`:

1. **Subprocess spawn failure** (binary not found, permission denied):
   `RPUnavailable("server binary not found at <path>")`

2. **`startup_timeout_s` expiry** (combined budget covering both `stdio_client.__aenter__`
   AND `session.initialize()` — the python-sdk #1452 hang path guard):
   `RPUnavailable("server startup/handshake timed out after <N>s")`

   A separate `initialize_timeout_s` path is infeasible (python-sdk #521 forbids nested cancel
   scopes; spike Criterion 7b confirms `RuntimeError` from nested `anyio.fail_after`). The
   combined `startup_timeout_s` budget IS the only guard for the #1452 hang path.

3. **Non-timeout `session.initialize()` failure** (specifically `McpError: Connection closed`
   from RP approval denial or RP-side socket close):
   `RPUnavailable("RP connection denied or closed: <reason>")`

   Spike § 4 confirms this is how RP 2.1.23 surfaces cross-session denial — the RP macOS app
   sends `McpError: Connection closed` within ~1 s of the subprocess requesting a new
   connection without approval. This is NOT a timeout; it is an `McpError` exception that MUST
   be caught at `session.initialize()` and re-raised as `RPUnavailable`.

4. **`call_timeout_s` expiry on `session.call_tool(...)`**:
   `RPUnavailable("MCP tool call timed out after <N>s: <tool>")`

   Spike Criterion 4 confirms `anyio.fail_after` around `call_tool` fires and reaps the
   subprocess correctly.

5. **Mid-call stdio / broken-pipe transport failure on `session.call_tool(...)`**
   (python-sdk #396 — RP subprocess dies mid-call without timeout firing):
   `RPUnavailable("MCP transport failure during <tool>: <reason>")`

**Required error mapping at `session.initialize()` (pseudocode for itd-6):**

```python
# AsyncExitStack opened lazily on first mcp_call; held warm through all tool calls until command exit
async with AsyncExitStack() as stack:
    # Startup phase: combined budget for spawn + initialize (path 2 guard)
    with anyio.fail_after(startup_timeout_s):    # TimeoutError → RPUnavailable (path 2)
        read, write = await stack.enter_async_context(stdio_client(params))
        session = await stack.enter_async_context(ClientSession(read, write))
        try:
            await session.initialize()
        except McpError as e:
            # path 3: "Connection closed" = RP approval denied or RP-side socket close
            raise RPUnavailable("RP connection denied or closed") from e
    # Session is now live — stack remains open for call_tool calls
```

**Caller cascade contract (once `RPUnavailable` exists in itd-6):**
`oracle.py` catches `RPUnavailable` and routes to `dispatch_agent(agent_name="codex", ...)`.
This is the RP-unavailable fallback path specified in ADR-01 § 4 and itd-6.

**IMPORTANT — downstream epics:** fn-2 (`fn-2-move-repoprompt-review-artifacts-into`) and
fn-3 (`fn-3-strengthen-intent-stage-abcdgrill-skill`) and any other downstream epic CANNOT
import `RPUnavailable` until itd-6 ships. Until then, catch `OSError`.

---

## Alternatives Considered

**In-place addendum to ADR-01 § 4.** Adding the implementation contract directly into ADR-01
by appending to § 4. Rejected because: ADR-01 is locked at "Accepted (Phase 0 lock)"; editing
its body implies the Phase 0 lock has been reopened. A separate ADR preserves the lock
semantics and makes the Phase 0 → Phase 1 boundary explicit in the decision record.

**Pluggy-style backend registry.** Using `pluggy` for server discovery and lifecycle hooks.
Rejected: pluggy is designed for large plugin ecosystems where third-party plugins register
hooks without modifying the host. abcd's MCPBridge has exactly one concrete implementation.
pluggy adds a runtime dependency, a hook specification ceremony, and a registration layer
that serves no purpose at this scale. Overkill for a 2-backend shim.

**Async-first re-litigation.** Re-opening ADR-01 § 3 (sync vs async interface). Rejected:
ADR-01 § 3 is a Phase 0 lock. The harness method signature remains synchronous (`def mcp_call`);
the concrete sync↔async bridge mechanism (e.g. a dedicated event-loop thread holding the
`AsyncExitStack` warm, a per-invocation `asyncio.run()` scoped to the full command, or another
pattern) is an implementation detail deferred to itd-6. The bridge must preserve the
command-scoped warm session across multiple `mcp_call` invocations within one `abcd-cli`
execution — per-call `asyncio.run()` is NOT viable because it tears down the event loop
(and thus the live `AsyncExitStack`/`ClientSession`) between calls.

**Per-call lifecycle (separate subprocess per `mcp_call` invocation).** A fresh subprocess
spawned on every `mcp_call`, no state held between calls. Rejected because: (a) `chat_id`
continuity is structurally impossible across subprocesses (Criterion 3b: cross-session
`chat_id` reuse FAILS); (b) subprocess spawn cost (~0.5 s) plus RP approval latency (~1 s)
would be incurred on every oracle tool call; (c) RP's macOS approval queue would see a new
approval request per call, which is user-visible and disruptive.

**Per-process lifecycle without spike evidence.** A persistent session shared across multiple
`abcd-cli` invocations (daemon model). Rejected: Criterion 3b eliminates the technical path
(each new subprocess requires fresh GUI approval; a daemon cannot obtain approval without user
interaction). Requires RP to add session persistence in a future version; any future move to
this model requires a new ADR.

**Untyped fallback / silent return of `None`.** On RP unavailability, return `None` or a
sentinel `McpResult` rather than raising an exception. Rejected: callers cannot distinguish
RP-unavailable (route to Codex) from tool-returned-empty (continue with empty result). The
cascade in `oracle.py` requires a typed exception.

**Concrete `RPUnavailable` declaration in this ADR.** Declaring the class here rather than
reserving the name for itd-6. Rejected: the concrete class belongs alongside the concrete
`MCPBridge` implementation in itd-6. This ADR establishes the contract (name, base class,
detection mechanism); the implementation belongs in the same phase as the first concrete
harness implementation. Declaring here would create an orphan class with no implementation,
risking premature import in fn-2/fn-3 before the full itd-6 contract is in place.

**Layered merge for server discovery.** Merging all matching config levels (env + project +
user) rather than first-match-wins. Rejected: no concrete use case for layered merge
has been identified; first-match-wins is unambiguous. Deferred to itd-6 if a real need emerges.

---

## Consequences

### Enables

- **Phase 1 oracle code** (`oracle.py`, `lifeboat-oracle`, `press-release-composer`,
  `intent-fidelity-reviewer`) can be written against the `MCPBridge` protocol with a fully
  specified implementation contract. No further ADR is needed to implement the concrete harness.
- **itd-6 close-out is unblocked.** ADR-02 provides the implementation contract; itd-6 delivers
  the concrete `MCPBridge` implementation + `RPUnavailable` declaration.
- **Test stubs** can simulate RP-unavailable by raising `OSError` (until itd-6); after itd-6,
  raising `RPUnavailable`.

### Forecloses

- **Per-process (persistent) lifecycle.** Criterion 3b evidence establishes that per-process
  lifecycle is not technically feasible with RP 2.1.23. Any future move to persistent sessions
  (e.g., if RP adds session persistence or an approval token mechanism) requires a new ADR.
- **Separate `initialize_timeout_s` parameter.** python-sdk #521 (anyio cancel-scope constraint,
  confirmed by spike Criterion 7b) makes a nested cancel scope around `session.initialize()`
  infeasible. The combined `startup_timeout_s` is the only viable pattern. This forecloses any
  future `initialize_timeout_s` without a change to the timeout mechanism itself.
- **Same-chat re-review across separate `abcd-cli` invocations.** Narrowed by Criterion 3b.
  "Same chat" = within one invocation. Resuming a prior chat across invocations requires GUI
  re-approval and is out of scope for autonomous operation.

### Risks

- **RP approval model changes.** If RP 2.1.23+ adds a token-based approval mechanism (API key,
  stored approval token), the per-command lifecycle commitment becomes unnecessarily conservative.
  The `MCPBridge` interface contract (`mcp_call(server, tool, args)`) is abstract enough that
  the concrete implementation in itd-6 can be upgraded without an interface change.
- **`startup_timeout_s = 10.0 s` headroom may prove insufficient** if RP's approval retry loop
  extends beyond the observed 1–2 s (spike § 5). The default is configurable; `ABCD_MCP_CONFIG`
  (or a future `ABCD_STARTUP_TIMEOUT_S` env override) provides an escape valve.
- **fn-2/fn-3 cannot import `RPUnavailable` until itd-6 ships.** If fn-2/fn-3 proceed before
  itd-6, they MUST catch `OSError` in the interim. Any code review for fn-2/fn-3 MUST verify
  this constraint.

---

## Implementation Notes

_Additive errata recorded as the fn-5 epic (`fn-5-rp-mcp-integration-declare`)
implemented this contract. The Decision sections above and the Status field are
unchanged — these notes record what the implementation clarified, not a re-decision.
Historical reservation prose (§ 6 "Reserved, NOT Declared") is preserved as the
record of the pre-fn-5 state._

### Implementation note (added by fn-5, 2026-05-17)

`RPUnavailable` is now declared. The concrete `class RPUnavailable(OSError)` lives in
`scripts/abcd/exceptions.py` (fn-5 `.1`). The spawn-mode `MCPBridge` implementation —
the lazy-spawn lifecycle, two-dimension timeout, and `RPUnavailable` error-mapping of
§§ 2–6 — is in `scripts/abcd/mcp_bridge.py` (fn-5 `.2` / `.5`). The host-reuse code path
is the ADR-03 extension (fn-5 `.6` / `.7`). Downstream epics fn-2
(`fn-2-move-repoprompt-review-artifacts-into`) and fn-3
(`fn-3-strengthen-intent-stage-abcdgrill-skill`) now hard-depend on fn-5 (wired via
`flowctl epic add-dep`) and may import `RPUnavailable` directly, or catch `OSError`
(which covers it as a subclass). The § 6 "IMPORTANT — downstream epics" caveat
("CANNOT import `RPUnavailable` until itd-6 ships") is superseded by this dependency.

### Implementation erratum (added by fn-5, 2026-05-17)

The § 4b timeout table lists a `tools/list = 2 s` row. As implemented, this is **not** a
per-tool `call_timeout_s` override for a public tool — `tools/list` is an **internal**
MCP protocol RPC (`ClientSession.list_tools()`), NOT a callable RP tool surface.
`MCPBridge.mcp_call(server, tool, args)` rejects `tool == "tools/list"` at the public
boundary: it raises `ValueError` ("tools/list is a protocol RPC, not a callable tool")
rather than dispatching it. `DEFAULT_TOOL_TIMEOUTS` in `scripts/abcd/mcp_bridge.py`
therefore carries no `tools/list` entry — only the genuine heavy-tool overrides
(`oracle_send`, `context_builder`, `chat-send`).

Erratum, doc-vs-code: an earlier draft of this note (and the fn-5 `.8` task spec it was
inherited from) referred to a `MCPBridge.list_tools_timeout_s = 2.0` attribute as the
home of the table's 2 s budget. **No such `list_tools_timeout_s` attribute exists in the
shipped fn-5 `MCPBridge`.** Because `tools/list` is rejected at the boundary, the 2 s row
has no implemented counterpart at all; it should be read purely as ADR-era guidance for
an internal capability probe, not as a live `MCPBridge` field. The table row is retained
above as the historical record; the implementation simply does not expose `tools/list`.

### Implementation note 2 (added by fn-5, 2026-05-17)

The example exception messages in § 6 (`"server binary not found at <path>"`,
`"server startup/handshake timed out after <N>s"`, etc.) are **behavioural intent**, NOT
a normative `__str__` format. The fn-5 implementation builds `RPUnavailable` from a
structured `Reason` enum and a narrow, deliberately terse `__str__`; the exception's
machine-meaningful detail is on `__cause__` and the `Reason` value, not the message
string. The test suite asserts on `__cause__` / `Reason`, never on message text. The
§ 6 strings document which failure path maps to which reason — but the
`messages are non-normative` and callers MUST NOT pattern-match on them.

### Implementation note 3 (added by fn-5, 2026-05-17)

The § 3 per-invocation session lifecycle ("one stdio session per `abcd-cli` invocation",
"`AsyncExitStack` opens lazily … closes at command exit") is specifically the
**spawn-mode** contract. ADR-03 (`03-mcpbridge-host-reuse-extension.md`, fn-5) adds a
host-reuse mode where the MCP session lifecycle is owned by the injected host harness
rather than by an `abcd-cli` invocation. The two are not contradictory: § 3 governs
spawn mode, ADR-03 governs host-reuse mode, and the `MCPBridge` selects between them at
construction. "Same chat" in both modes still means same-MCP-session — see ADR-03 for
the host-reuse lifecycle boundary.

---

## Related ADRs

- [ADR-01: Harness Interface Design](01-harness-interface.md) — Phase 0 lock on the
  `Harness` Protocol, `MCPBridge`, `mcp_call` signature, and result type design. This ADR
  extends ADR-01 § 4 without superseding it.
- Extended by [ADR-03: MCPBridge Host-Reuse Extension](03-mcpbridge-host-reuse-extension.md)
  (`adr-3-mcpbridge-host-reuse-extension`, fn-5) — adds an additive host-reuse code
  path selected at construction when a caller injects a host harness. The
  spawn-mode contract in this ADR is unchanged.
- itd-6 (`rp-mcp-only-integration`) — architectural lock: RP is MCP-only; same-chat semantics;
  oracle cascade contract. ADR-02 narrows itd-6's same-chat semantics to within-one-invocation.

---

## Related Documentation

- `.abcd/development/research/phase/0/spike-mcp-evidence.md` — spike evidence this ADR is
  derived from; Criteria 1–7b results
- `.abcd/development/research/phase/0/spikes/mcpbridge_probe.py` — spike implementation
- `scripts/abcd/harness.py` — the interface file governed by ADR-01 and extended by this ADR
- `.abcd/development/intents/planned/itd-6-rp-mcp-only-integration.md` — RP MCP
  architectural lock; same-chat semantics (narrowed by this ADR)
- `.abcd/development/intents/drafts/itd-22-opencode-portability.md` — portability target for a later phase
  target; concrete MCPBridge implementation must remain swappable
