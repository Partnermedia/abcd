---
schema_version: 1
id: "iss-2608220150151397"
slug: "brief-zero-live-bundles-contradiction"
severity: "minor"
category: "inconsistency"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: ".abcd/development/brief/01-product/03-mental-model.md"
resolution: "mental-model brief reworded to declared-but-not-yet-delivered, pointing at active and dissolved bundles"
impact: internal
resolved_by:
  commit: "554f97f"
---

brief 01-product/03-mental-model.md claims the corpus has zero live bundles and points at retired-bundle history, but four planned intents (itd-20/24/63/69) declare bundle-member of spc-83-operator-surfaces and intents/README heads it Active bundles; residual of resolved iss-123 which fixed the README half only