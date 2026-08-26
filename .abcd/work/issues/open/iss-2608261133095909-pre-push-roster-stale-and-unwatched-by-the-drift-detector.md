---
schema_version: 1
id: "iss-2608261133095909"
slug: "pre-push-roster-stale-and-unwatched-by-the-drift-detector"
severity: "minor"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: ".githooks/pre-push:11"
---

the pre-push hook restates a three-gate preflight roster while the recipe has five, and the preflight-gates drift detector omits the hook from its surface list, so the iss-182 drift recurred on the one unwatched surface