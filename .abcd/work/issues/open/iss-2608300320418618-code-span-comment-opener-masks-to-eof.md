---
schema_version: 1
id: "iss-2608300320418618"
slug: "code-span-comment-opener-masks-to-eof"
severity: "minor"
category: "bug"
source: "impl-review"
found_during: "itd-177 third-round ruthless review, 2026-08-30"
found_at: "internal/core/intent/claims.go (opensComment, maskLines)"
---

opensComment reads a comment opener inside an inline code span as a real opener and the span runs to EOF through every later heading, so a draft whose prose quotes the opener idiom in backticks loses its Scope Conditions and Acceptance Criteria sections: ready reports no section, plan refuses on the AC discipline with a message naming the wrong fault, and the audit body hash would cover an empty AC section. Consult codeSpanRanges in opensComment and skip an opener inside a span; keep the to-EOF rule for a genuinely unclosed opener. Also cosmetic: the stamp trims a two-space hard line break from the bullet's first line.
