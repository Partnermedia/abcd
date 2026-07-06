# ADR-03: MCPBridge Host-Reuse Extension

## Status

Accepted; Extends ADR-02 — additive only; spawn-mode contract unchanged.

ADR-02's spawn-mode `MCPBridge` contract is preserved verbatim. This ADR adds a
second, additive code path (host-reuse) that is selected at construction time
when a caller injects a host harness. ADR-01's Status field remains "Accepted
(Phase 0 lock)" and ADR-02's Status field remains "Accepted" — neither is
modified by this ADR.

---

## Context

ADR-01 § "MCPBridge" (the Protocol docstring at `scripts/abcd/harness.py:295-306`)
records the dominant call path: abcd-cli runs from the bash wrapper
(`scripts/abcd-cli`) *outside* any host tool-use loop. For that caller, the
host's tool-use loop is not available, so a direct in-process MCP client
(subprocess-spawn) is the only viable transport. ADR-02 then specified the full
spawn-mode implementation contract — server discovery, lazy spawn, per-command
session lifecycle, two-dimension timeout, and the `RPUnavailable` error mapping.

Subprocess-spawn is, and remains, the primary v1 path. But it is not the only
caller shape abcd's `MCPBridge` will ever see. A *different* caller — a Claude
Code skill that imports abcd's Python directly, or a future OpenCode harness —
runs *inside* a host tool-use loop and already holds a live harness object that
can reach RepoPrompt. For such a caller, spawning a second RP subprocess is
both redundant (the host already has an MCP path) and harmful (each new RP
stdio session requires GUI-level approval per ADR-02 § 3 / spike Criterion 3b).

This ADR codifies the asymmetry: spawn is the production path for the dominant
abcd-cli caller; host-reuse is an additive path, plumbed for callers that
construct `MCPBridge` with a host harness. It does not contradict the
spawn-default — it extends the same Protocol to a second caller shape.

The code half of this work shipped in fn-5 `.6` (`scripts/abcd/mcp_bridge.py`
dual-mode selection + `tests/abcd/test_mcp_bridge_host_reuse.py`). This ADR is
authored against that landed implementation.

---

## Decision

Six points define the host-reuse extension.

1. **Optional `harness` constructor parameter.** `MCPBridge.__init__` accepts
   an optional `harness: Optional[Harness] = None`. The concrete signature is
   `harness: Optional[Any] = None` because `harness.py`'s `Harness` Protocol is
   `@runtime_checkable` and a structural-typing import is not required at the
   `mcp_bridge.py` boundary; the *contract* is "an object exposing `mcp_call`".

2. **Optional injectable `host_probe`.** `MCPBridge.__init__` accepts an
   optional `host_probe: tuple[str, dict] | None = None`, defaulting to
   `("oracle_utils", {"op": "models"})`. **The default is an ASSUMPTION about
   the canonical RP tool surface, not formally committed spike evidence** — the
   fn-4 spike verified subprocess lifecycle, not the `oracle_utils op=models`
   tool. The probe is injectable for forward-compat and for callers needing a
   probe verified against their own RP version. Host callers SHOULD pass an
   explicit `host_probe` unless they have verified `oracle_utils op=models`
   against their live RP. The probe is the canonical RP tool surface, NOT the
   MCP-protocol-level tool-listing RPC.

3. **`harness is None` → spawn-mode.** When no harness is injected, the bridge
   is in spawn-mode and behaves *exactly* as ADR-02 specifies — verbatim, no
   change. The constructor performs NO I/O on this path (lazy spawn on first
   `mcp_call`).

4. **`harness is not None` AND probe succeeds → host-reuse-mode.** When a
   harness is injected and the host probe returns a conformant `McpResult` with
   `is_error == False`, the bridge locks into host-reuse-mode. In this mode
   every `mcp_call` is a **pure passthrough** to `harness.mcp_call` — the bridge
   validates input, applies a per-thread re-entrancy guard, then delegates
   verbatim. The host harness owns timeout, cancellation, and error mapping.

