---
schema_version: 1
id: "iss-2608300929274006"
slug: "report-and-gate-filename-grammars-diverge-for-rdi-and-dsp"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-189 build review, 2026-08-30"
found_at: "internal/core/lint/readingoutstanding.go, internal/core/issueschema/disposition.go"
---

The outstanding report's filename grammar for reading items and dispositions (readingItemFileRe, DispositionFileID) is stricter than the record_schema gate's FilenameNumRe, so a hand-written rdi-N-slug.md or dsp-N-slug.md passes the gate and is then invisible to the report; for those two families the divergence fails toward silence rather than a false claim. One grammar, the resolver's, for every family the report walks.
