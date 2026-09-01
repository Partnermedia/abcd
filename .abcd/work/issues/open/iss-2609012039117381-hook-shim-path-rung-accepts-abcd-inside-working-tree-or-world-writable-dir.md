---
schema_version: 1
id: "iss-2609012039117381"
slug: "hook-shim-path-rung-accepts-abcd-inside-working-tree-or-world-writable-dir"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/hooks.json"
---

Sub-finding of GHSA-gx3m-3224-qqcv that does not need the design decision: the hook shims' PATH rung accepts an `abcd` that `command -v` resolves from inside the working tree or from a world-writable directory. With a dot or empty PATH entry — or a PATH entry naming a directory under the checkout — `command -v abcd` in `hooks/hooks.json` (UserPromptSubmit, PreToolUse, PreCompact, SessionEnd) returns a file the checkout controls, so a hostile clone carrying an executable `abcd` becomes the guard and rules loader for a session whose plugin-root binary is missing; a world-writable PATH directory hands the same role to any local user. The documented contract (docs/how-to/install.md: plugin root first, then an abcd on PATH, the one-liner installs into `~/.local/bin`) never places the binary in either location, so refusing both keeps the documented rescue intact. The fix must establish, for all four events, that a PATH abcd whose directory is under the shim's working directory or is world-writable is not executed, that the shim degrades to its existing loud line (PreToolUse: UNGUARDED, exit 1) plus one line saying the PATH binary was ignored and why, and that an abcd in an ordinary PATH directory still serves as the fallback. SessionStart is unchanged (no PATH rung). The full owned-only rung stays with the parent record.
