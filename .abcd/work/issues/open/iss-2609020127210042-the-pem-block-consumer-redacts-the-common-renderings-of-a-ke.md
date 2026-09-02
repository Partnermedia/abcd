---
schema_version: 1
id: "iss-2609020127210042"
slug: "the-pem-block-consumer-redacts-the-common-renderings-of-a-ke"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/adapter/scanner/pem.go"
---

The PEM block consumer redacts the common renderings of a key body but not every one, and one armour form is not detected at all. scanner.pemBodyLineRe accepts an optional gutter then a single base64 run, so a body rendered as log-prefixed lines (a timestamp/level prefix), as CSV cells, as one element per XML tag, with a trailing hash comment, as a source-string concatenation, as a JSON array on one line, or with double-escaped backslash-n separators is not body-shaped, the block ends at that line and the rest is stored verbatim; the tail of a block longer than maxPEMBodyLines survives by the same route. The RFC4716 armour (four dashes, BEGIN SSH2 ENCRYPTED PRIVATE KEY) does not match the bundled pem_private_key pattern at all, so no block is ever opened for it. Evidence: scanner.pemBodyLineRe and maxPEMBodyLines in internal/adapter/scanner/pem.go, DefaultPatterns pem_private_key in patterns.go. A fix must either widen the body-shape rule over these renderings or arm an entropy rule for a headerless base64 run (iss-96), and add the RFC4716 armour to the header pattern.
