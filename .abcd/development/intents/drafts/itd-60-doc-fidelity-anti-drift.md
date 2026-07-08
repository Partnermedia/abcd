---
id: itd-60
slug: doc-fidelity-anti-drift
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
related_adrs: []
prd_path: ".abcd/intents/itd-60/prd.md"
grill_session_id: 60d0f1de-0001-4a60-9c0d-000000000060
glossary_terms_used:
- core/brief
- core/intent
- core/spec
- core/oracle
grilled_intent_hash: bfaa672163edddb5d859bbdfc50169512349da38405c295729ce42c04b894401
prd_grandfathered: false
severity: major
---

# When Something Is Built, The Brief And The Public Docs Reflect It — Or The Spec Cannot Close

## Press Release

> **abcd gains a doc-fidelity pass: a forward-direction check that grades built reality against the brief, and the brief against the audience-adapted public docs, so a shipped capability is never left undocumented and the public framing never out-promises what the code actually does.** abcd already grades delivery against intention (the fidelity reviewer). It does not yet grade documentation against delivery. This pass closes that loop. After each task it advises (soft, non-blocking) where the brief or public docs may now lag the code; before a spec closes it becomes a hard gate. When drift is found it auto-drafts the brief delta and the audience-adapted public-doc delta — one view for end-users *applying* what was built, optionally one for developers *extending* it — and surfaces them for the product thinker to approve or edit. The brief stays the single source of truth; the public docs are derived, audience-specific explanations of it.

> "The gap that scares me isn't a bug — it's a README that quietly describes a capability we never shipped, or a brief that's six specs behind the code," said Iris, a product thinker shipping with abcd. "I read the verdict, not the source. If the docs drift, my whole picture of what I've built is wrong. I want the framework to refuse to call a spec done until the brief and the docs say what's actually true — and to hand me a draft of the change so I'm reviewing, not writing."

## Why This Matters

abcd's external honesty is part of its safety proposition: a product thinker who cannot read code trusts the brief and the public docs to tell them what they have built. An external assessment (2026-06-26) found exactly this failure mode latent in abcd's own corpus — the internal brief honestly scopes what abcd does, while the public README's softer framing could lead a reader to believe abcd enforces safety guarantees it leaves to downstream projects. That is documentation drift: built reality and stated reality diverging, silently.

abcd already has the machinery to grade one artifact against another deterministically — the fidelity reviewer grades delivery against acceptance criteria, fails closed, and never infers a pass from absent evidence. The same discipline applied to documentation closes the loop the assessment named: code → brief (accuracy) and brief → public docs (audience-adapted honesty). Without it, the framework that most rigorously grades *delivery vs intention* has no guard on *documentation vs delivery* — the one surface a non-expert actually reads.

This is the **forward** direction: built reality drives the brief and the docs. The **reverse** direction — a human editing the brief, with the implied roadmap changes drawn out — is a separate, paired intent ([[itd-61-brief-change-derivation]]).

## What's In Scope

- **Two layers, one gate — a deterministic floor beneath a semantic pass.** The pass is not one mechanism but two, stacked. Layer 1 grades *mechanics*; layer 2 (everything below) grades *meaning*.
- **Layer 1 — a deterministic doc-currency lint.** A language-agnostic check family in abcd's native lint engine (the same engine as the record and launch gates, config-driven per repo), it mechanically enforces the present-tense discipline — no change-narration in doc bodies ("previously X, now Y", "deprecated", "to be implemented") — plus resolvable cross-links and no stray or misplaced top-level documents. It needs no oracle, so it is cheap enough to run every commit on *any* abcd-managed repo whatever the language. This is the always-on floor: it catches the drift that is decidable without reading the code.
- **Layer 2 — the semantic doc-fidelity pass.** A doc-fidelity pass that grades, in the forward direction: (1) built reality against the brief (is the brief accurate?), and (2) the brief against the public docs (do the docs honestly explain the brief for their audience, without over- or under-claiming?). This is where meaning — not mechanics — must be judged, so it is host-delegated, not deterministic.
- A tiered enforcement weight, mirroring abcd's existing pre-commit → CI → runtime tiering. Layer 1 (deterministic) is the **pre-commit** floor: it blocks cheaply and always. Layer 2 (semantic) is tiered above it — **per task** it advises and flags possible drift, non-blocking; **per spec** it is a hard gate that must resolve before the spec closes.
- Auto-drafting of the corrective delta: when drift is found, the pass drafts the brief change and the audience-adapted public-doc change(s) and surfaces them for human review (approve / edit), reusing the "thin LLM core inside a deterministic shell, never auto-commit" posture of the fidelity reviewer.
- Audience separation in the drafted public docs: an **end-user** view (applying/using what was built) and an optional **developer** view (extending what was built), both derived from the brief — adapting content per audience is permitted; duplicating specification is not (single source of truth).
- Fail-closed behavior consistent with the existing reviewers: no backend / no reliable comparison → the gate refuses rather than passing silently.

