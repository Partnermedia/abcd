---
schema_version: 1
id: "iss-2608261437046287"
slug: "capture-blocked-by-writes-an-unverified-cross-reference-its"
severity: "major"
category: "bug"
source: "agent-observation"
found_during: "bughunt-b-round-9"
found_at: "internal/core/capture/workflow.go"
---

capture --blocked-by writes an unverified cross-reference its own blocker then refuses