---
schema_version: 1
id: "iss-51"
slug: "context-scoped-successor-bans"
severity: "minor"
category: "process"
source: "agent-finding"
found_during: "2026-07-09 practice/MVP/tool extraction"
found_at: ".abcd/record-lint.json"
resolution: "Strict banned_tokens schema: required successor + non-empty allow_context, rejected at load; finding auto-cites the successor; record-lint and docs-lint banlists migrated"
---

Extend the banlist so that every banned_tokens entry carries a successor mapping (old token to new token) and names the context in which the ban applies, with allow-contexts kept narrow and enumerable. The convention serves vocabulary hygiene without collateral damage: a term retired in one context may legitimately persist in another, and global bans over polysemous terms are false-positive generators that train people to ignore the lint. This extends the iss-36 banlist work. Detector: the record-lint banlist schema rejects entries lacking a successor or a named context, and acceptance is a lint run where every ban fires only inside its declared context and each hit reports the successor to use.