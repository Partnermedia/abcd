---
schema_version: 1
id: "iss-2608230751270193"
slug: "site-health-panel-wrap-and-unlabelled-summary"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "internal/core/site/explorer.go"
---

Record health panel entries wrap mid-phrase and the summary line '8 / 8 · 148' is three unlabelled numbers. Restructure to two lines per entry (fact, then explanation) and label the summary (report 3a).