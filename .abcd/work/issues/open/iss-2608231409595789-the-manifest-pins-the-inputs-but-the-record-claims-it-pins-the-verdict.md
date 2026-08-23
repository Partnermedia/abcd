---
schema_version: 1
id: "iss-2608231409595789"
slug: "the-manifest-pins-the-inputs-but-the-record-claims-it-pins-the-verdict"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "v0.6.2 release-gate double run 2026-08-23"
found_at: ".abcd/development/release-gate/manifest.json"
---

`.abcd/development/release-gate/manifest.json` describes itself as "the
reproducibility anchor (iss-122): two honest runs of the same tier mean the same
thing because the doc list, the directions, the checker count, and the prompt are
fixed here rather than chosen per run".

Two full-tier runs on 2026-08-23, same manifest, same 28 checkers, trees
differing only by seven shipped-doc fixes that touch none of the brief:

| | run 1 (7a4ee00) | run 2 (a2c77e5) |
| --- | --- | --- |
| total | 125 | 126 |
| false-claim | 58 | 48 |
| undocumented-surface | 38 | 50 |
| stale-count | 15 | 18 |
| criterion-violation | 7 | 3 |
| fictional-layout | 7 | 7 |

Per chapter the swing is larger than the totals suggest: `01-ahoy.md` 8 to 12,
`04-launch.md` 3 to 7, `19-identity.md` 1 to 4, `09-reflect.md` 6 to 3,
`10-docs.md` 3 to 1.

What IS stable is the location: 29 of 33 flagged files appear in both runs, 4 in
run 1 only and 3 in run 2 only. So the detector agrees on WHICH chapters have
drifted and disagrees on HOW MANY discrepancies each holds and WHAT CLASS each
one is.

The manifest is not wrong to exist and it is not lying about what it pins. It
pins the inputs, and it does: scope, depth, doc list, prompt. What it cannot pin
is the output, because the checkers are LLM agents. The defect is the sentence
claiming the consequence — "two honest runs of the same tier mean the same
thing" — which is a statement about outputs that the mechanism cannot deliver.

The consequence is practical, not theoretical:

1. A receipt's `failing` count is not a measurement. Comparing 37 (v0.6.1) to
   125 cannot support "drift got worse", and a later run showing 90 would not
   evidence improvement.
2. The class distribution cannot drive scope decisions. An intent that scoped
   itself to "the 58 false-claims" would have scoped itself to 48 a few hours
   later.
3. The stable half is still useful, and is the half to build on: which chapters
   drift is reproducible, and that is what locates a generation seam.

Fix is a record correction, not a mechanism change: the manifest's comment
should claim what it delivers — comparable SCOPE, not comparable findings — and
the runbook should say a receipt's finding count is an observation from one run
rather than a metric. Refines iss-2608231346137587, whose corpus this governs;
that issue's own class counts carry the same caveat.

This is the enforcement-claims-are-facts class again (iss-2608230847432286): a
record stating an assurance the machinery underneath it does not provide. It was
found only because a release recovery happened to run the same gate twice in one
afternoon, which is not something the process would otherwise ever do.