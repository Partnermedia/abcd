---
schema_version: 1
id: "iss-2608230752040334"
slug: "site-svg-inliner-strips-image-dimensions"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "user-observation"
found_at: "internal/core/site"
resolution: "Fixed in the 2026-08-23 manual-test triage pass; verified by rebuild, the seven site-check gates, the overflow audit, and a screenshot of the affected route compared against the report that raised it."
impact: fix
---

The site build's SVG inliner strips width/height attributes from svg image elements: docs/assets/img/process-loop.svg embeds cast portraits at width=42 height=42, but the inlined landing-page copy has no size attributes, so the WebP figures render at intrinsic size and the r=21 circle clips catch only slivers (report 8). Root-cause the optimiser and sweep every inlined SVG carrying image elements.