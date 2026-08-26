---
schema_version: 1
id: "iss-2608261437041050"
slug: "a-ledger-record-missing-schema-version-vanishes-from-capture"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "bughunt-b-round-9"
found_at: "internal/core/lint/schema.go"
---

a ledger record missing schema_version vanishes from capture surfaces while every gate stays green