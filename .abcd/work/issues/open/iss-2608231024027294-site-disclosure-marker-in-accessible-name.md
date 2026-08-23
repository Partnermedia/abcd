---
schema_version: 1
id: "iss-2608231024027294"
slug: "site-disclosure-marker-in-accessible-name"
severity: "nitpick"
category: "bug"
source: "agent-observation"
found_during: "agent-verification"
found_at: "site-src/site.css"
---

The disclosure marker glyph is part of the accessible name on every folded panel: .panel.fold summary h3::before renders a triangle that screen readers announce as text, so the panel is read out as 'triangle References and sources 28'. aria-hidden cannot apply to a pseudo-element; the remedies are a background-image marker or content:\"triangle\" / \"\" alt-text syntax. Pre-existing pattern elsewhere in the stylesheet, newly multiplied by the folded panels.