---
schema_version: 1
id: "iss-2608211849463840"
slug: "lint-walk-uncontained-cloned-repo-read"
severity: "major"
category: "security"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/core/lint/lint.go:206"
resolution: "lint walk routed through containedRepoPath/resolvedInsideRoot/containedRealPath/ReadGuarded"
impact: fix
resolved_by:
  commit: "51ed7e1"
---

docs-lint/record-lint main walk lacks the root containment and guarded read its sibling citation collector applies to the same cfg.Roots, so a cloned repo reads and lints files outside the repository