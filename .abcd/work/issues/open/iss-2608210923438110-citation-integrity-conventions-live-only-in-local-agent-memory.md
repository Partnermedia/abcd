---
schema_version: 1
id: "iss-2608210923438110"
slug: "citation-integrity-conventions-live-only-in-local-agent-memory"
severity: "minor"
category: "tech-debt"
source: "user-observation"
found_during: "memory-portability audit"
---

Citation-integrity conventions exist only in one user's local agent memory, not in the record or the cite/ingest surfaces — any other agent building a references baseline in a managed repo would repeat both corrected mistakes: (1) admitting entries with initial-only author names plus a to-be-checked caveat when the entry's own URL/DOI would resolve the full names in one fetch (caveats are for genuinely unreachable facts, not skipped lookups); (2) taking author names from anywhere but the publisher's CURRENT record — a published name that has changed must never be reverted to the former name (the maintainer flagged a live near-miss hard). Both belong in the product: a convention note for docs cite / ingest now, and a validator rung on the ingest metadata extraction later