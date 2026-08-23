# SOTA research protocol — capture and verb assessment

Dated 2026-08-22. Captures the research procedure run three times today —
producing
[`2026-08-22-context-window-management-sota.md`](2026-08-22-context-window-management-sota.md)
(two passes) and
[`2026-08-22-local-models-mlx-sota.md`](2026-08-22-local-models-mlx-sota.md)
(one pass) — and assesses whether conducting SOTA research needs its own
`/abcd:research` verb or is already covered by `/abcd:ideate`. Per
[`script-first-mvp`](../../principles/script-first-mvp.md), this documented
protocol is itself the MVP rung.

## The protocol, as run

1. **Parallel research pass.** Two independent read-only agents: an
   external web-research agent (primary sources preferred, evidence tiers
   demanded, "flag where no established practice exists" instructed
   explicitly) and a repo-survey agent (what the record already implements,
   plans, or has rejected on the question). Neither sees the other's
   output.
2. **Synthesis into a dated `*-sota.md` note** in the established genre:
   ranked techniques, an evidence-tier label on every load-bearing claim,
   a *Fit* paragraph per technique challenging it against this repo's
   conventions and shipped state (per
   [`prefer-sota`](../../principles/prefer-sota.md)), an
   investigated-and-rejected list, a gap list, and a sources list.
3. **Deterministic gates.** `abcd docs lint` and `make record-lint` on the
   note before any review.
4. **Independent adversarial review, report-only.** Fresh-context
   reviewers with *disjoint lenses*: one on source fidelity (re-fetch the
   cited sources, refute the numbers), one on repo fidelity and reasoning
   (verify every repo-facing claim against the files, hunt internal
   contradictions and tier inflation). Reviewers never edit; they return a
   verdict (SHIP / NEEDS_WORK / MAJOR_RETHINK) and numbered findings.
   Escalate reviewer count with the note's weight (one for a first pass,
   two once the note is load-bearing).
5. **Apply every finding, re-run the gates.** Unverifiable claims are
   findings, not passes: soften, source, or delete.
6. **Record the review inline.** The note itself carries a review record —
   reviewer lenses, verdicts, finding counts, disposition — because review
   output otherwise lives in gitignored ephemera, leaving a future session
   unable to tell a reviewed note from an unreviewed one. Without this
   field the drift named in the revisit condition below is undetectable.

## What the three runs measured

The reviewer yields are the argument for step 4, and for the disjoint-lens
design specifically. (Session testimony: these counts were observed
in-session and were not, at first writing, recorded anywhere the
repository can verify — the exact defect step 6 exists to close; each
named note now carries its yields in its own review record.)

- Round 1 (one reviewer, context-window note v1): 11 findings, 3 major —
  every in-repo claim verified, every falsified claim was an external
  number (a figure attributed to a benchmark that does not contain it, a
  threshold contradicted by its own citation, a range hardened beyond its
  source).
- Round 2 (two reviewers, context-window note v3): 7 source-fidelity plus
  7 repo/reasoning findings with no overlap between the lenses.
- Round 3 (two reviewers, local-models note): 12 source-fidelity plus 8
  repo/reasoning findings, again nearly disjoint — including one cited
  paper whose abstract states the opposite of the claim it anchored, and
  two cited issues that had closed (one the day before the note's date).

Three durable lessons: **research-agent output is not citable until
adversarially verified** (every round falsified at least one load-bearing
claim); **the two lenses catch disjoint failure classes** (redundant
same-lens reviewers would not); and **citations go stale in days**, so a
note's date is part of every claim it makes.

## Does `/abcd:ideate` cover this? No — different unit, different output

Ideate is an *idea-admission gauntlet*: one idea in; three ordered legs
(primary-source research of the idea's load-bearing claims, a grill
against the existing record, an adversarial review made of kill attempts);
one recorded verdict out (survives / killed / reframed), validated and
written by the binary with resolving citations.

SOTA research takes a *question* in and produces a *survey* out: ranked
techniques, tiers, fit challenges, gaps — no verdict, because nothing is
being admitted. (itd-104 itself names ideate's leg 1 "SOTA research", and
accurately so for claim-scoped verification — the distinction is scope and
output, not subject matter.) Forcing a survey through ideate would require inventing a
pseudo-idea and a verdict where none belongs; conversely, ideate's leg 1
is scoped to one idea's claims and cannot produce a survey. The expected
relationship is **feeder, not overlap**: a `*-sota.md` note is the
primary-source substrate a later ideate run consumes when a specific idea
("adopt X from the survey") is proposed for admission — an expectation,
not yet demonstrated practice: no ideate run in the record has consumed a
`*-sota.md` note as leg-1 substrate to date. The shared DNA — primary
sources, fresh-context adversaries (an extension, with itd-104 precedent,
of
[`evaluator-outside-the-loop`](../../principles/evaluator-outside-the-loop.md),
whose letter governs gate ownership rather than reviewer freshness), a
record either way — is a family resemblance, not coverage.

Adjacent surfaces already cover the rest: `/abcd:consult` and
`/abcd:ingest` own the sources-corpus and provenance side; research
*execution* is host-delegated (research agents and skills at the host
layer), matching the repo's host-delegated-by-default boundary; and the
`*-sota.md` genre is the established capture format — roughly thirteen
`*sota*` notes reaching back to `2026-07-06-docs-and-record-ia-sota.md`,
including today's two.

## Verdict on a `/abcd:research` verb: not yet

The only rung the binary could natively own — by analogy with
`abcd ideate record` — is a **validator**: `research record` checking a
composed survey (sources present and resolving, tier labels attached,
fit sections present, reviewer verdicts recorded) and writing the note
plus a decision-log pointer. That is a real candidate, but minting it now
fails `script-first-mvp` on the principle's actual trigger: contract
uncertainty. The schema the validator would enforce is still moving —
today's two notes each declared their *own* evidence-tier ladder, and that
divergence is calibration data for a future schema, not a defect to
freeze. Corpus size is deliberately *not* the argument: the genre corpus
is around thirteen `*sota*` notes reaching back to 2026-07-06, and
ideate's own validator was admitted after three manual protocol runs
(itd-104) — the repo's precedent says a settled contract, not a large
corpus, earns the rung.

**The documented protocol above is the current rung.** Revisit condition:
when hand-run drift appears — notes skipping adversarial review, tier
vocabularies diverging without declaration, citations landing unverified —
promote the validator through the ordinary admission path (an ideate run
on the verb idea, then an intent — the same path itd-104 took), with this
note and the in-tree review records (step 6) as its evidence. The
condition is detectable only because step 6 puts review outcomes in the
tree; without it, a skipped review leaves no trace and "not yet" would
harden into "never".

## Review record

2026-08-23 — the proposal itself adversarially reviewed: two fresh-context
reviewers, disjoint lenses, authorship stripped per ideate's leg-3
discipline. Kill-attempt reviewer: REFRAME (five attempts — two survived,
three partial: thin-corpus reasoning not grounded in `script-first-mvp`,
revisit condition undetectable without in-tree review records, session
yields pre-registered as evidence they could not become). Record-grill
reviewer: two major (corpus count wrong by ~2.5×; feeder claim stated as
fact with no confirming instance), one unverifiable (the yields), three
minor. Convergent repairs applied in this revision: thin-corpus argument
withdrawn (deferral rests on contract uncertainty alone), corpus count
corrected, yields marked session testimony, feeder downgraded to
expectation, step 6 added, `evaluator-outside-the-loop` citation
tightened. The no-verb verdict itself survived both reviews.
