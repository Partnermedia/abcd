---
schema_version: 1
id: "iss-2608270930248619"
slug: "pack-go-s-within-is-dead-production-code-after-the-fsutil-co"
severity: "nitpick"
category: "tech-debt"
source: "agent-finding"
found_during: "issue-sweep-review-2026-08-27"
found_at: "internal/core/lifeboat/pack.go"
---

pack.go's within() is dead production code after the fsutil consolidation — its only caller is a test while pathOverlaps calls fsutil.PathsOverlap directly; remove the wrapper and point the test at the production path