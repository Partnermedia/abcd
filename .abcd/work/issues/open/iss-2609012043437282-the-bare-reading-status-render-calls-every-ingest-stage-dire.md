---
schema_version: 1
id: "iss-2609012043437282"
slug: "the-bare-reading-status-render-calls-every-ingest-stage-dire"
severity: "minor"
category: "bug"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/status.go"
---

The bare reading status render calls every ingest stage directory an orphan, and promises a rollback that a committed run will never get. Describe (internal/core/reading/status.go) lists IngestStageDir and reports every well-formed run-id directory under orphaned_ingests, described as a run with no commit marker whose reading records the next validating ingest will roll back. But a stage also survives when the commit path's RemoveAll fails AFTER run.json has landed (the Degraded branch at the end of write in ingest.go): that run is complete, and the sweep's rollbackRun already tells the two apart by probing ReadingsRecordDir/<id>/run.json and leaving a committed run's records alone. So for that case the CLI interrupted line (renderReadingStatus in internal/surface/cli/reading.go) and the orphaned_ingests paragraph on commands/reading.md state something false: the records are not for a run that never happened and will not be rolled back; only the stage will be cleared. Both reviewers found it independently. The fix must make Describe probe run.json and report the two cases apart — an orphan whose records will be rolled back, versus a committed run whose leftover stage will merely be cleared — and make the CLI render and the plugin page say the right thing for each.
