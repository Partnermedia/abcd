---
schema_version: 1
id: "iss-13"
slug: "brief-stale-status-narration"
severity: "minor"
impact: internal
category: "drift"
source: "review-followup"
found_during: "roadmap-consistency-review"
found_at: ".abcd/development/brief"
resolution: "Stale status narration stripped from brief contracts: 'shipped —' status cells corrected to design target (no agents/ exist in this repo), '(spc-N — shipped)' parentheticals reduced to spec attribution, dated capture-history prose removed from 04-scope.md, and a delivery-state provenance note added to the brief README covering residual spc-N references (per adr-5)."
---

current-state brief files carry stale status narration (spc-X shipped, today only stubs ship) inside canonical design contracts, against ADR-5.