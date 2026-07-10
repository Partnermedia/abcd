---
id: itd-3
slug: modular-rules-loader
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
severity: major
---

# abcd Loads Its Own Rules On Demand

## Press Release

> **abcd ships a modular rules loader so the plugin's discipline never bloats your context.** Instead of force-loading a monolithic CLAUDE.md every session, abcd ships a default rule set bundled with the plugin and a per-repo `.abcd/rules.json` override. A `UserPromptSubmit` hook scans each `UserPromptSubmit`-event payload, matches recall keywords against domains (COMMITTING, DOCUMENTATION, ROADMAP, ISSUES, INTENTS, LIFEBOAT), and injects only the rules relevant to what you're actually doing. Sessions about React rendering pay zero tokens for the ADR-format rules; sessions about a PR description load only the commit-and-PII rules. CLAUDE.md becomes a small marker block, not a 1000-line wall.
>
> "I had a 978-line CLAUDE.md silently loading into every conversation," said Bob, staff engineer. "abcd's rules loader took it down to a 30-line marker block plus a JSON file. The full rules still apply — they just turn up when they're relevant. My token bill dropped, my keyboard inputs started landing more accurately, and I never have to maintain CLAUDE.md by hand again."

## Why This Matters

abcd already commits to "single source of truth" and "no scaffolding outside the repo." But the existing methodology document (`~/ABCDevelopment/.claude/CLAUDE.md`) is 978 lines and is force-loaded into every Claude Code session in any abcd-managed repo. Most of those rules don't apply to most prompts — yet every `prompt` pays the token cost.

Prior art ([CARL][carl]) demonstrates the fix: domain-keyed rules with `prompt`-keyword recall, loaded just-in-time. abcd takes the mechanism into its own plugin (rather than depending on CARL the runtime) and ships abcd-shaped defaults out of the box. See [`research/related-work.md`](../../research/related-work.md) for the full comparison.

This is also the keystone for retiring `~/ABCDevelopment/.claude/` entirely. Without modular loading there's no way to relocate the methodology rules without losing them; with modular loading, the rules ship with the plugin, override per repo, and the dev-root tree can be archived.

## What's In Scope

- **`hooks/prompt_router_hook.py`** — `UserPromptSubmit` hook. Reads merged rules (plugin defaults + `<repo>/.abcd/rules.json`), keyword-matches the `prompt` against each domain's `recall` list, injects matching rules as a system message. Signature-based dedup; force full re-inject every N turns (N defaults to 5, configurable).
- **Binary-bundled defaults** embedded in the Go binary (`internal/core/...`). Domains: COMMITTING, DOCUMENTATION, ROADMAP, ISSUES, INTENTS, LIFEBOAT, PII. Each carries `recall` keywords + `rules` array. Content harvested in a one-time manual session from the legacy `~/ABCDevelopment/.claude/CLAUDE.md`.
- **`<repo>/.abcd/rules.json`** schema (JSON Schema embedded in the Go binary, `internal/core/...`). Per-repo file. Can extend defaults, override individual rules, disable a domain entirely, or add custom domains. `/abcd:ahoy` writes a minimal skeleton.
- **CLAUDE.md marker block** owned by `/abcd:ahoy`. ~30 lines. Identifies that abcd is active, points at `.abcd/rules.json` and the plugin defaults, lists the active domains, gives the developer the explicit-activation syntax for star-commands.
- **Star-command bypass** — `*<DOMAIN> …` (e.g. `*ROADMAP draft a milestone`) explicitly activates a domain regardless of keyword match.
- **Dedup + refresh discipline** — same rule signature isn't re-injected within the same session; full refresh every N turns to recover from compaction.
- **Per-domain `state` field** (`active` / `dormant`) — lets users toggle a domain off without deleting the rules.
- **`abcd rules [domain]`** CLI subcommand — bare `abcd rules` prints the full active rule set; `abcd rules <domain>` scopes to one domain. Diagnostic / explainability. Aligns with the bare-command-as-render discipline (see `02-constraints/04-naming.md`); no `show` sub-verb (collapses to bare-with-positional-argument).

## What's Out of Scope

- **Global `~/.abcd/rules.json`** — explicitly rejected. Keeping a personal cross-repo rules file recreates exactly the "scaffolding accumulates outside the repo" failure mode this intent exists to fix. Per-repo only.
- **MCP server for runtime rule editing** — CARL ships one (`carl_v2_add_rule`, `toggle_domain`, etc.). abcd ships JSON-only; the agent edits `rules.json` via the standard file-edit tools. MCP integration is a candidate if friction is real.
- **Star-command bibliography** — only domain activation; no arbitrary `*<command>` macros (CARL's broader capability).
- **Cross-`prompt` rule chaining** — keyword recall is independent per `prompt`. No "session sticky" mode.
- **Migration tooling** — the harvest from legacy `~/ABCDevelopment/.claude/CLAUDE.md` to plugin defaults is a one-time manual session, not a `/abcd:cull` command. See note in the brief.

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** an abcd-installed repo with `.abcd/rules.json` declaring the COMMITTING domain active, **when** the developer sends a `prompt` containing "commit" or "PR" or "git add", **then** the COMMITTING rules appear as injected system context within that turn.
- **Given** the same repo, **when** the developer sends a `prompt` unrelated to any active domain's `recall` keywords, **then** zero abcd rules are injected and the token overhead is < 200 tokens (header only).
- **Given** the same `prompt` repeated three times in a session, **when** dedup is enabled, **then** the rules inject on `prompt` 1 and skip on prompts 2-3 (until forced refresh on `prompt` 5).
- **Given** a developer runs `*ROADMAP draft a milestone` with no roadmap-related keywords, **then** the ROADMAP domain's rules inject regardless of keyword match.
- **Given** a fresh `git clone` of an abcd repo with only the plugin installed, **when** Claude Code starts a new session, **then** CLAUDE.md is < 50 lines and contains the abcd marker block.
- **Given** the legacy `~/ABCDevelopment/.claude/CLAUDE.md` (978 lines), **when** harvested into `defaults.json`, **then** every methodology rule with broad applicability is preserved (manual review, recorded in `research/legacy-harvest.md`).

## Open Questions

- What's the right N for forced refresh? CARL uses 5; abcd's typical session might be longer (lifeboat sessions especially).
- Should the hook respect a `.abcdignore` for prompts that shouldn't trigger any injection (e.g. paste-of-foreign-text)?
- How does this interact with [Claude Code Skills][claude-skills-docs]? Skills are procedural-workflow shaped; rules are declarative. Boundary needs documenting in the marker block.
- Does the hook need to coexist gracefully with CARL if a developer has both installed? (Probably yes — separate hook scripts, separate JSON files, no conflict expected, but worth verifying in the corpus tests.)

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

[carl]: https://github.com/ChristopherKahler/carl "CARL — Context Augmentation & Reinforcement Layer, just-in-time rule injection for Claude Code (Kahler)"
[claude-skills-docs]: https://code.claude.com/docs/en/skills "Claude Code Skills (Anthropic docs)"
