---
schema_version: 1
id: "iss-3"
slug: "phase-membership-vs-lifecycle"
severity: "critical"
impact: internal
category: "inconsistency"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/intents"
resolution: "ADR-34: lifecycle (commitment, directory) and scheduling (phase Scope) are orthogonal; invariant runs one way — scheduled implies planned/. itd-43 and itd-7 moved drafts->planned (itd-43 kind bound standalone per its suggested_kind); intents README planned/ definition and listings corrected."
---

phase membership and lifecycle directories disagree: planned intents (itd-20, itd-24, itd-29, itd-46) appear in no phase Scope, while draft itd-43 is scoped into Phase 0/1 yet listed later-phase. ADR-bound.