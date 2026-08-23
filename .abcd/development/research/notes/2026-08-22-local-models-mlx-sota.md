# SOTA survey — local models (MLX / Apple silicon) in an agentic dev setup

Dated 2026-08-22. Compiled from one host-run web research pass. Method per
[`prefer-sota`](../../principles/prefer-sota.md): each pick is named as the
presumptive SOTA and challenged for fit against this repo's conventions.
The evidence-tier vocabulary (vendor docs / issue-tracker evidence /
research tier / community consensus / practitioner-measured / anecdote) is
this note's own ladder, not a canon from the principle. Companion to
[`2026-08-22-context-window-management-sota.md`](2026-08-22-context-window-management-sota.md).

**Why this note.** Two standing questions: (a) can local models — initially
MLX on Apple silicon — take auxiliary work off the frontier host so its
context and token budget are preserved, with abcd as the integration layer;
and (b) how far is complete autonomy from frontier models for the coding
loop itself. Source-quality warning carried over from the research pass:
this space is heavily colonised by SEO content farms; numbers here trace to
vendor docs, issue trackers, arXiv, and named practitioners (Willison,
Raschka, Unsloth), and anything from marketing-tier aggregators is flagged
as anecdote.

## (a) Local models supporting the frontier host — ranked by value-for-effort

### 1. Semantic recall via local embeddings — ready, hybrid-only

The most mature offload: small embedding models (the nomic-embed-text
family, bge-m3, the Qwen3-Embedding family) run fast on any M-series
machine and are served over standard `/v1/embeddings` endpoints by LM
Studio, Ollama, and llama.cpp's `llama-server` (community consensus). The
critical caveat: the retrieval evidence on technical corpora is mixed —
the one study checked at primary level (arXiv 2505.13129, on the Object
Constraint Language corpus) found learned-sparse SPLADE best with plain
BM25 underperforming — while the broader community consensus is that
hybrid sparse+dense beats either alone. Embeddings are a *second* recall
signal beside exact matching, never a replacement.

*Fit:* two consumers exist. The memory substrate's unconsumed `recall:`
field and the itd-39 retrieval draft are the natural first wiring. The
rules loader is the harder case: the itd-3 design of record deliberately
rejected embeddings for rule recall ("ranking is meaningless and adds
false-positive bloat",
[`../../plans/2026-07-11-itd-3-rules-loader.md`](../../plans/2026-07-11-itd-3-rules-loader.md) §3)
— the survey does not overturn that decision: itd-3's curated boolean
trigger is a precision-first control over a small curated rule-set, not
ranked retrieval over a corpus, so the corpus-retrieval evidence (mixed,
above) neither endorses nor refutes it. The survey adds only the revisit
condition — a hybrid second signal for the recall-miss case, where today
the mitigation is the explicit `*DOMAIN` override.

### 2. Classification, routing, and triage — ready

Closed-set classification (issue labels, triage categories) is well
within quantised Qwen3-4B/8B-class models; it is the least demanding
auxiliary task and failure is cheap and visible (community consensus). For
PII/secret redaction specifically, the measured pattern is a deterministic
layer (regex + NER + checksums, Presidio-style) as the floor — ~69%
tag-match recall alone in one practitioner benchmark — with a local-model
pass on top lifting the hybrid to ~83% on English (higher on some
languages). The same write-up found a general instruction-following model
beating both a domain-specialist fine-tune and a BERT-NER stack at the
constrained-extraction task (practitioner-measured tier: an AWS Builders
write-up; RedactionBench, arXiv 2606.18782, is the emerging benchmark and
publishes no comparable recall figures).

*Fit:* the transcript store redacts on write today; the survey's verdict
matches the repo's shape — the deterministic layer stays primary and
remains the gate, with a local model as an assist for unstructured edge
cases. Issue triage is a candidate consumer at the same opt-in tier.
Rule-domain routing is deliberately excluded: itd-3's rejection of
probabilistic recall (§1) applies to a small-model classifier exactly as
it does to embeddings.

### 3. Summarisation and distillation — ready with a safety rail

Rolling-summary compaction and transcript distillation offloaded to a
small model is established community practice, and current 4B–30B models
(Qwen3-30B-A3B, Gemma-3-class, Phi-4-class) clear the bar for lossy
summarisation of dev transcripts. Structured distillation of resolved
exchanges into typed compact objects (summary + distinguishing detail +
file refs) reports ~11× token reduction while preserving retrieval
(research tier: arXiv 2603.13017). The rail: small models drop details
*silently* — always retain the raw transcript and treat the distillate as
an index into it, not a replacement.

