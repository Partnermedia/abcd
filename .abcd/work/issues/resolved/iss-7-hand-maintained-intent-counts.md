---
schema_version: 1
id: "iss-7"
slug: "hand-maintained-intent-counts"
severity: "major"
category: "drift"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief/01-product/04-scope.md"
blocked_by: [iss-3]
resolution: "Hand-maintained 'thirteen phased intents' counts in 04-scope.md and 03-mental-model.md replaced with derivation-by-pointer to the phase docs' Scope sections (adr-9): the phased set is the union across phase docs, never transcribed into the brief."
---

04-scope.md hand-maintains intent counts (thirteen phased intents) that the phase docs and planned corpus have outgrown.