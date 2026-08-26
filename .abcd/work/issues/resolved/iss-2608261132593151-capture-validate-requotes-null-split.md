---
schema_version: 1
id: "iss-2608261132593151"
slug: "capture-validate-requotes-null-split"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/core/capture/validate.go:111"
resolution: "validateStrict tests nullness as the empty string alone, so quoted nulls are refused exactly where record-lint blocks them; the parse-layer loop that pinned the bug now asserts refusal"
impact: fix
resolved_by:
  commit: "eafcd89b"
---

capture validateStrict re-applies IsNull to the already-unquoted impact value, so quoted nulls pass capture while record-lint blocks them, and the IsNull widening added NULL and Null to the split its fix claimed closed