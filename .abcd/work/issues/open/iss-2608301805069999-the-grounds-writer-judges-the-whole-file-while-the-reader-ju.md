---
schema_version: 1
id: "iss-2608301805069999"
slug: "the-grounds-writer-judges-the-whole-file-while-the-reader-ju"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "itd-179-delta-security"
found_at: "internal/core/capture/validate.go"
---

the grounds writer judges the whole file while the reader judges the body so a frontmatter Grounds comment makes triage report success and record nothing

Found by the itd-179 delta security review, which reproduced it end to end
through the production path. BRANCH-INTRODUCED. Fail-OPEN.

Root cause, stated by that review better than the symptom states it: the WRITER
(`AppendToSection`) is handed the whole file, frontmatter and body, while the
READER (`groundsEntries`) is handed the body alone. The read-back guard is
therefore a property of the whole file rather than of the section the reader
will actually consult, so writer and reader can agree that an entry landed while
disagreeing about where.

The trigger is a legal YAML comment. `parseFrontmatterBlock` skips a line whose
trimmed form starts with `#`, so `# Grounds` sits happily in frontmatter -- and
`headingRe` (`^#{1,6}\s+Grounds\s*$`) matches it as the section heading. The
bullet is appended into that pseudo-section, the read-back runs over the whole
file and agrees, and the verb reports success. Then `issueFromFrontmatter`
computes grounds over the BODY only, finds no heading, and returns nothing.

The record moves to `resolved/` carrying a conjecture no surface can read. That
is precisely the outcome adr-57 argues against in its own consequences: a value
nothing reads.

Reachability is lower than its sibling: no committed record carries such a
comment and capture never writes one, so it needs a hand-edited or
contributor-supplied record. It is filed because it shares a root cause with
iss-2608301803423101 and ONE fix closes both -- run the append and its read-back
over the BODY and splice the result back onto the frontmatter, after which
`headingRe` can never match a frontmatter comment at all.

The sibling shape (`# Grounds` followed by `# Notes`) puts the bullet inside the
frontmatter and IS caught downstream, fail-closed, but with a message naming
neither cause nor remedy.

