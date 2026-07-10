---
schema_version: 1
id: "iss-60"
slug: "agent-git-revert-hazard"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "autonomous-run"
found_at: "internal/core/lint"
---

Subagent incident during the context_status_free build: the implementation agent reverted temporary test mutations with 'git checkout -- lint.go' and wiped its own uncommitted implementation (recovered verbatim, gates re-verified). Agent-hygiene rule for the run seam and prompts: agents revert edits with the editor, never with git; git-based revert over an uncommitted tree is destructive. Same hazard family as iss-28's ambient-git exposure.