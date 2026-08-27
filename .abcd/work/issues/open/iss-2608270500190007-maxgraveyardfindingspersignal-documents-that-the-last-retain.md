---
schema_version: 1
id: "iss-2608270500190007"
slug: "maxgraveyardfindingspersignal-documents-that-the-last-retain"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/lifeboat (graveyard signal cap)"
---

maxGraveyardFindingsPerSignal documents that the last retained finding notes the cap, but 6 of 10 signals truncate silently with no such note. GitHub mirror: #292