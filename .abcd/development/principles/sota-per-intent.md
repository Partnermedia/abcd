# SOTA Per Intent

**The rule.** Every intent declares the **current state of the art** for the
capability it promises: the alternatives that already exist and each one's rough
maturity. From that declaration the intent recommends exactly one of three
paths, and each path has its own gate:

1. **Build on the SOTA alternative** — adopt the existing tool/library. Because
   this adds a dependency, it is a **hard stop**: the user must approve before
   adoption. (This is the standing new-dependency gate, now reached *through*
   the SOTA declaration rather than at `go get` time.)
2. **Build a basic native version, seam for SOTA later** — a minimal internal
   solution that achieves the intent now, **designed so the SOTA alternative can
   be plugged in later** behind a seam. This path **proceeds without a stop** —
   its licence to skip approval is precisely that it keeps adoption open. The
   seam must be real, not asserted.
3. **Bespoke, no swap possible** — the SOTA alternative cannot be used now **and**
   cannot be plugged in later either (no viable seam). This is a **hard stop**:
   the user reviews to decide *how much complexity* the permanent internal
   solution should carry. A merely *deferred* seam is still path 2; path 3 is
   only for a genuine commitment to a permanent bespoke solution.

The load-bearing idea: **the SOTA alternative always stays on the table.** A
native build earns its way past the approval gate by keeping the option to adopt
open (path 2); the only way to close that option (path 3) is a decision the user
makes deliberately, having chosen the complexity level.

**Why.** [prefer-sota](prefer-sota.md) makes SOTA the default bar and supplies
the fit-challenge that stops "it's SOTA" from laundering a preference violation.
This principle is what that verdict *produces at the intent boundary*: a written,
per-intent record of where the field is, and a decision that is honest about
which of the three postures the intent takes. It closes two failure modes at
once. Silent over-adoption — pulling in a heavy dependency without the user
weighing it — is blocked by gate 1. Silent lock-in — building a bespoke thing
that quietly forecloses ever adopting the better external tool — is blocked by
gate 3; the default, path 2, is abcd's house pattern (a native floor with an
easy opt-in to a superior backend, per [adr-22](../decisions) /
[adr-26](../decisions)), and it is allowed to proceed *because* the seam keeps
the field's best answer reachable. The declaration also ages well: maturity is
rough and dated, so a path-2 intent can be revisited when the SOTA alternative
matures enough to be worth adopting.

**Bounds.**

- The SOTA declaration is per intent and rough — existing alternatives + a
  coarse maturity read (e.g. experimental / usable / mature / de-facto
  standard), not a survey. It is the input to the path choice, not a deliverable
  of its own.
- Path 2 is the default and the only stop-free path, and only while the seam is
  genuine. If, during build, the seam turns out not to be viable, the intent has
  become path 3 — stop and take the complexity decision to the user; never
  quietly harden a bespoke solution that can no longer adopt SOTA.
- Gate 1 and the standing "new dependency ⇒ ask first" rule are the same gate;
  this principle routes the dependency decision through the SOTA declaration so
  the user approves *with the alternatives and maturities in front of them*.
- A path-2 seam is a design constraint, not a promise to build the adapter now:
  the adapter lands only when a phase consumes it ([wired or it isn't done]),
  but the interface must not foreclose it.

**Scope.** A propagated default, like [prefer-sota](prefer-sota.md): it holds for
every abcd-managed repo as one of the opinionated conventions abcd installs, and
its home in a managed repo is the marked working-conventions section abcd writes
into `AGENTS.md`. abcd's own record carries the normative statement here.

**Composes with.** [prefer-sota](prefer-sota.md) (the adversary→verdict protocol
that *produces* the SOTA read this principle declares); [script-first-mvp](script-first-mvp.md)
and [one-canonical-primitive](one-canonical-primitive.md) (what a good path-2
native floor looks like); [loud-staging](loud-staging.md) (a path-2 build says
loudly what is basic-now-vs-SOTA-later); [verifier-selects-gates-decide](verifier-selects-gates-decide.md)
(gates 1 and 3 are the human's decision, not the recommender's).

**Promotion.** MVP: this principle plus a `## SOTA` section in the intent
template (the alternatives + maturity + the path-1/2/3 recommendation), carried
in `intents/README.md` and `brief/04-surfaces/05-intent.md`. Discipline: an
`intent_sota` record-lint gate that a promoted intent must carry a non-empty
`## SOTA` declaration naming a path, with gate 1 / gate 3 intents blocked from
`plan`/`ship` until an approval or complexity-decision receipt is recorded —
promoted to a discipline-kind intent when the mechanical gate is built.
