---
schema_version: 1
id: "iss-2608281222011114"
slug: "guard-health-still-reports-unchecked-after-fail-safe"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "batch F guard fail-safe fix (2026-08-28)"
found_at: "internal/core/ahoy/guard_health.go"
---

the guard health report still says a broken repo guard.json means commands run unchecked, but the fail-safe fix (iss-2608261551087492) keeps the bundled hazards armed in that case: guard_health.go's guardRegistryUnloadableReason and the RegistryLoadable bool model a broken repo layer as fully-unguarded, which is now false. The health check should distinguish 'repo overrides dropped, bundled hazards still armed' (a mild, expected state) from 'no registry at all' (the only genuinely-unguarded state, which the embed makes unreachable). Refine the health model and reword the reason. Follow-up to the fail-safe change.