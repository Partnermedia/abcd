---
schema_version: 1
id: "iss-2608270500194738"
slug: "backfilllegacy-gates-frontmatter-presence-on-hasprefix-text"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/memory/writer.go"
---

backfillLegacy gates frontmatter presence on HasPrefix(text, '---') while the canonical parsers tolerate a leading HTML comment, demoting real source: provenance. GitHub mirror: #288