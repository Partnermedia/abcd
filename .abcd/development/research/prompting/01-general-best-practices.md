# General Prompting SOTA — Baseline for abcd Agents

> **Role of this document.** Per the design brief: *"Research is the **gate** (audit reference), not the **source** (prompt template). Author writes prompt informed by research; oracle audits alignment."* This file is the gate against which every abcd agent prompt will be audited (`lifeboat-oracle` periodic SOTA-audit, prompt linter cross-check, golden-test fixtures).
>
> **Scope.** the plugin's 15 agents (Pass A: `flow-essence`, `decision-archaeologist`, `review-collator`; Pass B: `chat-distiller`; Pass C: `principle-distiller`, `artefact-curator`, `brief-composer`, `press-release-composer`, `lifeboat-oracle`; plus `code-rescuer`, `issue-scout`, `embark-scaffolder`, `launch-gatekeeper`, `intent-fidelity-reviewer`).
>
> **Iteration.** This is the first revision and is **edited in place** as new material accrues (new Anthropic guidance, new survey paper, new injection-defence benchmark, agent-specific pitfalls discovered during epic work). Once the doc has accumulated enough churn to make a clean snapshot useful (rule of thumb: 5+ material edits), promote the current text to `archive/01.md` and start a new `01-general-best-practices.md` as iteration 02 — same pattern as the design brief, but only after history has earned it. Per-agent research lives at `prompting/agents/<name>.md` and references this doc as baseline.

---

## 1. The big shift: prompt → context engineering

The headline 2026 consensus, articulated most clearly by Anthropic's *Effective Context Engineering* and adopted into Thoughtworks' Tech Radar (April 2026, Adopt ring), is that **prompt engineering is now a subset of context engineering**. The unit of design is no longer "the system prompt" but the **token budget across the agent's whole loop**: system prompt + tool descriptions + retrieved content + scratchpad + message history.

**Concrete principle (Anthropic):** *"find the smallest set of high-signal tokens that maximize the likelihood of some desired outcome."* Minimal does not mean short — it means sufficient without excess.

**Why this matters for abcd specifically.** Pass B's `chat-distiller` is the canonical case. Naïvely loading 401 specstory transcripts will trigger **context rot** (Chroma 2025: 18 frontier models, accuracy drops *off a cliff* unpredictably, often well before the advertised window) and **lost-in-the-middle** (Stanford: ≥30% accuracy drop for mid-context information). The brief's design — git-blame epic windows → time-window index → filtered transcripts → one call per unresolved spine entry — is already aligned with SOTA. Treat it as the template for any agent that touches large corpora (`review-collator` over prompt-exports, `decision-archaeologist` over git log, `intent-fidelity-reviewer` over shipped artefacts).

## 2. System-prompt structure: the Goldilocks zone

Anthropic's explicit anti-pattern is **"complex, brittle logic in their prompts"** at one extreme, and **"vague, high-level guidance"** at the other.

Adopted shape for abcd agent prompts:

- **Distinct sections** via XML tags or Markdown headers: `<background_information>`, `<instructions>`, `## Tool guidance`, `<output_format>`, `<examples>`. Models follow structured prompts measurably better than monolithic prose.
- **Few-shot examples**: 2–5 *diverse, canonical* examples that span the behaviour envelope. **Anti-pattern:** stuffing edge cases. Anthropic: *"For an LLM, examples are the 'pictures' worth a thousand words."*
- **Explicit effort-scaling rules** when the agent dispatches sub-work. Anthropic's research-system prompt literally says: *"simple fact-finding requires just 1 agent with 3-10 tool calls"* vs *"complex research might use more than 10 subagents."* `chat-distiller` should encode an analogous budget cap.
- **Place high-priority instructions at the start AND the end** to dodge lost-in-the-middle. Already done for verdict-tag emission in fn-3 T4 — generalise.

## 3. Sub-agent prompt design (mapped to the 14 abcd agents)

The Anthropic multi-agent-research post is the single best reference for what abcd is building. Lessons mapped:

