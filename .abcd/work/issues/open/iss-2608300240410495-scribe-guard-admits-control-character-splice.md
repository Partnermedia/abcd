---
schema_version: 1
id: "iss-2608300240410495"
slug: "scribe-guard-admits-control-character-splice"
severity: "minor"
category: "security"
source: "impl-review"
found_during: "itd-188 fifth-round security review, 2026-08-30"
found_at: "internal/core/lint/scribecontract_test.go (code-point loop)"
---

The scribe guard's code-point loop refuses format code points and every non-ASCII code point but not ASCII control characters, so a carriage return, vertical tab, form feed, NUL, escape or DEL splits a path match the way a format code point would; a trailing traversal on the far side of the control character stands alone, produces no finding, and a viewer that renders the control invisibly shows .abcd/work/issues/.. as allowed material. Refuse unicode.IsControl outside LF and TAB with its own class marker; the test should also assert the canary exemplar omits the control shape.
