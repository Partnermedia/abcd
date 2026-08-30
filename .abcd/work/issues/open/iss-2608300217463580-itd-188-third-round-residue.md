---
schema_version: 1
id: "iss-2608300217463580"
slug: "itd-188-third-round-residue"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-188 third-round adversarial reviews, 2026-08-30"
found_at: "internal/core/lint/scribecontract_test.go, agents/scribe/fixtures/injection-canary.json, agents/scribe.md"
---

itd-188 third-round residue, all in the guard's fold step and the canary exemplar: a single non-hex percent sign anywhere in the definition makes url.PathUnescape fail and the whole percent-decoding step is skipped for the file, contrary to its comment; decoding runs once, so double-encoded separators survive; the separator replacer enumerates forward solidus lookalikes but not their reverse twins (U+2572, U+29F9, U+2216, U+29F5) nor eleven further lookalikes, while its comment claims exhaustiveness; a separator-free shipped-tree filename is outside the guard's reach while spc-66 says any repository path is refused; the canary exemplar emits a reading record for item-1 only against its own one-record-per-item behaviour; the definition's Delivery line 'and nothing else' contradicts the reported outstanding state and the canary's refused entry; two British spellings sit in code-side comments. Remedy is class-closing: refuse every non-ASCII code point outside a small typographic allow-list, decode tolerant and to a fixpoint, refuse residual encoded shapes.
