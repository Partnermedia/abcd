---
schema_version: 1
id: "iss-2608220750029987"
slug: "meta-ahoy-skeleton-copy-claim"
severity: "minor"
category: "drift"
source: "user-observation"
found_during: "2026-08-22 filing session (NEXT.md handover)"
found_at: ".abcd/development/brief/00-meta.md"
resolution: "00-meta.md § Brief vs lifeboat restated: ahoy scaffolds the config pair; the skeleton leg is a staged design target — the all-blank coverage state, not a copied file tree."
impact: internal
resolved_by:
  commit: "425ec89"
---

00-meta.md § Brief vs lifeboat claims /abcd:ahoy copies an empty brief skeleton into a fresh repo; shipped ahoy scaffolds only .abcd/config.json + .abcd/rules.json (plus banlist artefacts), never the skeleton. Restate as a design target aligned with the interview-entry design — greenfield is the all-blank coverage state, not a copy-a-skeleton step.