| Anthropic lesson | abcd application |
|---|---|
| Lead-agent delegation needs *"objective, output format, guidance on tools and sources, clear task boundaries"* | The skill `workflow.md` for `disembark` is already structured this way; verify each Pass-A/B/C dispatch carries all four. |
| *"Bad tool descriptions can send agents down completely wrong paths"* — they got a 40% task-time reduction by having Claude rewrite its own tool descriptions | Run a one-shot self-improvement pass over each abcd agent prompt before locking it: ask the oracle to rewrite for clarity, run golden tests, keep the diff if green. |
| *"If a human engineer can't definitively say which tool should be used in a given situation, an agent can't be expected to do better"* | Audit the 15 agents for overlap. `principle-distiller` vs `decision-archaeologist` vs `review-collator` all touch "rationale" — make the boundary a single sentence in each `description`. |
| Sub-agent returns *"condensed, distilled summary (typically 1,000-2,000 tokens)"* | Set explicit output-token targets per agent in the spec. This is also the defence against parent-context rot. |
| Cost: research-style multi-agent burns ~15× chat tokens | Document expected token order-of-magnitude per command in the brief. `disembark` is *expected* to be expensive; embark is normally cheap. |

**Claude Code subagent specifics (relevant to `agents/*.md` files):**

- The `description` field is the **trigger, not the label**. Anthropic's docs: write it like a routing rule, include "Use proactively when…" or "Trigger when…". Most public subagent libraries get this wrong — they write descriptions for humans, not for the dispatcher.
- Tighten `tools:` per agent. Reviewers/auditors must be read-only (no `Edit`, `Write`, `NotebookEdit`). Already noted as a constraint in `.work/issues.md` for itd-2 (in-session). Apply uniformly: `lifeboat-oracle`, `intent-fidelity-reviewer`, `launch-gatekeeper` (read-only); `brief-composer`, `embark-scaffolder` (read/write, no Bash); `code-rescuer` (read-only — principles only, never edits source).
- `model: inherit` is the safe default; pin Opus only where reasoning quality is load-bearing (`lifeboat-oracle`, `principle-distiller`, `intent-fidelity-reviewer`).

## 4. Repos worth cloning rather than re-creating

Hand-picked, all MIT, all alive in 2026:

