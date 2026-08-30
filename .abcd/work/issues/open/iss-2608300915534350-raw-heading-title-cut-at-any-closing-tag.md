---
schema_version: 1
id: "iss-2608300915534350"
slug: "raw-heading-title-cut-at-any-closing-tag"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 seventh-round ruthless review, 2026-08-30"
found_at: "internal/core/reading/project.go (rawHeadingEndRe, blockCloser, quotedSpanRe, setext loop)"
---

The raw-heading title is cut at the first closing tag of ANY element, so inline markup inside a heading element — an anchor then text, or emphasis on the first word — truncates or empties the title and the excluded section is admitted (refused before the seventh-round commit); bound the title at the opening element's own closing tag, the next heading open, or a blank line. Also: blockCloser trims indentation before testing three dots so an indented ellipsis inside a block scalar closes the block and refuses the file (test at column 0 only); the double-quoted span does not honour an escaped quote; the setext loop reads the frontmatter closer as an underline of a column-0 last frontmatter line.
