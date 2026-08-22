---
schema_version: 1
id: "iss-2608220136597127"
slug: "ahoy-guard-health-zero-value-serialised"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/ahoy/ahoy.go"
resolution: "DetectionResult.Guard is a pointer with omitempty, assigned only for a repo, matching the Banlist sibling"
impact: fix
resolved_by:
  commit: "4d06ab1"
---

ahoy DetectionResult.Guard is a value field with a plain json tag but is only assigned for a managed repo, so an unmanaged folder serialises a never-computed all-false GuardHealth (a broken guard) beside plugin_root_status resolved; the Banlist sibling is already a pointer with omitempty for this reason