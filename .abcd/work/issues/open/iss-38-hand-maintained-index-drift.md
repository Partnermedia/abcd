---
schema_version: 1
id: "iss-38"
slug: "hand-maintained-index-drift"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/README.md"
---

hand-maintained index drift: intents/README.md corpus listings have drifted from the filesystem; commands/README.md lists three of the seven plugin verb files; internal/README.md still describes core as two capabilities with adapter/scanner as a planned seam; the repo README Layout omits skills/ from the plugin surface. Detector (per adr-5 derive-dont-store): hand enumerations of sibling files are generated or deleted — a lint flags a README that enumerates directory contents by hand, or the enumeration is emitted by tooling with a drift test. Acceptance corpus: the four stale indexes above.