---
schema_version: 1
id: "iss-2608270500184205"
slug: "abcd-memory-ask-rendercitedmatches-prints-citation-fields-to"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/memory/ask.go"
---

abcd memory ask (RenderCitedMatches) prints citation fields to the terminal without termsafe.Sanitize, so an escape/control sequence in a stored page reaches the terminal raw. The termsafe sweep covered history/intent/record-lint/launch/disembark/cite render sites but not the memory-ask surface. GitHub mirror: #250