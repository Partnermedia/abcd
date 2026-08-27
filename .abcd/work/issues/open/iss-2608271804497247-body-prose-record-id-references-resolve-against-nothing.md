---
schema_version: 1
id: "iss-2608271804497247"
slug: "body-prose-record-id-references-resolve-against-nothing"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "structural consistency review of .abcd/ and docs/ (2026-08-27)"
found_at: ".abcd/record-lint.json"
---

body-prose record-id references resolve against nothing: the context_citation_currency rule's record_stores mapping covers one file, and the site record export extracts only the eight typed frontmatter edges, so adr-N/itd-N/spc-N/iss-N tokens in body prose across the durable record are checked by no gate (the mechanism behind this review's dangling-id findings). Widen the existing gate rather than adding a second: extend the export's reference extraction to body-prose tokens, flow them into the same Unresolved set and site-baseline ratchet, and seed the baseline with the first run's backlog.