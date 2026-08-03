# SOTA: abcd as a research platform / benchmark authority

Desk research into extending abcd toward authoritative benchmarking — (a) enhancing
implementation quality autonomously, (b) assuring a product thinker that what shipped
matches the brief and intents they can read. Recency anchored to 2026-07-14.

**Outcome: the benchmark-from-our-own-data idea is not salvageable. A different, stronger,
and more timely contribution is sitting right next to it.**

## 1. Correction to the prior record

The [cross-model review research](2026-07-14-cross-model-review.md) claimed no published
precision/recall exists for spec-conformance. **That was wrong**, and the error is recorded
here rather than quietly fixed.

- **A peer-reviewed requirement-conformance benchmark exists**: *"Are LLMs Reliable Code
  Reviewers? Systematic Overcorrection in Requirement Conformance Judgement"*
  (arXiv 2603.00539, Springer *Automated Software Engineering*). The task is exactly ours:
  binary *"does this code meet this natural-language requirement?"*, no test execution,
  ~1,400 instances (HumanEval-X-Bugs, Buggy-MBPP, QuixBugs). **Publishing without citing it
  is a desk rejection.**
- **SpecBench does not "defer" conformance to future work.** It scopes code out
  permanently — *"Agents produce no code when evaluating on SpecBench"* — and its roadmap is
  spec → revised spec. The earlier claim put words in its authors' mouths.

**The gap that remains is real but narrow:** everything published is **function-level and
synthetic** (standalone functions, injected bugs). Nobody benchmarks *a real design
document + a real multi-file diff → does this satisfy acceptance criterion #3.*

## 2. The finding worth more than any benchmark we could build

From 2603.00539, measured across ~1,400 paired instances:

> **Asking a reviewer to explain and propose a fix is what manufactures false
> non-conformance verdicts.**

- GPT-4o false-negative rate: **26.2% → 73.2%** (HumanEval) and **35.9% → 87.9%** (MBPP),
  moving from a *direct* verdict prompt to a *full* explain-and-fix prompt.
- *"FP drops from 19 under Direct to 0 under Direct+Explain, but the same change more than
  doubles false rejection of correct code (FN 184 → 451 on MBPP)."*
- **93%** of GPT-4o's self-contradictions are "No verdict with positive rationale".
- Models match *symptoms* 94–100% but *root-cause bug type* only 44–75%.

**Every abcd reviewer currently does the forbidden thing** — verdict + explanation +
suggested fix in one call.

**Design consequence.** Split the conformance review into a **terse binary verdict call**
and a **separate explanation call**, taken only when the verdict is non-conformant. This is
a prompt change. It needs no benchmark, no telemetry, no ethics review, and it is the only
item in this research with an effect size large enough to matter at our scale.

**Their mitigation is also cheap and we have the pieces.** The *fix-guided verification
filter* treats the model's proposed fix as a **counterfactual to execute**: run the original
and the "fixed" version against the tests; if the fix changes nothing observable, the
rejection was hallucinated. FNR **54.8% → 16.3%** (HumanEval), **69.0% → 28.9%** (MBPP),
**51.0% → 24.0%** (QuixBugs). Depends on test coverage.

## 3. The field just blew up — verified against the primaries

**Both figures below were read first-hand from OpenAI's own posts** (the earlier secondary
sourcing was wrong on one material point — see § 3.1).

**SWE-bench Verified** — *"Why SWE-bench Verified no longer measures frontier coding
capabilities"*, 2026-02-23:
- Audited a **27.6% subset** of often-failed problems; **"at least 59.4% of the audited
  problems have flawed test cases that reject functionally correct submissions."**
- The audit covered **138 problems** o3 failed to solve consistently over 64 runs, each
  reviewed by **≥6 experienced software engineers**.
- Progress had stalled: **74.9% → 80.9% in six months.**
- On contamination: *"models that have seen the problems during training are more likely to
  succeed, because they have additional information needed to pass the **underspecified
  tests**."* **Contamination and underspecification interact** — vagueness is what makes
  memorisation pay.

