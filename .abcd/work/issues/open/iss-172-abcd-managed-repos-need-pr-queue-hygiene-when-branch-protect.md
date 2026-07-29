---
schema_version: 1
id: "iss-172"
slug: "abcd-managed-repos-need-pr-queue-hygiene-when-branch-protect"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "explainability-plan merge stall"
found_at: ".github/workflows"
---

abcd-managed repos need PR-queue hygiene when branch protection requires up-to-date branches: a queued auto-merge PR whose base moved sits BEHIND forever because GitHub auto-merge never updates branches. Two delivery rungs per basics-built-in/SOTA-delegated: (1) protocol-level now — after arming auto-merge, poll mergeStateStatus and run gh pr update-branch on BEHIND (works with the agent's own token, CI triggers normally); (2) scaffolded CI later — abcd scaffolds an update-branch workflow into managed repos, which requires a PAT/GitHub-App secret because the default GITHUB_TOKEN's pushes do not trigger workflows (the recursion suppression trap), or points at a merge queue where the plan allows. Same scaffolding family as itd-93's release gate; same merged-work-hygiene family as itd-95.