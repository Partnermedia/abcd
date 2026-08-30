---
schema_version: 1
id: "iss-2608300914552505"
slug: "quote-blanker-pairs-stray-quotes"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 seventh-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (quotedSpanRe, blankQuoted)"
---

The quote blanker treats any quote character as a scalar opener, so a quote inside a plain scalar (an apostrophe in ordinary prose) or an escaped quote inside a double-quoted scalar pairs with a later quote and blanks the excluded key sitting between them; the flow-key scan then never sees a nested origin and the value travels while the manifest asserts refusal — demonstrated end to end on three shapes. Blank only quotes in scalar-opening position (line start, after a colon-space, comma, brace, bracket or dash-space) with escape-aware double-quote matching, or refuse a quote that is not in opening position on any flow line as unresolvable; add the three shapes as tests.
