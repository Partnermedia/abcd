---
schema_version: 1
id: "iss-2608231350127745"
slug: "by-links-layout-publishes-overlapping-positions"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "agent-verification"
found_at: "internal/core/site/layout.go"
promoted_to: itd-157
---

The by-links arrangement's non-settling has a LAYOUT half that is still open, and it is a redesign rather than a fix. The renderer half is fixed (the collision resolver no longer demands 1.5 units of padding the published positions never promised, iss-2608231243286557), but the layout itself still publishes overlapping positions, so the chart still never comes to rest on that arrangement — measured 2026-08-23: by date comes to rest, by links does not. Two causes found and NOT fixed, because fixing them changes the picture: the islands are settled by a spring layout with no collision pass at all, and the two rim rows map each record's arc-width onto a circle whose circumference is smaller than the sum of those widths, so the rows are overpacked by construction. Attempted fixes measured: relaxing the linked records alone left 600 overlaps (the rest are on the rim); adding a rim-row growth rule left 367 and pushed the outermost radius from 0.97 to 1.83, visibly changing the picture without settling it. Both reverted. A correct fix places every record — islands, pairs and rim — under one packing rule, which is a design of the arrangement, not a patch to it.