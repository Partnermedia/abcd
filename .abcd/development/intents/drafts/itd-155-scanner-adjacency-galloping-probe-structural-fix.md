---
id: itd-155
slug: scanner-adjacency-galloping-probe-structural-fix
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-229
---

# implement the galloping/exponential-doubling probe (trueMatchEnd) that replaces the fixed 512-byte adjacency window in scanAllPatterns/probeAt: the window grows only while a match keeps running into its edge and stops the moment the match ends short of the edge or reaches the real end of line, so a match end is never a truncation artifact. This is the structural fix designed in the 2026-08-08 DECISIONS entry after bug-hunt rounds 6-8 each BLOCKed on local patches for the same root cause (a truncated window-edge view driving a discard/skip decision reaching further than the ambiguity). It dissolves iss-189 (trailing boundary satisfied by the artificial window edge, spurious over-redaction) and iss-190 (window-capped recovery truncating a token and breaking the recovery chain) outright, without reintroducing the round-6 cost regression: the window grows only for matches genuinely still growing, at the amortized cost class the top-level unbounded match already pays. iss-189 and iss-190 (both superseded by this capture) are its acceptance corpus: their repro shapes must pass without a boundary classifier or a clipped-so-skip special case.

## Press Release

> _Seeded by promotion from iss-229. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-229`: implement the galloping/exponential-doubling probe (trueMatchEnd) that replaces the fixed 512-byte adjacency window in scanAllPatterns/probeAt: the window grows only while a match keeps running into its edge and stops the moment the match ends short of the edge or reaches the real end of line, so a match end is never a truncation artifact. This is the structural fix designed in the 2026-08-08 DECISIONS entry after bug-hunt rounds 6-8 each BLOCKed on local patches for the same root cause (a truncated window-edge view driving a discard/skip decision reaching further than the ambiguity). It dissolves iss-189 (trailing boundary satisfied by the artificial window edge, spurious over-redaction) and iss-190 (window-capped recovery truncating a token and breaking the recovery chain) outright, without reintroducing the round-6 cost regression: the window grows only for matches genuinely still growing, at the amortized cost class the top-level unbounded match already pays. iss-189 and iss-190 (both superseded by this capture) are its acceptance corpus: their repro shapes must pass without a boundary classifier or a clipped-so-skip special case.. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
