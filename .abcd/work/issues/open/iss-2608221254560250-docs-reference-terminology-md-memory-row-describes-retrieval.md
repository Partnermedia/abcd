---
schema_version: 1
id: "iss-2608221254560250"
slug: "docs-reference-terminology-md-memory-row-describes-retrieval"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "context-window SOTA investigation"
found_at: "docs/reference/terminology.md"
---

docs/reference/terminology.md Memory row describes retrieval as 'recall-matched and budget-bracketed', but no retrieval engine exists: the recall: frontmatter field is parsed and round-tripped yet never consumed, and the budget brackets live only in itd-39 (draft). The row overstates shipped behaviour.