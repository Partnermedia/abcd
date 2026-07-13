# SOTA — capture-time intent decomposition + interdependency analysis

**Date:** 2026-07-13. **Drawn from:** one `sota-researcher` run (recency anchored
to 2026-07-13). **Backs:** the `itd-84` intent-decomposition discipline.

> **Citation-confidence caveat.** Gathered in a single research pass; five sources
> were opened and read, the rest rest on search-result snippets and are marked
> `[snippet]` below — verify before quoting a figure. This note records the
> verdict, not independently re-derived evidence.

## The finding: the two halves bifurcate sharply

Splitting **deterministic candidate-finding** (which existing records does this
touch?) from **semantic decomposition** (route each part to a record type; flag
reversals) is correct — and it is exactly where the evidence divides:

- The deterministic half is a *retrieval* problem: mature, solved.
- Semantic routing/classification is *usable-with-a-human*.
- Contradiction/reversal detection is **research-only and unreliable even on
  frontier models** — a warning, not a green light.

## Adopt / adapt / bespoke (per `sota-per-intent`)

| Half | Maturity | Call |
|---|---|---|
| Deterministic candidate-finding | Mature (IR / embedding retrieval) | **Adapt-with-seam** — Go lexical shortlister (BM25/trigram) always-on; embeddings an optional adapter. No new heavy dep. |
| Semantic decomposition (part → record type) | Usable (LLM/ML + Diátaxis classifiers) | **Adapt-with-seam** — host-delegate to the existing agent fleet; bespoke = *taxonomy + prompt only*. |
| Contradiction / reversal reasoning | Research-only | **Bespoke-with-seam, advisory-only** — reason over the shortlist; human confirms every verdict; false-positive budget mandatory. Never auto-file. |
| Advisory-gate mechanics | Practitioner-proven pattern | **Adopt the pattern** — three outcomes (not pass/fail), sample-size floor, safety veto, per-`(target, prompt_version)` windowing. |

No genuinely bespoke **no-seam** build is warranted anywhere — the outcome the
hard-stop rule wants.

## Load-bearing specifics

- **Candidate-finding has a shipped precedent.** Linear generates an embedding on
  a new issue, cosine-searches the corpus, and surfaces similar issues **advisory,
  at creation time** ([linear], opened). The always-on default for abcd should be
  a plain-Go lexical shortlist; embeddings are the seam, not the floor.
- **A deterministic pre-pass can carry real load.** Paska detects nine requirement
  "smells" with pure NLP at **89% precision / 89% recall** ([paska], opened, kappa
  0.89). Takeaway: "is this proposal actually *several* records?" (an atomicity
  smell) is detectable **without an LLM**. Borrow the technique; **reject** the
  controlled-natural-language authoring it rides on (high friction, wrong for
  free-form capture).
- **Doc-type routing already does the split.** Diátaxis classifiers (Katara)
  *flag a page that blends categories so it can be split* `[snippet]` — literally
  the decomposition behaviour, demonstrated for a 4-way taxonomy.
- **The advisory-gate pattern is directly transferable.** An evidence-driven
  release gate emits **PROMOTE / HOLD / ROLLBACK, never binary pass/fail**, stays
  advisory behind a flag, enforces a sample-size floor + a safety veto, and windows
  verdicts per prompt-version to stop calibration drift ([gate], opened). Its gate
  and its human agreed at only **kappa = 0.13** — treated as *complementary
  detectors*, not "the gate is wrong". For abcd: the deterministic overlap-finder
  and the semantic reasoner are complementary; report both, never collapse to one
  score.

## Don't build (SOTA says it fails or doesn't exist)

- **No bespoke contradiction/reversal detector as a gate.** ContraDoc: GPT-4
  struggles with subtle internal inconsistency; cross-*document* contradiction is
  an open gap `[snippet]`. A hand-built Go/grep check would be strictly worse.
  Advisory + human-confirmed only.
- **No LLM verdict as a blocking gate.** "A Coin Flip for Safety" ([coinflip],
  opened): LLM judges near-random on adversarial-robustness judging; clinical
  study found the dominant failure was **over-flagging** (50–81% of false
  positives) — exactly the "this contradicts an invariant" false alarm to budget
  for.
- **No reinvented vector storage / similarity.** pgvector + cosine is mature and
  Linear-proven; if embeddings are wanted, that is an adapter, not Go we write.
- **No formal RTM / traceability tooling, no goal/DSM impact-analysis machinery.**
  Enterprise-scale, assumes a hand-curated trace database — anti-fit for a
  solo-to-small-team config CLI. Steal the *idea* (typed interdependency graph →
  "what breaks if reversed"); skip the tooling.
- **No off-the-shelf NLP4RE library.** The field's own survey ([nlp4re], opened)
  finds 126 tools but a reuse crisis: **44.6% unlicensed, ~23% archived** —
  i.e. there is no drop-in to adopt. This *is* the build-thin-yourself finding.

## Design steal worth naming

Type the cross-record links — `supersedes` / `reverses` / `duplicates` /
`refines` — instead of a vague "related" edge. It is what makes the deterministic
interdependency pass tractable and feeds the (advisory) reversal flag.

## Sources

Opened and read:
- [Linear — Using AI to detect similar issues][linear]
- [Paska/Rimay — Automated Smell Detection in NL Requirements (arXiv 2305.07097)][paska]
- [Evidence-Driven Release Gates for LLM Agents][gate]
- [A Coin Flip for Safety — LLM Judges Fail to Reliably Measure (arXiv 2603.06594)][coinflip]
- [NLP4RE Tools: Classification, Overview, and Management (arXiv 2403.06685)][nlp4re]

Snippet-only (verify before leaning on): Automating Requirements Classification
(EASE 2025); NFR-classification benchmark (arXiv 2510.18096); Contradiction
detection in RAG (arXiv 2504.00180); LegalWiz (arXiv 2510.03418); Diátaxis+AI
(Katara); ADR tooling index; ADR dependencies (nilus.be); change-impact /
traceability overview (Springer).

[linear]: https://linear.app/now/using-ai-to-detect-similar-issues
[paska]: https://arxiv.org/html/2305.07097
[gate]: https://vadim.blog/evidence-driven-release-gates-llm-sales-agents/
[coinflip]: https://arxiv.org/pdf/2603.06594
[nlp4re]: https://arxiv.org/html/2403.06685v1
