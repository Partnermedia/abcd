# Less, but better

**The rule.** Reach for the subtraction first. Before adding a field, a flag, a
gate, or a judgement the user has to make, ask what could be removed instead so
the question never arises. A design that demands a decision at every use has
moved work onto whoever uses it; a design that needs a special case has admitted
that the general case differs from the one in hand.

Two questions decide most of it:

- **Which case am I designing for?** Say it out loud. Designing for the mess in
  front of you produces a mechanism that outlives the mess and becomes permanent
  architecture nobody can justify.
- **What does this let me delete?** An addition that removes two other things is
  a subtraction. An addition that only guards an existing addition is a third
  copy of the problem.

The phrase is Dieter Rams's — *Weniger, aber besser*. The "better" is doing the
work: this is not minimalism, and not a licence to ship less capability. Fewer
knobs, each of which is right, beats a knob for every case.

**Why.** The expensive failures in this repository were not missing features.
They were mechanisms that each looked reasonable and together made a claim
impossible to keep true.

`resolved_by.commit` is the clearest one. It named the commit that fixed an
issue, and its value could not exist while the record was being written — the
record and the fix are the same change. So stamping became a second act after the
merge, and a second act has no forcing function: 363 records in `resolved/`
against 89 carrying a stamp. The fix was not a better stamping workflow. It was
making the field optional and letting resolution ride the fixing commit, which
deleted the ordering problem instead of automating around it. See
[the-record-lands-with-the-act](the-record-lands-with-the-act.md).

The changelog's `impact: internal` carries the same shape. The field decides the
version bump, which it must; it also silences a record's changelog line, which is
a second job bolted onto one value. That second job turns a judgement about
product impact into a judgement about publication, and `checkIssueImpact`'s own
comment names the consequence: a rule that refused `internal` would "either force
that work into a user-facing changelog or push authors into a mislabel".

**Bounds.**

- Subtraction is not deletion of capability. Removing a gate that catches real
  defects is not "less, but better", it is less. The test is whether the removed
  thing was answering a question the design should never have asked.
- A special case for a migration is legitimate, and must SAY it is one. Name it
  in the mechanism itself — the field's own documentation, not a note elsewhere —
  or it becomes permanent by default and the next author builds on it.
- Bundling is not curation. One changelog line citing five records that were one
  user-visible change is fewer lines carrying the same information; that is the
  "better" half. Dropping a record because someone judged it dull is not.
- Composes with [one-canonical-primitive](one-canonical-primitive.md) (never a
  third copy) and [script-first-mvp](script-first-mvp.md) (earn each rung). Those
  two say *do not add yet*; this one says *check what adding lets you remove*.

**Live instance.** The release derivation. The cut is a set-difference of folder
membership, read by every consumer as "what shipped in this release". Those
coincide only while resolution lands in the fixing commit — which is now the
rule, so in a repository abcd manages from the start they coincide by
construction.

They did not coincide while abcd was being built. A 2026-08-24 hygiene sweep
closed 33 records for work released as far back as v0.2.0, and the v0.6.4 cut
derived its version from four of them. `shipped_in` (iss-2608241612087533) is the
accommodation for that history, and it is a MIGRATION mechanism: a repository
managed by abcd from its first commit should never need it, and its own
documentation says so rather than leaving a future reader to infer it.

The design decision that followed is the subtraction: every record that reaches a
terminal folder earns its changelog line, with no inclusion judgement at all. The
rationale already lives in the record; the changelog is the index and the trace
from record to version. That deletes the inclusion decision, deletes `impact`'s
second job, and deletes the reason `shipped_in` exists — at the price of a longer
changelog, which is the honest trade.

**Promotion.** Unenforced, and probably unenforceable — "could this have been
removed instead" is a question for review, not a lint. The rung above it is
review discipline: a change that adds a field, a flag or a gate says in its own
description what it lets the design delete, or says plainly that it deletes
nothing. A reviewer who cannot find that sentence has found the finding.
