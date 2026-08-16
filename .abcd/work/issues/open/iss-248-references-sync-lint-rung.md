---
schema_version: 1
id: "iss-248"
slug: "references-sync-lint-rung"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "academic-references-baseline"
found_at: ".abcd/development/research/references.csl.json"
---

references_sync lint rung: cross-check references.csl.json keys/DOIs/URLs against the ACKNOWLEDGEMENTS.md References & sources section, so a store entry without a prose entry (or vice versa) is a finding; today the sync is the documented protocol in research/_references.md