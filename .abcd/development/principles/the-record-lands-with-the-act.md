# The record lands with the act

**The rule.** A record about a change is written by the same commit that makes
the change true. When completing that record depends on a fact which does not
exist until after the act — a merge sha, a release tag, a published URL — the
dependency is the defect, not an inconvenience to schedule around. Drop the
field, make it optional, or derive it downstream; never let it hold the record
open. A step scheduled after the act has no forcing function, because the work
it records is already done and already rewarded, so it is the step that gets
skipped.

The failure is silent by construction. An unwritten record leaves nothing behind
to find it by, so the omission has no denominator and nobody can count what was
never filed. That is what separates this from ordinary debt, which is at least
visible in the place it accumulates.

**Why.** Measured in this repository on 2026-08-24: 363 records in
`issues/resolved/` against 89 carrying a `resolved_by` stamp, and 214 records
still in `open/`, of which a single triage pass found 33 already fixed — some
of them fixed as far back as v0.4.0 and sitting open across four minor
releases.

One field caused it. `resolved_by.commit` names the commit that fixes an issue,
and at the moment the record is edited that sha does not exist, because the
record and the fix are the same change. Stamping therefore had to be a second
act after the merge, and the second act was performed for roughly a quarter of
the records that needed it.

The sampled evidence rules out inattention as the explanation. Of the ten
commits that fixed the swept records, every one names its issue id in the commit
message and not one moves the record out of `open/`. The intent to record was
present in every case. Only the opportunity was missing, and intent without
opportunity produces nothing.

**Bounds.**

- A fact that genuinely postdates the act, such as which release carried a fix,
  is recorded by derivation rather than by a promised later edit. Where nothing
  derives it, it goes unrecorded rather than pending.
- Optional-but-verified is the shape that works. The field stays optional so it
  can never block the record, and a gate checks it whenever it is present.
- The rule binds bookkeeping that describes an act. It does not bind genuinely
  staged work, where a later phase is the plan rather than an omission.
- Composes with [fix-the-detector](fix-the-detector.md): The deferred step is
  the finding class, and the detector is whatever makes an absent record loud.
- Distinct from [loud-staging](loud-staging.md), which governs a stage that
  degrades or no-ops. Nothing degrades here; the record simply never appears.

**Live instance.** iss-2608241347321757 dissolves the dependency. `make
lint-issues` (RS001) requires a commit carrying a `Resolves: iss-N` trailer to
move that record out of `open/` in the same diff, so resolution lands inside the
fixing commit and no post-merge step exists to forget. `resolved_by.commit`
becomes optional-but-verified: RS002 checks that a stamp added in a range names
a reachable commit, and RS003 checks that every stamp already in the ledger
stays reachable under squash and rebase merges.

The residual gap is recorded as iss-2608241612007530. RS001 is triggered by the
trailer, so a fix that lands with no trailer at all is invisible to every one of
the three rules. iss-202 is the standing example: Its fix merged on 2026-08-24
in a commit ending with a bare `iss-202` line rather than the trailer form, and
its record sits in `open/` with the fix shipped.

**Promotion.** The enabling convention is the `Resolves:` trailer plus the
same-diff ledger move. The armed rung is `make lint-issues`, a CI step and a
pre-push gate. The rung this principle still lacks is a detector for the
inverse direction — a fix that names no issue at all — which is what
iss-2608241612007530 carries, and which promotes this from a rule that catches
dishonest records to one that catches missing ones.
