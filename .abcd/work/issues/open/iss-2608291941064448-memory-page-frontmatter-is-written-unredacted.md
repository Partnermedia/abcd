---
schema_version: 1
id: "iss-2608291941064448"
slug: "memory-page-frontmatter-is-written-unredacted"
severity: "major"
category: "security"
source: "impl-review"
found_during: "ultra-v0.6.8-followup"
found_at: "internal/core/memory/writer.go"
---

ultra-v0.6.8 follow-up (review of the branch): WritePages in internal/core/memory/writer.go redacts only PageWrite.Body. PageWrite.Frontmatter is rendered by renderWrites and written verbatim, and on ask --file-back it is host-supplied: source.citation.title, recall entries and any other frontmatter string a distiller echoes reach the committed .abcd/memory store unscanned (reproduced with a synthetic token and a home path in citation.title). Pre-existing on main — Ingest never redacted frontmatter either — but the door beside GHSA-j5f5-phgm-9m73 remains open one field over. Fix: redact every string leaf of the frontmatter mapping through the store redactor inside WritePages before dumpFrontmatter, with the same fail-closed gate as the body; IngestEvent.Origin reaching registry.md is the same class.
