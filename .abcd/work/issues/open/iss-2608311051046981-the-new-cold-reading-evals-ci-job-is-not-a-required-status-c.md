---
schema_version: 1
id: "iss-2608311051046981"
slug: "the-new-cold-reading-evals-ci-job-is-not-a-required-status-c"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

The new cold-reading-evals CI job is not a required status check: .abcd/work/rulesets/main-protection.json lists attribution, check (macos-latest), check (ubuntu-latest), external-review, gitleaks, record-lint, smoke and zizmor, and the job spc-64 lands reports a conclusion without gating the merge. The whole point of the always-run lane is that a record-only pull request cannot reach main with warm content in included material, and an unrequired check does not stop one. Add cold-reading-evals to the ruleset and to its committed mirror.
