---
schema_version: 1
id: "iss-2608211432483805"
slug: "work-tier-roster-omits-issue-ledger"
severity: "nitpick"
category: "documentation"
source: "user-observation"
found_during: "bughunt-b/round-5"
found_at: "AGENTS.md"
resolution: "Name the issue ledger, reviews charter, and ruleset mirror in the work-tier roster"
impact: internal
resolved_by:
  commit: "1c3ae21"
---

the .abcd/work tier roster in AGENTS.md and .abcd/README.md names only CONTEXT.md and DECISIONS.md, omitting the committed issue ledger (adr-32), reviews, and rulesets that also live there