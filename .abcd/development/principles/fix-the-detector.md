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

**Lifecycle.** A detector is itself debt: the moment it arms it becomes a
maintained artifact with its own failure modes, and the record treats it as
such.

- Each detector carries a false-positive budget and a kill criterion,
  recorded beside its acceptance corpus. Flags dismissed above roughly one
  in ten kill trust in the detector wholesale; a detector that stays green
  while its finding class recurs has a wrong definition, not a clean
  codebase; a detector whose guarded class has disappeared is pruned.
- Saturation triggers succession review. A violation stream from adaptive
  producers that reaches zero and stays there means either the class is
  extinct or the producers have learned to satisfy the letter of the check —
  either way the response is deliberate: harden the criterion, layer a
  stricter one, or archive the detector as regression-only. The Red Queen
  Gödel Machine (Iacob et al., 2026) demonstrates the stagnation of a frozen
  evaluator empirically; the GLM-5 report (2026) shows the corpus-decay
  side — retire fixtures nothing current fails, and mine fresh cases
  continuously.
- Succession is anchored, never argued. Challenger and incumbent are scored
  on the same fixed, detector-independent anchor corpus; the challenger is
  promoted only when it strictly beats the incumbent there, and ties keep
  the incumbent. A detector is never replaced or loosened on argument alone.
- Escapes feed the successor. A defect that slips past an armed detector
  lands as a fixture in the same change as its fix, and a displaced
  detector's demonstrated blind spots become adversarial fixtures in its
  successor's corpus. The acceptance corpus is a harvest, not a birth-time
  snapshot.
- Judgment-shaped detectors need held-out proof. The arming test above —
  flags every instance it was built from — fully proves only a detector
  that is specified by its instances, such as a banlist entry. For a
  detector that generalises (a rubric, a checklist, a judge oracle),
  founding-instance recall is training evidence, not proof: adequacy is
  demonstrated on held-out instances the detector never saw, and the
  evidence its builder iterates against stays disjoint from the evidence
  that accepts it.
- The criterion freezes within a pass and changes only between passes. One
  frozen criterion scores everything inside a review, audit, or cleanup
  pass, so the pass's findings are comparable evidence; detector changes
  batch at pass boundaries, and the re-audit a change forces is lazy —
  items re-score as work returns to them — never an immediate whole-corpus
  big bang, composing with [ratchet-not-big-bang](ratchet-not-big-bang.md).
- Persisted verdicts name their detector version. A baseline entry,
  suppression, or "passes gate X" claim records which detector produced it;
  when the detector changes, dependent verdicts are erased and recomputed
  under the new version, never rescaled or carried forward. The raw
  instances are durable and detector-independent, so recomputation is cheap
  re-scoring, not redone work.

**Live instance.** The 2026-07-08 multi-agent review's findings are recorded
as clustered ledger issues, each naming its detector and carrying its
instances as the acceptance corpus.

**Promotion.** A capture convention (or lint) requiring review-sourced issues
to name a detector would make this a discipline. The lifecycle promotes the
same way: the MVP is per-detector metadata (false-positive budget, kill
criterion, per-fixture last-failed date) recorded beside the acceptance
corpus; the tool is a gate that reruns the anchor corpus on any change to a
detector file and blocks on regression, plus a runner that records the
detector version per run and refuses to aggregate mixed-version findings
into one baseline.
