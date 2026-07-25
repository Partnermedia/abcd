---
schema_version: 1
id: "iss-127"
slug: "repo-local-lmw-sites-unguarded"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "iss-101/102 class sweep (2026-07-24 run queue, burst 3)"
found_at: "internal/core/intent/lifecycle.go"
---

repo-local load-modify-write sites remain unguarded after the iss-101/102 class fix: intent transitions (review.go, lifecycle.go — strongest candidate), ahoy config.json stepConfigValues, gitignore/marker block rewrites, and the CHANGELOG release path all do load-mutate-write without a lock; they lack the home-global cross-worktree exposure the fixed sites had, so risk is same-worktree concurrency only