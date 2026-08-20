---
schema_version: 1
id: "iss-370"
slug: "agent-worktrees-commit-without-the-private-name-guard-abcd-w"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "manual-capture"
found_at: "internal/core/ahoy/defaults/pre-commit"
---

Agent worktrees commit without the private name-guard: .abcd/.work.local/ is per-worktree, so every isolated-worktree agent commit runs with the banlist layer absent — loudly warned, per design, but the isolated-agent pattern now systematically bypasses a protection the main checkout has. Candidate remedies: the worktree-creation path seeds a pointer to the primary checkout's store, or the hook falls back to reading the primary worktree's local tier