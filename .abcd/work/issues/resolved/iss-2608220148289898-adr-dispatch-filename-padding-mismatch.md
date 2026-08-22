---
schema_version: 1
id: "iss-2608220148289898"
slug: "adr-dispatch-filename-padding-mismatch"
severity: "nitpick"
category: "bug"
source: "agent-finding"
found_during: "bughunt-b-round-7"
found_at: "internal/core/record/record.go"
resolution: "describeADR routes by the filename numeric ordinal, padding-agnostic; spc-26 amended"
impact: fix
resolved_by:
  commit: "7ccd966"
---

record.describeADR routes ADR filenames by a fixed %04d- prefix, so an ADR file with non-four-digit padding (00099-, 77-) is lint-green and citation-resolvable but reported not-found by abcd adr-N; the filename-ordinal residue of iss-2608211140410761 which fixed the frontmatter-id half; spc-26 spec line amended to numeric-ordinal probe