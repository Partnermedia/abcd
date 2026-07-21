---
schema_version: 1
id: "iss-57"
slug: "capture-source-agent-provenance"
severity: "minor"
impact: additive
category: "observation"
source: "agent-finding"
found_during: "autonomous-run"
found_at: "internal/core/capture/capture.go"
resolution: "Added agent-observation as a valid --source value (validSources map) plus the brief enumeration; detector test asserts accept + bogus-reject."
---

capture --source has no autonomous-agent provenance value: agent-finding means a review agent's finding, but an autonomous run's self-observation has no honest source enum; the run had to reuse agent-finding. Candidate: add an agent-observation (or autonomous-run) source.