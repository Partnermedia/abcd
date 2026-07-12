---
id: itd-81
slug: judge-calibration
kind: discipline
kind_notes: "Cross-cutting calibration gate over every agent that renders a verdict on code or artefacts. No user moment of its own; imposes an acceptance gate every reviewer-agent spec inherits. Sits below itd-5 (which mandates that a prompt is tested) and answers the question itd-5 leaves open: tested against what, and scored how."
suggested_kind: null
spec_id: null
reclassification_history: []
blocked_by: [itd-1, itd-5]
severity: major
---

# No Judge Ships Unmeasured, And A Clean Diff Is A Result

## Rule

Every agent abcd ships that renders a **verdict** on code or artefacts — today
`intent-fidelity-reviewer`, and every reviewer, auditor, linter-with-judgement,
and oracle that follows — is calibrated against a labelled corpus before its
prompt is locked, and re-scored whenever the prompt changes.

Four requirements, enforced at agent-spec close-time:

1. **A calibration corpus exists for the agent**, at
   `agents/<name>/corpus/`, containing labelled cases of **both** classes:
   - **known-bad** — an artefact with a defect the agent is supposed to catch,
     with the defect class named and the expected verdict recorded;
   - **known-good** — an artefact that is *correct*, where the expected verdict
     is the agent finding **nothing**.

   Known-good cases are **not optional and not a minority**. They are at least
   40% of the corpus.

2. **The agent is scored on both axes, and both are recorded**: recall (share of
   known-bad cases caught) and **true-negative rate** (share of known-good cases
   the agent correctly leaves alone). A judge with high recall and an
   unmeasured TNR is an unmeasured judge.

3. **A false-positive ceiling gates the lock.** The agent's spec declares the
   TNR floor it must clear before its prompt locks at `1.0.0`. A prompt change
   that improves recall while dropping TNR below the floor does not ship.

4. **Every finding the agent emits carries a failure scenario** — concrete
   inputs or state, and the wrong output, crash, or violated invariant that
   results. A finding that cannot state one is not emitted. This is the
   admissibility bar the corpus scores against.

The corpus is built by **injection, from our own history**: take a merged
`fix:` commit, keep the merged state as a known-good case, and reconstruct the
pre-fix state as the known-bad case. The defect class is already named in the
closed issue. Nothing needs to be invented.

## Why

abcd's whole proposition is that a configuration layer can raise the quality of
LLM-assisted work. A reviewer that cries wolf destroys that proposition
quietly: the operator learns to skim it, and then the one real finding is
skimmed too. We currently have no way to detect that happening, because we have
never measured it.

The 2026 evidence (surveyed in
[`research/notes/2026-07-12-judge-calibration-sota.md`](../../research/notes/2026-07-12-judge-calibration-sota.md))
is directional and consistent:

- LLM code judges **systematically over-flag**, and roughly a third of their
  errors trace to statements **hallucinated into the code** ([overcorrection]).
- Judges **underestimate human-written code and overestimate LLM-written code**
  ([judge-codegen]) — a self-preference bias that lands squarely on abcd, where
  the reviewing model is routinely the authoring model.
- Prompts that demand **explanations and proposed corrections** measurably
  **increase** misjudgment ([overcorrection]) — which is why requirement 4 is a
  bar on admissibility, not an invitation to elaborate.
- Ground truth for code review is manufactured by **injecting** defects; this is
  how CriticGPT was evaluated ([criticgpt]) and how the credible review
  benchmarks are built ([qodo]).

**This discipline is the missing half of itd-5.** itd-5 mandates that every
agent prompt is versioned, pre-flighted, canaried, and backed by golden
fixtures. Golden fixtures establish that an agent *does something*. They do not
establish how often it is *wrong about a clean artefact*, and that is the
failure mode that actually erodes trust in a review layer.

**It also corrects an unsafe rule in itd-5.** itd-5's one-shot pre-flight
accepts an oracle-rewritten prompt when it "passes the same goldens and is
shorter by >10%". That tiebreak **selects for brevity bias** — the mechanism
[ace] identifies as progressively destroying instruction quality, where
optimisers converge on short generic instructions that drop the domain detail
that was load-bearing — and it does so against goldens that do not measure
false positives at all. A shorter prompt passing the same goldens may simply
have shed detail the goldens never tested. **Under this discipline the >10%
length tiebreak is struck; the pre-flight gate is the corpus score.**

