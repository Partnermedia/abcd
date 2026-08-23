---
schema_version: 1
id: "iss-2608230752151670"
slug: "site-roles-table-th-labels-mobile"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
resolution: "Fixed in the 2026-08-23 manual-test triage pass; verified by rebuild, the seven site-check gates, the overflow audit, and a screenshot of the affected route compared against the report that raised it."
impact: fix
---

Roles comparison table column labels (th spans beside the portraits) render badly on iPhone; visually hide them on the composed landing surface, keeping them screen-reader-accessible (report 13).