| Repo | What to lift | Caveat |
|---|---|---|
| **[Piebald-AI/claude-code-system-prompts](https://github.com/Piebald-AI/claude-code-system-prompts)** | Anthropic's *own* sub-agent prompts (Explore, Plan, Task), tool descriptions, system reminders. Tracks 167 versions of Claude Code with a CHANGELOG. **Closest thing to a canonical reference for sub-agent prompt shape.** | Reverse-engineered, not officially endorsed. Use as *style guide*, not as drop-in. |
| **[VoltAgent/awesome-claude-code-subagents](https://github.com/VoltAgent/awesome-claude-code-subagents)** | 130+ subagents across 10 categories. Same frontmatter shape (`name`, `description`, `tools`, `model`). | **Explicitly disclaims quality/security review.** Mine for patterns, do not trust for production. No versioning, no evals. |
| **[EliFuzz/awesome-system-prompts](https://github.com/EliFuzz/awesome-system-prompts)** & **[tallesborges/agentic-system-prompts](https://github.com/tallesborges/agentic-system-prompts)** | Production system prompts from Cursor, Devin, Codex, Augment, Kiro, Cluely. Useful for *competitive prompts*: see how Devin frames code-review handoff, how Cursor scopes Edit. | Scraped, no licence claim on the underlying content — treat as reference only, do not copy verbatim into a public repo. |
| **[microsoft/prompty](https://github.com/microsoft/prompty)** | The **`.prompty` asset format**: YAML frontmatter (model, inputs, tools) + Markdown body with `system:`/`user:`/`assistant:` role markers + Jinja templating. MIT, VS Code extension, Python/TS runtimes. | Still v2 alpha. ~1.2k stars — niche, not a standard. Worth lifting the *format* even if not adopting the runtime. |
| **[promptfoo/promptfoo](https://github.com/promptfoo/promptfoo)** | Declarative eval configs, GitHub Actions integration, built-in red-team / prompt-injection tests. **OpenAI uses this internally; Anthropic uses it.** | Recently acquired by OpenAI — still MIT, still open, but governance signal worth tracking. |
| **[promptslab/Awesome-Prompt-Engineering](https://github.com/promptslab/awesome-prompt-engineering)** | Reading list, papers, tooling index. | Curation only; nothing directly cloneable. |

**The Prompt Report** ([arXiv 2406.06608](https://arxiv.org/abs/2406.06608)) — 58 LLM prompting techniques, 33-term vocabulary. Treat as a **dictionary** for the prompt linter (B+C+D infrastructure): when the linter flags an agent prompt as "low-rigour", it can suggest specific named techniques (CoT, self-consistency, ReAct, plan-and-solve, etc.) by reference.

## 5. Versioning: what to actually do

The 2026 consensus on prompt versioning, distilled:

- **Solo developer / prompts-as-code** (abcd's situation) → **default to git**. Workbench platforms (Langfuse, Braintrust, Latitude, Maxim) are for teams of 4+ where non-engineers edit prompts.
- **Prompt files are immutable once shipped.** Each release is a snapshot. Amend by adding a new version, not editing the old one. Consistent with the brief's existing `archive/01.md` pattern.
- **Capture more than the text.** A complete prompt version captures: prompt body, model ID, parameters (temperature, max tokens), change rationale, eval result. Prompty's frontmatter handles this naturally.
- **Semantic versioning per agent prompt** (`v1.0.0` → `v1.1.0` for behaviour change, `v1.0.1` for typo). The Liu et al. 2026 taxonomy paper notes this is now the dominant convention.
- **Rationale beats diff.** *"A diff tells you what changed. The change rationale tells you why. Version history that captures only the what becomes illegible after three months."* Commit messages already do this; keep doing it for prompt edits.

**Concrete recommendation for abcd:**

1. Each `agents/*.md` file carries a `prompt_version: 1.0.0` frontmatter field alongside `name`, `description`, `tools`, `model`.
2. A `CHANGELOG.md` per agent (or one consolidated `agents/CHANGELOG.md`) records each bump with a one-line rationale and the eval delta.
3. Golden-test fixtures (B+C+D plan already has this) live at `agents/<name>/fixtures/` and are versioned together with the prompt — fixtures are part of the prompt's contract.
4. The prompt linter (cross-cutting infra item) checks: frontmatter completeness, examples count (≥2, ≤5), required sections present, length budget per agent, no PII / personal paths.
5. The **periodic SOTA-audit oracle prompt template** (D component) re-runs against this research doc every quarter. Its job: flag agents whose prompt structure has fallen behind the moving SOTA target.

## 6. Evaluation: how to tell if a prompt is good

The 2026 evidence on LLM-as-judge has matured fast and is worth knowing before betting on it:

- **Four canonical biases in every untreated judge:** position, verbosity, self-preference, authority. Three more recent: rubric-order, score-ID, reference-answer-score. Krippendorff's α ≈ 0.8 is the new agreement target.
- **Frontier judges exceed 50% error rate on hard bias benchmarks** (RAND 2026). No single judge is uniformly reliable.
- **Prompt-design effect on judges:** detailed structured rubrics → systematically *harsher* scores. Minimal prompts → *more lenient*. Calibrate accordingly; never compare scores across rubric versions.
- **Anthropic's current production guidance:** *combine unit tests for correctness with LLM rubrics for overall quality*. Do not use LLM rubric alone for a pass/fail gate.

**For abcd specifically:** `lifeboat-oracle` and `intent-fidelity-reviewer` are LLM-judge agents. Apply:

- Pairwise comparison where possible (less biased than absolute scoring).
- Position-shuffle: judge the same artefact in both orders; require agreement.
- Use the structured `<verdict>SHIP|NEEDS_WORK|MAJOR_RETHINK</verdict>` tag (already in use) — that's a 3-bin discrete scale, far less bias-prone than 1–10 floats.
- Hard gates remain structural (PII scan, schema validation, file existence) not judge verdicts.

## 7. Prompt injection — the risk to budget for

OWASP LLM01 (Prompt Injection) is **#1 for the third year running** as of 2026. For abcd specifically, the threat surface:

- `disembark` reads `.specstory/` transcripts that may contain attacker-controlled content (a malicious commit message, an injected MCP server output captured in a transcript). The `chat-distiller` agent then summarises them — classic indirect injection vector.
- `embark` unpacks lifeboats that could be hostile (someone hands abcd a "lifeboat" as a Trojan).
- `launch` scrubs files for the public sibling — an injection could try to *re-introduce* PII the scrubber removed.

**Defence stack (in priority order, per 2026 SOTA):**

1. **Structured prompt formatting.** Never concatenate untrusted content into the system prompt. Use clear delimiters and explicit role markers. Free; always adopt.
2. **Output schema validation.** Every agent that produces JSON validates against a schema. Already done for `epic-essence.json`, `_provenance.json`, etc. — make universal.
3. **Dual-LLM pattern (Simon Willison)** for the most sensitive paths. Privileged LLM holds tools (`launch-gatekeeper`'s scrub-and-commit), reads only structured summaries from a quarantined LLM that processed the untrusted artefact. Worth considering for `embark` when source is `--from <unknown>`.
4. **PromptArmor-style filter** (ICLR 2026, <1% FP/FN on AgentDojo) on transcripts before `chat-distiller` reads them — an off-the-shelf cheap-model preprocessor that detects and strips injection. Likely an enhancement for a later phase, not now.
5. **Read-only enforcement on auditor agents** — already in plan.

## 8. Trade-offs to choose explicitly

| Trade-off | Lean toward | Why |
|---|---|---|
| Long detailed prompt vs. short focused prompt | **Short focused** | Goldilocks zone is biased toward minimal; long prompts compound context-rot risk and amortise badly across the 15 agents. |
| Pin specific models vs. `model: inherit` | **`inherit` by default**, pin only where load-bearing | Future-proof; but requires acknowledging verdict-tag drift across model versions (already flagged for fn-3 T4). |
| LLM-judge rubric vs. structural test | **Structural where possible, judge where necessary** | Judges are bias-laden. Structural tests are deterministic. The brief's "round-trip + oracle = phase acceptance" already does this hybrid. |
| Few-shot vs. zero-shot for agents | **Few-shot, ≤5 diverse examples** | The single biggest single-lever quality boost in the literature; cheap to add; ages well. |
| Bundle prompts with code vs. separate prompt-management platform | **Bundle (git-native)** | Solo-developer scale, prompts-as-code, no non-engineer editors. Defer Langfuse/Braintrust until a contributor without git literacy needs to edit a prompt. |
| Centralised "prompt library" vs. per-agent prompts in-place | **Per-agent in-place** (current plan) | Prompts are tightly coupled to their agent's tools, output schema, and golden tests. A shared library creates indirection that ages worse than duplication. |
| One mega-agent vs. fan-out to specialists | **Fan-out** (current plan) | Anthropic's research-system result + Cognition's "writes single-threaded, intelligence multi-agent" rule both support this. But: budget the 15× token cost. |

## 9. Pitfalls specific to the abcd design

Reviewing the brief and in-flight `itd-2` (in-session subagent dispatch) work against SOTA:

- **`research is the gate, not the source`** (existing rule) is exactly right and matches Anthropic's "examples are pictures" framing. Hold the line — do not let agents quote research blocks verbatim.
- **`chat-distiller` is the highest context-rot risk.** Make the spine-entry-per-call discipline non-negotiable. If it ever drifts toward "load and summarise this whole transcript", performance will collapse silently.
- **Specstory transcripts are an indirect-injection vector.** `chat-distiller` reads transcripts that may have captured attacker-influenced content (a malicious commit message quoted in chat, an MCP server's hostile output, a pasted artefact). This is OWASP LLM01 even when nothing is being unpacked; defence is § 7 rung 1 (spotlighting / delimiting) plus citation-bound output (every rationale claim resolves to a real transcript line-range). Surfaced from `prompting/agents/chat-distiller.md`.
- **`press-release-composer` is the highest verbosity-bias risk.** LLM judges (and humans) reliably prefer longer marketing copy. Cap output tokens hard; have `lifeboat-oracle` audit on *fidelity* and *concision* separately.
- **Verdict-tag drift across models** (flagged for fn-3 T4 + memory) is the *systemic* risk of `model: inherit` everywhere. Build the multi-model verdict-emission smoke matrix already noted as future work.
- **Subagents-cannot-spawn-subagents** (`.work/issues.md` capture) limits one-level-deep dispatch. The `claude -p` headless escape pattern is the documented workaround if a later phase needs it.
- **The 14-agent count itself.** Anthropic's research system spawns 3–5 subagents per query — not 14. abcd is not running them all at once (Pass A: 3 parallel, Pass B: 1 looped, Pass C: ~5 sequential), so the tool-discrimination problem is bounded. But: audit the `description` fields explicitly for routing-rule clarity; that is where the 14-agent design will fail first.
- **Provenance-washing on `embark`.** A lifeboat's `_provenance.json` can claim a trustworthy origin without abcd having any way to verify it cryptographically. The mitigation is procedural: the `embark-scaffolder` agent neither validates nor trusts provenance — it surfaces the fields verbatim in `embark-report` and requires user confirmation. Cryptographic verification comes in a later phase, tracked as `itd-16` (hash-chain Merkle audit). Surfaced from `prompting/agents/embark-scaffolder.md`.

## 10. Three concrete additions to prompt-quality infrastructure (B+C+D)

1. **A `prompt_version` field** in every agent frontmatter, plus `agents/CHANGELOG.md`. Per §5 above. Versioning is currently implicit in git; making it explicit costs nothing and makes the SOTA-audit oracle's job easier.
2. **A self-improvement pre-flight** before locking each agent prompt at v1.0.0: ask the oracle to suggest improvements, accept only changes that pass golden tests. Anthropic got 40% task-time reduction this way; even a fraction of that across 15 agents is worth a one-shot pass.
3. **An injection-canary fixture** in `chat-distiller`'s and `embark-scaffolder`'s golden tests. A specstory transcript / lifeboat with a deliberately-injected `IGNORE PREVIOUS INSTRUCTIONS` payload. The agent must not honour it. The cheapest possible regression test for OWASP LLM01.

These three are tracked as a candidate intent (`itd-N`); see `roadmap/intents/drafts/`.

---

## Sources

**Primary (Anthropic / first-party):**
- [Effective context engineering for AI agents](https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents) — Anthropic, 2026
- [How we built our multi-agent research system](https://www.anthropic.com/engineering/multi-agent-research-system) — Anthropic, 2026
- [Building Effective AI Agents](https://www.anthropic.com/research/building-effective-agents) — Anthropic
- [Claude Code subagent docs](https://code.claude.com/docs/en/sub-agents) — Anthropic, 2026
- [Prompting best practices](https://platform.claude.com/docs/en/build-with-claude/prompt-engineering/claude-prompting-best-practices) — Anthropic API docs

**Surveys & academic:**
- [The Prompt Report (arXiv 2406.06608)](https://arxiv.org/abs/2406.06608) — 58-technique taxonomy, 33-term vocabulary
- [Liu et al. 2026 — A comprehensive taxonomy of prompt engineering techniques](https://link.springer.com/article/10.1007/s11704-025-50058-z) — Frontiers of Computer Science
- [A Survey on LLM-as-a-Judge (arXiv 2411.15594)](https://arxiv.org/abs/2411.15594)
- [Understanding Prompt Management in GitHub Repositories (arXiv 2509.12421)](https://arxiv.org/html/2509.12421v1)

**Repositories worth cloning / mining:**
- [Piebald-AI/claude-code-system-prompts](https://github.com/Piebald-AI/claude-code-system-prompts) — MIT, versioned, closest canonical reference
- [VoltAgent/awesome-claude-code-subagents](https://github.com/VoltAgent/awesome-claude-code-subagents) — MIT, 130+ subagents
- [EliFuzz/awesome-system-prompts](https://github.com/EliFuzz/awesome-system-prompts) — Cursor / Devin / Codex production prompts
- [tallesborges/agentic-system-prompts](https://github.com/tallesborges/agentic-system-prompts)
- [microsoft/prompty](https://github.com/microsoft/prompty) — MIT, asset format
- [promptfoo/promptfoo](https://github.com/promptfoo/promptfoo) — MIT, evaluation + red-team

**Versioning & evaluation:**
- [What is prompt versioning? — Braintrust](https://www.braintrust.dev/articles/what-is-prompt-versioning)
- [How to version prompts: 2026 guide — Prompt Assay](https://promptassay.ai/blog/how-to-version-prompts-2026-guide)
- [Rubric-Based Evals & LLM-as-a-Judge — Adnan Masood (Apr 2026)](https://medium.com/@adnanmasood/rubric-based-evals-llm-as-a-judge-methodologies-and-empirical-validation-in-domain-context-71936b989e80)

**Security:**
- [OWASP LLM01:2025 Prompt Injection](https://genai.owasp.org/llmrisk/llm01-prompt-injection/)
- [OWASP Top 10 for Agents 2026 (DeepTeam)](https://www.trydeepteam.com/docs/frameworks-owasp-top-10-for-agentic-applications)
- [Prompt Injection Defense 2026: 8 Tested Techniques Ranked](https://tokenmix.ai/blog/prompt-injection-defense-techniques-2026)

**Context / 2026 ecosystem:**
- [State of Context Engineering in 2026 — Aurimas Griciūnas](https://www.newsletter.swirlai.com/p/state-of-context-engineering-in-2026)
- [Multi-Agent or Not? Context-First Insights from Anthropic and Cognition](https://agenticspace.dev/multi-agent-or-not-context-first-insights-from-anthropic-and-cognition/)

---

## Related Documentation

- [`.abcd/development/brief/README.md`](../../brief/README.md) — design brief; § "Research-driven prompts" mandates this doc
- [`.abcd/development/research/related-work.md`](../related-work.md) — prior-art comparison (PAUL, CARL, claude-skills)
- [`prompting/agents/`](agents/) — per-agent SOTA research (one file per agent, references this baseline)