**SWE-bench Pro** — *"Separating signal from noise in coding evaluations"*, 2026-07-08:
- **731-task public split**; frontier models **23.3% → 80.3% in eight months.**
- Automated pipeline flagged **200 (27.4%)**; a **five-engineer** human campaign flagged
  **249 (34.1%)**; initial filter surfaced **286** candidates. OpenAI estimate **~30% of
  tasks are broken.**

### 3.1 CORRECTION — underspecification is NOT the dominant defect

An earlier draft of this research claimed *"35.5% of tasks demanded implementation details
never stated in the problem"* and framed underspecification as the killer. **That figure
does not appear in the primary, and the framing is wrong.** OpenAI's actual taxonomy, as a
percentage of the full dataset (agent-flagged / human-flagged):

| Issue type | Agent | Human |
|---|---|---|
| **Overly strict tests** | **14.4%** | **17.8%** |
| Low-coverage tests | 4.1% | 9.4% |
| Misleading prompt | 6.3% | 7.5% |
| Miscellaneous | 1.9% | 1.2% |
| **Underspecified prompt** | **0.6%** | **0.8%** |

*(The agent column sums to 27.3%, matching the reported 27.4% — the ordering is confirmed.)*

**"Underspecified prompt" is the smallest category. The dominant defect, by roughly 20×, is
"overly strict tests".**

### 3.2 The reframed — and stronger — claim

OpenAI's definition of the dominant defect is the key sentence:

> *"**Overly strict tests** enforce specific implementation details **not specified in the
> prompt**, invalidating many functionally correct submissions."*

This **is** a specification failure — but the blame sits on the **oracle**, not the prompt.
The disease is: **the ground truth demands things the specification never stated.**

- SWE-bench's oracle is a merged PR's test suite, **authored independently of, and after,
  the task statement.** It can therefore drift arbitrarily far from what the task asked.
  That is *structurally* why 14–18% of tasks punish functionally correct code.
- In abcd, **the acceptance criteria ARE the oracle**, and they are written before the code.
  **The oracle cannot exceed the specification, because it IS the specification.** This
  failure mode is structurally excluded, not merely reduced.

**The honest flip side — abcd's own structural risk is the mirror image.** With no
executable oracle, criteria can be vague and the judge fills the gap. That is the
**low-coverage** failure (4–9% in SWE-bench Pro: *"incomplete fixes can pass"*). It compounds
with two biases already on record: judges over-accept AI-written code by up to **1.91×**, and
abcd's judge shares a family with the author.

**Therefore the claim is not "our specs are better."** It is:

> **A spec-derived oracle cannot over-reject; it can only under-check — and under-checking
> is measurable.**

Falsifiable, testable, and a genuinely more interesting position than the one it replaces.

*(A methodological gift is buried in the Pro audit: the automated and human passes disagreed
by ~7 points — de facto evidence that automated task-quality auditing **under-detects**
relative to humans. Any instrument we build must report its own gap against human review.)*

## 4. Why a benchmark from abcd's own data cannot work

### 4.1 The corpus does not exist
Harvestable (specification, diff, conformance verdict) triples **today: 2** — and one has no
diff attached. Across all 13 criterion-level labels: ~8 MET, 2 MET_WITH_CONCERNS, **1
NOT_MET**, 1 INCONCLUSIVE. A classifier that always predicts MET scores ~85%. Both verdicts
were produced by Claude Opus 4.8 on code Claude Opus 4.8 wrote, with **no human
adjudication and no second judge**. One grades criteria MET on the grounds that *the act of
grading demonstrates the criterion*.

The **unlabelled** pool is the real asset: **80 intents, 440 Given/When/Then criteria, 216
commits, 67 merged PRs**, 59 commits tagged `itd-N`. Strong specification substrate, strong
implementation substrate, **no independent labels.**

### 4.2 Circularity is the objection that sinks it
The system would (a) generate the specs, (b) produce the code, (c) judge conformance, and
(d) supply the labels. *"Benchmarking is Broken — Don't Let AI be its Own Judge"*
(arXiv 2510.07575) names this directly; *The Leaderboard Illusion* (arXiv 2504.20879,
NeurIPS 2025) has already primed the community to treat a benchmark authored by a party with
a stake in the result as prima facie suspect. **abcd would be a benchmark authored by the
tool whose workflow it validates.** Same shape.

