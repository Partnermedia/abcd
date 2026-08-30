---
schema_version: 1
id: "iss-2608300402598718"
slug: "itd-183-fifth-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-183 fifth-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go, .abcd/development/readings/README.md"
---

itd-183 fifth-round residue, all Low: a heading nested in a blockquote or list item and a homoglyph or format-character title leak, and neither residue is disclosed in the charter or the code although the fourth-round record named them; the raw-HTML refusal's comment says the site's markdown subset admits h1-h6 when the site refuses raw HTML blocks, and the per-line pattern misses an unclosed, multi-line, self-closing or attribute-carrying heading and a div with a heading role; a tagged key, an anchored key and a block-scalar explicit key are spellings the escape-is-the-signal rationale covers but the pattern set does not; the setext scan runs over frontmatter lines so a block scalar ending in the excluded title before the closing --- is refused with the wrong diagnosis.
