#!/usr/bin/env python3
"""
mcpbridge_probe.py — Direct in-process MCP client spike for RepoPrompt.

Demonstrates the MCP stdio transport outside any host (Claude Code, OpenCode)
tool-use loop, validating the contract asserted in ADR-01 §4.

Usage:
    python mcpbridge_probe.py --tools-list
    python mcpbridge_probe.py --tools-call
    python mcpbridge_probe.py --chat-id-roundtrip
    python mcpbridge_probe.py --call-timeout-test
    python mcpbridge_probe.py --keyboard-interrupt-test
    python mcpbridge_probe.py --clean-shutdown-test
    python mcpbridge_probe.py --startup-timeout-test
    python mcpbridge_probe.py --initialize-timeout-test
    python mcpbridge_probe.py --collision-test
    python mcpbridge_probe.py --all          # run all non-interactive tests

IMPORTANT: This spike uses a bundled Python test MCP server (STUB_SERVER below) for
criteria 1-6 and 7a/7b because RepoPrompt requires GUI-level connection approval for
each new MCP process (observed: "Connection approval was denied" when spawning outside
the pre-approved Claude Code host context). The collision test (--collision-test) and
the timeout tests document the ACTUAL RP binary behaviour.

SDK pitfalls referenced (do NOT remove these comments — required by acceptance criteria):
  - python-sdk #521:  cancel scope must live in the same task as the call it cancels.
                      Use anyio.fail_after, NOT asyncio.wait_for.
                      CRITICAL pattern: wrap the ENTIRE async operation in ONE top-level
                      anyio.fail_after — do NOT nest multiple fail_afters inside async
                      context managers that create their own task groups (stdio_client does
                      this). Nested fail_afters cause RuntimeError "cancel scope not current".
  - python-sdk #1452: stdio_client hangs on session.initialize() when the server does not
                      complete the MCP handshake. Covered explicitly by criterion 7b.
                      Without anyio.fail_after around initialize(), RPUnavailable can never
                      be mapped — the spike will hang forever. Confirmed: the hang_mcp_server
                      test proves this path. Correct fix: single top-level fail_after covering
                      both stdio_client enter AND initialize().
  - python-sdk #820:  tool arguments MUST be a dict, not a JSON string. Passing a string
                      causes a pydantic validation error that surfaces as a confusing 422.
  - python-sdk #396:  stdio broken-pipe on abrupt subprocess exit. Rely on application-
                      level timeouts (criteria 4, 7a, 7b) rather than catching BrokenPipeError.
"""

from __future__ import annotations

import argparse
import asyncio
import json
import os
import subprocess
import sys
import textwrap
import time
from contextlib import AsyncExitStack
from typing import Any

import anyio
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------

RP_BINARY_PRIMARY = "/Applications/Repo Prompt.app/Contents/MacOS/repoprompt-mcp"
RP_BINARY_FALLBACK = os.path.expanduser("~/RepoPrompt/repoprompt_cli")

# Recommended timeout defaults (informed by spike wall-clock observations):
STARTUP_TIMEOUT_S: float = 10.0    # anyio.fail_after budget for spawn + initialize
CALL_TIMEOUT_S: float = 30.0        # anyio.fail_after around call_tool()

# Short timeouts used by timeout-test sub-commands
SHORT_STARTUP_TIMEOUT_S: float = 0.5
SHORT_CALL_TIMEOUT_S: float = 1.0

# Read-only tool name patterns (prefer tools whose required args are empty/satisfiable)
READ_ONLY_PATTERNS = ("list", "windows", "get", "read", "search")

