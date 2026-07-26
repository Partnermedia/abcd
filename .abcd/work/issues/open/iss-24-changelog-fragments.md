---
schema_version: 1
id: "iss-24"
slug: "changelog-fragments"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "sources-ingest session 2026-07-08"
---

changelog fragments for abcd-managed repos: per-change entries land as files in a fragments directory (towncrier pattern) and are assembled into CHANGELOG.md at release, so concurrent PRs never conflict on the Unreleased block; candidate abcd verb/discipline — assembly could hook the release gate

---

**Evidence (2026-07-26, the 2026-07-24 run-queue merge window):** the monolithic
`[Unreleased]` block produced three server-side merge conflicts in a single day
while the run's eleven PRs merged. The `.gitattributes` `CHANGELOG.md merge=union`
stopgap resolves every one of them locally but is ignored by the hosting
platform's server-side merge engine, so each round needed a manual
merge-main-into-branch commit (the 2026-07-21 record already names this
asymmetry). Any two open PRs that both carry an Unreleased entry conflict at the
same insertion point; the run protocol already assumes `changelog.d/<slug>.md`
fragments — the repo has never grown the directory or the assembly step.

**Design constraint discovered since capture:** the release gate now *reads* the
changelog records to derive the release tier (`releaseImpact` →
`ShippedSince(base).Impact()`, landed with iss-122), so fragment assembly cannot
be cosmetic — it must preserve the impact-derivation seam those readers depend
on. Together with the original "verb or discipline?" question this makes iss-24
a grill-then-implement item, not an autonomous fix.