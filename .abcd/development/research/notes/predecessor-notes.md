# Predecessor Notes — Phase 0 Task 1

## TL;DR

- [REC] abcdZero's chezmoi-template approach conflates distribution with configuration; abcd should keep configuration as plain YAML/JSON in-repo and treat distribution as a separate concern.
- [REC] abcdSubZero's `CLIAdapterBase` abstraction is the right shape for harness.py — one ABC with `_build_command()` override per CLI tool, shared subprocess lifecycle; adapt, don't reinvent.
- [REC] abcdSubZero's FastMCP server (`abcd/mcp/server.py`) shows the integration pattern: `claude mcp add --transport stdio abcd -- python3 -m abcd.mcp.server`; abcd harness entry should follow the same stdio/JSON-RPC convention.
- [DROP] abcdSubZero's 88-file `orchestrator/` with 6 classifier files is the primary anti-pattern; the brief explicitly drops multi-classifier orchestration — abcd must not re-grow this surface area.
- [REC] Both predecessors used `flowctl` for task tracking (confirmed by CLAUDE.md in both); keep this — it is load-bearing, not incidental.

---

## Keep

### K1 — CLIAdapterBase subprocess pattern
**Source:** `abcdSubZero/abcd/orchestrator/cli_adapter_base.py:58-142`
**Tag:** [OBS] → [REC]
The abstract base class with `_build_command()` as the single override point, shared `_spawn_subprocess()`, SIGTERM/SIGKILL escalation (`SIGTERM_GRACE_SECONDS=10`), stdin piping toggle (`_STDIN_PROMPT`), and `WORKER_TIMEOUT=3600` is a clean, tested pattern. Task 5's harness.py shim should adapt this directly rather than reinventing subprocess management.

### K2 — flowctl as task tracker (confirmed in both predecessors)
**Source:** `abcdZero/.claude/CLAUDE.md:1-50`; `abcdSubZero/.claude/CLAUDE.md:33-50`
**Tag:** [OBS] → [REC]
Both predecessors' `CLAUDE.md` explicitly reference `flowctl` for task tracking and prohibit TodoWrite/markdown TODOs. This is not incidental — it is a proven practice. abcd's `CLAUDE.md` should carry this forward verbatim.

### K3 — FastMCP stdio server entry pattern
**Source:** `abcdSubZero/abcd/mcp/server.py:1-8`
**Tag:** [OBS] → [REC]
The `claude mcp add --transport stdio abcd -- python3 -m abcd.mcp.server` entry-point convention is the right shape for abcd's MCP surface. It keeps the server co-located with the plugin, avoids daemon management, and is what Claude Code natively expects. The optional-import guard (`try: from mcp.server.fastmcp import FastMCP`) is also worth porting — it allows `import abcd.mcp.server` to succeed even without the MCP SDK installed.

### K4 — Persona system as a separate module (not embedded in CLI)
**Source:** `abcdSubZero/abcd/persona/__init__.py:1-30`
**Tag:** [OBS] → [REC]
abcdSubZero separated persona logic (`BuilderArchetype`, `EndUserPersona`, `CommunicationAdapter`) into a standalone `abcd/persona/` module. This is the right boundary: communication style decisions belong in persona, not scattered across CLI handlers. abcd should plan for this separation from Phase 1.

### K5 — Session manager facade pattern
**Source:** `abcdSubZero/abcd/session/manager.py:1-60`
**Tag:** [OBS] → [REC]
`SessionManager` wraps `SessionState`, `CheckpointStore`, and `RecoveryEngine` behind a single facade with typed `WorkflowPosition` transitions. This state-machine approach to session lifecycle prevents ad-hoc state mutations. abcd should adopt a similar facade for its `ahoy`/`disembark`/`embark` session state rather than letting session state leak across commands.

### K6 — Non-root container user convention
**Source:** `abcdZero/.devcontainer/devcontainer.json:20` (`"remoteUser": "vscode"`)
**Tag:** [OBS] → [REC]
abcdZero already enforced non-root container execution. abcd's Docker isolation must continue this — run as unprivileged user, pass `ANTHROPIC_API_KEY` via environment, never embed credentials in image layers.

---

## Drop

