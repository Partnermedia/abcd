---
schema_version: 1
id: "iss-2608270500196428"
slug: "disembark-review-writes-untrusted-provenance-json-fields-int"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/lifeboat (disembark review)"
---

disembark review writes untrusted _provenance.json fields into the durable review .md via termsafe.Sanitize, which does not neutralise CommonMark HTML-block openers (<), allowing a hostile lifeboat to inject markdown structure into a persisted file. GitHub mirror: #325