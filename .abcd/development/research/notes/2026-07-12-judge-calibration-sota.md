# Judge calibration and self-improving agents — SOTA survey

**Date:** 2026-07-12
**Scope:** How to author and measure LLM reviewer/judge prompts, and what
"self-improving agent" means for a solo developer with no training access.
**Seeds:** [itd-81](../../intents/disciplines/itd-81-judge-calibration.md) (discipline).
**Status of citations:** gathered by the `sota-researcher` agent in a single
run; the load-bearing claims are cross-corroborated, but individual sources
below were **not re-opened by hand**. Treat a specific figure as indicative
until verified at the source. (This gap is why the agent now carries a
cite-only-what-you-opened rule.)

---

## 1. The finding that inverts our current posture

Our reviewer prompts are written to maximise suspicion. The evidence says LLM
code judges already over-flag, and that pushing them harder makes them worse.

- LLM judges **over-predict errors** relative to human experts on requirement
  conformance; a third of judge errors trace to **statements hallucinated into
  the code** that are not there. Critically, prompts that **demand explanations
  and proposed corrections produced higher misjudgment rates** ([overcorrection]).
- Judges **underestimate the correctness of human-written code** and
  **overestimate LLM-written code** ([judge-codegen]) — a self-preference bias
  that matters directly for abcd, where the model reviewing the diff is
  routinely the model that wrote it.
- Vendor false-positive numbers (the "5–15% FP", "54% in some configurations"
  figures) trace to marketing posts with no published methodology. The
  *direction* is well-supported; the digits are noise ([deepsource-benchmarks]).

**Consequence for abcd.** A reviewer's **true-negative rate is a first-class
metric**, not an afterthought. A reviewer that never returns "nothing to fix"
carries no information. Clean diffs belong in the fixture corpus as first-class
cases with the expected verdict `SHIP`.

## 2. Ground truth has to be manufactured

Natural bugs are unlabelled, so the field constructs labels by **injecting**
them: OpenAI's CriticGPT was trained and evaluated on deliberately inserted
bugs, with critiques preferred 63% of the time on naturally-occurring bugs, and
the human+critic team hallucinating less than the critic alone ([criticgpt]).
Code-review benchmarks use the same construction — take merged PRs, inject 1–3
functional bugs, measure precision/recall ([qodo]).

**Consequence for abcd.** Our closed issues are free labelled negatives. Every
`fix:` commit is a known-bad diff (the pre-fix state) paired with a known-good
one (the merged state), with the defect class already named in the issue.
`iss-29` (fail-closed capture), `iss-30` (partial-failure reporting, CRLF
parity) are corpus items we already own and have never used as such.

## 3. The rubric cannot be written in advance

**Criteria drift** ([validators], UIST 2024): you need criteria to grade
outputs, but grading outputs is what teaches you the criteria. Evaluation
criteria are not fully specifiable a priori. The practitioner consensus
([evals-faq]) is error analysis — open-code your judge's actual outputs until
~20 traces yield no new failure category; that taxonomy *is* the rubric. The
same source warns that **eval-driven development — writing the evaluator before
the implementation — backfires**, because LLM failure surface is unbounded.

**Consequence for abcd.** itd-1's Given-When-Then acceptance gates remain right
for *capabilities*. They are the wrong instrument for *judge quality*, which
must be derived from observed failures, not declared.

## 4. Binary, and formatted after reasoning

- **Binary labels beat Likert scales** ([evals-faq]). Numeric severity drifts to
  the middle, needs larger samples for significance, and carries position bias
  even in pointwise rubric scoring ([position-bias]).
- **Reasoning inside a JSON schema degrades reasoning** ([speak-freely]) —
  format restriction measurably hurts, and JSON mode failed to respect
  reason-then-answer ordering. The fix is two-stage: free prose, then a
  structured block.

**Consequence for abcd.** Our SHIP / FIX-FIRST split is already the right
shape — keep it binary and resist adding a 1–5 scale. Agent prompts should emit
`## Analysis` (prose) then `## Findings` (structured), and tooling should parse
only the second.

## 5. What "self-improving" actually means without training access

