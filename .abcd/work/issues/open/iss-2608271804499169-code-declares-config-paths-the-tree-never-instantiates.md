---
schema_version: 1
id: "iss-2608271804499169"
slug: "code-declares-config-paths-the-tree-never-instantiates"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "structural consistency review of .abcd/ and docs/ (2026-08-27)"
found_at: "internal/core/repolint/rule_privacy.go"
---

the binary declares two per-repo config paths the tree never instantiates: rule_privacy reads .abcd/config/pii.json and the launch includes-closure reads .abcd/config/scripts-closure.json, but .abcd/config/ holds neither and no doc mentions them. Both read as optional overrides, so nothing is broken — but code-declared record paths and tree-instantiated ones have no reconciliation check in either direction. Document the two optional files where the config/ members get their index entry, and consider a parity sweep between code path literals and the documented namespace.