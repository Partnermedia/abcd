---
schema_version: 1
id: "iss-2608290810033503"
slug: "the-forge-reported-a-pull-request-as-conflicting-when-it-mer"
severity: "nitpick"
category: "process"
source: "agent-observation"
found_during: "intent-implementation-run"
found_at: ".github"
---

The forge reported a pull request as conflicting when it merged cleanly against both the current main and the tip of the pull request queued ahead of it, and dropped its armed auto-merge as a result. Verified by merging locally against both refs with zero conflicting paths, twice, several minutes apart. The state cleared only after pushing a merge of main into the branch, which forced the forge to recompute, after which it reported mergeable and auto-merge could be re-armed. Cost roughly twenty minutes of diagnosis on the assumption that a real conflict existed. Recording it so a future run recognises the shape: a conflicting verdict that no local merge reproduces is a stale computation, and updating the branch is the cheap remedy rather than hunting for a conflict that is not there.