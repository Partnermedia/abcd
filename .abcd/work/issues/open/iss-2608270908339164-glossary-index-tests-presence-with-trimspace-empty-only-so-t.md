---
schema_version: 1
id: "iss-2608270908339164"
slug: "glossary-index-tests-presence-with-trimspace-empty-only-so-t"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-2026-08-27"
found_at: "internal/core/glossary/index.go"
---

glossary index tests presence with TrimSpace-empty only, so term: NULL or status: ~ is accepted as a literal value and mints a term named NULL — a YAML null becoming data rather than a diagnosis