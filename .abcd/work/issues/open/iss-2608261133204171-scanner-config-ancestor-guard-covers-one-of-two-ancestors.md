---
schema_version: 1
id: "iss-2608261133204171"
slug: "scanner-config-ancestor-guard-covers-one-of-two-ancestors"
severity: "nitpick"
category: "security"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "internal/adapter/scanner/scanner.go:126"
---

scanner.New lstats only the .abcd ancestor while its config lives two levels down, so a symlinked .abcd/config walks the pii.json read out of the tree; the read belongs in the ReadGuardedInRoot form