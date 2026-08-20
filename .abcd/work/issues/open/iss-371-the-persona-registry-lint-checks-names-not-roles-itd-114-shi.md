---
schema_version: 1
id: "iss-371"
slug: "the-persona-registry-lint-checks-names-not-roles-itd-114-shi"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
found_at: ".abcd/development/personas.json"
---

The persona_registry lint checks names, not roles: itd-114 shipped 'Bob, a maintainer' (Bob is registered staff engineer; the maintainer role is Kira's) and itd-115 has 'Carol, a facilitator' (Nia's role) — both passed lint. The selection-by-role rule (personas.json, itd-79) has no detector for role-name agreement