---
schema_version: 1
id: "iss-2608211914592726"
slug: "lint-residual-config-derived-unguarded-reads"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "internal/core/lint/persona.go:26"
---

residual config-derived os.ReadFile sites in internal/core/lint read cloned-repo-controlled paths without the containment/guarded-read stack the roots and glossary walks now use: persona.go registry, lint.go spec-doc and record-store reads, speclinks.go, contextcurrency.go, deliverystate.go changelog, indexdrift.go, subverbs.go snapshots. Triage each by attacker-controllability under a cloned repo and guard those in scope