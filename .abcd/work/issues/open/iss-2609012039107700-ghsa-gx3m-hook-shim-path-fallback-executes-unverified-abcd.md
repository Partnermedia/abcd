---
schema_version: 1
id: "iss-2609012039107700"
slug: "ghsa-gx3m-hook-shim-path-fallback-executes-unverified-abcd"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/hooks.json"
---

GHSA-gx3m-3224-qqcv (CWE-426, advisory severity medium): the hook shims' PATH fallback executes an unverified `abcd` as the guard and rules loader. In `hooks/hooks.json`, UserPromptSubmit, PreToolUse and PreCompact carry `if [ -z "$g" ] && command -v abcd; then g=abcd; fi` and SessionEnd `exec abcd hook session-end` when the plugin-root binary is missing; PreToolUse's fence passes exit 0/1/2 through, so a foreign binary exiting 0 approves every shell command. SessionStart has no PATH rung and fails closed. Reproduced at v0.7.0 with the shipped PreToolUse command, a failing bootstrap and a hijack directory first on PATH: the planted binary received `guard hook` plus the tool input and its exit 0 was accepted as armed. The rung was introduced by iss-254 (20968bdf, v0.5.1), is pinned as desired behaviour by `internal/surface/cli/hooks_selfprovision_test.go:TestBinaryHooksFallBackToAPathBinary`, and is a documented contract (docs/how-to/install.md: "the plugin root first, then an abcd on PATH … the install one-liner restores it"; brief 04-surfaces/01-ahoy.md section 9; brief 05-internals/03-configuration.md). The record contradicts itself: .abcd/work/DECISIONS.md 2026-08-01 (itd-105 / spc-21) states "Rejected: a PATH fallback in the hook commands (spc-21 forbids it)" and "A PATH fallback is forbidden by the intent's Decisions and is NOT added"; iss-254 later added it with no DECISIONS entry reversing that.

DESIGN DECISION NEEDED. Options:

A. Owned-only PATH rung (RECOMMENDED; the advisory's suggestion and the iss-377 seam): each shim resolves `p=$(command -v abcd)` and accepts it only when `~/.abcd/path-entry` exists and its `path=` equals `$p` (no hashing on the fast path, per adr-46); a foreign PATH abcd degrades to the existing loud one-line refusal (PreToolUse: UNGUARDED, exit 1, never a silent 0). Contract change: the one-line install writes NO path-entry today, so the documented rescue stops working unless both one-liners also write `path=` and `binary_sha256=` (they know both; plugin_root is optional per readPathEntry). Touches hooks/hooks.json, hooks_selfprovision_test.go (invert the fallback test into owned-runs / foreign-refused / PreToolUse-says-UNGUARDED / SessionStart-unchanged), docs/how-to/install.md (both one-liners and the hooks paragraph), brief 01-ahoy.md section 9, 03-configuration.md, and a DECISIONS entry closing the spc-21-versus-iss-254 contradiction.

B. Drop the PATH rung from PreToolUse only — the guard is the one consumer where a wrong binary fails open on every command; the prompt-router, precompact and session-end shims keep it. Smallest change; leaves the prompt channel and transcript_path exposed; still needs the DECISIONS entry.

C. Accept the residual: PATH is the operator's environment (the same shims resolve sh, find and printf from PATH; DECISIONS.md 2026-08-29 and 2026-08-01 already say so for gitleaks and curl); document it in the brief and adr-46, close won't-fix.

No mechanical hardening keeps the documented one-liner rescue AND refuses a foreign binary, because the one-liner leaves no ownership record — that is the decision. Sibling sites: only the four hooks.json events; commands/*.md invoke abcd through the harness's shell by design; hooks/bootstrap.sh never executes a PATH abcd. The strict hardening that keeps the contract intact (a PATH abcd that resolves inside the working tree or from a world-writable directory is never used) is captured and fixed separately and leaves this record open for the decision. Cross-lane: iss-323 (salvage timeouts) edits the same command strings.
