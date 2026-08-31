---
schema_version: 1
id: "iss-2608311632439831"
slug: "spc-64-s-positions-exercised-paragraph-is-stale-against-the"
severity: "minor"
category: "documentation"
source: "impl-review"
found_during: "reviewing the evals audit fix round"
origin: researcher-authored
production_mode: hand-written
found_at: ".abcd/development/specs/closed/spc-64-the-read-block-eval-falsifies-the-firewall-planted-warm-cont.md"
---

spc-64's Positions exercised paragraph is stale against the eval it specifies. It says widening, entailment and detection are asserted in full and that at the comparative position the eval asserts only the read-block over the instrument's own stored prior outputs, making no claim about the in-cycle candidate set. The delivered eval asserts every sentinel class at every position, which the first review round required after finding that carving comparative out made an assembler leak at that one position invisible. The spec therefore understates its own delivery, and understating is not harmless here: a reader deciding what the eval covers would conclude that a comparative-only leak goes unnoticed, which is exactly the defect the round closed.
