---
schema_version: 1
id: "iss-2609020154404742"
slug: "prioritise-and-openidset-internal-core-capture-workflow-go-b"
severity: "minor"
category: "bug"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/workflow.go"
---

prioritise and openIDSet (internal/core/capture/workflow.go) build the blocked_by projection from scanLedger's ISSUES alone and drop its Skipped roster, so an open record the guarded reader refused — a FIFO, a body over the read cap, a symlinked leaf — is invisible to the predicate. A dependent whose blocked_by names that record renders with an empty blocked_by_open and sorts ahead of genuinely unblocked work in both list and the status board, while its own blocked_by still names the blocker: the board understates the blocking, in the direction that invites work nobody can start. Both call sites are affected — List's openIDSet(ir) and Status's inline idSet(open). The fix must establish that a record skipped from open/ still counts as open for the projection, resolving the unreadable case toward still-blocking, with a test that plants an unreadable blocker and asserts the dependent still reports it blocking in list and in status.
