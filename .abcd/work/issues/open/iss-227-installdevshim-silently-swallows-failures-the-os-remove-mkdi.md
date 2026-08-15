---
schema_version: 1
id: "iss-227"
slug: "installdevshim-silently-swallows-failures-the-os-remove-mkdi"
severity: "minor"
category: "observation"
source: "agent-observation"
found_during: "ahoy install dogfood"
found_at: "internal/core/ahoy/apply.go"
---

installDevShim silently swallows failures: the os.Remove, MkdirAll, and WriteFileAtomic error paths are bare returns (internal/core/ahoy/apply.go, installDevShim), so a failed shim write yields status=partial or clean with no note explaining what was not done or why; only the success path calls a.note