---
schema_version: 1
id: "iss-2608270500193735"
slug: "planembark-s-claimed-map-dedups-targets-case-sensitively-so"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/lifeboat/embark.go"
---

planEmbark's claimed map dedups targets case-sensitively, so two lifeboat files differing only in case both plan Create and the second silently overwrites on a case-folding filesystem. GitHub mirror: #326