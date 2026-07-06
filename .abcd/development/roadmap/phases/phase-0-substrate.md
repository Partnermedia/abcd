# Phase 0 — Substrate & disciplines

## Expectation

By the end of this phase, abcd has a working oracle backend and an honest
account of its own state. A contributor running `flowctl specs` sees real
numbers; a reviewer asking "what's shipped" gets an answer that matches the
filesystem. The three cross-cutting disciplines are in force, so every spec
planned after this point inherits Given-When-Then acceptance, prompt-quality
gates, and a modification-grammar section from the moment it is written — no
retrofitting. The three-intent-kinds taxonomy is codified, and the spec-terminology
rename is complete before any further doc is written against the old word.

This phase ships **no product-capability user moment** — no press-release-worthy
capability a user reaches for — and that is why it is numbered 0. It is the
floor the five capability phases stand on: the oracle the lifeboat pipeline
calls, the disciplines the later specs are measured against, and the vocabulary
every later doc is written in. Numbering it 0 rather than 1 states honestly that
no *product capability* lands here; the first such moment is `/abcd:ahoy` in
Phase 1. The qualifier is "product capability", not "user-typed surface": fn-12
adds the `/abcd:intent review` discipline-audit *sub-verb* — a
substrate/maintenance surface, the substrate inspecting itself, not a product
capability. So Phase 0 does carry a user-typed verb; what it does not carry is a
product-capability user moment.

## Milestone

- `flowctl specs` reports true per-task counts: `.git/flow-state/` holds the
  rehydrated runtime state, restored from the `.flow/.checkpoint-*.json` files
  (it was never persisted — see `.work/issues.md`). With that store in place,
  "what is shipped, what is next" is answerable, which every other item in this
  phase depends on.
- `fn-3`'s spec JSON `status` is `done` — its six tasks are complete on disk
  with a `ship` completion review, and the JSON now matches.
- The intent-lifecycle automation has since shipped — **not** as a Claude Code
  `intent_lifecycle_hook` (no such hook exists in `hooks/`), but as a flowctl
  `spec close` close-hook (fn-36, registered in
  `scripts/abcd/tools/flow-next/verbs.json`) that, after a successful upstream
  close, calls `scripts/abcd/intent_lifecycle.py`'s `reconcile` (fn-28). On a
  linked spec's close the reconcile moves the intent (and its bundle-mates)
  `planned/` → `shipped/` and enqueues a review-queue entry; the
  `intent-fidelity-reviewer` itd-1 audit then runs off that queue (via
  `review-queue run`/`drain` — fn-28.2, with the opt-in `review.autodrain` edge
  per fn-43). So the *lifecycle move* is automatic; the reviewer fires from the
  queue, not synchronously inside the close. fn-12 shipped the
  `intent-fidelity-reviewer` agent and its `/abcd:intent review` surface. (At
  Phase 0 close this automation had not yet landed and these moves were tracked
  reconciliation work; that backlog cleared with fn-36/fn-28.)
- `fn-5` is shipped: `RPUnavailable` and `MCPBridge` exist; the RP MCP leg of
  the oracle cascade works (subject to RP approval-denial caveats per the
  fn-5 spec).
- All three disciplines (`itd-1`, `itd-5`, `itd-37`) are registered as active
  gates in `disciplines/`, and `itd-34`'s three-kinds taxonomy is the schema
  they are registered under.
- The `itd-43` spec-terminology rename is complete: schema, reviews, and prose
  use "spec", with `terminology/core/spec.md` canonical. No half-renamed state
  remains in the corpus.

## Phase Acceptance

> _Roll-up acceptance per [adr-9 amendment](../../decisions/adrs/0009-phase-as-product-layer.md). Each bullet asserts an emergent, cross-intent truth or a phase-spanning user journey — never a copy of an intent's own `## Acceptance Criteria`._

- **Given** the three disciplines are registered, **when** any spec is planned
  in a later phase, **then** that spec inherits all three discipline
  gates at once — Given-When-Then acceptance, `prompt_version`, and a
  `## Modification Grammar` section — with no per-spec retrofit. (Emergent: the
  *every later spec is born correctly-shaped* guarantee is a property of the
  three disciplines being in force together, owned by no single discipline.)
- **Given** a discipline gate has a mechanical half (section/field presence,
  well-formedness) and a judgement half (is the criterion *actually met*? is
  the Modification Grammar boilerplate?), **when** Phase 0 closes, **then** the
  mechanical half is hard-enforced by `intent_lint.py` / `lint_prompts.py`
  (the `IL`/`MG`/`VR` lint families, `prompt_version`/`capability_scope`
  checks) and the judgement half is formalised by the dedicated
  `intent-fidelity-reviewer` agent — **fn-12 ships that agent in Phase 0**, as
  the last Phase 0 epic. fn-12 builds the agent's Role 1: the itd-1
  per-criterion `MET`/`MET_WITH_CONCERNS`/`NOT_MET`/`INCONCLUSIVE` acceptance
  pass and the itd-37 `MG004` boilerplate pass (the discipline-judgement
  subset of `/abcd:intent review`; the broader press-release-prose /
  term-drift / PRD-fidelity outputs are deferred to a later epic). This
  supersedes the earlier plan-record statement that the reviewer agent was
  *not* a Phase 0 deliverable — that statement predated the fn-8/fn-12 split
  that made the judgement half buildable inside Phase 0. Before fn-12 lands,
  the judgement half is covered by the existing dual-backend oracle at
  `/flow-next:plan-review` and `/flow-next:impl-review`; once fn-12 ships, "the
  disciplines are in force" is true at full strength — **lint-enforced (fn-8)
  AND judgement-enforced by the dedicated reviewer (fn-12)**.
