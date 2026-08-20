---
schema_version: 1
id: "iss-330"
slug: "parallel-mint-id-collision-observed-live-the-maintainer-s-bu"
severity: "minor"
category: "architectural-insight"
source: "user-observation"
found_during: "v0.6.1 re-cut"
---

Parallel-mint id collision observed live: the maintainer's bughunt session and the release session each minted iss-291..293 concurrently; the collision surfaced only at the merge-queue's record-lint on the second branch, costing a full release re-cut. Direct field evidence for itd-114 (collision-proof record ids) — the ledger's next-free-integer mint is unsafe under exactly the parallel-agent workflows the repo now runs daily