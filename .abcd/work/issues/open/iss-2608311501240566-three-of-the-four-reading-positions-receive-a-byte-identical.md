---
schema_version: 1
id: "iss-2608311501240566"
slug: "three-of-the-four-reading-positions-receive-a-byte-identical"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "end-to-end reading rehearsal, cold-reading cycle 1"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/include.go"
---

Three of the four reading positions receive a byte-identical item set. Comparing the manifests of a real assembly at one commit, widening, comparative and detection each carry the same 951 items with the same fields, and only entailment differs, by the 346 draft and planned items the drafts asymmetry admits. That asymmetry is itd-183's and it works exactly as designed. The other three do not. The definitions state four distinct objects: widening reads the brief, glossary, disciplines, specs and shipped tree; comparative reads the candidate set, which is the widening reading's pre-admission output, against the declared selection criteria; detection reads the shipped tree against the claim record. What the assembler actually passes at comparative and at detection is the widening object unchanged. The comparative case is the sharper one, because its object is not a subset of the repository at all but the output of a previous reading, and the assembler has no channel for it -- spc-64 says as much in passing, that the in-cycle candidate set arrives by whatever channel spc-61 defines and the eval makes no claim about it, but nothing states that the position therefore receives the wrong object. No criterion measures cross-position differentiation, and the evals run on a fixture too small for the difference to show.
