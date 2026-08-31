---
schema_version: 1
id: "iss-2608311418202794"
slug: "a-stable-reversal-of-the-field-order-within-one-path-changes"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "delta review of itd-187"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/assemble.go"
---

A stable reversal of the field order within one path changes the assembled bundle's bytes and is caught by nothing: not the cold-reading determinism eval, whose order oracle compares paths only and whose byte comparison sees both runs agree, and not the assembler package's own tests. Five items can share one path when a record is projected field by field, so the ordering of those fields inside a path is a real degree of freedom in the artefact a reading is handed. A map-iterated field order is caught, because it differs across processes and fails the byte comparison; a deterministic but reversed one is not. This is the assembler's property rather than the amnesia eval's, so it belongs to spc-61 and predates the eval that exposed it.
