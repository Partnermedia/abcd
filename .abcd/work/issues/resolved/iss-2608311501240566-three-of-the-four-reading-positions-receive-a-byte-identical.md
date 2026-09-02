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
resolution: "The include table gains a per-position section: rows that replace the shared pile at one position, absent means the shared pile. The default stays one shared assembly, because the readers' definitions already tell each position what to attend to and a shared pile keeps the four readings comparable; a position can now be handed only its own object when a run needs it. itd-194's validation binds a per-position list through the same validator, and every manifest carries a pile stamp (shared or own, with the pile's hash) so a run is reproducible and the closing-run comparison can tell the two apart. The comparative position's natural own pile — the widening reading's admitted output — stays a documented example in the charter: it sits inside a record family the table may not name and inside the .abcd deny, and the assembler has no channel for a prior run's output, so that position goes on refusing to assemble (itd-199)."
impact: additive
resolved_by:
  commit: "a8f345b4"
---

Three of the four reading positions receive a byte-identical item set. Comparing the manifests of a real assembly at one commit, widening, comparative and detection each carry the same 951 items with the same fields, and only entailment differs, by the 346 draft and planned items the drafts asymmetry admits. That asymmetry is itd-183's and it works exactly as designed. The other three do not. The definitions state four distinct objects: widening reads the brief, glossary, disciplines, specs and shipped tree; comparative reads the candidate set, which is the widening reading's pre-admission output, against the declared selection criteria; detection reads the shipped tree against the claim record. What the assembler actually passes at comparative and at detection is the widening object unchanged. The comparative case is the sharper one, because its object is not a subset of the repository at all but the output of a previous reading, and the assembler has no channel for it -- spc-64 says as much in passing, that the in-cycle candidate set arrives by whatever channel spc-61 defines and the eval makes no claim about it, but nothing states that the position therefore receives the wrong object. No criterion measures cross-position differentiation, and the evals run on a fixture too small for the difference to show.

## Grounds

- pursued: the expectation is that a shared pile is the right default and that a per-position include list is enough configuration to hand a position its own object; the falsifier is a run where a position needs its own object and the include table cannot express it — the comparative position is that case today, and it is documented rather than solved.