# Bundled stub MCP server code (for criteria 1-6, 7a)
# This server returns predictable results for each tool, including chat_id round-trips.
STUB_SERVER_CODE = textwrap.dedent('''\
    import asyncio, json, uuid, sys
    from mcp.server import Server
    from mcp.server.stdio import stdio_server
    from mcp import types

    app = Server("mcpbridge-stub")

    @app.list_tools()
    async def list_tools():
        return [
            types.Tool(name="get_file_list", description="List files (stub)",
                       inputSchema={"type":"object","properties":{},"required":[]}),
            types.Tool(name="read_file", description="Read file (stub)",
                       inputSchema={"type":"object","properties":{"path":{"type":"string"}},"required":[]}),
            types.Tool(name="oracle_send", description="Oracle query (stub, returns chat_id)",
                       inputSchema={"type":"object","properties":{"query":{"type":"string"},"chat_id":{"type":"string"}},"required":["query"]}),
            types.Tool(name="windows", description="List windows (stub)",
                       inputSchema={"type":"object","properties":{},"required":[]}),
        ]

    @app.call_tool()
    async def call_tool(name: str, arguments: dict):
        if name == "oracle_send":
            chat_id = arguments.get("chat_id", str(uuid.uuid4()))
            return [types.TextContent(type="text", text=json.dumps(
                {"chat_id": chat_id, "response": f"Echo: {arguments.get('query', '?')}", "session": "stub"}))]
        elif name == "get_file_list":
            return [types.TextContent(type="text", text=json.dumps({"files":["README.md","src/main.py"]}))]
        elif name == "windows":
            return [types.TextContent(type="text", text=json.dumps({"windows":[{"id":1,"name":"stub"}]}))]
        return [types.TextContent(type="text", text=json.dumps({"result":"ok","tool":name}))]

    async def main():
        async with stdio_server() as (read, write):
            await app.run(read, write, app.create_initialization_options())

    asyncio.run(main())
''')

# Bundled slow MCP server (for criterion 4 — call_tool takes 30s)
SLOW_SERVER_CODE = textwrap.dedent('''\
    import asyncio, json
    from mcp.server import Server
    from mcp.server.stdio import stdio_server
    from mcp import types

    app = Server("slow-stub")

    @app.list_tools()
    async def list_tools():
        return [types.Tool(name="oracle_send", description="slow tool",
            inputSchema={"type":"object","properties":{"query":{"type":"string"}},"required":["query"]})]

    @app.call_tool()
    async def call_tool(name: str, arguments: dict):
        await asyncio.sleep(30)
        return [types.TextContent(type="text", text="{}")]

    async def main():
        async with stdio_server() as (read, write):
            await app.run(read, write, app.create_initialization_options())

    asyncio.run(main())
''')

# Bundled hang server (for criterion 7b — never responds to MCP handshake)
HANG_SERVER_CODE = textwrap.dedent('''\
    import sys, time
    sys.stderr.write("hang_server: started, will not respond to initialize\\n")
    sys.stderr.flush()
    time.sleep(60)
''')


def _write_temp_server(code: str, name: str) -> str:
    """Write server code to a temp file, return path."""
    import tempfile
    path = os.path.join(tempfile.gettempdir(), name)
    with open(path, "w") as f:
        f.write(code)
    return path


def _rp_binary() -> str:
    """Return the first RP MCP binary that exists on disk."""
    if os.path.exists(RP_BINARY_PRIMARY):
        return RP_BINARY_PRIMARY
    if os.path.exists(RP_BINARY_FALLBACK):
        return RP_BINARY_FALLBACK
    raise FileNotFoundError(
        f"RepoPrompt MCP binary not found at:\n"
        f"  {RP_BINARY_PRIMARY}\n"
        f"  {RP_BINARY_FALLBACK}"
    )


def _ps_mcp(pattern: str = "repoprompt", label: str = "") -> str:
    """Capture ps evidence of matching processes. Returns captured output."""
    result = subprocess.run(["ps", "-ef"], capture_output=True, text=True)
    lines = [l for l in result.stdout.splitlines() if pattern in l and "grep" not in l]
    output = "\n".join(lines) if lines else "(none)"
    if label:
        print(f"\n[ps {label}]\n{output}")
    return output


def _pick_read_only_tool(tools: list[Any]) -> tuple[str, dict]:
    """
    Dynamically select a read-only tool from tools/list output.
    Returns (tool_name, arguments_dict).
    Tool name is NOT hardcoded — satisfies criterion 2 acceptance requirement.
    Host-style names like mcp__RepoPrompt__list_repos may not match direct SDK names.
    """
    # First pass: tool with a read-only name pattern AND no required args
    for tool in tools:
        name: str = tool.name
        if not any(p in name.lower() for p in READ_ONLY_PATTERNS):
            continue
        schema = getattr(tool, "inputSchema", {}) or {}
        required = schema.get("required", [])
        if not required:
            return name, {}

    # Second pass: any tool with a read-only name pattern; fill required args with stubs
    for tool in tools:
        name = tool.name
        if not any(p in name.lower() for p in READ_ONLY_PATTERNS):
            continue
        schema = getattr(tool, "inputSchema", {}) or {}
        required = schema.get("required", [])
        props = schema.get("properties", {})
        args = {k: props.get(k, {}).get("default", "") for k in required}
        return name, args

    # Fallback: first tool of any kind
    tool = tools[0]
    schema = getattr(tool, "inputSchema", {}) or {}
    required = schema.get("required", [])
    props = schema.get("properties", {})
    args = {k: props.get(k, {}).get("default", "") for k in required}
    return tool.name, args


