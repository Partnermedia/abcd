---
schema_version: 1
id: "iss-2609012037432966"
slug: "ghsa-x46m-mw9h-5jwj-cwe-312-writepages-redacts-only-pagewrit"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/writer.go"
---

GHSA-x46m-mw9h-5jwj (CWE-312): WritePages redacts only PageWrite.Body; the frontmatter is dumped verbatim by renderWrites, so host-supplied recall, contradicts, citation scalars, sources[].licence and weighting_note land in the committed store unscanned, and the core-copied licence (DetectLicence from an SPDX line or a License: header), the redirect-controlled origin and title (materialFromFetched FinalURL, query string included) and every registry placement (the MergeIngest fresh entry and its fill-if-empty origin, backlinkOtherHashes, the ask --file-back event) reach .sources_index.json, contradictions.md and the ingest --json and ask --json results raw (reproduced at v0.7.0: a PAT in an SPDX header is returned in licence and stored three times in the page and once in the registry). Strict superset of iss-2608291941064448, which names host frontmatter through WritePages only. The fix must walk every string leaf a write introduces — the page frontmatter in WritePages and the registry leaves the merge adds, never the cached ones, so a dirty cached citation cannot lock a re-ingest out — through the store redactor with the same fail-closed discipline as the body (redact, sweep the home, refuse only a BlockingResidual), redact licence, citation and title at the source in Ingest so the returned result matches what was written, and build the redactor on the registry-only fast path too, where today none is built.
