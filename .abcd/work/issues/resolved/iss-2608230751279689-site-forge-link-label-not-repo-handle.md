---
schema_version: 1
id: "iss-2608230751279689"
slug: "site-forge-link-label-not-repo-handle"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "internal/core/site/compose.go"
resolution: "Fixed in the 2026-08-23 manual-test triage pass; verified by rebuild, the seven site-check gates, the overflow audit, and a screenshot of the affected route compared against the report that raised it."
impact: fix
---

Header and footer external links print the repo handle (owner/repo) where users expect the forge name (GitHub). Label must come from a ui.json interface string with the handle as generic-side fallback, never hardcoded in Go (reports 2, 17).