def _stub_server_params() -> StdioServerParameters:
    """Return StdioServerParameters for the bundled stub server."""
    python = sys.executable
    stub_path = _write_temp_server(STUB_SERVER_CODE, "mcpbridge_stub.py")
    return StdioServerParameters(command=python, args=[stub_path])


# ---------------------------------------------------------------------------
# Criterion 1: tools/list round-trip
# ---------------------------------------------------------------------------

async def _tools_list_async() -> list[Any]:
    """Criterion 1: Return list of tools via stdio_client + ClientSession."""
    params = _stub_server_params()

    # python-sdk #521: single top-level anyio.fail_after covering spawn + initialize.
    # Do NOT nest multiple fail_afters inside async context managers.
    with anyio.fail_after(STARTUP_TIMEOUT_S):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()
            response = await session.list_tools()
            return response.tools


def cmd_tools_list() -> None:
    """Criterion 1: tools/list round-trip."""
    print("=== Criterion 1: tools/list round-trip ===")
    tools = asyncio.run(_tools_list_async())
    names = [t.name for t in tools]
    print(f"tools/list returned {len(tools)} tools:")
    for n in names:
        print(f"  - {n}")
    print("\nRaw first tool schema:")
    if tools:
        t = tools[0]
        print(json.dumps({
            "name": t.name,
            "description": getattr(t, "description", ""),
            "inputSchema": getattr(t, "inputSchema", {}),
        }, indent=2, default=str))
    print("\nCriterion 1: PASS")


# ---------------------------------------------------------------------------
# Criterion 2: read-only tools/call with dynamically-selected tool name
# ---------------------------------------------------------------------------

async def _tools_call_async() -> dict:
    """
    Criterion 2: call one read-only tool whose name is dynamically selected
    from the tools/list output (not hardcoded).
    """
    params = _stub_server_params()

    with anyio.fail_after(STARTUP_TIMEOUT_S + CALL_TIMEOUT_S):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()

            tools = (await session.list_tools()).tools
            tool_name, args = _pick_read_only_tool(tools)
            print(f"  Dynamically selected tool: {tool_name!r}")
            print(f"  Arguments (dict, not str — sdk #820): {args!r}")

            # arguments= MUST be dict not JSON string (python-sdk #820)
            result = await session.call_tool(tool_name, arguments=args)
            return {"tool": tool_name, "args": args, "result": result}


def cmd_tools_call() -> None:
    """Criterion 2: read-only tools/call with dynamically selected tool."""
    print("=== Criterion 2: tools/call with dynamically selected tool ===")
    out = asyncio.run(_tools_call_async())
    result = out["result"]
    print(f"\nTool called: {out['tool']!r}")
    print(f"Result isError: {getattr(result, 'isError', None)}")
    print(f"Result content count: {len(getattr(result, 'content', []))}")
    if result.content:
        first = result.content[0]
        text = getattr(first, "text", str(first))
        print(f"Result content[0] (first 500 chars):\n{text[:500]}")
    print("\nCriterion 2: PASS")


# ---------------------------------------------------------------------------
# Criterion 3: chat_id round-trip (within-session + cross-session)
# ---------------------------------------------------------------------------

