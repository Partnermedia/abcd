---
schema_version: 1
id: "iss-2608220147106835"
slug: "json-empty-collections-null-not-array"
severity: "nitpick"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/memory/bare.go"
resolution: "spec/intent/memory/capture constructors seed their collections non-nil so an empty store marshals []"
impact: fix
resolved_by:
  commit: "1b1563e"
---

abcd spec/intent/memory (and capture recent_open) emit bare null for empty --json collections instead of [], contradicting the class-wide invariant history list fixed; a consumer iterating the value (jq .[], an agent following the command doc) errors on null