## What's In Scope

- The corpus layout, the label schema (case → defect class → expected verdict),
  and the scoring harness that reports recall and TNR per agent.
- The TNR floor as a declared, spec-level acceptance gate.
- The two-stage output contract every judge agent emits: free-prose analysis
  first, structured findings second — reasoning inside a schema measurably
  degrades reasoning ([speak-freely]), and unstructured prose cannot be scored.
  Both, in that order, satisfies both constraints.
- Binary verdicts. Severity is structural (a FIX-FIRST list vs a NOTES list),
  never a 1–5 score from the model: numeric scales drift to the middle and carry
  position bias even pointwise ([position-bias]).
- Corpus growth as a by-product of ordinary work: every closed `fix:` issue is a
  candidate case, harvested at close-time.

## What's Out of Scope

- **Automatic prompt rewriting.** A reflect→delta→re-score loop is the right
  mechanism ([gepa]) but it runs **human-approved, one delta at a time**. An
  unattended loop optimising against our own reviewer teaches the reviewer to
  pass, not the code to be good — measured at **73.8% spurious gains** on
  KernelBench ([reward-hacking]). Never autonomous.
- **Judge panels / juries.** A nine-judge, seven-family panel yielded 2.18
  effective independent votes and matched-or-underperformed the single best
  judge ([nine-judges]); errors across models are correlated, so the Condorcet
  argument does not apply. Cross-family disagreement is a **triage signal on
  contested findings**, not a voting scheme. Out of scope here.
- **Training, fine-tuning, process reward models.** No weight access; not a
  solo-scale practice.
- **A rubric written in advance.** Criteria drift ([validators]) says the rubric
  is *derived* from graded outputs, not declared before them. This discipline
  mandates the corpus and the score; the taxonomy of failure classes emerges
  from error analysis of the agent's real outputs, and is not front-loaded.
- **Non-judging agents.** Agents that transform or scaffold rather than render a
  verdict (`chat-distiller`, `embark-scaffolder`) remain under itd-5 only.

## Acceptance Criteria

> _BDD format, per the [itd-1 discipline](itd-1-acceptance-gates.md)._

- **Given** any agent spec that ships a verdict-rendering agent, **when** the
  spec closes, **then** `agents/<name>/corpus/` exists with labelled cases of
  both classes, and known-good cases are ≥40% of the corpus.
- **Given** a calibration run of that agent over its corpus, **when** the scoring
  harness reports, **then** it emits both recall and true-negative rate, and the
  agent's spec records the TNR floor it had to clear.
- **Given** a known-good case (a correct, merged diff), **when** the agent runs,
  **then** it returns a clean verdict and emits no findings — and a run that
  invents a finding on a known-good case is scored as a failure, not as
  thoroughness.
- **Given** any finding the agent emits, **when** the harness validates it,
  **then** the finding carries a failure scenario (inputs/state → wrong result);
  a finding without one fails validation.
- **Given** a prompt change to a locked judge agent, **when** it is proposed,
  **then** the agent is re-scored on the corpus, and the change is rejected if
  TNR falls below the declared floor — regardless of recall gains.
- **Given** an agent's one-shot itd-5 self-improvement pre-flight, **when** the
  oracle-rewritten variant is compared to the candidate, **then** the decision is
  made on corpus score alone; **length is not a tiebreak**.
