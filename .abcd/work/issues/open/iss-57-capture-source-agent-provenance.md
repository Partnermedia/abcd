---
schema_version: 1
id: "iss-57"
slug: "capture-source-agent-provenance"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "autonomous-run"
found_at: "internal/core/capture/capture.go"
---

capture --source has no autonomous-agent provenance value: agent-finding means a review agent's finding, but an autonomous run's self-observation has no honest source enum; the run had to reuse agent-finding. Candidate: add an agent-observation (or autonomous-run) source.