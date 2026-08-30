---
schema_version: 1
id: "iss-2608300232312077"
slug: "itd-188-fourth-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-188 fourth-round ruthless review, 2026-08-30"
found_at: ".abcd/development/specs/open/spc-66, agents/scribe.md, agents/scribe/fixtures/injection-canary.json, internal/core/lint/scribecontract_test.go"
---

itd-188 fourth-round residue: spc-66's Tests section still says the guard folds every separator spelling to a slash, while the guard now folds only the ASCII reverse solidus and spaced separators and refuses every other non-ASCII spelling — a maintainer trusting the spec could allow a lookalike in prose believing it still folds; the definition's untrusted-input banner says an embedded instruction is a thing a fidelity flag reports while the flag section requires two disagreeing pieces of material, and the canary's behaviour text names a different candidate flag than its exemplar emits; the residual-entity regex refuses any ampersand followed by a letter (Q&A) without the comment saying so; the bypass table asserts only that some finding exists, so the Cf branch's message is pinned by one case.
