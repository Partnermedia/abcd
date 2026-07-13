# Recurrence is signal

**The rule.** A finding that reappears after it was dispositioned is evidence
about the disposition, not a duplicate of the finding. A detector re-raises a
tension it has surfaced before whenever that tension is still present, and the
ledger escalates the recurrence rather than folding it back into the closed
entry. Suppressing a repeat because "we covered this before" destroys the only
signal that tells you the resolution was weak or false.

**Why.** A closed finding carries an implicit claim: *this was dealt with.* The
cheapest available test of that claim is whether the condition that produced the
finding is still there. If a blind detector — one that cannot see the ledger and
therefore cannot be influenced by the fact of closure — surfaces the same tension
again, the closure is contradicted by evidence, and that contradiction is more
informative than the original detection was. The failure mode this guards against
is a ledger that grows quieter as it grows less accurate: every dedupe silently
converts "we were wrong to close this" into "we already know about this". The
ABCD sensemaking method makes the same point from the detector side and states it
as a prohibition on the reader — *be deliberately amnesiac; re-raise previously
surfaced tensions* — because a tension that keeps re-surfacing after it was
supposedly resolved is the signal that the resolution was false
([the method](../research/notes/2026-07-13-abcd-sensemaking-method.md)).

**Bounds.**

- **Recurrence is not reopening.** The escalation is a claim that the closure is
  in question, not an automatic reversal of it. The human's disposition remains
  the gate — a recurrence can itself be dispositioned "still wontfix, and here is
  why", which is a *stronger* record than the original wontfix, because it was
  made against evidence of persistence.
- **This is not a licence to re-raise noise.** The rule earns its keep only where
  the detector is deterministic or blind. A detector that re-raises because it
  forgot, rather than because the condition persists, is producing duplicates and
  the rule does not protect it.
- **A recurrence after a `resolve` is a different fact from a recurrence after a
  `wontfix`.** The first says the fix did not hold (or was never made); the
  second says the accepted cost is still being paid. Both are worth surfacing;
  they are not the same signal and should not be collapsed into one.
- Composes with [`verifier-selects-gates-decide`](verifier-selects-gates-decide.md):
  the escalation is a proposal. It composes with
  [`evaluator-outside-the-loop`](evaluator-outside-the-loop.md) in the strong
  form — a detector that *can* read the ledger it feeds will learn to fall silent
  about closed entries, which is precisely the gaming this rule exists to prevent.

**Promotion.** The MVP is a convention: when a finding recurs, file it as a new
finding citing the closed one, rather than reopening or deduping — the citation is
the escalation. The tool is the ledger doing this itself: `abcd capture` matching
an incoming finding against dispositioned entries and escalating on a hit instead
of deduping, which is [itd-87](../intents/drafts/itd-87-recurrence-escalation-in-capture.md).
As a maintained detector it is demotable to advisory if its matching produces
false recurrences, and archivable to regression-only if closures stop recurring at
all — the saturation case, which is the outcome the rule wants.