### D1 — abcdZero's chezmoi-template scaffolding approach
**Source:** `abcdZero/templates/` (10+ `run_onchange_*.sh.tmpl` files); `abcdZero/.abcd.yaml:1-12`
**Tag:** [OBS] → [REC]
abcdZero used chezmoi's `run_onchange_` hooks to sync CLAUDE.md across stages, install logging hooks, set up Docker, and link project skills. This created a maintenance burden: every sync required a template re-run, template logic was hard to test, and the chezmoi dependency was non-obvious to new contributors. abcdSubZero dropped chezmoi entirely (no `templates/` directory, no `.chezmoi*` files). abcd should not revive it. Plain file copy + git hooks achieves the same result without the dependency.

### D2 — 88-file orchestrator with 6 classifier files
**Source:** `abcdSubZero/abcd/orchestrator/` (88 files); classifier files: `classifier.py`, `deployment_target_classifier.py`, `documentation_style_classifier.py`, `error_recovery_classifier.py`, `quality_strategy_classifier.py`, `user_personas_classifier.py`, `_classifier_utils.py`
**Tag:** [OBS] → [REC]
The brief explicitly drops multi-classifier orchestration (brief § 6.7 confirmed). The 88-file orchestrator surface area is the primary cause of complexity in abcdSubZero. Do not copy this pattern. abcd's harness.py should be a single file or a tight 3–4 file module (base + one adapter per CLI tool + registry). If the orchestrator grows beyond ~10 files, treat it as a scope violation.

### D3 — Gitpod integration
**Source:** `abcdZero/.gitpod.yml`, `abcdZero/.gitpod.Dockerfile`
**Tag:** [OBS] → [REC]
abcdZero shipped Gitpod support (prebuild config, Dockerfile, VS Code extension list). abcdSubZero dropped it. abcd should not revive it — it adds maintenance surface for a workflow that doesn't match the brief's Docker-isolation model.

### D4 — `version_check` / auto-update logic in CLI startup
**Source:** `abcdSubZero/abcd/cli/app.py:45-80`
**Tag:** [OBS] → [REC]
abcdSubZero checks for updates on every CLI invocation (`_show_update_notice`, `_show_plugin_update_notice`). This adds latency to every command, and the brief does not include an update mechanism. Drop for now; add via the `disembark` / `embark` surface when the distribution story is clearer.

### D5 — Inline `PipelineContext` threading throughout adapter base
**Source:** `abcdSubZero/abcd/orchestrator/cli_adapter_base.py:41` (`from abcd.orchestrator.pipeline import PipelineContext`)
**Tag:** [INF] → [REC]
`CLIAdapterBase` imports `PipelineContext` from a sibling module in the same 88-file package. This tight coupling forces the harness to carry the full pipeline graph even for simple subprocess calls. abcd's adapter base should accept a plain dataclass (or dict) context rather than importing from a wider pipeline module.

---

## Open Questions

### OQ1 — harness.py location (resolved by Task 5 spec)
**Affects:** Phase 1 Task 5 (harness.py scaffolding)
**Owner:** Phase 1 implementer (see Task 5 spec for canonical answer)
abcdSubZero's `cli_adapter_base.py` is a module inside a Python package (`abcd.orchestrator`). abcd is a Claude Code plugin. Task 5's spec already locates harness at `scripts/abcd/harness.py` (a package-style layout within `scripts/`). The adapter-boundary pattern from abcdSubZero is still worth respecting within that location; this OQ is recorded for traceability, not for re-litigation.

### OQ2 — codex `mcp-server` entrypoint: noted, not used
**Affects:** Phase 1 Task 5 (oracle backend wiring in harness.py)
**Owner:** Phase 1 implementer (decision already made — record only)
`codex mcp-server` is a confirmed entrypoint (verified at task time: `codex-cli 0.129.0`, `codex mcp-server --help` returns valid usage). **This is NOT an open question.** Per itd-6, RP MCP is the only LLM transport surfaced via `mcp_call`; Task 5's spec explicitly invokes Codex CLI as a subprocess via `dispatch_agent(agent_name="codex")`, not via `mcp_call`. Decision: Codex subprocess confirmed; `codex mcp-server` is NOT used in abcd.

---

## Tooling Verification

- `codex` path: `/opt/homebrew/bin/codex`
- `codex --version`: `codex-cli 0.129.0` (task spec expected `0.120.0`; actual is newer — OK)
- `codex mcp-server`: confirmed available (see OQ2 above)
- `rp-cli` path: `/usr/local/bin/rp-cli`
- `rp-cli --version`: `rp-cli (repoprompt-mcp) 2.1.23` (task spec required `≥ 2.1.15` — satisfied)

---

## Raw Notes

### abcdZero (Go CLI + chezmoi scaffold)

