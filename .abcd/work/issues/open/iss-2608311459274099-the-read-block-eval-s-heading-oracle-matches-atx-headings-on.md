---
schema_version: 1
id: "iss-2608311459274099"
slug: "the-read-block-eval-s-heading-oracle-matches-atx-headings-on"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "itd-186 fidelity audit rcp-e3f354adbbee"
origin: researcher-authored
production_mode: hand-written
found_at: "evals/coldreading_oracle_test.go"
---

The read-block eval's heading oracle matches ATX headings only, so a setext-underlined or raw-HTML excluded heading arriving in a bundle item would satisfy the field-absence assertion completely. That state is unreachable today, but only because the assembler refuses it through the redaction verifier -- which is itself the mechanism no plant exercises and whose call can be deleted with the lane green. So both sides of one seam rest on a single unfalsified guard: the assembler's refusal is untested by the eval, and the eval's own oracle would not catch what that refusal is keeping out. Neither half is visible from the other, which is why three review rounds over the vacuity of this eval did not surface it and a fidelity audit did.