*Fit:* the capture-then-distil design is already this shape — `abcd
history` stores redacted transcripts at SessionEnd, and the chat-distiller
agent (`../prompting/agents/chat-distiller.md`) is researched but not
built. A local model is a candidate engine for that distiller; the
raw-transcript-retained rail is already the store's design.

### 4. Cascade / small-model-first routing — real evidence, wrong scale

RouteLLM-class learned routing showed ~85% cost reduction holding ~95% of
strong-model quality by routing only ~14% of queries upward — but on chat
benchmarks, not agentic coding, and it needs a trained router plus traffic
volume (research tier). Practitioner convergence for solo scale is *static*
routing: task type → model. A configuration layer routing "summarise this"
locally and "write this code" to the frontier host captures most of the win
with none of the machinery.

### 5. Local judges and draft-review — weakest; never a gate

A 2026 systematic evaluation (21 judges, ~541k judgments) finds judge
reliability without validity — exact-match agreement overstates
discriminative ability (Cohen's κ deflation of 33–41 points on MT-Bench)
— and small judges are markedly worse-aligned than 70B+ models (research
tier: arXiv 2606.19544; corroborated by "Judging the Judges", arXiv
2406.12624). A local 8B judge gating frontier output inverts the competence
hierarchy. Acceptable only as a cheap pre-filter whose *negative* verdicts
escalate rather than block.

*Fit:* consistent with `verifier-selects-gates-decide` (a local pre-filter
proposes; it never gates) and compatible with `evaluator-outside-the-loop`
(independence of the gate). The capability requirement — the judge must be
at least as capable as the producer — is the survey's own finding, not
part of either principle.

## Runtime and hardware facts

- **MLX is the default fast path on Apple silicon, including inside
  Ollama** — Ollama's Mac backend rebuilt on MLX (the 0.19 preview,
  March 2026;
  requires ≥32GB unified memory for the MLX path), with reported decode
  roughly doubling on MoE models. Independent benchmarks put MLX at ~1.4×
  llama.cpp on dense models and up to ~3× on MoE; one single-author report
  says the advantage erodes past ~40k context (anecdote tier).
- **Ballpark throughput** (anecdote tier, converging): 30B-A3B-class MoE at
  4–8-bit runs ~40–55 tok/s on M4-class machines at real context, ~30GB
  footprint at long context; a named practitioner (practitioner-measured
  tier) reports ~40 tok/s at 50k
  context on an M4 Mac mini and judges 20–30+ tok/s workable for agent use.
- **`mlx_lm.server` is OpenAI-compatible but explicitly not
  production-grade by its own documentation, and its tool-calling layer is
  a recurring source of per-model-family breakage** (issue-tracker
  evidence: JSON-assumption crashes on XML-style tool calls and undetected
  multi-token delimiters — both since fixed, the latter the day before
  this note's date — plus still-open empty-message and special-token
  issues on other model families; the pattern of breakage recurring family
  by family is the finding, not any single bug). A local model cannot yet
  reliably drive a tool-use loop
  through `mlx_lm.server`; per-model-family parser fragility is the failure
  mode, not model capability. LM Studio's server and llama.cpp are the more
  battle-tested fronts for tool calling; BFCL is the per-model eval to
  consult.

## (b) Complete autonomy from frontier models

**Model landscape**: the strongest open-weight models cluster in the
high-60s on SWE-bench Verified — Moonshot reports Kimi K2-Instruct-0905 at
69.2% agentic (the circulating ~71% figures come from third-party
harnesses or parallel test-time compute), with Qwen3-Coder-480B-A35B and
GLM-4.6 reported nearby — frontier-adjacent on paper, with the caveat that
these figures churn quickly and aggregator pages update underneath
citations. Later 2026 releases claim higher still, via marketing-tier
aggregators (anecdote until reproduced).