Self-preference bias appears to be driven by **perplexity** — judges over-reward text that
is familiar/low-perplexity to them — which means it **survives swapping to a different model
in the same family or one trained on similar data.**

### 4.3 N=1 has no methodological remedy
Epoch AI attacks **SWE-bench Verified** for concentration: **Django alone is nearly half the
issues; five repositories are over 80%** of the 500. If a 12-repo, 500-issue,
Princeton-authored benchmark is rejected for low diversity, a **one-person, one-repo**
corpus does not survive contact with a reviewer. Selection bias here is not a limitation to
disclose — **it is the entire dataset.**

### 4.4 The arithmetic does not close
- **Corpus size.** LiveCodeBench states that at n≈349, **1–1.5% performance differences are
  noise**. SWE-MERA's living-benchmark funnel yields **0.27%** of candidates (110K repos →
  300 tasks). At 10 usable tasks/month, a two-year runway reaches 240 — **below the noise
  floor for distinguishing frontier models.**
- **In-situ A/B is undecidable.** Detecting a small effect needs **2,200–15,500 paired
  observations**. Solo development produces a few hundred reviewable diffs a year.
- **Telemetry cannot rescue it.** Go's transparent telemetry needs **~16,000 reports/week**
  for 1% accuracy at 99% confidence. And **Audacity's opt-in, default-off telemetry still
  triggered a revolt that killed it**; .NET's opt-out drove Alpine/Arch/Fedora/Homebrew to
  patch it out permanently. The backlash tracks **perceived betrayal, not data volume**.

### 4.5 "Disagreement = defect" is not demonstrated
The most principled version (MUSE, multi-LLM uncertainty via Jensen-Shannon divergence)
reports **AUROC ≈ 0.62** — barely better than a coin flip at the thing we care about, and it
*assumes* the disagreement↔defect link rather than proving it.

