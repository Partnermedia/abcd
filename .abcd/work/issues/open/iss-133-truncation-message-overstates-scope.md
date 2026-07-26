---
schema_version: 1
id: "iss-133"
slug: "truncation-message-overstates-scope"
severity: "nitpick"
category: "tech-debt"
source: "impl-review"
found_during: "iss-112/114/116 review (2026-07-24 run queue, burst 9)"
found_at: "internal/core/lifeboat/sources_conventions.go"
---

the probe adapters' truncation message says the rest of the tree was not walked, but the per-directory read bound added for iss-112 can set truncated while the walk completed everything else; the over-warning is in the conservative direction, not a loud-staging violation, but the message is imprecise about WHAT was truncated