**What runs on which Mac** (quantised; unified memory is the binding
constraint): at 32GB, Qwen3-Coder-30B-A3B 4-bit is the consensus sweet
spot; at 64GB, GLM-4.5-Air 3-bit MLX (44GB, built for 64GB Macs — a named
practitioner got working code first try on an M2 MacBook Pro,
practitioner-measured tier) or Qwen3-Coder-30B at 8-bit (the same
practitioner's *non-Coder* 30B 8-bit attempt failed their test; prefer the
Coder variant); at 128GB, GLM-4.6 quantised and fast 30B-class MoE; on
512GB-class machines, Kimi-K2-class fits at 2–4-bit quantisation (the
quantiser's guide puts the floor near ~247GB combined memory for 5+
tok/s), with reported speeds in the single digits to low tens of tok/s
(anecdote tier) — demonstrated, not practical for interactive agent loops.

**Harness support:** a Claude-Code-style harness points at non-vendor
backends via its base-URL override plus a translation layer (LiteLLM proxy,
documented first-party, or claude-code-router for the solo case). What
breaks, per practitioner reports: tool-call format translation (the same
parser fragility as above), interleaved-thinking and streaming edge cases,
and sheer token appetite — one task consumed ~578k input tokens
(practitioner-measured tier)
over 25 turns, and at local prefill speeds that appetite, not intelligence,
is often the binding constraint. Leaner local-first harnesses exist
(opencode, aider, goose, OpenHands, Codex CLI — the last measured beating a
model vendor's own harness even for that vendor's models). Supply-chain
note: a LiteLLM PyPI release pair (1.82.7/1.82.8) shipped
credential-stealing malware — confirmed by the project's own advisory and
independent security vendors; pin a clean release.

**Honest gap:** the optimistic practitioner claim is that local covers
~80% of routine sessions (completion, refactors, tests, explanation) with
the frontier model needed for multi-file architecture and subtle
concurrency; the skeptical counter — also practitioner-sourced — is that
demos skip the real failure modes (long-horizon drift, non-determinism
even at temperature 0, verification capacity as the actual bottleneck). A
named practitioner's micro-benchmark cuts finer: top local models scored
4/5 under a leaner local harness and 5/5 under the two strongest harnesses
(small sample) — the gap measured there was as much harness-dependent as
model-dependent, so the junior-not-peer verdict rests on the long-horizon
failure modes above, not on that sample. **Local is a competent junior for
bounded tasks, not a peer for long-horizon autonomous work.** Privacy, offline, and
no-rate-limits are the honest primary motivations; the cost argument is
weak at solo scale (a ~$10k 512GB-class machine delivering
single-digit-to-low-tens tok/s costs years of frontier API usage).

## Synthesis for this repo

**Ready today to offload** — integration seam: OpenAI-compatible HTTP on
localhost (`/v1/chat/completions` + `/v1/embeddings`), plain `net/http`
from Go with no SDK dependency; served by LM Studio headless, Ollama ≥0.19
(MLX path, ≥32GB), or `llama-server`; built as an **opt-in adapter that
degrades loudly to the current deterministic path when no endpoint
answers** — the repo's "host-delegated by default, native oracles opt-in"
boundary, with the degraded path announcing itself rather than
manufacturing a false green (the spirit of `loud-staging`, whose letter
governs staged code, not runtime fallback):

1. Embedding-based recall as a hybrid signal for the memory substrate
   (nomic-embed-text or Qwen3-Embedding class).
2. Transcript and subagent-return distillation to structured notes, raw
   transcript retained (Qwen3-30B-A3B 4-bit on ≥32GB; Qwen3-8B on 16GB).
3. Issue triage and similar closed-set classification (4B–8B suffices) —
   candidate-consumer tier: the seam is ready, a consumer is not yet
   named. Rule-domain routing is excluded (§2).
4. Redaction assist layered over the deterministic layer, never replacing
   it.

**Premature:** local models driving tool-use loops through `mlx_lm.server`
(issue-tracker evidence of per-family parser breakage); local judges
gating frontier output (validity evidence is against it); learned routers
at solo scale; cross-vendor speculative decoding (the blocker is
token-level draft verification, which no frontier API exposes — infeasible
through current APIs, though heterogeneous-vocabulary variants exist in
the literature; the workable pattern is cascade routing).

**Verdict on (b):** full local autonomy is a working *fallback tier*, not
parity. On a 64GB Mac, GLM-4.5-Air or Qwen3-Coder-30B in a lean harness
(or the frontier harness through a translation proxy) genuinely completes
bounded tasks offline — worth knowing as a privacy/offline escape hatch,
though no record names a consumer for it yet; wiring one would need its
own intent through the ordinary admission path. It is not a replacement for the frontier host on the long-horizon,
multi-file work this repo's playbook gates; the token appetite of
frontier-style harnesses alone makes local prefill the wall. Revisit when a
~70%-SWE-bench-class model lands in the ≤70GB-quantised range with a robust
native tool-call parser.

