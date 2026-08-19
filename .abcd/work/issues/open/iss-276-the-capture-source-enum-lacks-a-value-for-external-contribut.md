---
schema_version: 1
id: "iss-276"
slug: "the-capture-source-enum-lacks-a-value-for-external-contribut"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "manual-capture"
found_at: "internal/core/capture/capture.go"
---

The capture source enum lacks a value for external-contribution intake: validSources in internal/core/capture/capture.go names nine sources, none fitting a finding that arrives from an outside PR's triage; the intake pipeline (.abcd/work/intake.md) will need one