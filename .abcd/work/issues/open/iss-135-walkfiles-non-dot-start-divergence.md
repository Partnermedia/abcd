---
schema_version: 1
id: "iss-135"
slug: "walkfiles-non-dot-start-divergence"
severity: "nitpick"
category: "tech-debt"
source: "impl-review"
found_during: "iss-112/114/116 review (2026-07-24 run queue, burst 9)"
found_at: "internal/core/lifeboat/probe.go"
---

WalkFiles start-boundary divergence for a non-dot start: a start dir named in the skip set would now be walked and a regular-file start returns nil; unreachable today (both callers pass dot) — flag for any future caller passing a non-dot start