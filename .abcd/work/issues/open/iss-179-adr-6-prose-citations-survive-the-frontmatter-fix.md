---
schema_version: 1
id: "iss-179"
slug: "adr-6-prose-citations-survive-the-frontmatter-fix"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "iss-39 record-schema validation"
found_at: ".abcd/development/decisions/adrs"
---

adr-29 and adr-25 still narrate adr-6 in prose, but adr-6 is in no store and is absent from all git history (the ADR set landed in one import commit; 0004/0006/0008/0014-0018 were never migrated, and every one but 0006 is accounted for by a successor's supersedes). adr-29 states in its own body that adr-6's decision stands and that adr-29 does not supersede it, so the supersession vocabulary cannot record it and neither ADR body should be rewritten to paper over it. iss-39 closed the machine-readable half (the two related_adrs entries, and the two brief pages that cited it). The surviving prose sites are adr-29 lines 24, 39, 43, 56, 60, 63, 71 and adr-25 lines 51, 80. Decide whether the record carries an unmigrated-predecessor marker, restores adr-6 as a reconstructed record, or accepts the prose as historical narration; record_schema is frontmatter-scoped and does not read prose, so nothing gates this either way.