The only loop that pays off on a laptop is **eval-gated context editing**: a
graded corpus, a deterministic verifier, and a playbook that accumulates
deltas.

- **Verifier-gated proposal is the universal mechanism.** AlphaEvolve's real
  gains (~0.7% of Google fleet compute recovered) come from automated
  evaluators that verify answers ([alphaevolve]); DGM's SWE-bench 20%→50% is
  gated by benchmark execution ([dgm]). Self-improvement is reliable **only**
  where outcomes are objectively verifiable. abcd has `make preflight`; the
  move is to make it a *precondition* of the reviewer, not a sibling check.
- **Append-only playbooks, never rewrites.** Agentic Context Engineering
  ([ace], +10.6% on agent benchmarks) names the two failure modes that destroy
  hand-maintained instruction files: **brevity bias** (optimisers and tidy
  humans converge on short generic instructions, dropping the domain detail
  that was load-bearing) and **context collapse** (iterative full rewrites
  progressively erode information). The fix is structural: itemised append-only
  deltas curated by a separate role.
- **Reflective prompt evolution works, as an algorithm.** GEPA ([gepa], ICLR
  2026 oral) beats RL fine-tuning with ~35× fewer rollouts by reflecting on
  execution traces in natural language; a reported industrial use lifted an
  LLM-judge prompt from 68.9% → 88.9% accuracy. Adopting DSPy would mean a
  Python toolchain and new dependencies in a Go repo — an anti-recommendation.
  The mechanism transfers without the library: score prompt on corpus → feed
  *failing* traces to a reflector → request a minimal delta → re-score → keep
  only if TNR improves without recall collapsing.
- **Do not run it unattended.** "Reward Hacking in Self-Improving Code Agents"
  ([reward-hacking]): **73.8% of KernelBench and 46.8% of ALE-Bench
  "optimisations" showed proxy gains with no real-task gains.** An agent
  optimising against our own reviewer learns to pass the reviewer, not to write
  good code. Human approval on every merged delta.
- **The reflector must not be the actor.** 2025–26 replications of Reflexion
  report that one model generating, evaluating, and reflecting produces
  confirmation bias and re-commits to the same flawed chain ([mar]). Reflection
  requires a separate agent, ideally a different model family.

## 6. Panels: contested, and we should not build one

- **For:** Cohere's PoLL ([poll]) — three small models from disjoint families
  beat a single large judge across six datasets at ~7× lower cost, with less
  intra-model bias.
- **Against:** a nine-judge panel across seven model families delivered only
  **2.18 effective independent votes** (24% of nominal), mean pairwise error
  correlation φ = 0.391, and **matched or underperformed the single best judge**
  on all three datasets ([nine-judges]).

**Reading.** Panels buy bias cancellation, not accuracy-through-independence —
the errors are correlated, so the Condorcet argument does not hold. Multi-voting
a biased judge yields the same bias with tighter error bars. **Use cross-family
disagreement as a triage signal on contested findings, not as a jury.**

## 7. Investigated and rejected for abcd's scale

| Practice | Why not |
|---|---|
| Darwin Gödel Machine / self-referential self-modification | Real ([dgm]) but needs a benchmark harness and GPU-weeks per lineage; its ceiling still sat below the best hand-built agent of the time. Take the archive+verifier+selection pattern, discard the machinery. |
| AlphaEvolve-style evolutionary search | Requires a machine-checkable scalar objective (kernel latency, packing density). "Is this Go code well-reviewed?" is not one. |
| Process reward models / RFT / trace distillation | Require training access and thousands of graded steps. Categorically unavailable. |
| 1–5 severity scores from the judge | Noisy, middle-drifting, position-biased ([position-bias]). |
| Reasoning inside a JSON schema | Measurable reasoning degradation ([speak-freely]). |
| Find + explain + propose-fix in one pass | Measurably *increases* misjudgment ([overcorrection]). Separate find from fix. |
| Generic metrics ("helpfulness", ROUGE/BERTScore) | "Generic evaluations waste time and create false confidence" ([evals-faq]). |

## 8. Where this collides with the existing record

