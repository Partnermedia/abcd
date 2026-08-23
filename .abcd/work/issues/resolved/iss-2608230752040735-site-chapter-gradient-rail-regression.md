---
schema_version: 1
id: "iss-2608230752040735"
slug: "site-chapter-gradient-rail-regression"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
resolution: "Fixed in the 2026-08-23 manual-test triage pass; verified by rebuild, the seven site-check gates, the overflow audit, and a screenshot of the affected route compared against the report that raised it."
impact: fix
---

Landing-page a/b/c/d chapters lost the colour-gradient connecting rail from the signed-off mock: .chapter::before/::after rules were never lifted from the prototype into site.css (report 7). Port the rail CSS.