5. **Probe failures fall back to spawn; other exceptions propagate.** A probe
   that fails in a *recognised* way — `NotImplementedError`, `RPUnavailable`, an
   MCP method-not-found error (`-32601`), a harness object with no `mcp_call`
   attribute, a non-`McpError` tool-level `is_error == True` result — sets the
   mode to spawn and caches the failure cause for the both-modes-fail terminal.
   Any *other* exception propagates from `__init__`: an unexpected failure is a
   bug and must not be auto-masked. A probe result that is reachable but not
   `McpResult`-shaped is a hard `TypeError` — selecting host-reuse for a
   non-conformant harness would leak raw SDK objects past the bridge boundary.

6. **Selection is fixed at `__init__`; immutable; no runtime mode-switching.**
   `_mode` is set once during construction and never changes for the bridge's
   lifetime. No public method switches it. A caller wanting the other mode
   constructs a new bridge.

### Reference pseudocode — `__init__` selection logic

This is the single source of truth that fn-5 `.6` implements against.

```python
def __init__(self, *, harness=None, host_probe=None, eager_probe=False, ...):
    # ... spawn-mode state init (ADR-02) ...
    self._harness = harness
    self._host_probe = host_probe if host_probe is not None else DEFAULT_HOST_PROBE
    self._mode = SPAWN                       # default
    self._select_reason = "spawn_default"
    self._cached_host_exc = None             # host-probe failure cause, if any

    if harness is not None:
        self._select_mode_via_host_probe()   # may lock HOST_REUSE, or fall back
                                             # to SPAWN caching the cause

    if eager_probe and self._mode == SPAWN:
        self._ensure_started_with_host_fallback()
```

`DEFAULT_HOST_PROBE = ("oracle_utils", {"op": "models"})`. `SPAWN` and
`HOST_REUSE` are the two `_mode` values; `_mode` is immutable post-`__init__`.

### Reference pseudocode — host-probe failure state machine

```python
def _select_mode_via_host_probe(self):
    harness = self._harness                  # caller guarantees not None
    probe_tool, probe_args = self._host_probe

    # #1 — harness object has no mcp_call attribute → fall back to spawn
    if not hasattr(harness, "mcp_call"):
        fall_back_to_spawn("no_mcp_call", NotImplementedError(...))
        return

    try:
        result = harness.mcp_call(RP_SERVER_KEY, probe_tool, probe_args)
    except NotImplementedError as exc:        # #2 — harness declares no MCP
        fall_back_to_spawn("NotImplementedError", exc); return
    except RPUnavailable as exc:              # #3 — host reached RP, RP is down
        fall_back_to_spawn("RPUnavailable", exc); return
    except BaseException as exc:              # #4 / #6
        if is_method_not_found(exc):          # #4 — MCP -32601, probe tool absent
            fall_back_to_spawn("ToolNotFound", exc); return
        raise                                 # #6 — unexpected: PROPAGATE the bug

    # #5 — reachable but result is not McpResult-shaped → HARD TypeError
    if not is_mcp_result_shaped(result):
        raise TypeError("host harness probe did not return a McpResult ...")

    # #7 — reachable host, probe tool reported a tool-level error
    if result.is_error:
        fall_back_to_spawn("is_error_true", RuntimeError(...)); return

    # SUCCESS — lock HOST_REUSE
    self._mode = HOST_REUSE
    self._select_reason = "host_probe_ok"
    self._cached_host_exc = None
```

State-machine summary:

| # | Probe outcome | Mode | Raises at `__init__`? |
|---|---------------|------|-----------------------|
| 1 | harness has no `mcp_call` | SPAWN (cached cause) | no |
| 2 | `NotImplementedError` | SPAWN (cached cause) | no |
| 3 | `RPUnavailable` | SPAWN (cached cause) | no |
| 4 | MCP method-not-found (`-32601`) | SPAWN (cached cause) | no |
| 5 | reachable, non-`McpResult` shape | — | yes — `TypeError` |
| 6 | any other unexpected exception | — | yes — propagated |
| 7 | reachable, `is_error == True` | SPAWN (cached cause) | no |
| ✓ | conformant `McpResult`, `is_error == False` | HOST_REUSE | no |

