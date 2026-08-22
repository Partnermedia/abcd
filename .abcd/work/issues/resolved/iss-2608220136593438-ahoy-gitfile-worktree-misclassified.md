---
schema_version: 1
id: "iss-2608220136593438"
slug: "ahoy-gitfile-worktree-misclassified"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/ahoy/detect.go"
resolution: "ahoy classify tests .git existence via gitPresent, so a gitfile worktree or submodule is a repo"
impact: fix
resolved_by:
  commit: "4d06ab1"
---

ahoy classify tests .git with isDir, so a linked worktree or submodule (where .git is a gitfile) is misread as unmanaged-folder, every gap detector is skipped, and ahoy install exits 0 aborted with a wrong not-a-git-repository reason; the last isDir(.git) holdout after iss-72