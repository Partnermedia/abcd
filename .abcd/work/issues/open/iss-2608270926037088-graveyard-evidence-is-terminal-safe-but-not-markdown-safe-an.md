---
schema_version: 1
id: "iss-2608270926037088"
slug: "graveyard-evidence-is-terminal-safe-but-not-markdown-safe-an"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "issue-sweep-review-2026-08-27"
found_at: "internal/core/lifeboat/graveyard.go"
---

graveyard evidence is terminal-safe but not markdown-safe and untyped: termsafe.Sanitize leaves CommonMark and raw-HTML openers for the layer-3 interpreter to read, and cap or shadow notices share the plain evidence string array with record-influenced text, so a crafted path can forge an omission notice — a typed field or CleanProse pass closes the channel