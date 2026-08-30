---
schema_version: 1
id: "iss-2608300402591658"
slug: "entity-decoding-order-is-nondeterministic"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "itd-183 fifth-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (renderedText)"
---

renderedText decodes entities by iterating a Go map, so the decoding order is random and the refusal verdict for a heading such as Audit&amp;nbsp;Notes is nondeterministic (200 calls: refused 77, admitted 123) — a determinism instrument with a coin-flip refusal; and a numeric or hex character reference inside an excluded title is outside the short list, slugs differently, and travels, while the comment claims an entity outside the list is a refusal not a leak. Replace the hand list with html.UnescapeString before slugging, which closes both, and correct the comment.
