---
schema_version: 1
id: "iss-2608300914553076"
slug: "frontmatter-shape-refusal-fires-on-body-rules"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "itd-183 seventh-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (unresolvableFrontmatterShape, inFirstBlock), internal/core/site (isFenceLine)"
---

unresolvableFrontmatterShape and the setext first-block bound run on the first --- found anywhere, so a frontmatter-less admitted doc with a thematic break refuses the whole assembly (a lone rule reads as an unclosed block; a rule followed by an image line reads as a YAML tag; a rule followed by an anchor-like line or a dots paragraph likewise); every docs page is admitted and several open with image lines. Scope both to a block opening at line 0 (BOM allowed), which is the only place the stripper recognises one. Residue, pre-existing and shared with the site renderer: a four-space-indented fence delimiter is accepted as a fence while CommonMark reads an indented code block, so a heading between two such lines is masked.
