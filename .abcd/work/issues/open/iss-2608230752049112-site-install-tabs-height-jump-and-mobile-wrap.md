---
schema_version: 1
id: "iss-2608230752049112"
slug: "site-install-tabs-height-jump-and-mobile-wrap"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

Install tab panel takes the height of the active tab, so switching to a thin tab (Windows) collapses the layout; on iPhone the tab strip wraps into three broken rows (Plugin / CLI+OS / orphaned Build). Stack panels in one grid cell so the tallest wins; give the strip a deliberate narrow-width layout per the existing <=700px nav pattern (reports 9, 14b).