**Design consequence.** Disagreement is defensible as a **triage flag** ("show me the diffs
where configs disagreed") and **indefensible as a quality metric** ("config B is better
because it disagrees more").

## 5. The version that survives — and it is stronger

**Drop "authoritative benchmark". The asset is not data; it is a process that produces
specifications before code — and the field has just publicly announced, twice, that missing
specifications are what kills benchmarks.**

### Step 1 — A task-quality instrument, validated against an external reference
Claim: *acceptance criteria authored **before** the code exists eliminate the
underspecification defect that caused ≥35.5% of SWE-bench Verified's audited task failures
and ~30% of SWE-bench Pro's.*

Test it by **applying the instrument to someone else's benchmark**: audit SWE-bench Pro's
**731 tasks** and compare our flag rate against OpenAI's published **27.4% (automated)** and
**34.1% (human)**. That is falsifiable, externally validated, needs **no new corpus and no
telemetry**, and lands in an open controversy with a ready-made reference standard.

**SWE-bench's ground truth is retrospective** — the task statement is derived *after the
fact* from a GitHub issue, which is why 35.5% of tasks demand implementation details the
issue never mentioned. **An issue was never a specification.** abcd's intents are written
*before* the diff exists. **That is the defensible novelty, and it targets the field's
acknowledged number-one defect.**

**Precondition:** pre-register each intent's acceptance criteria with a **timestamp/hash at
spec time, before any code is written**. Without that commitment artefact we cannot prove the
criteria were not back-fitted, and the claim collapses. *(abcd already hashes the AC section
into `receipt_id`, excluding `## Audit Notes` — the primitive exists.)*

### Step 2 — Only then, a corpus, and only with a real oracle
**Ground truth must come from outside the tool.** Where there is no test oracle, the next
most defensible thing is **not an LLM judge** — it is a **retrospective event oracle**: was
the diff reverted? did a later commit fix it? did an armed detector fire? *(abcd's
post-review recording workflow is already this shape.)*

Required from day one: a **release-date field per task** (contamination-resistance is nearly
free for us — our data is unpublished and post-cutoff; LiveCodeBench showed a model drop
~60% → ~0% across its cutoff), a **datasheet**, **Croissant metadata** (missing/invalid =
**desk rejection** at NeurIPS 2026), an ML-host artefact, and a **published label-error
rate**.

### Step 3 — N=3 before anything is called a benchmark
**Aider polyglot is the existence proof**: a single developer's tool became an authority
tracked by Epoch and cited by model providers — **because it did not use its own data.** It
wrapped **225 external, human-authored Exercism exercises with unit tests.** *The tool
supplied the harness and the distribution; the ground truth came from outside.* That is the
template.

*Copilot Arena* is the warning: a genuine benchmark from an IDE's own telemetry needed
**1,642 users and 11,604 votes**, and found in-the-wild rankings correlate with static
benchmarks at **Spearman ≤ 0.1**.

## 6. The review we would get (write the paper against this)

1. *"The dataset is one person's project."* No remedy but more contributors.
2. *"The ground truth is the system's own output."* Specs, diffs and verdicts all from the
   same pipeline, judged by the same model family. **Not independent measurement.**
3. *"The construct is undefined."* "Does this diff satisfy this criterion" is exactly the
   contested construct **47.8%** of 445 surveyed benchmarks failed to operationalise
   (arXiv 2511.04703, 29 domain experts).
4. *"No uncertainty quantification."* Only **16.0%** of benchmarks report it. We would be
   held to the 16%.
5. *"No label reliability evidence."* No κ, no second annotator, no adjudication protocol.
   And note SE papers routinely report the **wrong coefficient** (agreement vs reliability) —
   a free hit for a reviewer.
6. *"Conflict of interest."* Post-*Leaderboard Illusion*, live and fashionable.
7. *"Task quality is unaudited."* Both retracted benchmarks died here. An OpenAI-style audit
   (automated pass + ≥5 independent engineers, reporting **both** flag rates **and their
   disagreement**) is now **table stakes**.
8. *"Croissant/RAI metadata missing."* **Desk reject, before anyone reads point 1.**

## 7. Not worth adopting

- **A leaderboard or pass@k score.** Below ~500 oracle-backed tasks it cannot resolve model
  differences. We will not have 500 for years.
- **LLM-as-judge as the primary label source.** κ ≈ 0.77–0.79 against humans *at best*, with
  measured self-preference, where the judge shares a family with the generator. Fine as a
  pre-filter; **fatal as ground truth**. SWE-MERA uses an LLM only as the *last* filter,
  after build validation and end-to-end execution — a garbage-collector on top of a real
  oracle, never the oracle.
- **Transmitted telemetry.** Cannot be powered at our scale; downside is the project's
  reputation. **Local-only recording gets 100% of the usable signal and 0% of the risk.**
- **A significance-tested in-situ A/B of review configs.** Not reachable from solo
  development. Reframe as *seeded-corpus offline eval* + *in-the-wild case archaeology*.
- **EARS / Gherkin as a machine-checkable oracle.** **Zero evidence found.** Adopt EARS if it
  helps humans write intents; **do not claim a conformance benefit.** The BDD literature
  *generates* Gherkin; nobody has shown it *judges* code reliably.
- **Formal specs (TLA+/Dafny/Verus) as the target.** The bottleneck is *writing* the spec,
  not checking it — best repo-level executable-spec generation is **20.2%**.
- **Citing SWE-bench Verified or Pro as a comparator without qualification.** Both retracted.

## 8. Decisions taken (2026-07-14)

1. **Split the conformance reviewer into a terse binary verdict call and a separate
   explanation call.** Highest-value change in this research; independent of everything else.
2. **Reject a benchmark built from abcd's own data as ground truth.** Circularity, N=1, and a
   corpus of 2 are each individually fatal.
3. **The contribution, if pursued, is a task-quality instrument + methodology, validated
   against an external benchmark** (SWE-bench Pro's 731 audited tasks), not a corpus.
4. **Ground truth must originate outside the tool** — an executable test or a retrospective
   event oracle, never an LLM judge.
5. **Any recording is local-only.** No transmitted telemetry.
6. **Pre-register acceptance criteria (timestamp + hash at spec time)** — without it the
   prospective-spec claim is unprovable.
