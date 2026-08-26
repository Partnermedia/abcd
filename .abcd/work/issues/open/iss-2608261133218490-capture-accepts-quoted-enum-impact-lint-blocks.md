---
schema_version: 1
id: "iss-2608261133218490"
slug: "capture-accepts-quoted-enum-impact-lint-blocks"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/core/capture/validate.go:104"
---

capture validateStrict accepts a quoted legal enum impact and a quoted empty impact that record-lint and the release derivation block; the quoted-scalar acceptance is wider than the nulls the round-8 fix closed