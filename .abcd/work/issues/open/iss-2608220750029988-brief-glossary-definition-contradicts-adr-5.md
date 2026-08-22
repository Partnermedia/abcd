---
schema_version: 1
id: "iss-2608220750029988"
slug: "brief-glossary-definition-contradicts-adr-5"
severity: "minor"
category: "inconsistency"
source: "user-observation"
found_during: "2026-08-22 filing session (NEXT.md handover)"
found_at: ".abcd/development/brief/glossary/core/brief.md"
---

glossary/core/brief.md's frontmatter definition says the brief defines the project before any implementation begins — contradicting adr-5 and the file's own body, which say the brief is the current state revised in place. Its When-to-use section also claims the brief lives at the root of the .abcd/ hierarchy; the path is .abcd/development/brief/.