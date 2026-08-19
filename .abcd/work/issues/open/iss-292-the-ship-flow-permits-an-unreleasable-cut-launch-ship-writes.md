---
schema_version: 1
id: "iss-292"
slug: "the-ship-flow-permits-an-unreleasable-cut-launch-ship-writes"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
found_at: "internal/surface/cli/ship.go"
---

The ship flow permits an unreleasable cut: launch ship writes the dated heading with no check that the receipts protocol can complete, so the receipt gate fires at tag time — the most expensive possible moment, unrecoverable once tagged. The gate must shift left into the cut: ship (or a launch receipts sub-verb) refuses to finish a cut whose branch lacks PROMOTE receipts for the content commit, prints the two-commit protocol as its own checklist, and a product-thinker-grade operator never learns the protocol from a red release run