---
schema_version: 1
id: "iss-72"
slug: "core-isgitrepo-misses-worktrees"
severity: "nitpick"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/core.go"
---

core.Status.IsGitRepo=false in a linked git worktree or submodule: isDir(.git) requires a directory but .git is a regular gitfile (gitlink) in worktrees/submodules, so a genuine checkout reports not-a-git-repo (core.go:47, C1). Detector: an exists-not-isDir check for .git with a worktree fixture. Corpus: C1.