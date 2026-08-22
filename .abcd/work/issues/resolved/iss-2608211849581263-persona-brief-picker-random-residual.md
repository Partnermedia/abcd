---
schema_version: 1
id: "iss-2608211849581263"
slug: "persona-brief-picker-random-residual"
severity: "minor"
category: "inconsistency"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: ".abcd/development/brief/01-product/05-personas.md:3"
resolution: "05-personas.md now states selection is by role, never by name"
impact: internal
resolved_by:
  commit: "1395e72"
---

brief 05-personas.md says the persona picker chooses at random, contradicting personas.json and itd-79 (selection is by role, never by name); a live residual of resolved iss-300