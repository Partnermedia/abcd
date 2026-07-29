---
schema_version: 1
id: "iss-173"
slug: "three-tier-layout-s-local-artefact-placement-check-matches-e"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "iss-155 adversarial review round 2026-07-29"
found_at: "internal/core/audit/rule_layout.go"
---

three-tier-layout's local-artefact placement check matches exact names directly under a committed tier only — residual evasions flagged at review: a nested artefact (a NEXT.md one directory below a tier root), a NEXT.md at the .abcd/ root itself (one directory off the modelled incident), and lowercase name variants on case-sensitive filesystems all pass clean. Widening (depth, roots, case folding) must not start flagging legitimate tier content.