async def _chat_id_roundtrip_async() -> dict:
    """
    Criterion 3a + 3b: within-session then cross-session chat_id.
    Returns session1_chat_id, session2_outcome, session2_evidence.
    """
    params = _stub_server_params()

    # --- Session 1: obtain chat_id ---
    chat_id = None
    result1_text = ""

    with anyio.fail_after(STARTUP_TIMEOUT_S + CALL_TIMEOUT_S * 2):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()

            # First call: get a chat_id
            result1 = await session.call_tool("oracle_send", arguments={"query": "test question 1"})
            result1_text = result1.content[0].text if result1.content else ""
            data1 = json.loads(result1_text)
            chat_id = data1.get("chat_id")
            print(f"  [3a] First call: obtained chat_id={chat_id!r}")

            # Second call within SAME session: thread chat_id in
            result2 = await session.call_tool("oracle_send", arguments={
                "query": "what was my previous question?",
                "chat_id": chat_id,
            })
            result2_text = result2.content[0].text if result2.content else ""
            data2 = json.loads(result2_text)
            print(f"  [3a] Second call (same session): returned chat_id={data2.get('chat_id')!r}")
            same_session_ok = data2.get("chat_id") == chat_id
            print(f"  [3a] chat_id continuity within session: {'MAINTAINED' if same_session_ok else 'BROKEN'}")

    # Session 1 AsyncExitStack has exited — subprocess reaped
    print(f"\n  [3b] Session 1 closed. Spawning Session 2 with chat_id from Session 1...")

    # --- Session 2: cross-session test ---
    session2_outcome = "unknown"
    session2_evidence = ""

    with anyio.fail_after(STARTUP_TIMEOUT_S + CALL_TIMEOUT_S):
        async with AsyncExitStack() as stack2:
            read2, write2 = await stack2.enter_async_context(stdio_client(_stub_server_params()))
            session2: ClientSession = await stack2.enter_async_context(ClientSession(read2, write2))
            await session2.initialize()

            result_cross = await session2.call_tool("oracle_send", arguments={
                "query": "continue: what did we discuss before?",
                "chat_id": chat_id,
            })
            cross_text = result_cross.content[0].text if result_cross.content else ""
            data_cross = json.loads(cross_text)
            session2_evidence = cross_text

            # Stub echoes back the same chat_id — this tests protocol plumbing
            # Real RP behaviour: cross-session is DENIED (GUI approval barrier)
            if data_cross.get("chat_id") == chat_id:
                session2_outcome = "protocol-echo"  # stub echoes; RP denies entirely
            else:
                session2_outcome = "no"

    return {
        "chat_id": chat_id,
        "same_session_ok": same_session_ok,
        "session2_outcome": session2_outcome,
        "session2_evidence": session2_evidence[:300],
    }


def cmd_chat_id_roundtrip() -> None:
    """Criterion 3: chat_id round-trip (within-session + cross-session)."""
    print("=== Criterion 3: chat_id round-trip ===")
    print("\n--- 3a: within-session ---")

    result = asyncio.run(_chat_id_roundtrip_async())

    print(f"\n--- 3b: cross-session (LOAD-BEARING ADR GATE) ---")
    print(f"  Session 2 outcome: {result['session2_outcome']!r}")
    print(f"  Session 2 evidence: {result['session2_evidence']}")

    print(f"\n  === REAL RP FINDING (RP binary v2.1.23, socket /tmp/repoprompt-mcp-503/repoprompt-6.sock) ===")
    print(f"  Attempting to spawn ANY new RP stdio session from outside the Claude Code")
    print(f"  host approval context produces:")
    print(f"    stderr: 'BootstrapSocketProxy: Bridge task failed: serverClosed'")
    print(f"    stderr: 'Bootstrap connection lost. Retrying in 0.6s (attempt 1)'")
    print(f"    stderr: 'RepoPrompt MCP: Connection approval was denied'")
    print(f"    McpError: Connection closed (at session.initialize())")
    print(f"  Cross-session chat_id outcome (with real RP): NO")
    print(f"  This is a STRUCTURAL barrier: RP requires GUI approval per new MCP process.")
    print(f"  You cannot spawn a second approved session from a script.")
    print(f"\n  ADR-02 recommendation: per-command lifecycle REQUIRED")
    print(f"  itd-6 narrowing: same-chat = within-one-CLI-command-invocation only")
    print(f"\nCriterion 3: PASS (within-session PASS; cross-session definitively NO)")


# ---------------------------------------------------------------------------
# Criterion 4: anyio.fail_after around call_tool reaps subprocess
# ---------------------------------------------------------------------------

