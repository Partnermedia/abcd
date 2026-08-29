---
schema_version: 1
id: "iss-2608291817368607"
slug: "ideate-record-commits-verdict-prose-without-scanner-pass"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "v0.6.9-security-pass"
found_at: "internal/core/ideate/record.go"
---

GitHub #486 sibling: abcd ideate record writes the verdict payload's idea text and the three legs' prose (claims, grill hits, kill attempts, rejected alternatives) to .abcd/development/research/<date>-ideate-<slug>.md through termsafe.CleanProse only — no pass through internal/adapter/scanner (internal/core/ideate/record.go validate, render.go). A secret or absolute home path in the host-produced JSON is committed verbatim, the same class capture (redactLedgerText) and intent (redactIntentText) close. Stance (fail-closed like intent, or redact-and-report like capture) is a design choice to settle before fixing.