- `abcdZero/.abcd.yaml:1-12` — top-level config: `version: "1.0"`, `mode: lite`, `autonomy_level: collaborative`, `tool: claude-code`, `transparency.enabled: true`, `privacy.pii_scan_enabled: true`. This config shape informed abcd's config schema but chezmoi delivery is dropped.
- `abcdZero/templates/` — 10 `run_onchange_*.sh.tmpl` files for chezmoi hook-driven setup; includes `setup-abcd-command.sh.tmpl`, `setup-autonomous-docker.sh.tmpl`, `sync-*-claude.sh.tmpl` for stage CLAUDE.md sync. Empty on inspection (templates not expanded here). The *intent* (stage-level CLAUDE.md sync) is valid; the chezmoi delivery is the drop.
- `abcdZero/.devcontainer/devcontainer.json` — Go 1.24 devcontainer, non-root `vscode` user, `postCreateCommand: go install ./cmd/abcd && abcd init`. Clean pattern; non-root convention is a Keep.
- `abcdZero/.gitpod.yml` — Gitpod prebuild config with `go install ./cmd/abcd && abcd init`. Drop (see D3).
- `abcdZero/.claude/CLAUDE.md` — single-file, references flowctl with `$FLOWCTL` env var pattern. Keep (see K2).
- `abcdZero/home/dot_abcd/README.md` — describes `~/.abcd/` as local-only storage for transparency logs, segments, hash chains, SQLite db. abcd's local storage model should follow this shape (local, gitignored, per-project subdirectory).
- `abcdZero/api/v1/` present (Go HTTP API); abcdSubZero dropped the HTTP API entirely in favour of MCP. abcd aligns with abcdSubZero — MCP over HTTP API.

### abcdSubZero (Python CLI + FastMCP)

- `abcdSubZero/abcd/cli/app.py:1-36` — Typer root app with 10 subcommands (`init`, `build`, `deploy`, `status`, `stats`, `brief`, `config`, `lifeboat`, `update`, `plugins`). Full-featured CLI; abcd ships 6 commands (`ahoy`, `disembark`, `embark`, `launch`, `intent`, `capture`) — scope difference is intentional.
- `abcdSubZero/abcd/orchestrator/cli_adapter_base.py:58-142` — `CLIAdapterBase` ABC; `_build_command()` abstract; `is_available()` probes via `--help`; `WORKER_TIMEOUT=3600`, `SIGTERM_GRACE_SECONDS=10`. Core harness pattern (Keep K1).
- `abcdSubZero/abcd/orchestrator/codex_adapter.py:42-60` — `CodexAdapter(CLIAdapterBase)`: `_STDIN_PROMPT=True`, `--full-auto` mode, `CODEX_DEFAULT_MODEL="o4-mini"`. Confirms subprocess-via-stdin is the right Codex CLI invocation pattern.
- `abcdSubZero/abcd/orchestrator/claude_code_max.py:1-60` — `ClaudeCodeMaxAdapter`: `--output-format stream-json`, SIGTERM/SIGKILL escalation, `AdapterRegistry` with auto-detection order (`claude-code-max` > `claude-code` > `codex` > `gemini`). Registry pattern is worth porting if abcd needs adapter selection.
- `abcdSubZero/abcd/orchestrator/gemini_adapter.py:43-50` — `GeminiAdapter(CLIAdapterBase)`: `_STDIN_PROMPT=True`, `--sandbox` mode, `GEMINI_DEFAULT_MODEL="gemini-2.5-pro"`. Third adapter in the family; confirms the pattern scales.
- `abcdSubZero/abcd/mcp/server.py:1-50` — FastMCP server; optional-import guard; `logging.basicConfig(stream=sys.stderr)` (stdout reserved for JSON-RPC). Entry: `claude mcp add --transport stdio abcd -- python3 -m abcd.mcp.server`. Confirms MCP-over-stdio is the integration pattern (Keep K3).
- `abcdSubZero/abcd/orchestrator/` — 88 files, 6 classifier files. Primary complexity anti-pattern (Drop D2).
- `abcdSubZero/.flow/epics/` — 65 epics (47 referenced in task spec, actual count is 65). Confirms flow-next/flowctl was load-bearing in abcdSubZero, not experimental.
- `abcdSubZero/abcd/session/manager.py:1-60` — `SessionManager` facade with typed `WorkflowPosition` FSM. Clean boundary pattern (Keep K5).
- `abcdSubZero/abcd/persona/` — 6-file module: `models.py`, `elicitor.py`, `communication.py`, `hooks.py`, `defaults.py`. Persona as isolated module, not mixed into CLI handlers (Keep K4).
