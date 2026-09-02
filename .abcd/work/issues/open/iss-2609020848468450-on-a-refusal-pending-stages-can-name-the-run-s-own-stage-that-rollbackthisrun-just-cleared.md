---
schema_version: 1
id: "iss-2609020848468450"
slug: "on-a-refusal-pending-stages-can-name-the-run-s-own-stage-that-rollbackthisrun-just-cleared"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "ingest-help-fix-2026-09-02"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/ingest.go"
---

On the refusal path the ingest result's pending_stages is set from the orphan list gathered before the refusal and is never recomputed after rollbackThisRun runs, so it can name the refused run's own stage as pending although that stage was just cleared. The disclosure the surface renders is then wrong by one entry in exactly the case it exists to describe. Found while pinning the ingest help's sentences (iss-2609020843067458). Fix: recompute or filter the pending list after the run's own rollback, and pin it with the same end-to-end refusal test that now proves the record set.
