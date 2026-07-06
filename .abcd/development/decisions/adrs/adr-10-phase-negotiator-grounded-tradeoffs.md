---
id: adr-10
slug: phase-negotiator-grounded-tradeoffs
status: accepted
date: 2026-05-16
supersedes: null
superseded_by: null
related_intents: [itd-41]
related_rfcs: []
related_adrs: [adr-9]
---

# ADR-10: The phase negotiator — a Socratic agent that proposes phases and grounds every trade-off

> **Terminology note.** The *how* layer is named the **spec**. This ADR's prose
> was updated by the spec-terminology-rename ADR
> ([adr-11](adr-11-spec-terminology-rename.md)).

## Context

[adr-9](adr-9-phase-as-product-layer.md) made the **phase** abcd's sequencing
layer: an ordered stretch of work that bundles intents and plumbing and ends in
a milestone, with a product-authored `## Expectation` and `## Phase Acceptance`.

adr-9 gave sequencing a *home*. It did not give it a *negotiator*. Deciding
which intents go in which phase, and in what order the phases run, is currently
the product thinker's unchallenged judgement. That has a known failure mode,
and the user named it directly: **the product thinker "wants everything."**
They hold a vision of the whole product; they have no natural adversarial
reader at the phase-planning grain who will tell them, honestly, *what they
give up* when they pull an intent forward, defer one, or split a phase.

abcd already has adversarial readers at the other grains — `intent-fidelity-
reviewer` (does reality match the intent's acceptance?), the `/abcd:intent
grill` skill (Socratic interrogation of an intent before it is planned), and
dual-backend review of specs (adr-8). The phase-planning grain has none.

The hard part is not "add an agent that lists trade-offs." An LLM asked "what
are the trade-offs of this phasing?" will fluently invent plausible ones. A
hallucinated trade-off is worse than no trade-off: it spends the product
thinker's trust on a fiction. The real design problem is **grounding** — every
trade-off the agent raises must trace to a fact that already exists in the
project, not to the model's generative fluency.

## Decision

abcd adds a **phase negotiator**: an agent that proposes a phase (or a
re-sequencing of phases) and then discusses the trade-offs with the product
thinker. It is captured as a user-facing capability in [itd-41](../roadmap/intents/drafts/itd-41-phase-negotiator.md).

The negotiator is built from two patterns abcd already has, deliberately —
it is **the same family as `intent-fidelity-reviewer` and the grill skill**,
applied one grain up, at the phase-planning grain:

1. **Socratic, not assertive (from the grill skill).** The negotiator does not
   *declare* "deferring itd-3 costs you X." It **interrogates** the proposed
   phasing with questions, each tagged with a named Socratic move, the same way
   `/abcd:intent grill` interrogates an intent. "You have placed itd-3 in
   Phase 3 — what in Phase 2 depends on the rules loader and now waits?" The
   product thinker's own answer surfaces the trade-off. An agent that only asks
   cannot hallucinate an answer it does not give.

2. **Grounded comparison (from `intent-fidelity-reviewer`).** Where the
   negotiator *does* assert, every assertion must cite a fact that already
   exists in the project — exactly as `intent-fidelity-reviewer` cites a
   specific `## Acceptance Criteria` bullet rather than a generated standard.
   The grounding sources, in priority order:
   - the **dependency DAG** in `brief/06-delivery/01-build-sequence.md` and
     intent/spec dependency declarations — "deferring itd-3 delays every item
     whose dependency chain runs through the rules loader" is a checkable
     graph fact;
   - each phase's **`## Phase Acceptance`** (per the adr-9 amendment) — moving
     an intent out of a phase means naming which `## Phase Acceptance` bullet
     can no longer be met;
   - the **`## Expectation`** prose — what the phase's user-truth would lose.

**The grounding rule (load-bearing):** the negotiator MUST NOT present a
trade-off it cannot tie to a named DAG edge, a named `## Phase Acceptance`
bullet, or a named `## Expectation` clause. If a concern has no such anchor, the
negotiator raises it as a *question* (Socratic mode), never as an asserted
trade-off. "Honest trade-offs only" is enforced by construction: assertions are
grounded; everything else is a question.

The negotiator's output is a *proposal plus a discussion*, not a decision. The
product thinker still owns the phase plan — they own `## Expectation` and
`## Phase Acceptance`, per adr-9. The negotiator is the adversarial reader that
makes "wants everything" meet "here, concretely and truthfully, is the cost."

## Alternatives Considered

1. **No negotiator — phase planning stays the product thinker's unchallenged
   call.** Rejected: it leaves "wants everything" with no counter-force. adr-9
   gave sequencing an artefact; without a negotiator the artefact is just an
   unreviewed wish-list, and the phase audit only catches the mismatch *after*
   the work is built.
2. **A trade-off agent that asserts trade-offs directly (no Socratic layer).**
   Rejected: this is the hallucination trap. An agent licensed to assert
   trade-offs will assert ungrounded ones fluently. Restricting assertions to
   grounded facts and routing everything else through questions is what makes
   the output trustworthy.
3. **A pure DAG/graph tool — compute the dependency consequences mechanically,
   no LLM, no discussion.** Rejected: the mechanical consequences are *part* of
   the grounding, but a graph diff is not a negotiation. The product thinker
   needs the *discussion* — the Socratic surfacing of what an Expectation would
   lose, which is a judgement exchange, not a graph query. The negotiator uses
   the graph as grounding and the Socratic loop as the interface.
4. **Fold it into the existing grill skill as another mode.** Rejected: grill
   operates on a single intent before it is planned; the negotiator operates on
   a phase (a set of intents plus plumbing) and on the ordering of phases —
   different input, different grain. Shared lineage (Socratic moves) does not
   mean shared surface. It is its own capability.

## Consequences

**Gains:**
- The phase-planning grain gains the adversarial reader the other grains
  already have. abcd's review family becomes complete: grill (intent, pre-plan)
  → phase negotiator (phase, pre-commit) → dual-backend review (spec) →
  `intent-fidelity-reviewer` / phase audit (delivered reality).
- "Honest trade-offs" is a structural guarantee, not a hope: assertions are
  DAG/acceptance-grounded, the rest are questions.
- The `## Phase Acceptance` blocks (adr-9 amendment) gain a second consumer —
  they are both the phase-audit target *and* the negotiator's grounding input.

**Costs / obligations:**
- The negotiator needs the dependency DAG to be accurate and machine-readable
  enough to ground claims against. Where the DAG is prose-only, the negotiator
  is limited to Socratic questioning for that area — an honest degradation, but
  a real limit.
- One more agent to maintain, with the grill skill's Socratic-move taxonomy as
  a shared dependency.
- The negotiator must not drift into a decision-maker. Lint or prompt-level
  discipline keeps its output advisory (proposal + discussion), preserving the
  product thinker's ownership of the phase plan per adr-9.

**Downstream decisions enabled:**
- itd-41 is the user-facing intent that specifies the negotiator's surface,
  flow, and acceptance criteria.
- A future shared "Socratic move taxonomy" module, if grill and the negotiator
  converge enough to warrant extracting one.
