# Fix the detector

**The rule.** After a review, the unit of fix is not the individual finding
but the *detector* — the gate, lint rule, or test convention that would have
caught the finding's whole class. Each captured issue records the class, the
proposed detector, and the found instances as the detector's acceptance
corpus: a detector is proven when it flags every instance it was built from.
Instances are then drained by a cleanup pass *behind* the armed detector,
never hand-fixed ahead of it.

**Why.** Hand-fixing instances treats symptoms at the cost of the cause and
leaves nothing behind to stop recurrence — the 2026-07-08 review demonstrated
this directly, finding drift that had survived a dedicated same-day
consistency pass because the pass fixed instances without arming anything.
Inverting the order pays twice: the review's findings stop being a chore list
and become a free, real-world test fixture for the enforcement layer, and the
cleanup itself becomes verifiable (run the detector; the count reaches zero
and stays there).

**Bounds.**

- Genuine one-off defects with no class (a single wrong constant) are fixed
  directly; the rule binds where two or more findings share a root cause.
- The detector need not be fully mechanical on day one — a review-blocking
  checklist rule is a valid detector — but the promotion path to a mechanical
  gate is named in the issue.
- Composes with [ratchet-not-big-bang](ratchet-not-big-bang.md): the detector
  arms immediately against the found instances as its baseline, and with
  [retire-the-name](retire-the-name.md): a banlist entry is the smallest
  detector.

**Live instance.** The 2026-07-08 multi-agent review's findings are recorded
as clustered ledger issues, each naming its detector and carrying its
instances as the acceptance corpus.

**Promotion.** A capture convention (or lint) requiring review-sourced issues
to name a detector would make this a discipline.