Every fallback branch (#1–#4, #7) caches its failure cause in
`_cached_host_exc`. When the bridge later falls through to spawn-mode and spawn
*also* fails, the both-modes-fail terminal raises
`RPUnavailable(HOST_AND_SPAWN_FAILED)` whose `__cause__` is an `ExceptionGroup`
carrying both the cached host-probe cause and the spawn failure — so a caller
can see why *neither* mode worked. If spawn-mode then succeeds, the cached
cause is discarded so a later transport failure is not misclassified.

---

## Lifecycle

Per-bridge-instance in either mode — this matches ADR-02 § 3's per-command
lifecycle narrowing applied to a single bridge object. Cross-bridge `chat_id`
reuse remains undefined behaviour, exactly as under ADR-02.

- **Spawn-mode:** the bridge owns the RP subprocess. The stdio session opens
  lazily on the first `mcp_call`, is held warm for the bridge's lifetime, and
  `close()` drains it and reaps the subprocess (ADR-02 § 2 / § 3 verbatim).
- **Host-reuse-mode:** the bridge owns NOTHING at the transport level. It
  spawns no subprocess, opens no stdio session, installs no signal handler, and
  starts no loop thread. The injected harness owns the subprocess and session
  lifecycle entirely. The bridge's `mcp_call` is a pure passthrough; `close()`
  on a host-reuse bridge is a no-op at the transport level (there is nothing to
  reap).

"Same chat" semantics scope to a single `MCPBridge` instance / single MCP
session: in spawn-mode that is one `abcd-cli` invocation's stdio session; in
host-reuse-mode it is the injected harness's session lifetime. `chat_id`
threading within a session works identically in both modes — the passthrough
neither adds nor strips `chat_id`.

---

## Host-reuse contract for harness implementations

A harness used as the host-reuse backend MUST satisfy the following contract.
fn-5's bridge in host-reuse-mode does **pure passthrough** — it does not wrap,
remap, or translate the harness's behaviour — so the contract is the harness's
responsibility, not the bridge's.

1. **Typed transport failures.** Host harnesses MUST raise `RPUnavailable`
   (declared in `scripts.abcd.exceptions`) for transport-layer failures —
   server unreachable, connection closed, broken pipe, timeout. fn-5's bridge
   does NOT catch and remap lower-level exception types. If a host raises
   `TimeoutError`, `BrokenPipeError`, or any other lower-level type for a
   transport failure, that is a **contract bug in the host harness**, not
   something fn-5 wraps. (Spawn-mode `MCPBridge` does this mapping itself per
   ADR-02 § 6; host-reuse-mode delegates it to the harness.)

2. **Result neutrality.** A host harness MUST return `scripts.abcd.harness.McpResult`
   (snake_case `is_error` / `content`), NOT a raw `mcp.types.CallToolResult`
   (camelCase `isError`). The host probe enforces this: a reachable harness
   whose probe result is not `McpResult`-shaped is rejected with a `TypeError`
   at `__init__` (state machine #5) precisely so non-conformant SDK objects
   never leak past the `MCPBridge` boundary.

3. **No re-entrant delegation.** A host harness MUST NOT delegate RepoPrompt
   MCP calls back to the *same* `MCPBridge` instance. Doing so would recurse
   unboundedly. The bridge guards against this with a per-thread re-entrancy
   flag: a same-thread re-entrant `mcp_call` raises `RuntimeError` to fail-fast
   rather than recurse. (The guard is per-thread so concurrent independent
   passthrough calls from different threads do not false-trip it.)

---

## Consequences

- **(a) Callers do not change.** Existing abcd-cli callers construct
  `MCPBridge()` with no harness and get the ADR-02 spawn-mode contract
  unchanged. Host-reuse is opt-in via the new `harness=` kwarg.
- **(b) Tests cover both modes.** fn-5 `.6` ships
  `tests/abcd/test_mcp_bridge_host_reuse.py` for the host-reuse path alongside
  `tests/abcd/test_mcp_bridge.py` for the spawn path — different failure modes,
  separately exercised.
- **(c) Future transports extend the same Protocol.** A third transport (e.g.
  an OpenCode harness) is another `harness=` injection, not a new bridge class
  or a Protocol change.
- **(d) The cascade resolver stays on top of `MCPBridge`.** The `oracle.py`
  RP→Codex cascade is built *on* `MCPBridge`, not *inside* it. Mode selection
  is a `MCPBridge` concern; backend cascade is a caller concern.

---

## Rejected alternatives

- **(a) Spawn-only with subprocess collision documented.** Keep `MCPBridge`
  spawn-only and just document that a host-side caller will collide with the
  host's own RP subprocess. Rejected: a documented footgun is still a footgun;
  the additive host-reuse path removes the collision entirely for host-side
  callers.
- **(b) Host-only with no spawn fallback.** Make `harness` mandatory and drop
  spawn-mode. Rejected: the dominant abcd-cli caller has no harness — it runs
  outside any host tool-use loop (ADR-01 § "MCPBridge"). Spawn-mode is the
  production path and cannot be dropped.
- **(c) Polymorphic dispatch via subclasses.** `SpawnBridge` and `HostBridge`
  subclasses with a factory. Rejected: it forces callers to know which subclass
  to construct, splits the `close()` / lifecycle surface, and obscures the
  shared input-validation path. A single class with an immutable `_mode` is
  simpler and keeps one Protocol surface.
- **(d) Amend ADR-02.** Fold host-reuse into ADR-02 by editing its Decision
  section. Rejected: the spawn-mode contract is genuinely unchanged; editing
  ADR-02's body implies its "Accepted" status was reopened. A separate
  extension ADR keeps the spawn contract's lock semantics intact.
- **(e) An adapter that remaps host-side errors to `RPUnavailable` inside
  fn-5.** Wrap the host passthrough in a try/except that catches
  `TimeoutError` / `BrokenPipeError` / etc. and re-raises `RPUnavailable`.
  Rejected: the host harness is *upstream* of the contract; remapping its
  errors inside fn-5 violates pure-passthrough and creates a dual source of
  truth for error semantics (the harness's mapping and fn-5's remapping could
  diverge). A host harness raising lower-level types is a host contract bug, to
  be fixed in the host — not papered over in the bridge.

---

## References

- **ADR-01 § 4** (`./01-harness-interface.md`) — `mcp_call` semantics; the
  "MCPBridge" Protocol docstring recording that abcd-cli runs outside the host
  tool-use loop.
- **ADR-02** (`./02-mcpbridge-implementation-contract.md`, full) — the
  spawn-mode implementation contract this ADR extends without superseding:
  server discovery, lazy spawn, per-command lifecycle, two-dimension timeout,
  `RPUnavailable` error mapping.
- **Spike evidence** — `.abcd/development/research/phase/0/spike-mcp-evidence.md`;
  Criterion 3b establishes cross-session `chat_id` reuse is impossible, which
  is why a host-side caller should reuse the host's session rather than spawn a
  second RP subprocess.
- **fn-5 epic spec** — `.flow/specs/fn-5-rp-mcp-integration-declare.md`.
- **fn-5 `.6`** (`fn-5-rp-mcp-integration-declare.6`) — the host-reuse backend +
  dual-mode selection implementation this ADR records
  (`scripts/abcd/mcp_bridge.py`, `tests/abcd/test_mcp_bridge_host_reuse.py`).
- **fn-5 `.7`** (`fn-5-rp-mcp-integration-declare.7`) — this task: authored ADR-03
  and the ADR-01/ADR-02 cross-links.

---

## Related ADRs

- [ADR-01: Harness Interface Design](01-harness-interface.md) — Phase 0 lock
  on the `Harness` Protocol and `mcp_call` signature. This ADR extends ADR-01's
  `MCPBridge` to a second caller shape (host-injected) without changing its
  signature.
- [ADR-02: MCPBridge Implementation Contract](02-mcpbridge-implementation-contract.md)
  — the spawn-mode contract. This ADR extends ADR-02 additively; the spawn-mode
  contract is unchanged.
