---
schema_version: 1
id: "iss-146"
slug: "guard-bash-gh-write-local-path-fp"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "2026-07-27 session"
found_at: "guard hook, agent shell usage"
---

guard-bash false positive class: a compound command chaining a gh write verb with a LOCAL command taking an absolute path argument (gh pr merge ... ; git worktree remove /abs/path) is blocked by the line-level privacy regex, which cannot see that the path never reaches GitHub. Second guard incident of 2026-07-27 (the first, blocking --no-verify, was a true positive). Both are fixture material for itd-103's calibration corpus: this one is a known-good case the TNR floor must protect; the workaround (split the compound) is recorded here so the friction is not silent.