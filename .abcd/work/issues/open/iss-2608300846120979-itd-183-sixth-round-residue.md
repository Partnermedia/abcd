---
schema_version: 1
id: "iss-2608300846120979"
slug: "itd-183-sixth-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-183 sixth-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go, internal/core/reading/assemble_test.go"
---

itd-183 sixth-round residue: the tag and anchor floor catches only the double-bang shorthand at line start — single-bang local and verbatim tags, and an explicit key carrying a tag, an anchor or an escape are all admitted (refuse any tag at line start and any explicit-key line the pattern cannot fully read); the raw-heading scan blanks fenced lines so an HTML block opened outside a fence that continues through fence delimiters is mis-titled and admitted (use the fence mask only to disqualify an open tag on a fenced line); a div with a heading role is neither caught nor disclosed; the renderedText comment does not say a double-encoded reference is deliberately left as the different heading it renders as; the amp-entity test still continues on error rather than asserting the refusal the commit and record claim.
