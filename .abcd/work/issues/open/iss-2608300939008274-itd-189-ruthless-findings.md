---
schema_version: 1
id: "iss-2608300939008274"
slug: "itd-189-ruthless-findings"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-189 adversarial ruthless review, 2026-08-30"
found_at: "internal/core/lint/readingoutstanding.go, .abcd/work/issues/README.md, internal/core/lint/schema.go"
---

itd-189 ruthless findings: a widening proposal whose admission sits behind an unreadable admissions tree is still reported outstanding with a remedy to write a disposition, contradicting the invariant the disposition side enforces (after the admitted check, break when the admissions tree is unreadable; add the mirror test); the issues README store contract still says the ledger root holds two sibling families while four roots are now declared; the held exclusion on the admission leg is untested; proposal is never resolved while occasioned_by is.
