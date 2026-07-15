---
schema_version: 1
id: "iss-102"
slug: "capture-orphan-sweep-commit-race"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "multi-agent-bughunt"
---

cleanOrphanPlaceholders (capture/alloc.go) runs locklessly from mutationPreamble at the start of every Capture/Resolve/Wontfix, while commitCapture (workflow.go) fills the reserved placeholder OUTSIDE the ledger flock. A capture stalled >60s between reservation and commit can have its just-committed issue file deleted: the sweep's Lstat passes (size 0, aged) then the stalled commit's atomic rename lands, then the sweep's os.Remove deletes the fully-committed file — after Capture returned success. Deferred from the multi-agent bug hunt (was B26): a pre-unlink re-check (orphanStillRemovable: re-Lstat + SameFile + size==0) already mitigates the common interleaving, but full elimination requires running the sweep AND the commit write under withLedgerLock (the commit write path is in workflow.go). Low severity; data-loss only under a >60s stall plus concurrency.