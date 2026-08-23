---
schema_version: 1
id: "iss-2608231003136455"
slug: "site-print-graph-canvas-aspect-squashed"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

The relationship chart prints squashed: .bstage is sized height:min(72vh,720px) with width:100%, and the canvas bitmap record.js drew for a wide viewport is then stretched into the narrower print page box, distorting the aspect. The expanded Browse-as-a-list content is also cut mid-row at the panel boundary rather than paginating. Found printing /record/graph/ from a browser with the list open (report A of the 2026-08-23 second pass).