---
schema_version: 1
id: "iss-2608220750029989"
slug: "evidence-placeholders-contradict-adr-5"
severity: "minor"
category: "inconsistency"
source: "user-observation"
found_during: "2026-08-22 filing session (NEXT.md handover)"
found_at: ".abcd/development/brief/03-evidence/"
related_issues: ["iss-2608220750029991"]
resolution: "Banners on pages 01-03 and the chapter README restated to the live adr-5 doctrine (extraction reads, never solely populates); stale related-sources block refreshed; first live open question recorded — where a hold-register-shaped record lives — pointing at the iss-2608220750029991 seed, which stays open to carry that question."
impact: internal
resolved_by:
  commit: "0360888"
---

The 03-evidence pages 01–03 carry leave-empty-until-disembark placeholder banners — during development, leave empty; populated post-build — contradicting adr-5's always-current doctrine. Open questions exist now and have no recordable home while the banner forbids entries. The resolution should also weigh where a hold-register-shaped record would live (see the hold-route triage seed).