## What's Out of Scope

- **The reverse direction.** Drawing implied intents/principles out of a *human-authored brief edit* is [[itd-61-brief-change-derivation]], a separate intent. This intent only flows built-reality → brief → docs.
- **Re-grading delivery against intention.** That is the existing fidelity reviewer's job; this pass consumes "what shipped" as an input, it does not re-derive it.
- **Authoring the public docs' prose voice / style from scratch.** The pass drafts deltas against existing docs; wholesale doc authorship and information architecture are not its concern.
- **Deciding sequencing or dependencies** relative to the other two intents — that is `/abcd:intent plan`'s job.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** a doc body carries change-narration ("previously X, now Y", "deprecated"), an unresolvable cross-link, or a stray top-level document, **when** the deterministic doc-currency lint (layer 1) runs at pre-commit, **then** it flags each one deterministically and blocks — with no oracle call — independently of, and without waiting for, the semantic pass.
- **Given** a task has completed, **when** the semantic doc-fidelity pass (layer 2) runs, **then** it reports (non-blocking) any place the brief or public docs may now lag the built reality, with evidence pointing at the divergence.
- **Given** a spec is about to close, **when** the doc-fidelity pass runs as a hard gate, **then** the spec cannot close while an unresolved brief-or-docs drift remains.
- **Given** drift is found, **when** the pass produces its output, **then** it auto-drafts the brief delta and the audience-adapted public-doc delta(s) and surfaces them for human approve/edit — it never commits a documentation change silently.
- **Given** the brief is the single source of truth, **when** the public-doc delta is drafted, **then** it adapts the brief's content for its audience (end-user applying; optionally developer extending) without duplicating specification text.
- **Given** no reliable comparison can be made (e.g. backend unavailable), **when** the gate runs, **then** it fails closed (refuses) rather than reporting no-drift.

## Open Questions

- Where does the pass hook? The fidelity reviewer already runs at spec close — does the hard-gate tier compose with it (one surface, two lenses) or run as a distinct gate?
- What is "built reality" as a concrete input — the shipped diff, the spec's acceptance verdicts, the changed surfaces, or a synthesis? The reviewer's existing notion of delivery is the natural candidate.
- Is the developer-extending public-doc view in scope from the first ship, or end-user-only first with developer docs as a follow-up?
- How does the per-task soft tier avoid noise, given most tasks are sub-capability and won't touch the brief/README? A cheap "could this plausibly have changed the brief?" pre-filter may be needed.
- Does this pass itself become a framework-provided discipline (see [[itd-62-pluggable-safety-gate]]'s two-discipline-kinds question), or remain a reviewer surface like fidelity?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Originating assessment: `~/Desktop/abcd-assessment.html` (2026-06-26) — the
  README-over-promises / brief-honest finding that motivates the forward
  doc-fidelity loop.
- Complements: the spc-12 fidelity reviewer (delivery-vs-intention) — this pass
  is the documentation-vs-delivery analogue, reusing its fail-closed,
  deterministic-shell, never-auto-commit posture.
- Paired with: [[itd-61-brief-change-derivation]] (the reverse direction) and
  [[itd-62-pluggable-safety-gate]] (whose brief change this pass would govern).
- Governing principle: single source of truth — the brief is canonical; public
  docs are audience-adapted derivations (`.abcd/development/brief/` and
  `CLAUDE.md` engineering conventions).
