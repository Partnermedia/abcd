---
id: itd-150
slug: agent-worktrees-commit-without-the-private-name-guard-abcd-w
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-370
---

# Agent worktrees commit without the private name-guard: .abcd/.work.local/ is per-worktree, so every isolated-worktree agent commit runs with the banlist layer absent — loudly warned, per design, but the isolated-agent pattern now systematically bypasses a protection the main checkout has. Candidate remedies: the worktree-creation path seeds a pointer to the primary checkout's store, or the hook falls back to reading the primary worktree's local tier

## Press Release

> _Seeded by promotion from iss-370. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-370`: Agent worktrees commit without the private name-guard: .abcd/.work.local/ is per-worktree, so every isolated-worktree agent commit runs with the banlist layer absent — loudly warned, per design, but the isolated-agent pattern now systematically bypasses a protection the main checkout has. Candidate remedies: the worktree-creation path seeds a pointer to the primary checkout's store, or the hook falls back to reading the primary worktree's local tier. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
