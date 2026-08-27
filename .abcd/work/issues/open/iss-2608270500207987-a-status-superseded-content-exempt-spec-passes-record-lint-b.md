---
schema_version: 1
id: "iss-2608270500207987"
slug: "a-status-superseded-content-exempt-spec-passes-record-lint-b"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/spec (spec.Load)"
---

a status: superseded (content-exempt) spec passes record-lint but spec.Load validates intent/id unconditionally, so a malformed intent/id aborts the entire load. GitHub mirror: #331