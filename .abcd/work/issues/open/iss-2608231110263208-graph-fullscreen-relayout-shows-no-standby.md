---
schema_version: 1
id: "iss-2608231110263208"
slug: "graph-fullscreen-relayout-shows-no-standby"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/record.js"
---

Entering full screen on the relationship chart leaves the reader with no feedback during the relayout: the standby indicator exists (ui.standby, the .bstandby element in graphpage.go, styled at site.css) and is shown on first load, but the full-screen transition re-lays-out 772 nodes without re-showing it, so the chart appears frozen. The indicator must be raised for any relayout the reader waits on — entering and leaving full screen, and any arrange change — not only the first paint (report E of the 2026-08-23 second pass).