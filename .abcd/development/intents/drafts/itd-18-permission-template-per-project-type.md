---
id: itd-18
slug: permission-template-per-project-type
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
---

# Sensible Permissions From Day One

## Press Release

> **abcd detects your project type and pre-populates `.claude/settings.local.json` with sensible permissions.** When `/abcd:ahoy install` runs in a Python project, you get pip/pytest/uv permissions out of the box. Swift project? `swift build`/`swift test`/`xcodebuild`. Go? `go test`/`go build`/`golangci-lint`. The permission-prompt friction during the first hour of using a new abcd-managed project drops to near zero — you spend time on the work, not on approving safe operations.
>
> "Every fresh project I'd hit the same dozen permission prompts before getting to actual work," said Henry, junior developer. "abcd handed me a settings.local.json that just worked. I edited a few entries; I didn't have to start from scratch."

## Why This Matters

abcd's `/abcd:ahoy` doesn't touch `.claude/settings.local.json`. Users start with whatever permissions they had (often nothing project-specific) and accumulate them by repeatedly approving Bash commands. This is friction that adds up, especially for autonomous-mode work where every prompt interrupts a run-seam loop or Claude Code session.

`~/.claude/templates/settings.local.json.template` from abcd v0 already shows the pattern: a base permission set + per-language additions. Lift the pattern; detect the project type during ahoy; install a sensible starting set with transparent confirm.

## What's In Scope

- Project-type detection (file-based heuristics: `pyproject.toml` → Python, `Package.swift` → Swift, `go.mod` → Go, `package.json` → TypeScript/JavaScript, `Cargo.toml` → Rust, `Dockerfile` → Docker)
- Per-type permission templates shipping in the plugin
- `/abcd:ahoy` step: detect type, show user the proposed `.claude/settings.local.json` content, transparent confirm
- Conflict handling: if `.claude/settings.local.json` exists, offer merge / replace / skip
- Templates committed and versioned in `scripts/abcd/templates/permissions/`

## What's Out of Scope

- Auto-detecting non-language project shapes (CI/CD repos, doc repos, config repos) — too many heuristics
- Updating existing settings.local.json on plugin upgrade (one-shot at install only)
- Permission-prompt analytics (covered by a different concern)

## Acceptance Criteria

> _BDD format, per `itd-1-acceptance-gates`. These gates are checked by `intent-fidelity-reviewer` when this intent moves to `shipped/`._

- **Given** a fresh repo containing `pyproject.toml` and no existing `.claude/settings.local.json`, **when** the user runs `/abcd:ahoy install` and accepts the proposed permissions, **then** `.claude/settings.local.json` is created with the Python permission template (pip, pytest, uv, ruff) AND the install summary lists exactly which entries were added.
- **Given** the same repo already has a `.claude/settings.local.json` with custom entries, **when** `/abcd:ahoy install` runs, **then** the user is offered three explicit options (merge / replace / skip) AND the chosen action is recorded in the install report.
- **Given** a multi-language repo (e.g., `pyproject.toml` AND `package.json`), **when** `/abcd:ahoy install` detects project type, **then** both Python and TypeScript permission templates are offered as a merged proposal AND the user can deselect either before applying.
- **Given** a Swift repo (`Package.swift` present), **when** `/abcd:ahoy install` runs, **then** the resulting `settings.local.json` includes `swift build`, `swift test`, `xcodebuild`, plus the base permissions common across all project types.
- **Given** the user wants to inspect what would be installed without committing, **when** they run `/abcd:ahoy dry-run`, **then** the proposed `settings.local.json` content is printed to stdout and no file is written.
- **Given** an abcd-managed project on a later plugin upgrade, **when** the user runs `/abcd:ahoy install` (idempotent re-install), **then** the existing `settings.local.json` is NOT modified — permission templates are install-time only, not auto-upgraded — AND a hint is shown about how to re-apply the latest template manually if desired.

## Open Questions

- What's the right default base permission set across all project types?
- How granular is the project-type detection (Python with FastAPI vs Python with PyTorch — different needs)?
- Where does the user see what was added — in the ahoy summary, in `.claude/settings.local.json` itself with comments, both?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
