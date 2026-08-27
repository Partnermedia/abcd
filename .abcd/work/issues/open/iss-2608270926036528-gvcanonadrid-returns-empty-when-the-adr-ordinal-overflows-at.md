---
schema_version: 1
id: "iss-2608270926036528"
slug: "gvcanonadrid-returns-empty-when-the-adr-ordinal-overflows-at"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-review-2026-08-27"
found_at: "internal/core/lifeboat/graveyard_abandoned.go"
---

gvCanonADRID returns empty when the ADR ordinal overflows Atoi, so an overlong id silently drops the finding the previous regexp path emitted — canonicalise textually by trimming leading zeros so overflow cannot unmake an identity