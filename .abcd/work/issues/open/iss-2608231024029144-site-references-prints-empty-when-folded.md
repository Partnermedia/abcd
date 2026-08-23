---
schema_version: 1
id: "iss-2608231024029144"
slug: "site-references-prints-empty-when-folded"
severity: "minor"
category: "bug"
source: "agent-observation"
found_during: "agent-verification"
found_at: "site-src/site.css"
---

The references page prints EMPTY as a consequence of the stacked-disclosure change (iss-2608231003138797): both bibliography panels are closed on load, CSS cannot force a details element open, and the ::details-content escape already used for .more is Chromium-only and not applied to .panel.fold. A reader printing /references/ from a fresh load gets one sheet carrying two headings and zero citations; before the change the panels were always-open blocks and the bibliography printed. Verified 2026-08-23: 1 page as loaded, 5 pages with both opened by hand.