---
schema_version: 1
id: "iss-67"
slug: "intent-lifecycle-fidelity-gaps"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/intent/review.go"
resolution: "Plan/ingest AC gate aligned (require top-level bullet, no perpetual dead-letter); Audit Notes placeholder cleared on first review block (both delimiter styles; itd-80 backfilled); ReEmitReview doc corrected. seed9 (byte-identical DEAD_LETTER re-write) deferred as cosmetic. security PASS + ruthless SHIP."
---

intent lifecycle fidelity gaps (itd-80): Plan accepts an Acceptance-Criteria section with no top-level bullet (hasAcceptanceCriteria only checks non-blank) so the intent plans then perpetually dead-letters at ingest (countAcceptanceCriteria==0) — Plan and ingest disagree on has-criteria (intent.go:108, C9/seed8); review ingest appends the audit below the template Empty-until placeholder instead of clearing it on first verdict (seed3); ReEmitReview documents a re-park but silently no-ops on an already-INGESTED/DEAD_LETTER receipt (review.go:188, C10/seed10); DEAD_LETTER re-ingest rewrites byte-identical content instead of a no-op (seed9). Detector: Plan and ingest share countAcceptanceCriteria; first-verdict-clears-placeholder; terminal-receipt re-emit; idempotent deadletter. Corpus: C9, seed3, C10, seed9.