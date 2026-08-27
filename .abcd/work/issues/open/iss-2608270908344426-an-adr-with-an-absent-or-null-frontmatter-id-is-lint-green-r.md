---
schema_version: 1
id: "iss-2608270908344426"
slug: "an-adr-with-an-absent-or-null-frontmatter-id-is-lint-green-r"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-2026-08-27"
found_at: "internal/core/lint/schema.go"
---

an ADR with an absent or null frontmatter id is lint-green — record_schema declares no required ADR fields and checkRecordFilename returns early — while describeADR cannot dispatch it and recordid.Resolver resolves it from the filename, so two readers of one store disagree and abcd adr-N reports not found: the ADR-family analogue of the resolved intent and spec loader-parity defects