---
schema_version: 1
id: "iss-274"
slug: "abcd-lint-red-on-clean-main-checkout"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "scheduled process-build run 2026-08-19"
found_at: "internal/core/repolint"
wontfix_reason: "duplicate of iss-2608231000561060: identical red-lint/not-gated facts about the same rule"
---

abcd lint exits 2 on a clean main checkout: four privacy-hygiene errors flag illustrative example paths (internal/core/ahoy/fsutil.go:40+42, internal/core/ahoy/store.go:170, internal/surface/cli/staleness_test.go:14 — comment prose and a test fixture like /home/alice/.local/bin/abcd), and no CI job runs the full conformance verb, so the repo's own lint sits red between releases. Either the lines take the sanctioned abcd-lint:allow marker (they are deliberately illustrative) or the rule learns the persona convention; and whether abcd lint belongs in CI is a separate maintainer call.