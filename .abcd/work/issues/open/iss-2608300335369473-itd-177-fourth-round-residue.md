---
schema_version: 1
id: "iss-2608300335369473"
slug: "itd-177-fourth-round-residue"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "itd-177 fourth-round security review, 2026-08-30"
found_at: "internal/core/intent/lifecycle.go (Plan draft face, writeIntentFile comment), internal/core/intent/claims.go (opensComment), internal/core/intent/audit.go, internal/core/intent/create.go"
---

itd-177 fourth-round residue, all Low: the write cap is applied per write but the final size is known before the first, so a draft within one byte of the cap passes the kind write, is moved to planned, then has its spec_id write refused, leaving a half-planned record neither Plan nor Link repairs (compute the largest intermediate before any write); opensComment resolves code spans before comment openers, so a backtick inside a live comment re-pairs the rest of the line and the mask diverges from CommonMark in both directions (one left-to-right cursor where the first construct wins); the three audit-block upserts and the create-path seed write still call WriteFileAtomic directly and are uncapped, so the helper's one-way claim overstates (pre-existing on the base branch).
