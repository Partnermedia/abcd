# The ABCD sensemaking method — cold reading, warm ledger, disposition

**Status:** durable record of the method abcd tools. Recorded 2026-07-13.

abcd's principles have been converging on a method that was never written down in
this record. The method exists; it was developed alongside abcd, and one of its
passes — the cold reading — exists as a working skill outside this repo. This note
is the parent the record has been missing: it states the method, maps it onto the
principles that already restate parts of it, and names which intents implement
which pass.

Without this note, each pass gets rediscovered piecemeal and its rules look
arbitrary. A detector that refuses to propose fixes reads as an unhelpful detector
unless you know a disposition step exists downstream to do that work.

---

## The method

Three parts, and the separation between them is the whole point.

### 1. The cold reading — the blind specification pass

A reader with broad knowledge of how designs of this kind work, and **zero
investment in how this particular design came to be framed**, reads a single
document against the constraints visible *in that document* and surfaces
**candidate detections**: places where the document's own stated or implied
constraints are inconsistent with its framing.

Its value comes from what it refuses to do:

- **Carries no project context.** Background about the wider programme, prior
  decisions, or why the design is shaped as it is, is set aside — even when known.
  An embedded reader's sympathy smooths over exactly the inconsistencies that
  familiarity has normalised.
- **Remembers no prior readings.** Deliberately amnesiac; re-raises tensions it
  surfaced before if they are still present in the document.
- **Proposes no fixes.** A detection presented with its solution attached has
  already been collapsed into a resolution and can no longer be *held*.
- **Ranks nothing.** Prioritisation is a meta-level judgement made against the
  ledger, on axes the cold reading cannot see.
- **Does not pre-judge intentionality.** No "this may be intentional, but…". A
  tension that turns out to be intentional is not a false positive — surfacing it
  forces the intention to be stated, which is the value.

Each detection carries exactly three things: the **tension**, the **constraint in
play** (explicit or implicit, and what implies it), and **why it is a tension**.
No fourth element.

### 2. The warm ledger — the accumulated record

The detections, their dispositions, and the history of both. It is warm precisely
because it carries everything the cold reading is denied: what was decided before,
what was accepted as an intentional constraint, what is still open. The cold
reading never sees it. **This is the separation that makes the method work.**

### 3. The disposition — the human judgement

A human takes each candidate detection and dispositions it: **accept**, **reject**,
or **hold** — that last one being the load-bearing option, and the one a
fix-proposing reviewer destroys. A tension can be real and not yet articulable;
holding it keeps it alive without forcing a premature resolution.

---

## What this method already is, inside abcd

The convergence is not a coincidence — it is the record re-deriving the method from
the other direction, one principle at a time.

| Method element | Where abcd already states it |
|---|---|
| Detector is denied the context that would make it sympathetic | [`evaluator-outside-the-loop`](../../principles/evaluator-outside-the-loop.md) — but abcd's version isolates the evaluator *structurally* (it cannot edit its own gate), not *epistemically* (it must not know why the design is shaped as it is). The method's cut is stronger. |
| A detection is a proposal; the human's disposition is the gate | [`verifier-selects-gates-decide`](../../principles/verifier-selects-gates-decide.md) — an exact restatement, arrived at independently. |
| A recurring detection is evidence the resolution was false | **Nothing.** This is the one element with no counterpart in the record. Now [`recurrence-is-signal`](../../principles/recurrence-is-signal.md). |
| The warm ledger | The capture ledger (`iss-N`) is the warm ledger's tooled form — detections land, a human dispositions them (`resolve` / `wontfix`). It is missing the recurrence behaviour (see itd-87). |

The deliberate omission: **no new principle was minted for the cold/warm split.**
It is already covered by the two principles above, and `one-canonical-primitive`
forbids a third near-copy. Only the genuinely absent element became a principle.

## Which intents implement which pass

- **itd-27 / itd-42** (grill; ADR-0007) — interactive challenge of acceptance
  criteria. *Planned, not built.*
- **itd-55** (foundations auditor) — audits where a claim's justification
  terminates. *Draft.*
- **itd-86** (cold-reading surface) — the blind detection pass. *Draft.*
- **itd-87** (recurrence escalation in capture) — makes the ledger warm in the
  sense the method requires. *Draft.*

These four have been accumulating as siblings with no stated parent. This note is
the parent.

---

## Provenance and attribution

The cold reading was developed by abcd's co-author in support of abcd. The
`ACKNOWLEDGEMENTS.md` entry is **deliberately deferred, not forgotten** — it is
held pending confirmation of how they wish to be credited, and must land before
itd-86 ships. Recording this openly rather than silently, per
[`loud-staging`](../../principles/loud-staging.md).

Its sibling skills (`socratic-grill`, `first-principles-analysis`) were evaluated
separately in
[`socratic-and-first-principles-skills-evaluation.md`](socratic-and-first-principles-skills-evaluation.md);
that note's dispositions stand.

## Scope of this note

This records the method as it is practised. The cold reading's own text is the
source for part 1; parts 2 and 3 are stated at the level the cold reading's scope
boundary implies them, since they are operated by a human at the meta-level. Where
this note describes the ledger's two disposition axes (frame-location and MoSCoW
priority) it is naming them, not specifying them — a full specification of the
warm ledger and the disposition step is **not** in this note and is not implied by
it.