- **Given** the reconciliation work is done, **when** a contributor asks "what
  is shipped and what is next", **then** `flowctl` and the phase docs give the
  same answer — the flow-state desync is fixed and `fn-3`'s spec status is
  honest. (At Phase 0 close the intent-lifecycle directories were not yet in
  step — itd-27/itd-28 sat in `planned/` although their specs were closed,
  pending the lifecycle automation. That automation later shipped as the fn-36
  close-hook → fn-28 reconcile, and fn-48 backfilled the lifecycle state across
  closed specs; the directories are now reconciled.)
- **Given** fn-5 has shipped, **when** any oracle-using code path runs, **then**
  the RP MCP leg is available as the cascade's preferred backend — the
  substrate every later phase's audits and reviews depend on is in place.
- **Given** the `itd-43` rename and the `itd-34` taxonomy have both landed,
  **when** any later phase's spec or intent is authored, **then** it is written
  in the settled vocabulary (one kind from three; the settled term, not the
  former word) from the first draft — no later phase inherits a terminology migration.

## Scope

**Intents:** itd-6 (RP MCP integration — the oracle backend leg; no user
moment, so it lives in Phase 0 even though it is what ahoy's install probe
checks for), itd-1 (acceptance gates — discipline), itd-5 (prompt-quality
additions — discipline), itd-37 (modification grammar — discipline), itd-34
(three intent kinds — the taxonomy the disciplines are registered under),
itd-43 (spec-terminology rename).

**Brief plumbing-phases:** the brief's "Phase 0 — Foundation" (specs `fn-1`,
`fn-4` done) is already complete; this phase consumes its output (the harness
interface, ADR-02) rather than re-running it. Note the numbering now aligns:
the brief's foundation work and this phase are both "Phase 0".

**Reconciliation work** (no intent — originally tracked in `.work/issues.md`):
the `.git/flow-state/` runtime store is rehydrated from the checkpoint files
and the `fn-3` spec JSON reads `done`. The lifecycle backlog that was pending at
Phase 0 close — the intent-lifecycle automation, and the `itd-27`/`itd-28` move
to `shipped/` with an `intent-fidelity-reviewer` audit on each — has since
cleared: the automation shipped as the fn-36 close-hook → fn-28 reconcile
(move + review-queue enqueue), and fn-48 backfilled lifecycle state across
closed specs. These were bugs and stale state, not verification steps.

The `intent-fidelity-reviewer` agent ships as fn-12, the last Phase 0 epic (see
the discipline-gate Phase Acceptance bullet above). The reviewer fires off the
review-queue (`review-queue run`/`drain`, fn-28.2, with the opt-in
`review.autodrain` edge per fn-43) rather than synchronously inside the close
hook; fn-12 also exposes the manual `/abcd:intent review` surface.

## Maps against

- **Brief:** `06-delivery/01-build-sequence.md` (Phase 0 foundation, itd-1/5/6
  in the execution order); `05-internals/01-agents.md` (oracle backend
  resolution); `05-internals/05-prompt-quality.md` (itd-5's home);
  `01-product/03-mental-model.md` (itd-34's three-kinds taxonomy).
- **Intents deliver the expectation:** itd-6 delivers the working oracle leg;
  itd-1/5/37 deliver the discipline gates that make every later spec auditable;
  itd-34 and itd-43 settle the taxonomy and vocabulary every later phase writes
  in.
- **ADRs realised:** adr-02 (MCPBridge contract — fn-5's implementation
  contract); adr-8 (dual-backend review, already in use for fn-5's plan
  review).

## Dependency rationale

- **fn-5 first among the build work** — it is already plan-reviewed (RP +
  Codex both SHIP, round 27) and depends only on `fn-4` (done). It is the
  lowest-risk, highest-unblock item.
- **Disciplines before Phase 1** — disciplines impose inherited acceptance
  gates on every *other* spec. Registering them before any later capability
  phase's surface spec is planned means those specs land correctly-shaped from
  day one. This is the single most important ordering constraint in the whole
  plan, and the reason this work is a phase of its own rather than folded into
  the ahoy phase.
- **itd-43 rename before the corpus grows** — a terminology rename is cheapest
  while the corpus is small; every later phase's docs are written in the new
  vocabulary, so the rename must precede them.
- **Reconciliation before anything** — the flowctl state desync makes "what's
  next" unanswerable until fixed; it gates trustworthy progress reporting for
  every later phase.

## Open questions

- The `phase:` spec anchor now HAS lint coverage: fn-66 delivered the `PA001`
  verify-exists lint (a `phase:` naming a non-existent phase is an error; a
  missing anchor stays legal) plus the phase-audit reviewer that reads it. The
  valid-phase set is derived live from the phase docs
  (`phase_resolution.valid_phase_ids()`), so the six phases 0–5 need no manual
  update. What remains deferred is the corpus anchor *backfill* (making `phase:`
  a standing convention) — a separate planning act, unblocked by fn-66.
- `IL011` (the cross-phase bundle check) is now a plan-time check rather than a
  frontmatter-field lint, since phase membership lives editorially in phase
  docs. Confirm the planner has access to target-phase context when `IL011` is
  eventually implemented.
- itd-34 was previously folded into this phase as a rider on the disciplines;
  it is now a named scope intent. Confirm itd-34 still ships *with* the
  disciplines (its taxonomy is exercised by them) rather than as a separable
  unit.
