---
schema_version: 1
id: "iss-2608270945464715"
slug: "banlist-s-declaresformat-strips-a-bom-from-any-line-with-no"
severity: "nitpick"
category: "inconsistency"
source: "agent-finding"
found_during: "issue-sweep-review-2026-08-27"
found_at: "internal/core/banlist/private.go"
---

banlist's declaresFormat strips a BOM from any line with no index guard while every other reader trims it at line 0 only — its misplaced-declaration caller fails closed, so an inconsistency in the BOM-position family rather than a defect