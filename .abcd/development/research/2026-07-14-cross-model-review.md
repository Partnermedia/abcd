# SOTA: cross-model adversarial review

Desk research into a proposed "super-reviewer" — a verb that, when a second model
backend is available, reviews every commit against the brief and the implementation
against its spec. Recency anchored to 2026-07-14.

**Outcome: the premise dissolves, favourably.** The debiasing the idea exists to buy is
achievable with **no second model, no second subscription, and no optional dependency** —
which also removes the two-tier UX problem the idea was built around.

## 1. The motivating rationale does not survive

The idea's premise: the reviewing model is routinely the authoring model, so it is soft on
its own work. itd-81 states this too, and it is the right instinct. But the 2026 evidence
splits it into a real effect and a false one.

- **Self-preference under genuine authorship is weak or absent.** Four model families, 85
  author-vs-fresh comparisons against a **deterministic verifier**: **no detectable
  self-preference**, gap **−5.1 pp, 95% CI [−12.9, +2.7]** (arXiv 2606.20093). Where authors
  *do* reject their own verified-good fix, **97%** of stated reasons are flaw-catching, not
  preference.
- **The bias that DOES exist on code is a context-framing artefact, not model identity.**
  Across four coding and tool-use datasets, monitors under-report high-risk actions **when
  the action appears in a previous assistant turn**, and stop doing so when the *same*
  action is presented neutrally as a **user-turn artefact of unknown authorship**
  ("off-policy attribution") — arXiv 2603.04582.

**Design consequence — the highest-value finding here.** Re-present the diff in a **fresh
session, as a user-supplied artefact of unknown authorship, same model**. That captures the
only *measured* debiasing effect, costs nothing, and is available to every user. A different
model family is a second-order refinement on top of it, and **no one has published a number
for cross-family benefit on code review specifically.**

**The widely-circulated "cross-model review catches 40–60% more issues" figure is
blogspam** — traced to an "Editorial Team" post citing a vendor blog, with no study, no
sample size, and no control group. It must not appear in a design doc.

> **Provenance warning.** While fetching one paper, PDF extraction **fabricated** a finding
> supporting cross-family review that appears nowhere in the source; the abstract says the
> opposite. It was caught only by re-fetching. Verify any cross-family claim against an
> abstract you have opened.

## 2. "Super" is the one shape the evidence forbids

- **Panels do not beat the best single judge.** Nine frontier judges, seven families →
  **n_eff ≈ 2.18**; panel accuracy falls **8–22 points short** of what independent voting
  would give; the **best single judge matches or beats the full panel in every condition**;
  smarter aggregation closes ≤11% of the gap *even given the correct answers*
  (arXiv 2605.29800).
- **On code, consensus actively hurts — the "popularity trap".** Ten models, five families:
  the cross-family ensemble's *oracle* upside is large (Defects4J +83% over the best single
  model), but **majority/consensus selection underperforms the naive baseline**, because
  models converge on the same syntactically plausible, semantically wrong answer and filter
  out the minority-correct one (arXiv 2510.21513).
- **The gains are *selection* gains, and review has no oracle.** Code repair can cash them
  because tests adjudicate. **Code review cannot.** A second model yields a second *opinion*,
  and the literature says explicitly not to resolve disagreement by consensus.
- Multi-agent debate fails to beat single-agent CoT + self-consistency at far higher compute
  (arXiv 2502.08788); heterogeneity is the only lever that helps, and its magnitude is
  unverified.

**This is already abcd's position.** itd-81's out-of-scope section reaches the same
conclusion from the same evidence: *"Cross-family disagreement is a triage signal on
contested findings, not a voting scheme."*

## 3. Precision is the product, and the real numbers are grim

| Source | Finding |
|---|---|
| **Factory.ai benchmark** — 50 real PRs, human-curated golden bug sets, 3 runs/model, disclosed methodology | Frontier precision **47.5–65%**. A third to a half of findings are noise. |
| **Google AutoCommenter** (arXiv 2405.13565) — tens of thousands of devs, production | Posts on only **3.9% of opportunities**. Hit 80% usefulness *only* after suppressing 22 non-actionable rule classes (54% → 66% → target). **40%** of posted comments resolved. *"Even a few negative user experiences can erode trust."* |
| **Does AI Code Review Lead to Code Changes?** (arXiv 2508.18771) — 22,326 AI comments, 178 repos, 16 tools | Addressing rate: **humans 60%**, best AI tool **19.2%**, worst **0.9%**. **15.4–15.7% of AI comments outright invalid.** |
| **Automated Code Review In Practice** (arXiv 2412.18531) — 238 practitioners, 4,335 PRs | 73.8% of comments resolved, but **mean PR closure time rose 5h52m → 8h20m**. AI review made the process *slower*. |
| Anthropic Code Review blog | Claims **<1% of findings marked incorrect** — but publishes **no methodology**, and "marked incorrect" measures how often busy engineers click a button. Its **dedicated verification agent that re-checks candidate findings before posting** is the architecturally load-bearing part. |

**Design consequence.** **Silence is the default.** Any reviewer needs a suppression list and
a confidence threshold from day one, must be allowed to say nothing, and must have its
**addressing rate instrumented** — killed if it falls below ~20%.

## 4. Per-commit is the worst available unit

- **Nobody credible does it.** CodeRabbit, Greptile, Cursor BugBot and Anthropic Code Review
  are all **per-PR**. The 16 GitHub Actions studied are PR-, file-, or hunk-level.
- **Findings on WIP commits are false positives by construction** — they flag what the next
  commit fixes. This is the single most noise-generating choice available.
