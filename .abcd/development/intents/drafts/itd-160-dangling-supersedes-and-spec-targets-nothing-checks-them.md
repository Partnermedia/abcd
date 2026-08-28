---
id: itd-160
slug: dangling-supersedes-and-spec-targets-nothing-checks-them
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-2608220150157498
---

# Eight typed cross-references point at targets absent from the tree — adr-22 supersedes adr-14, adr-15 and adr-17; adr-25 supersedes adr-8; adr-27 supersedes adr-16; adr-28 supersedes adr-18; adr-35 supersedes adr-4 (all retired under retire-the-name); itd-3 names spec_id spc-1 which has no file — and nothing checks supersedes targets today. The 2026-08-21 site investigation counted six; the in-session grep found eight. The planned site build arms the detector via the .abcd/site-baseline.json ratchet, seeded with what the build finds, and the tombstones-or-stubs question (itd-136/itd-137) decides how they render

## Press Release

> _Seeded by promotion from iss-2608220150157498. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-2608220150157498`: Eight typed cross-references point at targets absent from the tree — adr-22 supersedes adr-14, adr-15 and adr-17; adr-25 supersedes adr-8; adr-27 supersedes adr-16; adr-28 supersedes adr-18; adr-35 supersedes adr-4 (all retired under retire-the-name); itd-3 names spec_id spc-1 which has no file — and nothing checks supersedes targets today. The 2026-08-21 site investigation counted six; the in-session grep found eight. The planned site build arms the detector via the .abcd/site-baseline.json ratchet, seeded with what the build finds, and the tombstones-or-stubs question (itd-136/itd-137) decides how they render. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
