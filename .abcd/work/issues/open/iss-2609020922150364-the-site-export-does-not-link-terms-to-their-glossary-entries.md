---
schema_version: 1
id: "iss-2609020922150364"
slug: "the-site-export-does-not-link-terms-to-their-glossary-entries"
severity: "minor"
category: "future-work-seed"
source: "agent-finding"
found_during: "glossary-fix-2026-09-02"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/site"
---

The glossary now names every sense of the record's colliding terms (iss-2609012245352480), but the site export reads no glossary: internal/core/site carries no reference to it, so a rendered page shows 'phase', 'plan' or 'surface' as plain words and a reader on the site cannot reach the entry that disambiguates them. The record's ask was that the site links terms to the glossary; that half is Go work and stays open here. Directions, none adopted: the site renderer wraps the first occurrence of each glossary term on a page in a link to its entry; or every rendered page carries a glossary sidebar listing the terms it uses; or the glossary is exported as its own page set and chapters link by hand.
