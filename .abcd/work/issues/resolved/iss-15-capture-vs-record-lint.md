---
schema_version: 1
id: "iss-15"
slug: "capture-vs-record-lint"
severity: "critical"
category: "tech-debt"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: "internal/core/capture"
resolution: "Fixed by ADR-32 / PR #4 (2be0b81): ledger moved to .abcd/work/issues and created/updated dropped, so capture no longer writes git-inferable metadata into the record-lint root."
---

capture wrote a schema-required created field into the record-lint root, so running capture broke record-lint. This drove moving the ledger to work/issues and dropping created/updated. Fixed by refactor/capture-ledger-schema-tier.