**Not worth adopting:** a 512GB-class machine for interactive
Kimi/DeepSeek-class agents (single-digit-to-low-tens tok/s); fine-tuning own small models for
these auxiliary tasks (off-the-shelf clears the bar); a small model as a
blocking reviewer; `mlx_lm.server` as the serving layer (its own docs say
not production-grade).

## Review record

2026-08-22, two fresh-context adversarial reviewers, disjoint lenses:
source fidelity NEEDS_WORK, 12 findings (4 major — including one citation
whose source states the opposite of the anchored claim); repo fidelity and
reasoning NEEDS_WORK, 8 findings (2 major — including an internal
contradiction between §1's defence of itd-3 and a routing recommendation).
All applied. (Session-reported counts; the reviewer transcripts are local
ephemera.)

## Sources

- [mlx-lm SERVER.md](https://github.com/ml-explore/mlx-lm/blob/main/mlx_lm/SERVER.md)
- [mlx-lm issue #607 — GLM XML tool calls](https://github.com/ml-explore/mlx-lm/issues/607)
- [mlx-lm issue #984 — multi-token delimiters](https://github.com/ml-explore/mlx-lm/issues/984)
- [mlx-lm issue #1307 — Devstral TOOL_CALLS](https://github.com/ml-explore/mlx-lm/issues/1307)
- [mlx-lm issue #875 — gpt-oss special tokens](https://github.com/ml-explore/mlx-lm/issues/875)
- [mlx-openai-server](https://github.com/cubist38/mlx-openai-server)
- [Using local coding agents — Sebastian Raschka](https://magazine.sebastianraschka.com/p/using-local-coding-agents)
- [GLM-4.5 Air on a 64GB laptop — Simon Willison](https://simonw.substack.com/p/my-25-year-old-laptop-can-write-space)
- [local-llms tag — Simon Willison](https://simonwillison.net/tags/local-llms/)
- [Ollama taps Apple's MLX — The New Stack](https://thenewstack.io/ollama-taps-apples-mlx/)
- [Ollama MLX preview — Ollama blog](https://ollama.com/blog/mlx)
- [Claude Code with non-Anthropic models — LiteLLM docs](https://docs.litellm.ai/docs/tutorials/claude_non_anthropic_models)
- [Claude Code + LiteLLM advisory — Morph](https://www.morphllm.com/claude-code-litellm)
- [Kimi K2 locally — Unsloth](https://unsloth.ai/docs/models/tutorials/kimi-k2-thinking-how-to-run-locally)
- [Open-weight coding models — Faros](https://www.faros.ai/blog/open-weight-models)
- [Berkeley Function-Calling Leaderboard](https://gorilla.cs.berkeley.edu/leaderboard.html)
- [Reliability without validity in LLM judges (arXiv 2606.19544)](https://arxiv.org/html/2606.19544v1)
- [Judging the Judges (arXiv 2406.12624)](https://arxiv.org/pdf/2406.12624)
- [Structured distillation for agent memory (arXiv 2603.13017)](https://arxiv.org/pdf/2603.13017)
- [Rate–distortion memory compaction (arXiv 2607.08032)](https://arxiv.org/html/2607.08032v1)
- [PII redaction without a frontier model — AWS Builders](https://dev.to/aws-builders/you-dont-need-a-frontier-model-to-redact-pii-3cme)
- [RedactionBench (arXiv 2606.18782)](https://arxiv.org/pdf/2606.18782)
- [Optimising RAG for Object Constraint Language (arXiv 2505.13129) — SPLADE best, BM25 underperforms](https://arxiv.org/pdf/2505.13129)
- [LLM coding workflow 2026 — Hacker News](https://news.ycombinator.com/item?id=46570115)
- [Local vs cloud coding models — Rohit Raj](https://rohitraj.tech/de/notes/best-local-llm-for-coding-replace-cloud-2026)
- [MLX vs llama.cpp long-context caveat — Towards AI](https://pub.towardsai.net/apples-mlx-runs-local-llms-3x-faster-than-llama-cpp-until-your-context-hits-40k-715ec441afbb)
- [RouteLLM routing guide — digitalapplied](https://www.digitalapplied.com/blog/llm-model-routing-2026-cost-quality-optimization-engineering-guide)
