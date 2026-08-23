---
schema_version: 1
id: "iss-2608230751433117"
slug: "site-references-grid-shatters-citations"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

References page citations shatter: .refs li{display:grid} turns every inline child (em, a, text runs) into alternating grid items, so journal names and DOI links land in the 40px marker column and wrap one character per line. Core content illegible on desktop; the overflow audit cannot see it because nothing overflows (report 6). Sweep for sibling li{display:grid} patterns.