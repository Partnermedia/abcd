---
schema_version: 1
id: "iss-2608210737265820"
slug: "audit-ingest-appends-below-trailing-link-refs"
severity: "nitpick"
category: "bug"
source: "impl-review"
found_during: "itd-114 ship, verdict ingest"
---

intent audit ingest appends the INGESTED verdict block at end-of-file, so when a shipped intent carries link-reference definitions after its Audit Notes heading (itd-114: the iss-80 link ref) the block lands below them — rendering fine but visually detached from the Audit Notes section it belongs to. The writer should insert under the Audit Notes heading (or above trailing link-ref definitions) rather than appending at EOF