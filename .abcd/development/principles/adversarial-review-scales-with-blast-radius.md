# Adversarial review scales with blast radius

**The rule.** An artefact that changes what gets built crosses its gate only
after independent adversarial review — two reviewers with different lenses
for the major crossings: an intent leaving `drafts/` for `planned/`, an ADR
moving to `accepted`, a plan before execution begins. Artefacts below that
line — ledger captures, comments, routine pull requests — are never blocked
on adversarial review: they are covered by the gates they already have, and
capture in particular must stay frictionless.

**Why.** Both halves are load-bearing, and both are empirical. The 2026-08-19
itd-92 extension went to two independent adversarial reviewers before its
planning interview; both returned NEEDS_WORK, and the findings — a verdict
vocabulary undecidable by the probes it specified, a gate promised with no
acceptance criterion, two trust rules parked in intent prose — would each
have surfaced as implementation failures instead
([the calibration entry](../research/notes/2026-08-15-decomposition-calibration.md)
records the overturn). The floor exists for the same reason the ceiling does:
a blanket review-everything rule would suppress the ledger (recording a
finding must cost less than fixing it, or findings go unrecorded) and would
train the skim that adr-42's warn-storm STOP names — reviewers who read two
mandatory verdicts on trivia skim the one that matters.

**Bounds.**

- "Independent" is [evaluator-outside-the-loop](evaluator-outside-the-loop.md):
  the reviewers did not produce the artefact, and for the two-reviewer
  crossings their lenses differ (design/feasibility vs record-discipline is
  the proven pair).
- The review is a proposal, never the gate itself
  ([verifier-selects-gates-decide](verifier-selects-gates-decide.md)): the
  human adopts, overrides, or rejects findings — but an unreviewed major
  crossing is a skipped stage, and per [loud-staging](loud-staging.md) a
  skipped stage says so where the crossing is recorded.
- Enforcement is the documented protocol for now (script-first); the
  armed-detector rung — the readiness checks refusing a `drafts/ → planned/`
  move without review receipts in the existing `.abcd/work/reviews/` shape —
  is a recorded seed, built once this protocol has real runs behind it.