- **Cost.** Anthropic publishes **$15–25 per review**. A 5–20-commit branch reviewed
  per-commit is **$75–500 per PR-equivalent**, for strictly worse signal.
- The one genuine signal in favour of *tight scope*: hunk-level comments are addressed
  **43.88%** of the time vs **13.89%** file-level. That argues for reviewing **small diffs**,
  not for reviewing every commit.

**Design consequence.** The unit is the **branch diff / PR**. Per-commit, if offered at all,
is explicit and on demand — never an automatic gate.

## 5. Spec-conformance is research, not a gate

- **SpecBench** (arXiv 2605.30314), built from real RFC processes (Kubernetes, React, Rust,
  TVM, vLLM): the best agent scores **44.4%** — and SpecBench tests only the *easier*
  adjacent task, **identifying deficiencies in a specification**. It explicitly **does not**
  evaluate whether an implementation conforms to a spec.
- The closest industrial work (arXiv 2605.17926) **publishes no precision, no recall, and no
  false-positive rate**, and makes no claim of gating suitability.
- ~~**No published precision/recall exists anywhere for "does this diff match the brief".**~~
  **SUPERSEDED 2026-07-14 — this was WRONG.** A peer-reviewed benchmark exists:
  arXiv **2603.00539** (Springer *Automated Software Engineering*), *"Are LLMs Reliable Code
  Reviewers? Systematic Overcorrection in Requirement Conformance Judgement"* — binary
  "does this code meet this natural-language requirement?", ~1,400 instances.
  **Publishing without citing it is a desk rejection.** Also: SpecBench does not "defer"
  conformance to future work — it scopes code out **permanently** (*"Agents produce no code
  when evaluating on SpecBench"*). The gap that remains is narrow but real: everything
  published is **function-level and synthetic**; nobody benchmarks *a real design document +
  a real multi-file diff*. See
  [`2026-07-14-research-platform-benchmarks.md`](2026-07-14-research-platform-benchmarks.md) § 1.

**The one transferable pattern.** Do **not** prompt a model with brief + diff and ask "does
this conform?". **Extract a structured intermediate artefact of checkable rules from the
brief first — explicitly separating the ambiguous and non-verifiable statements — then check
the code against those rules.**

**Design consequence.** Spec-conformance ships **advisory only**. Gating a merge on a
~44%-accurate signal costs more in false blocks than it saves in caught drift.

## 6. The optional-dependency gate is a structural error

A gate that no-ops when the second backend is absent is **fail-open**, and fail-open is
operationally indistinguishable from **no enforcement** for the majority of users who never
install it. Worse, contributors will *believe* the gate ran. It would make the convention
"every commit is adversarially reviewed" true for one machine and false everywhere else.

This is `loud-staging` in the record: *"an explicit not-wired refusal, **never a stub that
half-works**."*

**Design consequence.** If it degrades to nothing it is **not a gate** — it is a local
enhancement and must be named one. The fresh-context reviewer (§1) is the always-available
floor that removes the need for the tier entirely.

## 7. Where the idea is genuinely novel

Neither practitioner thread surveyed mentions **spec/brief-conformance review** or
**cross-model review** at all. That cuts both ways: it is not a reinvented wheel, and there
is **no field evidence that it works or that anyone wants it**.

What the proposal has that the noise-plagued incumbents lack is **a reference document to
check against** — a plausible answer to the dominant practitioner complaint that AI
reviewers do not understand intent or architecture. That is worth **prototyping as advisory
output on branch diffs**, measured by addressing rate, long before it goes near a gate.

## 8. Decisions taken (2026-07-14)

1. **Fresh-context, off-policy, same-model reviewer** is the design. No second subscription,
   no optional dependency, no two-tier UX.
2. **A second model, if present, is used for disagreement-as-triage — never voting.** This
   is adr-25's asymmetric trust (scoped gates, broad is mined) and itd-81's out-of-scope,
   both already decided.
3. **Advisory + receipt; a deterministic gate decides.** Per
   `verifier-selects-gates-decide`: an LLM verdict *"may advise a gate — flag, rank,
   annotate, propose — but the gate's blocking decision stays deterministic."*
4. **Unit of review is the branch diff, not the commit.**
5. **Rejected: a "super-reviewer" panel/vote, per-commit gating, and spec-conformance as a
   blocking gate.**

## 9. Relationship to the existing record

Most of this is already owned. The idea's motivating scenario — *"wire a second-model CLI so
gates are reachable headlessly"* — **is itd-47, which was written and then superseded** by
adr-22/25/27, precisely because it hard-required an external tool.

- **adr-25** already decided the two-model shape and **demoted it from a hard-wired cascade
  to "advice the adapter layer offers, not a cascade the core imposes."**
- **itd-53** already owns continuous review at a safe boundary, with `review.autodrain` and a
  *deferred-not-failed* rule when no backend is reachable.
- **itd-83** owns reviewers firing themselves; **itd-58** owns the unforgeable verdict
  channel; **itd-28** owns the SHA-pinned review store; **itd-17** owns capability-aware
  backend selection.
- **itd-2** (in-session dispatch) is the substrate and is unshipped.
- **The seam does not exist**: there is no `internal/adapter/oracle` and no
  `internal/registry`. `oracle.backend` is a config key ahoy writes and **nothing reads**.
- Per-commit hook-driven review is pre-flagged as *"the canonical anti-pattern"* (itd-28).
- A user-facing verb naming a specific bundled tool **fails the docs gate** at blocker
  severity.

**The record's answer to "I have a second model" is not a new command.** It is: finish the
oracle seam, and adr-25 already tells you how to combine the two.
