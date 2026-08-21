---
schema_version: 1
id: "iss-2608211850070318"
slug: "history-list-json-null-on-empty-store"
severity: "nitpick"
category: "bug"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/surface/cli/cli.go:2750"
resolution: "history list --json normalises a nil slice to []"
impact: fix
resolved_by:
  commit: "9ea9177"
---

abcd history list --json emits bare null on an empty store while the command doc promises an empty list