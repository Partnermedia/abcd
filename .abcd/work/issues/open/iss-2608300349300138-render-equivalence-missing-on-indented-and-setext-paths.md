---
schema_version: 1
id: "iss-2608300349300138"
slug: "render-equivalence-missing-on-indented-and-setext-paths"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 fourth-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (verifyRedaction indented and setext loops)"
---

The render-equivalence comparison added for column-0 headings does not reach the indented-ATX and setext refusal loops, which compare by case-fold only, so an indented or underlined heading spelled with bold, a code span, emphasis or a non-breaking space travels while the manifest asserts it was refused — the class the last two commits claim to close, closed on one of three paths. Use the same equality (fold or same rendering) in both loops and add the eight probe spellings to the equivalence test.
