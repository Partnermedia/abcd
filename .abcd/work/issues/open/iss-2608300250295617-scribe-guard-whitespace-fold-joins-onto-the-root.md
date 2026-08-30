---
schema_version: 1
id: "iss-2608300250295617"
slug: "scribe-guard-whitespace-fold-joins-onto-the-root"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-188 sixth-round security review, 2026-08-30"
found_at: "internal/core/lint/scribecontract_test.go (scribeSpacedSeparatorRe, scribeFold)"
---

The scribe guard's spaced-separator fold is a join, not only a fold: horizontal whitespace after an allowed root's trailing slash glues the next token onto it, so a shipped-tree path written in plain prose after .abcd/work/issues/ (space or TAB, bare or in one code span) becomes a segment under the allowed root and passes the prefix check; demonstrated against the real test. Scan both the folded and the decoded-but-uncollapsed views and union the findings.
