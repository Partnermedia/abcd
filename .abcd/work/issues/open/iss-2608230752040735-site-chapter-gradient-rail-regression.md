---
schema_version: 1
id: "iss-2608230752040735"
slug: "site-chapter-gradient-rail-regression"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

Landing-page a/b/c/d chapters lost the colour-gradient connecting rail from the signed-off mock: .chapter::before/::after rules were never lifted from the prototype into site.css (report 7). Port the rail CSS.