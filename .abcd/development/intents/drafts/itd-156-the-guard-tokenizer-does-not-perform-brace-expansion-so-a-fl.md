---
id: itd-156
slug: the-guard-tokenizer-does-not-perform-brace-expansion-so-a-fl
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-2608221457227161
---

# The guard tokenizer does not perform brace expansion, so a flag wrapped in a single-element brace group with an empty alternative (git push {--force,} origin main) expands in bash to byte-identical argv --force yet the guard reads the literal token {--force,} and allows it — a Tier-1 blocker miss of the same mutate-the-flag-token shape as the round-6 redirection fix. Distinct from the $'...' quoting gap (this is expansion, not quoting, and breaks no written invariant); recorded for a scoped follow-up because a correct bounded brace-expander is larger than this round's scope.

## Press Release

> _Seeded by promotion from iss-2608221457227161. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-2608221457227161`: The guard tokenizer does not perform brace expansion, so a flag wrapped in a single-element brace group with an empty alternative (git push {--force,} origin main) expands in bash to byte-identical argv --force yet the guard reads the literal token {--force,} and allows it — a Tier-1 blocker miss of the same mutate-the-flag-token shape as the round-6 redirection fix. Distinct from the $'...' quoting gap (this is expansion, not quoting, and breaks no written invariant); recorded for a scoped follow-up because a correct bounded brace-expander is larger than this round's scope.. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
