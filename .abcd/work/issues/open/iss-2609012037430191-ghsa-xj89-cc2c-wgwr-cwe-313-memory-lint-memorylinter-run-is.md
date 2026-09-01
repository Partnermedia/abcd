---
schema_version: 1
id: "iss-2609012037430191"
slug: "ghsa-xj89-cc2c-wgwr-cwe-313-memory-lint-memorylinter-run-is"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/lint.go"
---

GHSA-xj89-cc2c-wgwr (CWE-313): memory lint (memoryLinter.run is source classes + licence + quotation) imports no scanner, so a page whose frontmatter or body already holds a PAT or a home path yields zero findings and a clean exit, and fileBackSource deep-copies citation and licence from every matched page onto the file-back page, so a contaminated store propagates itself on ask --file-back (reproduced at v0.7.0: lint --json reports one MS001 info and blockers 0 over a page carrying a PAT in citation.title, recall and the body). The fix must add a blocker-severity MR001 finding built from the store redactor (BlockingResidual over the scan plus the literal-home backstop) over every stored page, .sources_index.json and each text kept-original under sources/ — report, never mutate, per adr-13 — carrying the kind and the line but never the span in the message or the report files, emitting a blocker MR001 naming the store when the scanner is degraded rather than aborting, and the file-back clone must land redacted through the frontmatter walk.
