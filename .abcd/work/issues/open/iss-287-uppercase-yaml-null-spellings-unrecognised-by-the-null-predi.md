---
schema_version: 1
id: "iss-287"
slug: "uppercase-yaml-null-spellings-unrecognised-by-the-null-predi"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "pr-294-review"
found_at: "internal/core/frontmatter/frontmatter.go"
---

Uppercase YAML null spellings unrecognised by the null predicates (reported as GitHub #290)

Retroactive capture of the bug reported as GitHub #290 so the ledger holds
the handle the CHANGELOG entry should cite: `frontmatter.IsNull` and lint's
`isNull` recognised `null` and `~` but not the YAML 1.2 core-schema
spellings `Null`/`NULL`, so records using an uppercase null were misread
(e.g. a superseded ADR's `superseded_by:` treated as a live handle). Fixed
in flight by PR #294, which widens both predicates with regression tests;
resolve this issue when that PR merges, citing its commit. Evidence:
`.abcd/work/reviews/2026-08-19-pr-294-null-predicate/` (F10).
Forge mirror: #374
