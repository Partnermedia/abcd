---
schema_version: 1
id: "iss-2608300352403199"
slug: "pre-write-size-check-skipped-for-multi-condition-drafts"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-177 fifth-round ruthless review, 2026-08-30"
found_at: "internal/core/intent/lifecycle.go (checkDraftFaceSize, sizeProbeMinter, Plan)"
---

checkDraftFaceSize's probe minter uses constant entropy, so on a draft with two or more unmarked scope-condition bullets the second mint exhausts its redraws with a collision error that the check swallows as though it were a structural refusal, and the pre-write size judgement is silently skipped for the common case; Plan then stamps, writes kind, moves to planned, mints a spec and refuses the spec_id write, leaving the half-planned record and dangling spec the resolution of iss-2608300335369473 claims can no longer happen. Give the probe an advancing entropy source keeping every id sixteen digits wide, narrow or remove the error swallow, and extend the pre-write test to two conditions. Also re-run the check with the real spec id immediately after spec.Create and before the first write, since NextID is predicted outside the mint lock and a width crossing costs the same half-planned state; correct the comment and the resolution text.