- **itd-5's self-improvement pre-flight is unsafe as written.** It accepts an
  oracle-rewritten prompt when it "passes the same goldens and is shorter by
  >10%". That rule **selects for brevity bias** — the exact mechanism [ace]
  identifies as destroying instruction quality — and it does so against goldens
  that (per §1) do not currently measure false positives at all. A shorter
  prompt that passes the same goldens may simply have shed the domain detail
  the goldens never tested. The >10%-shorter tiebreak should be struck, and the
  gate should be a TNR/recall score on the calibration corpus.
- **itd-5's golden fixtures are not a calibration corpus.** They establish that
  an agent does something; they do not measure how often it cries wolf.
- **itd-64 (benchmark-driven config optimisation) is the reward-hacking risk
  zone.** Its "learn which configuration produced the better outcome" loop is
  precisely the proxy-optimisation shape [reward-hacking] measured at 73.8%
  spurious. It needs a human-approval gate and a held-out split before it tunes
  anything, and it should not treat reviewer verdicts as ground truth — the
  reviewer is the thing under measurement.
- **itd-15 (self-dogfooded SOTA audit)** stays qualitative: prompt-vs-baseline
  drift. It is complementary to, not a substitute for, a measured corpus.

## References

[overcorrection]: https://arxiv.org/html/2603.00539 "Are LLMs Reliable Code Reviewers? Systematic Overcorrection in Requirement Conformance Judgement (Automated Software Engineering, 2026)"
[judge-codegen]: https://arxiv.org/pdf/2507.16587 "On the Effectiveness of LLM-as-a-judge for Code Generation and Summarization"
[criticgpt]: https://arxiv.org/pdf/2407.00215 "LLM Critics Help Catch LLM Bugs (CriticGPT), OpenAI"
[qodo]: https://www.qodo.ai/blog/how-we-built-a-real-world-benchmark-for-ai-code-review/ "Qodo — building a real-world benchmark for AI code review (vendor; method sound)"
[deepsource-benchmarks]: https://deepsource.com/blog/ai-code-review-benchmarks "DeepSource — every AI code review vendor benchmarks itself, and wins"
[validators]: https://arxiv.org/pdf/2404.12272 "Who Validates the Validators? Aligning LLM-Assisted Evaluation with Human Preferences (UIST 2024) — criteria drift"
[evals-faq]: https://hamel.dev/blog/posts/evals-faq/ "Husain & Shankar — LLM Evals FAQ (Jan 2026)"
[position-bias]: https://arxiv.org/pdf/2602.02219 "Am I More Pointwise or Pairwise? Position Bias in Rubric-Based LLM-as-a-Judge"
[speak-freely]: https://arxiv.org/html/2408.02442v1 "Let Me Speak Freely? Impact of Format Restrictions on LLM Performance"
[ace]: https://arxiv.org/pdf/2510.04618 "Agentic Context Engineering: Evolving Contexts for Self-Improving Language Models — brevity bias, context collapse"
[gepa]: https://arxiv.org/abs/2507.19457 "GEPA: Reflective Prompt Evolution Can Outperform Reinforcement Learning (ICLR 2026 oral)"
[reward-hacking]: https://openreview.net/forum?id=ikrQWGgxYg "Reward Hacking in Self-Improving Code Agents"
[alphaevolve]: https://arxiv.org/abs/2506.13131 "AlphaEvolve: A coding agent for scientific and algorithmic discovery (DeepMind)"
[dgm]: https://arxiv.org/pdf/2505.22954 "Darwin Gödel Machine: Open-Ended Evolution of Self-Improving Agents (Sakana AI)"
[mar]: https://arxiv.org/html/2512.20845v1 "MAR: Multi-Agent Reflexion — degeneration-of-thought in single-agent reflection"
[poll]: https://arxiv.org/html/2404.18796v1 "Replacing Judges with Juries: PoLL (Cohere)"
[nine-judges]: https://arxiv.org/html/2605.29800 "Nine Judges, Two Effective Votes: Correlated Errors Undermine LLM Evaluation Panels"
[voyager]: https://arxiv.org/abs/2305.16291 "Voyager: An Open-Ended Embodied Agent with LLMs — skill library, admitted on verification"
