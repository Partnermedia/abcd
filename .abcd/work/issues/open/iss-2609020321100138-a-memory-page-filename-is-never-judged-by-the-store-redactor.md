---
schema_version: 1
id: "iss-2609020321100138"
slug: "a-memory-page-filename-is-never-judged-by-the-store-redactor"
severity: "major"
category: "observation"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/writer.go"
---

A memory page FILENAME is never judged by the store redactor, so a secret-shaped slug reaches the committed tree as a filename, in index.md and log.md, and now also as the registry back-link. slugRe admits [A-Za-z0-9_-] and pageNameRe admits <type>_<domain>_<slug>.md, so a host distiller returning slug ghp_<40 chars> writes topic_auth_ghp_....md verbatim; redactLeaves judged only leaf VALUES and KEYS, and excluding the registry back-links (the pruneOrphans data-loss fix) removed the accidental partial coverage the registry rewrite gave. Evidence: validatePageFilename in internal/core/memory/writer.go and DistilledPage.Filename in schema.go; reproduced on this branch (filename, index.md and .sources_index.json all carry the token). The fix must judge the filename at the write boundary and REFUSE rather than rewrite, as judgeKey does for a map key. It needs a decision first: storeRedactor.residue uses BlockingResidual, which treats a warn-severity network span as blocking, and an ordinary slug such as [redacted-hostname] matches net_device_hostname — so the filename rule must run at a narrower bar (hard_fail secrets only) or an ordinary page would be refused.
