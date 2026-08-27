---
schema_version: 1
id: "iss-2608270500207078"
slug: "record-lint-and-docs-lint-test-the-md-extension-case-sensiti"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/record-lint, docs-lint"
---

record-lint and docs-lint test the .md extension case-sensitively, so renaming a record to .MD flips the gate from blockers to exit 0, voiding every blocking rule including record_schema's own malformed-filename check while the lifeboat still packs the file. GitHub mirror: #333