---
schema_version: 1
id: "iss-2608300935215868"
slug: "admitted-set-keyed-on-proposal-alone"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-189 adversarial security review, 2026-08-30"
found_at: "internal/core/lint/readingoutstanding.go (admittedProposals), internal/core/lint/schema.go (scanRecordStores), scripts/check-issue-resolution-cases.sh"
---

The outstanding report's admitted set is keyed on proposal alone — run and existence are never checked — so an admission under one run naming a proposal from another silences the other run's item, and a proposal naming nothing passes record_schema because proposal is not resolved the way occasioned_by is (key by run and proposal; resolve proposal on the same terms). Pre-existing, not this branch's: scanRecordStores reads with os.ReadDir and os.ReadFile, no symlink refusal and no size cap, so a symlinked record's frontmatter key names can appear in lint output; the new RS fixtures carry a cited_by property the closed schemas do not know.
