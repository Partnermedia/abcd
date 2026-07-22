---
schema_version: 1
id: "iss-117"
slug: "abcd-capture-resolve-and-intent-seeddraft-cannot-set-impact"
severity: "major"
category: "inconsistency"
source: "agent-finding"
found_during: "itd-73 phase 1 derived versioning"
found_at: "internal/core/capture/workflow.go"
---

abcd capture resolve and intent seedDraft cannot set impact, so the tool's own path produces a record the new issue_impact_valid and intent_impact_valid blockers reject