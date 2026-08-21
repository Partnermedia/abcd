---
schema_version: 1
id: "iss-2608211850070541"
slug: "identity-init-fault-exit-code"
severity: "nitpick"
category: "bug"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/surface/cli/identity.go:83"
resolution: "identity init faults now exit 2, matching sibling verbs"
impact: fix
resolved_by:
  commit: "c09aeb1"
---

abcd identity init maps a structural fault to exit 1 (the rendered-refusal code) while its sibling verbs and the tree-wide usage-error contract use exit 2; init renders nothing on that path