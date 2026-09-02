---
schema_version: 1
id: "iss-2609021153269181"
slug: "the-ingest-verb-refuses-an-output-carrying-no-items-where-th"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "final compliance review of the Iteration 2 materials, 2026-09-02"
origin: researcher-authored
production_mode: dictated-and-formatted
found_at: "internal/core/reading/ingest.go"
---

The ingest verb refuses an output carrying no items, where the framework's clean-run contingency records a null result as a run with an empty item set