- **Given** a closed `fix:` issue in any project running abcd, **when** the issue
  closes, **then** the pre-fix and post-fix states are available as a candidate
  corpus pair (harvest is offered; adoption is the maintainer's call).

## Open Questions

- **Where the TNR floor starts.** A floor set before we have a baseline is a
  guess. Proposal: run the current `intent-fidelity-reviewer` against a first
  corpus, take the measured TNR as the floor, and ratchet — never regress.
  Resolve at the first judge spec's T1.
- **Corpus size before the number means anything.** 30–50 cases is the working
  figure from the review-benchmark literature; below ~20 the TNR estimate is
  too noisy to gate on. Confirm empirically.
- **Does the harness run in `preflight`?** Leaning no — judge scoring is the
  expensive, non-deterministic tier and belongs in an async lane; `preflight`
  stays deterministic. Confirm at T1.
- **Reflector separation.** The reflect step (proposing a prompt delta from
  failing traces) must not be performed by the agent under test — single-agent
  reflection produces confirmation bias and re-commits to the same flawed chain
  ([mar]). Whether the reflector must also be a *different model family*, or
  merely a different context, is unresolved.

## Dependencies

- **Hard prerequisite:** [itd-1](itd-1-acceptance-gates.md) — this discipline's
  gates use its Given-When-Then shape.
- **Extends and corrects:** [itd-5](itd-5-prompt-quality-additions.md) — itd-5's
  golden fixtures and pre-flight are the substrate; this discipline supplies the
  labelled-corpus scoring they lack, and strikes the >10%-shorter tiebreak.
- **Constrains:** [itd-64](../drafts/itd-64-benchmark-driven-config-optimization.md)
  — itd-64 proposes learning tuned defaults from run outcomes. It must not treat
  reviewer verdicts as ground truth (the reviewer is the instrument under
  measurement), and it inherits the never-autonomous rule above.
- **Complements:** [itd-15](../drafts/itd-15-self-dogfooded-sota-audit.md) —
  itd-15 audits prompt-vs-research-baseline drift qualitatively; this discipline
  measures behaviour quantitatively. Neither substitutes for the other.
- **Related:** [itd-14](../drafts/itd-14-prompt-registry-versioning.md) — corpus
  scores are the natural payload of a prompt registry entry.

## Audit Notes

_Empty. Populated by `intent-fidelity-reviewer`'s single-document role when this
discipline is first audited. Like itd-1 and itd-5, audited continuously via
rule-applies-to-every-judge-agent-spec semantics rather than a planned→shipped
transition._

## References

- [`../../research/notes/2026-07-12-judge-calibration-sota.md`](../../research/notes/2026-07-12-judge-calibration-sota.md) — the survey this discipline is drawn from, including the citation-confidence caveat.

[overcorrection]: https://arxiv.org/html/2603.00539 "Are LLMs Reliable Code Reviewers? Systematic Overcorrection in Requirement Conformance Judgement (Automated Software Engineering, 2026)"
[judge-codegen]: https://arxiv.org/pdf/2507.16587 "On the Effectiveness of LLM-as-a-judge for Code Generation and Summarization"
[criticgpt]: https://arxiv.org/pdf/2407.00215 "LLM Critics Help Catch LLM Bugs (CriticGPT), OpenAI"
[qodo]: https://www.qodo.ai/blog/how-we-built-a-real-world-benchmark-for-ai-code-review/ "Qodo — building a real-world benchmark for AI code review"
[validators]: https://arxiv.org/pdf/2404.12272 "Who Validates the Validators? (UIST 2024) — criteria drift"
[position-bias]: https://arxiv.org/pdf/2602.02219 "Am I More Pointwise or Pairwise? Position Bias in Rubric-Based LLM-as-a-Judge"
[speak-freely]: https://arxiv.org/html/2408.02442v1 "Let Me Speak Freely? Impact of Format Restrictions on LLM Performance"
[ace]: https://arxiv.org/pdf/2510.04618 "Agentic Context Engineering — brevity bias, context collapse"
[gepa]: https://arxiv.org/abs/2507.19457 "GEPA: Reflective Prompt Evolution Can Outperform Reinforcement Learning (ICLR 2026 oral)"
[reward-hacking]: https://openreview.net/forum?id=ikrQWGgxYg "Reward Hacking in Self-Improving Code Agents"
[nine-judges]: https://arxiv.org/html/2605.29800 "Nine Judges, Two Effective Votes: Correlated Errors Undermine LLM Evaluation Panels"
[mar]: https://arxiv.org/html/2512.20845v1 "MAR: Multi-Agent Reflexion — degeneration-of-thought in single-agent reflection"
