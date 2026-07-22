---
schema_version: 1
id: "iss-5"
slug: "disembark-embark-phase-banners"
severity: "major"
impact: internal
category: "inconsistency"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief/04-surfaces/02-disembark.md"
blocked_by: [iss-1]
resolution: "Disembark and embark banners re-anchored to Phase 6 (phase-6-lifeboat.md) per ADR-33; the agent catalog's lifeboat-pipeline rows (Pass A/B/C, embark-scaffolder, documentation-auditor) re-anchored from old Phase 4/5 numbering to Phase 6 + itd-65."
---

disembark and embark status banners point to the wrong phase: 02-disembark.md says Phase 4, 03-embark.md says Phase 5, but phase-6-lifeboat.md owns both.