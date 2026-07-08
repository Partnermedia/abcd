---
schema_version: 1
id: "iss-1"
slug: "launch-phase-ownership"
severity: "critical"
category: "inconsistency"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief/04-surfaces/04-launch.md"
resolution: "ADR-33: phase index is the sole ownership source; Phase 1 owns the curated-release cut, deepenings are intent-attributed (itd-65/66/70/72/73). 04-launch.md banner and Phase-5 references rewritten; launch-gatekeeper row re-anchored."
---

launch phase ownership contradicts across the record: phase-1-ahoy.md and build-sequence.md make install plus launch the first milestone, but 04-surfaces/04-launch.md says full launch builds in Phase 5. ADR-bound.