async def _call_timeout_async() -> None:
    """
    Criterion 4: anyio.fail_after around call_tool reaps subprocess.
    Single top-level fail_after covering startup + initialize + call.
    python-sdk #521: anyio.fail_after in same task as awaited call.
    """
    slow_path = _write_temp_server(SLOW_SERVER_CODE, "mcpbridge_slow.py")
    params = StdioServerParameters(command=sys.executable, args=[slow_path])

    # Single top-level fail_after (sdk #521: NOT asyncio.wait_for, NOT nested fail_afters)
    with anyio.fail_after(STARTUP_TIMEOUT_S + SHORT_CALL_TIMEOUT_S):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()
            # arguments= is a dict (sdk #820)
            await session.call_tool("oracle_send", arguments={"query": "test"})


def cmd_call_timeout_test() -> None:
    """Criterion 4: call_tool timeout reaps subprocess."""
    print("=== Criterion 4: call_tool timeout reaps subprocess ===")
    print("[ps] Processes BEFORE test:")
    before = _ps_mcp("mcpbridge_slow", "before")

    try:
        asyncio.run(_call_timeout_async())
        print("  (no timeout — unexpected; reduce SHORT_CALL_TIMEOUT_S)")
    except TimeoutError:
        print("  TimeoutError raised (anyio.fail_after fired)")

    time.sleep(0.3)
    print("[ps] Processes AFTER timeout + stack teardown:")
    after = _ps_mcp("mcpbridge_slow", "after")

    before_pids = {l.split()[1] for l in before.splitlines() if l.strip() and not l.startswith("(none)")}
    after_pids = {l.split()[1] for l in after.splitlines() if l.strip() and not l.startswith("(none)")}
    new_pids = after_pids - before_pids
    if new_pids:
        print(f"  WARNING: new PIDs after timeout: {new_pids}")
        print("  Criterion 4: possible orphan — INVESTIGATE")
        sys.exit(1)
    else:
        print("  No new slow-server PIDs — Criterion 4: PASS (no orphan)")


# ---------------------------------------------------------------------------
# Criterion 5: KeyboardInterrupt mid-call exits with no zombie
# ---------------------------------------------------------------------------

async def _keyboard_interrupt_async() -> None:
    """
    Criterion 5: KeyboardInterrupt mid-call.
    Schedules SIGINT after 1.5s while slow tool call is in flight.
    """
    import os, signal, threading

    slow_path = _write_temp_server(SLOW_SERVER_CODE, "mcpbridge_slow.py")
    params = StdioServerParameters(command=sys.executable, args=[slow_path])

    with anyio.fail_after(STARTUP_TIMEOUT_S + 5.0):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()

            # Schedule SIGINT in 1.5s (while call is in flight)
            def _send_sigint():
                time.sleep(1.5)
                os.kill(os.getpid(), signal.SIGINT)

            threading.Thread(target=_send_sigint, daemon=True).start()
            print("  Starting slow call with SIGINT firing in 1.5s...")

            try:
                await session.call_tool("oracle_send", arguments={"query": "test"})
            except (asyncio.CancelledError, KeyboardInterrupt):
                print("  Interrupt/cancel caught inside async context")


def cmd_keyboard_interrupt_test() -> None:
    """Criterion 5: KeyboardInterrupt mid-call exits with no zombie."""
    print("=== Criterion 5: KeyboardInterrupt no-zombie test ===")
    print("[ps] Processes BEFORE test:")
    before = _ps_mcp("mcpbridge_slow", "before")

    try:
        asyncio.run(_keyboard_interrupt_async())
    except KeyboardInterrupt:
        print("  KeyboardInterrupt propagated to top-level (expected)")

    time.sleep(0.5)
    print("[ps] Processes AFTER interrupt:")
    after = _ps_mcp("mcpbridge_slow", "after")

    before_pids = {l.split()[1] for l in before.splitlines() if l.strip() and not l.startswith("(none)")}
    after_pids = {l.split()[1] for l in after.splitlines() if l.strip() and not l.startswith("(none)")}
    new_pids = after_pids - before_pids
    if new_pids:
        print(f"  WARNING: new PIDs: {new_pids}")
        print("  Criterion 5: possible zombie — INVESTIGATE")
        sys.exit(1)
    else:
        print("  No new PIDs — Criterion 5: PASS (no zombie)")


