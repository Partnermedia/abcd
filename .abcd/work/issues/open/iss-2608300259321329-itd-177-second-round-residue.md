---
schema_version: 1
id: "iss-2608300259321329"
slug: "itd-177-second-round-residue"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "itd-177 second-round reviews, 2026-08-30"
found_at: "internal/core/intent/ready.go (scopeConditionsCheck), internal/core/intent/claims.go (malformedMarkerRe), internal/core/intent/lifecycle.go (stampPlanned), commands/intent.md"
---

itd-177 second-round residue: the structural faults (duplicated heading, fenced section) are checked only after the claim-state switch, so a nullity or empty first section hides a second Scope Conditions heading carrying real bullets and the gate passes while the stamp refuses; the malformed-marker guard is case-sensitive so a capitalised near-miss gets a real marker glued beside it; no size check after stamping, so a record near the read cap can be written larger than its own reader accepts and the next Load refuses the whole corpus; a marker inside an inline code span is accepted and excised; the plugin page says every fault is refused by the stamp where two-marker and duplicated-id bullets are skipped rather than refused.
