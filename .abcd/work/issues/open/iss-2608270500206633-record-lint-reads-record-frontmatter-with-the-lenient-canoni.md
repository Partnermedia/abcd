---
schema_version: 1
id: "iss-2608270500206633"
slug: "record-lint-reads-record-frontmatter-with-the-lenient-canoni"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/record-lint"
---

record-lint reads record frontmatter with the lenient canonical scanner while capture's strict parser refuses the same bytes, so a duplicated key or 'id : ' (space before colon) ships lint-green but is skipped by every capture verb, silencing the armed issue_impact_valid blocker and bypassing record_schema. GitHub mirror: #357