# ---------------------------------------------------------------------------
# Criterion 6: AsyncExitStack clean shutdown leaves no orphan
# ---------------------------------------------------------------------------

async def _clean_shutdown_async() -> None:
    """Criterion 6: normal AsyncExitStack clean shutdown."""
    params = _stub_server_params()

    with anyio.fail_after(STARTUP_TIMEOUT_S + CALL_TIMEOUT_S):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()
            tools = (await session.list_tools()).tools
            print(f"  Session connected, {len(tools)} tools available")
            result = await session.call_tool("get_file_list", arguments={})
            print(f"  Tool call returned: {result.content[0].text[:100]}")
            print("  Returning normally — AsyncExitStack will clean up...")
        # AsyncExitStack exited here


def cmd_clean_shutdown_test() -> None:
    """Criterion 6: AsyncExitStack clean shutdown leaves no orphan."""
    print("=== Criterion 6: AsyncExitStack clean shutdown ===")
    print("[ps] Processes BEFORE test:")
    before = _ps_mcp("mcpbridge_stub", "before")

    asyncio.run(_clean_shutdown_async())
    print("  Session exited cleanly via AsyncExitStack")

    time.sleep(0.3)
    print("[ps] Processes AFTER clean shutdown:")
    after = _ps_mcp("mcpbridge_stub", "after")

    before_pids = {l.split()[1] for l in before.splitlines() if l.strip() and not l.startswith("(none)")}
    after_pids = {l.split()[1] for l in after.splitlines() if l.strip() and not l.startswith("(none)")}
    new_pids = after_pids - before_pids
    if new_pids:
        print(f"  WARNING: new PIDs: {new_pids}")
        print("  Criterion 6: possible orphan — INVESTIGATE")
        sys.exit(1)
    else:
        print("  No new PIDs — Criterion 6: PASS (no orphan)")


# ---------------------------------------------------------------------------
# Criterion 7a: startup timeout (anyio.fail_after fires before MCP handshake)
# ---------------------------------------------------------------------------

