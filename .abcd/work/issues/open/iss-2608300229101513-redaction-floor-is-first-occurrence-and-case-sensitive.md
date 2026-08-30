---
schema_version: 1
id: "iss-2608300229101513"
slug: "redaction-floor-is-first-occurrence-and-case-sensitive"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 adversarial security review, 2026-08-30"
found_at: "internal/core/reading/project.go"
---

The exclusion floor's key and heading redaction is first-occurrence and exact-match: a duplicated origin key keeps its second copy, a frontmatter closed with four dashes is cut by StripFrontmatter but not read by Fields so the key survives in the body, and a heading spelled in another case is not redacted; the manifest still asserts the key and heading as refused. record_schema refuses the first two shapes in this repository, the third it does not. Make the key/heading half fail-closed like the path half: refuse the file on a duplicate excluded key, on an excluded key still at column 0 inside the first --- region, and on any heading matching an excluded title case-insensitively.
