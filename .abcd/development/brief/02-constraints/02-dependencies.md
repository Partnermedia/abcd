# Dependencies

> **Status: PLACEHOLDER.** Dependency rules are currently scattered through the plugin-shape, ahoy, oracle-backend, and adapter sections. Future iterations may consolidate them here. For now, treat this file as a navigational pointer.

## Required external tools

Detected and offered for install by `/abcd:ahoy` (see [`04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)):

- **`gitleaks`** — secret scanner (hard dependency, hard-fail in launch preflight)
- **`presidio`** — PII scanner (hard dependency, hard-fail in launch preflight)
- **`trufflehog`** — deep secret scan (optional, only when `scan.deep=true`)

**Missing-scanner explanation (itd-63 / fn-83).** When the fn-76 validation gate
refuses because a required scanner is absent, the **setup wizard**
(`scripts/abcd/setup_wizard/`) turns the bare fail-closed hint into four
explained elements — tool name (with the version floor), the requiring
capability, what fails without it, and the exact install step — sourced from the
gate's typed `MissingToolPayload` (single source; the curated registry supplies
prose only). It is DISPLAY-ONLY in v1 and NEVER weakens the gate: declining still
fails closed and records a decline in the logbook that surfaces as "previously
declined" on the next trigger. The standalone `explain` CLI runs outside Claude
Code. See [`scripts/abcd/setup_wizard/README.md`](../../../scripts/abcd/setup_wizard/README.md).

## Plugin dependencies

abcd prefers calling other installed plugins over reimplementing. ahoy probes for known plugins on install and records detection in `.abcd/config.json` → `plugins.<name>.detected = true|false` (see [`05-internals/04-universal-patterns.md § 3`](../05-internals/04-universal-patterns.md#3-plugin-preferred-internal-fallback) for the plugin-preferred + internal-fallback pattern).

- **`flow-next`** — preferred provider for `/flow-next:plan`, `/flow-next:work`, `/flow-next:plan-review`, `/flow-next:github-scout`. abcd's intent surface ([`04-surfaces/05-intent.md`](../04-surfaces/05-intent.md)) calls into flow-next; abcd never reimplements that surface.
- **`RepoPrompt`** (RP) — preferred oracle backend (macOS-only; see [`05-internals/01-agents.md` § Oracle backend resolution](../05-internals/01-agents.md#oracle-backend-resolution)). Cascade fallback to Codex CLI → in-session subagent.
- **Codex CLI** — alternative oracle backend for non-Mac users.

## Banned / out of scope

- **Direct API integrations** to Anthropic/OpenAI/etc. — the only transports are RP MCP, Codex CLI subprocess, and in-session subagent (per itd-6 + itd-2). Direct API calls and `claude -p` subprocess spawning are explicitly out of scope.
- **Cross-version schema migration** — abcd stamps `schema_version: 1` everywhere; migrators added if/when a later phase changes the shape (itd-9, a later phase).
- **Scheduled `dev-sync`** (cron / launchd) — itd-13, a later phase.
