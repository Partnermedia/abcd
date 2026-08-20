# SOTA survey — decision, validity, outcome, knowledge, provenance, and framing records

Dated 2026-08-19. Compiled from two host-run web research passes; every source
below was fetched during the runs. Evidence tiers per
[`prefer-sota`](../../principles/prefer-sota.md): each pick is named as the
presumptive SOTA and then challenged for fit against this repo's conventions.

**Why this note.** Several record extensions are under design: scope/validity
conditions on intents, selection grounds at plan time and capture triage,
per-record provenance of contributed material, a knowledge record distinct from
the brief, and a framing passage in the brief. Each has an in-repo precedent to
extend — the ideate rejected-alternatives shape, the `wontfix_reason`
invariants, the `Assisted-by` trailer convention — and this note positions each
against the nearest published practice.

## 1. Selection grounds / alternatives considered

1. **MADR 4.x** ([adr.github.io/madr](https://adr.github.io/madr/)) — the only
   major template making the **pursued** option's grounds a mandatory,
   structured field: `Chosen option: "{title}", because {justification}`, plus
   per-option good/bad-because arguments and an optional Confirmation section
   (how compliance will be verified). De facto standard.
2. **Y-statements** ([design-practice-repository](https://socadk.github.io/design-practice-repository/artifact-templates/DPR-ArchitecturalDecisionRecordYForm.html)) —
   smallest form carrying both directions in one sentence: decided for X **and
   against** Y, **to achieve** [benefit], **accepting that** [trade-off].
3. **Oxide RFDs** ([RFD 1](https://rfd.shared.oxide.computer/rfd/0001)) — the
   gate-moment answer: determinations recorded when a direction is determined,
   "even if it feels somewhat tentative", with a machine-readable state enum
   whose `abandoned` is a first-class terminal state — the closest published
   analogue to a wontfix-with-reasons ledger.
4. **Rust RFC template** ([rust-lang/rfcs](https://github.com/rust-lang/rfcs/blob/master/0000-template.md)) —
   asks the positive question directly ("Why is this design the best in the
   space of possible designs?"), plus Drawbacks, Prior art, and staged
   Unresolved questions.
5. **Google design docs** ([Ubl](https://www.industrialempathy.com/posts/design-docs-at-google/))
   and **Amazon PR/FAQ** ([workingbackwards.com](https://workingbackwards.com/concepts/working-backwards-pr-faq-process/),
   anecdote tier for template details) — alternatives subordinate to trade-off
   rationale; the PR/FAQ gate records approve/reject/defer **with reasons**.

*Fit:* every surveyed form carries pursued-option grounds as free prose; only
MADR's `because` clause and Oxide's state enum are trivially lintable, and
nobody records grounds at the gate in lintable form. A selection-grounds
section written at plan time and a pursued/deferred/declined triage vocabulary
— on the `wontfix_reason` invariant pattern (state-scoped, required, non-empty)
and the ideate `{alternative, why}` shape with its explicit-empty marker —
would be ahead of published practice, not behind it.

## 2. Validity / scope conditions on a promise

1. **Hypothesis-driven development card** ([Thoughtworks, O'Reilly](https://www.thoughtworks.com/en-us/insights/articles/how-implement-hypothesis-driven-development)) —
   the smallest widely-used form: three fixed stems ("We believe … / Will
   result in … / We will know we have succeeded when …"), one regex away from
   lintable. Encodes the falsification signal, not the validity envelope.
2. **GQM** ([Basili/Caldiera/Rombach, PDF](https://www.cs.umd.edu/~mvz/handouts/gqm.pdf)) —
   the classic form with explicit scope slots: a goal is stated *from a
   viewpoint, relative to a particular environment* — literally validity
   conditions on a goal. Evidence tier (case-study lineage).
3. **Rust RFC unresolved questions** — per-item validity staging
   (resolve-before-merge / before-stabilisation / out-of-scope).
4. Product-level design-by-contract preconditions: no widely-used template
   found — genuinely uncolonised territory.

*Fit:* a `## Scope Conditions` section on intents (the circumstances under
which the promise is expected to hold, read by the intent audit) has the HDD
stems and GQM's environment/viewpoint slots as its two parents, and no direct
published competitor at the per-promise gate.

## 3. Outcome-vs-promise review ("built as specified; the spec was wrong")

1. **Lean validated learning / pivot** ([Ries 2009](http://www.startuplessonslearned.com/2009/06/pivot-dont-jump-to-new-vision.html)) —
   the SOTA vocabulary: the *hypothesis* failed, not the execution; "pivot" is
   the named event, keeping one foot in validated learning.
2. **OKR outputs-vs-outcomes doctrine** ([whatmatters.com](https://www.whatmatters.com/faqs/outputs-vs-outcome-okr)) —
   names the failure mode (outputs completed, outcome unmoved) but has **no
   scoring value** for "the objective itself was wrong"; the verdict lives in
   prose reflection.
3. **Amazon PR/FAQ kill/defer** — the gate records rejection with reasons even
   for a well-executed doc. Anecdote tier.
4. Google postmortems ([sre.google](https://sre.google/sre-book/postmortem-culture/))
   are the wrong neighbour despite surface similarity: they review incidents,
   not promises.

*Fit:* no surveyed practice has a machine-readable verdict separating
*built-wrong* from *wrongly-specified*. If promise-questioning is ever built it
is its own surface with its own ADR (per the 2026-08-19 decision-log entry and
adr-40's closed bucket list), borrowing Lean's hypothesis-invalidated
vocabulary and the Oxide `abandoned`-state pattern; the disposition stays
human, on the `UNACHIEVABLE` replan-invitation precedent.

## 4. Knowledge / learning registries distinct from living docs

1. **Google's postmortem repository** ([sre.google](https://sre.google/sre-book/postmortem-culture/)) —
   the one large-scale survivor, and the chapter is explicit that it survives
   by **mandatory triggers, senior review before publication, and continuous
   reinforcement** — incentives, not structure.
2. **Why most fail** ([Sillito & Pope, arXiv:2402.09538](https://arxiv.org/abs/2402.09538),
   small-n) — failure learning is "informal, ad hoc, and inconsistently
   integrated"; repositories exist, the loop back into work doesn't.
3. **Per-entry confidence metadata** — the best-developed sustained convention
   is [Gwern's page metadata](https://gwern.net/about): a closed confidence
   vocabulary (Kesselman estimative-probability scale), an importance decile,
   and completion status. Single practitioner, 15+ years running; trivially
   lintable because the vocabularies are closed.
4. **Oxide's never-delete rule** — "the rationale of the past helps assure that
   hard-won wisdom is not discarded."

*Fit:* survivors pair a closed, lintable metadata vocabulary with a social
review gate. A knowledge record here must not be a free-standing registry (the
rot evidence): entries carry scope, evidence pointers **into existing record
ids**, and a closed confidence vocabulary, and are written at mandatory trigger
moments in the existing lifecycle rather than ad libitum. Routed per itd-84:
intent + ADR (tier, id scheme) + minimal record-lint.

## 5. Provenance of contributed ideas

1. **Linux kernel `Assisted-by:`** ([docs.kernel.org](https://docs.kernel.org/process/coding-assistants.html)) —
   the anchor standard: disclosure, not authorship; certification (DCO) is
   human-only; a `Co-developed-by` proposal for machine assistance was
   **rejected** in favour of the dedicated tag. checkpatch support is arriving,
   so machine enforcement has precedent. This repo's existing trailer
   convention already matches the endpoint of that debate.
2. **Fedora Council policy** ([LWN](https://lwn.net/Articles/1039623/)) —
   allow-with-disclosure via the same trailer; the "significant assistance"
   threshold is the contested part. The wider field spans ban (Gentoo) /
   disclose / structured provenance (Apache `Generated-by` feeding a
   machine-parsable file).
3. **W3C PROV** ([primer](https://www.w3.org/TR/prov-primer/)) — the right
   *naming* source (`wasAttributedTo`, `actedOnBehalfOf` — a tool acting on
   behalf of a human is directly expressible); the wrong *format* (RDF weight,
   no sub-document mechanism).

*Fit:* commit-level disclosure is solved. **Record-level provenance — a
frontmatter key saying where a draft's material came from (product thinker,
facilitator, a review output, triage) — has no prior art in this survey.** The
repo already practises it informally as the "Facilitator-seeded draft"
blockquote in several drafts; formalising that into a closed-vocabulary
frontmatter field inherits kernel semantics (disclosure, not authorship) and
PROV naming, and would be novel.

## 6. Framing artefacts ("what the situation is being treated as")

1. **How Might We** ([NN/g](https://www.nngroup.com/articles/how-might-we-questions/)) —
   the only convention whose wording itself marks provisionality; the NN/g
   checklist (not a solution in disguise, not symptom-focused, tied to observed
   insight) is effectively a lint spec for framings.
2. **Opportunity Solution Trees** ([Torres](https://www.producttalk.org/opportunity-solution-trees/)) —
   framing explicitly provisional ("crummy first draft"), restructuring *is*
   the reframe, stabilisation is the learning signal — but the reframe is a
   state change, not a logged event.
3. **Lean pivot** — the only surveyed practice where a reframe is a **named,
   announceable event** with conservation semantics.
4. **Google design docs' Non-Goals** — a framing boundary made contestable by a
   document lifecycle.
5. **Gwern's epistemic-status tags** — a working convention for "provisional
   and contestable" at page level.

*Fit:* framing is the least record-shaped area surveyed — everything lives on
whiteboards or in prose. A framing passage in the brief — optional, visibly
marked under the same loud-staging discipline the brief already uses, with a
reframe recorded through the existing typed supersession links — would be a
genuine synthesis, with Nygard supersession and the pivot-as-event as the two
precedents to cite when the ADR is written.

## Not worth adopting

- **Full long-form MADR everywhere** — MADR itself ships four grades because
  teams abandon the long form; take the `because` clause and the Confirmation
  idea only.
- **`Co-developed-by` for machine assistance** — proposed and rejected by the
  kernel community; this repo's convention already forbids it, correctly.
- **A free-standing lessons-learned database** — the rot evidence plus the
  absence of Google-scale incentive machinery at solo-maintainer scale; fold
  learnings into records with triggers and confidence fields instead.
- **OKR scoring as an outcome-review mechanism** — ceremony without the one
  verdict that matters.
- **PROV serialisation as a record format** — borrow the terms, not the RDF.

**Three most load-bearing sources:**
[Oxide RFD 1](https://rfd.shared.oxide.computer/rfd/0001) (end-to-end decision
records with machine-readable states in a Git repo),
[the kernel coding-assistants policy](https://docs.kernel.org/process/coding-assistants.html)
(provenance semantics), and [MADR](https://adr.github.io/madr/) (pursued-option
grounds and the Confirmation hook a lintable record system can enforce).
