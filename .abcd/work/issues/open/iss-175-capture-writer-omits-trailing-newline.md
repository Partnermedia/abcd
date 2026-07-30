---
schema_version: 1
id: "iss-175"
slug: "capture-writer-omits-trailing-newline"
severity: "nitpick"
category: "bug"
source: "review-followup"
found_during: "iss-156 adversarial review 2026-07-30"
found_at: "internal/core/capture"
---

abcd capture writes ledger files without a trailing newline: 170 of 174 files under .abcd/work/issues/ lack one, including entries the verb writes today. The writer is the source; individual files fixed by hand only mask it. Found while fixing the iss-156 entry's EOF during review follow-up.