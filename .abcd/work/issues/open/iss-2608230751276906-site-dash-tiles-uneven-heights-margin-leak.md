---
schema_version: 1
id: "iss-2608230751276906"
slug: "site-dash-tiles-uneven-heights-margin-leak"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "site-src/site.css"
---

Site dashboard/contributors stat tiles render unevenly: .panel+.panel{margin-top} leaks into the .dash grid (every panel but the first is pushed down 14px) and .dash align-items:start lets each panel take content height, so row-mates never equalise. Seen on /record/, /contributors/, and at phone widths (reports 1, 4, 10, 14a of the 2026-08-23 manual test pass).