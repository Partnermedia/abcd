---
schema_version: 1
id: "iss-149"
slug: "the-positioning-check-bounds-each-surface-read-to-1-mib-but"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "itd-102-security-review"
found_at: "internal/core/positioning/config.go"
---

The positioning check bounds each surface read to 1 MiB but nothing bounds the number of surfaces a committed registry may declare, so one abcd audit run over a hostile repo can be made to hold an unbounded multiple of that cap. Detector: a Config.Validate test that refuses a registry declaring more than a fixed surface count.