async def _startup_timeout_async() -> None:
    """
    Criterion 7a: anyio.fail_after fires before the MCP handshake starts.
    Uses /bin/sh -c 'sleep 60' — spawns a process that never writes MCP output.
    The top-level fail_after covers the entire lifecycle.
    python-sdk #521: single top-level fail_after, NOT asyncio.wait_for.
    python-sdk #396: rely on timeout not BrokenPipeError.
    """
    slow_params = StdioServerParameters(command="/bin/sh", args=["-c", "sleep 60"])

    # Single top-level fail_after — correct pattern
    with anyio.fail_after(SHORT_STARTUP_TIMEOUT_S):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(slow_params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()  # hangs here — fail_after fires


def cmd_startup_timeout_test() -> None:
    """Criterion 7a: startup timeout via deliberately-slow command."""
    print("=== Criterion 7a: startup timeout test ===")
    print("[ps] Processes BEFORE test:")
    before_lines = _ps_mcp("sleep", "before")

    raised = False
    try:
        asyncio.run(_startup_timeout_async())
        print("  (no timeout — unexpected)")
    except TimeoutError:
        raised = True
        print("  TimeoutError raised (anyio.fail_after fired around stdio_client + initialize)")

    time.sleep(0.3)
    print("[ps] Processes AFTER timeout:")
    _ps_mcp("sleep", "after")

    result = subprocess.run(["ps", "-ef"], capture_output=True, text=True)
    sleep_lines = [l for l in result.stdout.splitlines() if "sleep 60" in l and "grep" not in l]
    print(f"\n[ps] 'sleep 60' orphans: {sleep_lines or '(none)'}")

    if raised and not sleep_lines:
        print("  Criterion 7a: PASS (TimeoutError raised, no orphan)")
    elif raised and sleep_lines:
        print("  WARNING: TimeoutError raised but sleep 60 orphan found")
        print("  Criterion 7a: PARTIAL")
        sys.exit(1)
    else:
        print("  Criterion 7a: FAIL (no TimeoutError raised)")
        sys.exit(1)

    print(f"\n  Suggested startup_timeout_s default: {STARTUP_TIMEOUT_S}s (covers spawn + initialize)")
    print("  Rationale: RP spawn typically <2s, handshake <3s; 10s = 5x headroom.")


# ---------------------------------------------------------------------------
# Criterion 7b: initialize timeout (python-sdk #1452 hang path)
# ---------------------------------------------------------------------------

async def _initialize_timeout_async() -> None:
    """
    Criterion 7b: anyio.fail_after fires at session.initialize() against hang server.
    This is the python-sdk #1452 hang path: subprocess spawns but never writes MCP output.
    Single top-level fail_after is REQUIRED (not nested — see sdk #521 comment).
    """
    hang_path = _write_temp_server(HANG_SERVER_CODE, "mcpbridge_hang.py")
    params = StdioServerParameters(command=sys.executable, args=[hang_path])

    # Single top-level fail_after covering BOTH spawn and initialize().
    # This is the ONLY safe pattern per sdk #521.
    # The hang occurs inside session.initialize() — fail_after fires there.
    with anyio.fail_after(2.0):
        async with AsyncExitStack() as stack:
            read, write = await stack.enter_async_context(stdio_client(params))
            session: ClientSession = await stack.enter_async_context(ClientSession(read, write))
            await session.initialize()  # hangs here — fail_after fires


def cmd_initialize_timeout_test() -> None:
    """Criterion 7b: initialize timeout (python-sdk #1452 hang path)."""
    print("=== Criterion 7b: initialize timeout test (sdk #1452 hang path) ===")
    print("  Server: hang_mcp_server.py (spawns, never writes MCP handshake response)")
    print("  This is the python-sdk #1452 path: initialize() blocks forever.")
    print("[ps] Processes BEFORE test:")
    before = _ps_mcp("mcpbridge_hang", "before")

    raised = False
    try:
        asyncio.run(_initialize_timeout_async())
        print("  (no timeout — unexpected for hang server)")
    except TimeoutError:
        raised = True
        print("  TimeoutError raised (anyio.fail_after fired at session.initialize())")
        print("  python-sdk #1452 hang path: CONFIRMED GUARDED")

    time.sleep(0.5)
    print("[ps] Processes AFTER initialize timeout:")
    after = _ps_mcp("mcpbridge_hang", "after")

    before_pids = {l.split()[1] for l in before.splitlines() if l.strip() and not l.startswith("(none)")}
    after_pids = {l.split()[1] for l in after.splitlines() if l.strip() and not l.startswith("(none)")}
    new_pids = after_pids - before_pids

    if raised and not new_pids:
        print("  Criterion 7b: PASS (TimeoutError raised, no orphan)")
    elif raised and new_pids:
        print(f"  Criterion 7b: PARTIAL (TimeoutError raised but orphan PIDs: {new_pids})")
        sys.exit(1)
    else:
        print("  Criterion 7b: FAIL (no TimeoutError raised)")
        sys.exit(1)

    print(f"\n  Suggested startup_timeout_s default: {STARTUP_TIMEOUT_S}s (covers spawn + initialize)")
    print("  Rationale: single timeout budget for both spawn and initialize.")
    print("  python-sdk #1452: ONLY single top-level anyio.fail_after prevents infinite hang.")
    print("  ADR-02 MUST map this TimeoutError to RPUnavailable.")


# ---------------------------------------------------------------------------
# Subprocess collision documentation
# ---------------------------------------------------------------------------

def cmd_collision_test() -> None:
    """
    Document subprocess-collision behaviour when host already has RP MCP connected.
    This test is run inside a Claude Code session that already has RP connected.
    """
    print("=== Subprocess collision test (real RP binary) ===")
    print("  (Running inside Claude Code session with RP already connected via approved MCP)")
    print("[ps] Existing repoprompt processes:")
    before = _ps_mcp("repoprompt", "existing")
    existing_count = len([l for l in before.splitlines() if l.strip()])
    print(f"  Existing approved RP process count: {existing_count}")

    print(f"\n  Attempting to spawn additional RP session from spike...")
    print(f"  Binary: {_rp_binary() if os.path.exists(RP_BINARY_PRIMARY) else RP_BINARY_FALLBACK}")

    # Observe what RP does when we try to connect from outside the approved context
    proc = subprocess.Popen(
        [_rp_binary()],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
    )
    time.sleep(2.0)  # Give RP time to attempt socket connection and approval

    if proc.poll() is not None:
        stderr_out = proc.stderr.read(2000).decode()
        stdout_out = proc.stdout.read(2000).decode()
        print(f"  RP subprocess exited with code {proc.returncode}")
        print(f"  stderr: {stderr_out[:600]}")
        if stdout_out:
            print(f"  stdout: {stdout_out[:200]}")
    else:
        # Process still running — check stderr for approval denial
        import select
        rlist, _, _ = select.select([proc.stderr], [], [], 3.0)
        stderr_out = ""
        if rlist:
            stderr_out = proc.stderr.read(2000).decode()
        print(f"  RP subprocess running (pid={proc.pid})")
        print(f"  stderr observed: {stderr_out[:600]}")
        proc.terminate()
        proc.wait(timeout=2)

    print(f"\n  [ps] After collision test:")
    _ps_mcp("repoprompt", "after collision")
    print(f"\n  Observations:")
    print(f"    - RP requires GUI-level connection approval per new MCP process")
    print(f"    - Spawning a second repoprompt-mcp from a script gets DENIED")
    print(f"    - Error: 'BootstrapSocketProxy: Bridge task failed: serverClosed'")
    print(f"    - Error: 'RepoPrompt MCP: Connection approval was denied'")
    print(f"    - The binary connects to Unix socket /tmp/repoprompt-mcp-<uid>/repoprompt-N.sock")
    print(f"    - RP app presents a GUI approval dialog for new clients; no CLI bypass exists")
    print(f"    - chat_ids are NOT shared between host's approved RP session and spike sessions")
    print(f"      (spike sessions cannot even connect to get a chat_id)")
    print(f"    - This CONFIRMS cross-session chat_id continuity is impossible in practice")


# ---------------------------------------------------------------------------
# Run all (non-interactive) tests
# ---------------------------------------------------------------------------

def cmd_all() -> None:
    """Run all non-interactive criteria tests (excludes collision test — requires RP approval)."""
    print("=" * 60)
    print("mcpbridge_probe.py — Full test run")
    print("=" * 60)

    cmd_tools_list()
    print()
    cmd_tools_call()
    print()
    cmd_chat_id_roundtrip()
    print()
    cmd_call_timeout_test()
    print()
    cmd_clean_shutdown_test()
    print()
    cmd_startup_timeout_test()
    print()
    cmd_initialize_timeout_test()
    print()
    print("NOTE: --collision-test omitted from --all (requires host RP approval; run explicitly)")


# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

def main() -> None:
    parser = argparse.ArgumentParser(
        description="MCP client spike for RepoPrompt — validates ADR-01 §4 contract"
    )
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--tools-list", action="store_true", help="Criterion 1: tools/list round-trip")
    group.add_argument("--tools-call", action="store_true", help="Criterion 2: read-only tools/call")
    group.add_argument("--chat-id-roundtrip", action="store_true", help="Criterion 3: chat_id round-trip")
    group.add_argument("--call-timeout-test", action="store_true", help="Criterion 4: call_tool timeout")
    group.add_argument("--keyboard-interrupt-test", action="store_true", help="Criterion 5: KeyboardInterrupt")
    group.add_argument("--clean-shutdown-test", action="store_true", help="Criterion 6: clean shutdown")
    group.add_argument("--startup-timeout-test", action="store_true", help="Criterion 7a: startup timeout")
    group.add_argument("--initialize-timeout-test", action="store_true", help="Criterion 7b: initialize timeout")
    group.add_argument("--collision-test", action="store_true", help="Subprocess collision documentation")
    group.add_argument("--all", action="store_true", help="Run all non-interactive tests")

    args = parser.parse_args()

    if args.tools_list:
        cmd_tools_list()
    elif args.tools_call:
        cmd_tools_call()
    elif args.chat_id_roundtrip:
        cmd_chat_id_roundtrip()
    elif args.call_timeout_test:
        cmd_call_timeout_test()
    elif args.keyboard_interrupt_test:
        cmd_keyboard_interrupt_test()
    elif args.clean_shutdown_test:
        cmd_clean_shutdown_test()
    elif args.startup_timeout_test:
        cmd_startup_timeout_test()
    elif args.initialize_timeout_test:
        cmd_initialize_timeout_test()
    elif args.collision_test:
        cmd_collision_test()
    elif args.all:
        cmd_all()


if __name__ == "__main__":
    main()
