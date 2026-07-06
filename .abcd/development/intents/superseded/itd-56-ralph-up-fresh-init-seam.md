---
id: itd-56
slug: ralph-up-fresh-init-seam
spec_id: fn-51-ralph-up-fresh-one-guarded-re-vendor
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: []
prd_path: null
---

> **⚠️ Superseded by [ADR-22](../../decisions/adrs/0022-bundled-deps-as-pluggable-adapters.md)** (flow-next, ralph-init, and the abcd overlay all cease to exist; there is no re-vendor cycle to guard — see also [ADR-26](../../decisions/adrs/0026-native-spec-layer-ccpm-backend.md), [ADR-27](../../decisions/adrs/0027-autonomous-run-pluggable-seam.md)). Preserved as historical record per the supersession lifecycle.

# One Command Re-Vendors Upstream And Restores The abcd Overlay In A Single, Guarded Step

## Press Release

> **abcd gains a single operator action that runs `/flow-next:ralph-init` (the upstream re-vendor) and `/abcd:ralph-up` (the overlay restore) as one guarded cycle — so a contributor updating flow-next never leaves their harness half-reverted between the two steps.** Today the two-step sequence is correct but manual: ralph-init silently reverts every abcd customization, and the harness is only whole again after a separate ralph-up. Anyone who runs the first and forgets the second is left with a pristine-upstream Ralph that has lost abcd's guard, patches, dispatcher, and normalization — a quiet, easy-to-miss regression. This intent closes that window with a script-layer `--fresh` mode (the slash layer cannot call another slash command, by platform rule), gated by an explicit destructive-action confirmation because re-vendoring overwrites upstream-owned files.

> "I bumped flow-next, ran ralph-init, got distracted, and started a Ralph loop — and only later noticed my guard wasn't even registered," said a contributor. "I want the re-vendor and the re-apply to be one thing I can't half-do, but I still want to be told, loudly, that it's about to overwrite upstream before it runs."

## Why This Matters

The standing rule — "after every `/flow-next:ralph-init`, run `/abcd:ralph-up`" — is correct but human-enforced, and the failure mode is silent: a forgotten ralph-up leaves Ralph running against pristine upstream with no abcd guard, no patches, no dispatcher. The cost lands exactly when the operator is least watching (mid-upgrade, context-switching). A single guarded cycle removes the gap without weakening the deliberate, visible nature of the destructive re-vendor.

## What's In Scope

- A **script-layer** entrypoint (e.g. `scripts/abcd/ralph_up.py --fresh`) that invokes flow-next's ralph-init machinery directly (NOT via the slash layer — `--with-ralph-init` already documents why a skill cannot call another slash command), then runs the standard overlay apply.
- A **destructive-action guard**: `--fresh` re-vendors (overwrites upstream-owned files), so it must require an explicit confirmation (interactive prompt, or a `--yes`/`--force` token in non-interactive contexts) before doing any re-vendor work. Default is refuse-and-explain.
- A **post-cycle `--check` summary** so the operator sees the resulting overlay state (and any patch `drift`/`needs_migration` from upstream rewrites) in the same run.
- Decide the **surface name**: `ralph_up.py --fresh` vs `--init` vs a distinct `/abcd:ralph-init` command/skill that orchestrates the documented sequence. Resolve which is canonical.

## What's Out of Scope

- Making a slash command call another slash command (platform rule forbids it — the seam is script-layer only).
- Auto-re-anchoring patches whose upstream anchor text changed in the new version (that stays a manual, visible step; `--fresh` surfaces the drift, it does not silently rewrite patches).
- Changing the default `/abcd:ralph-up` behavior (it stays apply-only; `--fresh` is strictly additive).

## Acceptance Criteria

- **Given** a confirmed `--fresh` invocation, **when** the operator runs it, **then** upstream is re-vendored AND the overlay is applied in one invocation, and the resulting overlay state is fully re-asserted (a test proves both legs ran).
- **Given** a `--fresh` invocation with no explicit confirmation (`--yes` absent and non-interactive), **when** it runs, **then** it performs NO re-vendor, exits non-zero, and prints a clear destructive-action diagnostic (a test proves the unconfirmed path is a no-op).
- **Given** that the new upstream rewrote a region a patch anchors on, **when** `--fresh` completes, **then** the affected patch is surfaced in the post-cycle summary as drifted/needing-migration rather than silently failing (a test proves the drift is reported).
- **Given** the chosen surface name (`--fresh` vs `--init` vs a separate command), **when** the work lands, **then** it is documented consistently across the command markdown, the skill prose, and the CLI binding (a doc check confirms parity across layers).

## Open Questions

- Does the installed flow-next expose a **callable ralph-init entrypoint** (a Python/script target) the wrapper can invoke directly, or only the slash command? If only the slash, `--fresh` may be limited to printing+staging rather than truly executing init — which would reduce it to a guidance improvement.
- Surface choice: `--fresh` flag on the existing command (minimal new surface) vs a named `/abcd:ralph-init` (clearer identity, more to maintain)?
- Should `--fresh` snapshot/backup the pre-re-vendor state (so an operator can diff what upstream changed), reusing the dispatcher-wrapper `.bak.<ts>` pattern from fn-33 `.9`?

## Audit Notes

Captured 2026-06-06 from an operator question ("would ralph-up auto-run ralph-init? should there be --init/--fresh?") immediately after a successful manual `ralph_up.py --apply` over flow-next 1.8.0. Hand-authored draft — validate via `/abcd:intent` or `intent_lint` before promotion (the `/abcd:intent` capture path is canonical; hand-authored drafts risk lint findings, per fn-33 `.5`).

## References

- `skills/abcd-ralph-up/SKILL.md` — the `--with-ralph-init` "guidance-only-and-exit" block (states the no-slash-from-slash constraint).
- `scripts/abcd/ralph_up.py` — the CLI binding (`--check`/`--apply` group, `--with-ralph-init`, `--strict-watcher`).
- `scripts/abcd/overlay/README.md` — the re-vendor → re-apply contract + dependency watcher.
- `abcdDev/CLAUDE.md` — the standing "after every ralph-init, run ralph-up" rule this intent automates.
