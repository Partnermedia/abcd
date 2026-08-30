# Review-round cost calibration

**Date:** 2026-08-30
**Status:** evidence in; the stance it feeds is
[adversarial-review-scales-with-blast-radius](../../principles/adversarial-review-scales-with-blast-radius.md).

## What was run, and why it was run that way

Cycle 1 of the cold-reading workstream ran the maximal review discipline
deliberately: every build round on every branch got two independent adversarial
reviews, fresh contexts, over the whole branch diff, repeated until both came
back clean. That is more verification than the principle asks for, and it was
chosen as a measurement rather than as a standard. The question was what the
ceiling costs and which part of it earns its keep, which cannot be answered by
reasoning about it — [prefer-the-experiment-to-the-inference](../../principles/prefer-the-experiment-to-the-inference.md).

The evidence below is the answer. It is recorded so the next phase, and any
abcd-managed repository, inherits the calibration rather than repeating the
experiment.

## What the intents cost

Merged, measured across each build branch's merge into the integration branch:

| intent | commits | round-tagged | diff |
|---|---|---|---|
| itd-181 | 5 | 0 | +1,091 |
| itd-178 | 11 | 1 | +1,948 |
| itd-182 | 14 | 0 | +686 |
| itd-188 | 15 | 4 | +1,089 |
| itd-177 | 22 | 3 | +3,143 |
| itd-180 | 29 | 6 | +5,469 |
| itd-183 | 47 | 15 | +7,275 |

In flight at the time of writing, and already past itd-183 on
commits-per-line without having merged: itd-179 at 36 commits and 25
round-tagged for +4,648, itd-189 at 42 commits for +4,590.

## What the reviews cost

Eight subagents completed in one orchestrating session, totalling
approximately 1.28M subagent tokens. Six of the eight were review or
review-driven. A review PAIR costs roughly 350-400k tokens. itd-189 had five
pairs.

The dominant cost of this cycle is not writing the code and not writing the
record. It is re-reading the same diff.

## The three findings

**1. The rounds kept finding real defects, so depth is not the waste.** Round 5
returned ten findings across the two branches, six of them branch-introduced
and one of them landing on a claim inside an already-resolved record. There is
no evidence here for reviewing less deeply.

**2. The re-reading is the waste.** Each round was briefed as
`<integration>...HEAD`, so itd-189's fifth pair re-read 42 commits of which
about 35 had been reviewed four times already. Nothing in the protocol asked
for that; it was how the briefs were written. Scoping a round to the delta
since the last clean review removes most of the cost without removing a single
question the reviewer asks about new code.

**3. Nits are re-review triggers, and that is where the money goes.** itd-179's
round 5 returned one FIX-FIRST, one SHOULD-FIX and five nits -- a misattached
godoc, a wrong noun in a detail string, a comment describing a Unicode table
loosely. Fixing them opens a fix round, which under the maximal discipline
opens another review pair: roughly 400k tokens spent to settle five sentences.
The nits were all correct. Their cost was in when they were paid, not in
whether they were worth paying.

## What follows

Three changes, adopted for the remainder of this phase and recorded as the
principle's build-round half:

- A build round is reviewed over the DELTA since its last clean review, never
  the whole branch. The full-branch pass happens once, at the integration
  review, where it is the point.
- A nit does not open a review round. It is captured and settled at the ship
  commit. Only a FIX-FIRST or a BLOCK re-opens the pair.
- The pair is for a trust boundary. A round that touches input parsing, auth,
  secrets, network, subprocess or file/DB access gets both lenses; a round that
  does not gets one. This is the repository's existing rule for changes,
  applied to rounds, which is where it had never been applied.

Applied to the four remaining intents (itd-184, itd-185, itd-186, itd-187) the
expected saving is 50-60% of review spend at unchanged correctness bar, because
every question about new code still gets asked, once.

## What this does not settle

The recurring defect class of this cycle -- a message, comment, spec sentence
or record claim asserting a mechanism that is not the case -- accounted for
five of round 5's ten findings and appeared across six different surfaces. It
is semantic, so no lint rule catches it, and it is the one class where more
review rounds genuinely did keep paying. Whether a discipline record reduces
its rate is the next thing worth measuring, and it is not measured here.

The token figures are from one session's completed subagents. They exclude the
session lost to a machine crash earlier the same day, so the cycle's true spend
is higher than the number above and is not recoverable. That irrecoverability is
itself a gap in the record rather than an accident of this cycle: nothing here
meters development acts, and the near-miss records estimate forward without ever
measuring. Parked as
[the token metering gap](2026-08-30-development-token-metering-gap.md) for the
next phase to file or reject.
