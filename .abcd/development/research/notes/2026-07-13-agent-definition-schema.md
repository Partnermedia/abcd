# SOTA: agent-definition schema

Desk research into how an individual agent should be specified — the content and
structure of an agent's definition file, not multi-agent orchestration theory.
Recency anchored to 2026-07-13. Sources are cited only where a page was actually
opened; index-only hits are marked.

This is the **gate** (the audit reference), not the **source** (a prompt template),
per the standing convention in [`prompting/agents/README.md`](../prompting/agents/README.md).

## 1. Why this research exists

abcd ships five agents. They were written ad hoc, and the record and the reality
have drifted apart:

- `05-internals/06-lint.md` states prompt-lint rules `PQ001`–`PQ006` are
  "**delivered**" in `internal/core/lint`. **No `PromptLinter`, no `PQ*` rule, and no
  `task_classes` enum exist in Go.** The lint is fiction.
- `05-internals/01-agents.md` makes `capability_scope` required and
  `reads_untrusted_input` conditional. **Zero of five agent files carry either.**
- itd-5 requires an injection-canary fixture per agent and an `agents/CHANGELOG.md`.
  **Neither exists.** All five agents sit at `prompt_version: 0.1.0`.
- The catalog declares 16 agents, the prompting README targets 14, the baseline says
  15, and disk holds 5 — with exactly **one** name common to catalog and disk.

iss-11 (agent-count drift) was closed by editing prose until the documents agreed
with *each other*, never against `ls agents/`. A surface with no mechanism to detect
its own drift is how those four numbers came to coexist without one failing test.

## 2. The governing constraint: brevity is not a style preference

The instinct when writing a "best-practice schema" is to mandate a rich template.
The evidence says that instinct is wrong, and this finding outranks every other
recommendation in this document.

**Gloaguen et al. (ETH Zurich, Feb 2026)** built a benchmark from real GitHub issues
in repositories with developer-committed context files. Finding: context files tend
to **reduce** task success rates versus no repository context, while increasing
inference cost by **>20%**. The mechanism is that agents *respect* instructions —
unnecessary requirements make tasks harder. Their recommendation, explicitly
"contrary to agent developers' recommendations", is that human-written context should
describe only **minimal** requirements.

Two independent results point the same way:

- **Lost in the Middle** (arXiv 2307.03172) and Chroma's *Context Rot* report: long
  inputs degrade reliability non-uniformly, even on trivial retrieval. A long agent
  definition is itself a reliability risk.
- Anthropic's context-engineering guidance: find "the smallest possible set of
  high-signal tokens".

### 2.1 The counter-evidence, which this repo already adjudicated

**This finding does not license a length gate, and abcd has already decided so.**

itd-81 (`intents/disciplines/itd-81-judge-calibration.md`, amended 2026-07-12) **struck**
the prior ">10% shorter" tiebreak on agent prompts, on the grounds that *"selecting on
length selects for the **brevity bias**"* identified by **Agentic Context Engineering**
(arXiv 2510.04618) as progressively destroying instruction quality: *"a shorter prompt
that passes the same goldens may simply have shed the domain detail the goldens never
tested."* Under that discipline, **length is not a tiebreak; the corpus score is.**

The two findings are compatible but say different things:

- **Gloaguen** is about *unnecessary* instructions — irrelevant requirements the agent
  dutifully honours. It argues against padding.
- **ACE / itd-81** is about *selecting on length as a proxy* — which sheds load-bearing
  domain detail that the tests never covered. It argues against a line-count gate.

**Design consequence.** Prefer minimal, high-signal content — but **never enforce that
with a length rule.** Rigour belongs in *mechanically-checkable frontmatter*, which
costs the model nothing at inference. Content quality is measured against a **labelled
corpus** (itd-81), not against a line count. A prompt-length lint is an
evidence-contradicting design and is out of scope for any abcd schema.

## 3. The convergent core (what every framework agrees on)

Field-level comparison of twelve agent-definition formats — Claude Code subagents,
Agent Skills/`SKILL.md`, OpenAI Agents SDK, Google ADK `LlmAgent`, GitHub Copilot
`*.agent.md`, crewAI, LangChain/LangGraph, AutoGen, Bedrock Classic, Bedrock AgentCore
harness, OpenCode, and the A2A `AgentCard`.

Five fields are near-universal:

| Field | Present in | Note |
|---|---|---|
| `name` | 12/12 | sometimes defaulted from filename |
| **routing description** | 9/12 | required in Claude Code, Skills, Copilot, OpenCode, A2A |
| instructions | 11/12 | A2A alone omits it, deliberately |
| tool grant | 11/12 | allowlist of names; MCP as the extension mechanism |
| model selector | 11/12 | increasingly with an `inherit` default |

A sixth, quieter convergence: **in every markdown format, frontmatter is metadata and
the body IS the instructions.** Do not invent an `instruction:` YAML key.

### 3.1 The routing/instruction split is the strongest signal available

**9 of 12** formats carry a routing field distinct from the agent's own instructions.
Google ADK states the rule most plainly: `description` is *"primarily used by **other**
LLM agents to determine if they should route a task to this agent"*, while `instruction`
tells the agent its own task. AutoGen: `description` is *"used by team to make decisions
about which agents to use"*.

The exceptions prove the rule. All three non-separators (crewAI, Amp, AgentCore harness)
are formats in which **no LLM ever chooses between agents** — single-agent configs, or
human-selected modes, or delegation that is off by default. Every format that does real
LLM-mediated delegation has the split.

The decisive case is **A2A's `AgentCard`**: a spec whose *only* job is inter-agent
selection. It has **no instructions field at all**, and it makes `name`, `description`,
`version`, and per-skill `tags` **required**, with `examples` available. When someone
designed a schema purely for "when to route to me", that is the field set they arrived at.

In four formats the routing description is the *only* required field. Their designers
concluded that an agent without instructions is recoverable and an agent without a
routing description is not.

**Design consequence.** `description` is a routing rule addressed to a dispatcher, not a
label addressed to a human. `description: security expert` is a defect; it must state
trigger conditions.

### 3.2 `use PROACTIVELY` is cargo cult — attribution corrected

All four of abcd's reviewer agents open their description with `Use PROACTIVELY…`.
**The current Claude Code subagents documentation contains no prose guidance to write
this.** The idiom's only appearance is lowercase, inside a JSON example. It survives in
the community from an older doc example.

The actual current guidance (Anthropic's subagents blog) is: **"Be specific about the
trigger conditions, not just the capability."** Better — *"Reviews code for security
issues before commits"*; worse — *"security expert"*.

**Design consequence.** Drop the shouty token; state the trigger. Our descriptions
already carry good trigger conditions, so this is a cheap edit, not a rewrite.

## 4. Evidence that contradicts common practice

### 4.1 Persona openers do not work (and perturb safety)

**Zheng et al., arXiv 2311.10054 (Findings of EMNLP 2024)** — 162 roles × 2,410 factual
questions × 4 model families: adding a persona to the system prompt **does not beat the
no-persona control**. Oracle-selecting the best persona per question helps, but
*automatically* choosing one is no better than random — the gain is unexploitable in a
static spec.

Corroborated for reasoning (ACM 10.1145/3688864.3689149: role-play does not improve and
can degrade mathematical reasoning) and shown to be a *safety* liability
(**arXiv 2509.08075**: refusal behaviour swings with persona on identical requests).

The one paper cited in favour — **arXiv 2308.07702**, reporting large gains — is, by its
own authors' admission, a two-stage protocol where the model first generates a
role-immersion turn that "serves as an implicit CoT trigger". It is a
reasoning-elicitation result wearing a persona costume, measured on 2023-era
non-reasoning models. It does not license *"You are a senior security engineer."* as a
line in a spec.

Independently, GitHub's analysis of **2,500+ repositories** with `agents.md` files names
**vague personas** as a top anti-pattern; the effective files carry commands, structure,
style examples, and three-tier boundaries.

**Design consequence.** Open with the task and what counts as done, not with who the
agent is pretending to be. Personas remain licensed only for style/register control on
subjective outputs. abcd's agents mostly comply already (`You verify that…`,
`You research…`) — this is a rule to *hold*, not a repair.

### 4.2 Never schema-constrain the reasoning span

**Tam et al., arXiv 2408.02442** found reasoning degrades under format restriction, with
stricter constraints costing more. **Fan, arXiv 2606.09410 (June 2026)** gives the most
actionable number: **Claude Opus 4.7 drops 96.2% → 91.0% on AIME when made to answer in
JSON**, and a **reason-first-then-format** protocol recovers **80–87%** of the loss.

The dottxt rebuttal argues the cause is a prompt confound rather than the constraint
itself. The honest verdict is *contested on cause, agreed on remedy*: **both camps
recommend free-form reasoning followed by a constrained emit.** Google ADK independently
documents that its `output_schema` **disables tool use**, which is the same lesson in
harness form.

**Design consequence.** Two-phase output contract: unconstrained analysis, then a single
schema-validated final block. Never make the model think inside the JSON.

### 4.3 "Review your answer and improve it" is a no-op or worse

**Huang et al., arXiv 2310.01798 (ICLR 2024)**: *intrinsic* self-correction — revising
with no external feedback — does not help and often **degrades** performance. Reflexion
(2303.11366) works precisely because it converts *environment* feedback (tests, task
reward) into verbal memory.

**Design consequence.** A reflect/self-check step is licensed **only** when a test run,
compiler, linter, or diff is fed back in. A prompt line saying "double-check your work"
is not a gate; it is a cost.

### 4.4 Description-based auto-activation is ~79% reliable

A Vercel evaluation (Jan 2026, cited by Hightower) found coding agents activated the
right skill only **79% of the time** even with explicit instructions; 100% required
brute-forcing an 8KB index into passive context. Static catalogs are also summarised
away under compaction.

**Design consequence.** A well-written `description` is best-effort routing, not a
guarantee. Anything that must *always* happen needs a deterministic trigger (a lint
rule, a hook, a gate), not a hopeful adjective. This is the load-bearing argument for
building the prompt-lint rather than documenting a convention.

## 5. Evidence for the reviewer agents specifically

Four of abcd's five agents are LLM-as-judge. The literature is unusually direct about
what helps.

- **Checklists beat adjectives.** TICK (arXiv 2410.03608): LLM-generated
  instruction-specific YES/NO checklists raised exact agreement with human preference
  **46.4% → 52.2%**. Kwok (`llmasaverifier`) adds the three axes that measurably improve
  verification: **score granularity, repeated evaluation, and criteria decomposition**
  (split one big rubric into many small criteria).
- **Judges are biased in known directions.** MT-Bench (arXiv 2306.05685): GPT-4 judges
  reach >80% agreement with humans — human-level — but exhibit **position, verbosity and
  self-enhancement bias**. **arXiv 2404.13076**: self-preference is *linearly correlated
  with self-recognition* — a model can tell its own output and rates it higher than
  humans do.
- **Judges are softer on machine-written work.** RedQueen (`iacob2026redqueen`): the
  strongest baseline reviewer **over-accepts AI-generated work at up to 1.91× the human
  rate**. Every abcd reviewer is judging AI-written code. This is the most
  operationally-relevant bias we face.
- **Panels buy less than they appear to.** *Replacing Judges with Juries* (2404.18796)
  shows a panel of three small models from **disjoint families** beats a single large
  judge at >7× lower cost. But Apple's *Nine Judges, Two Effective Votes* (2605.29800)
  finds nine frontier judges give **n_eff ≈ 2.18** (independence ratio 24.2%), and the
  **best single judge matched or beat the full panel on every dataset**. Both are
  unreplicated preprints. Verdict: **[CONTESTED]** — a second model *family* buys more
  than a third instance of the same one.

**Design consequence.** Reviewer leverage is: a decomposed checklist, a **required quoted
evidence span per finding**, a fixed severity scale, an explicit `INCONCLUSIVE` verdict,
and **never letting the author model be the sole judge**. Not persona strength, and not
a fifth reviewer.

## 6. The failure taxonomy worth designing against

**Cemri et al. (Berkeley), *Why Do Multi-Agent LLM Systems Fail?* — arXiv 2503.13657.**
7 frameworks, 200+ tasks, 1,600+ annotated traces, 6 expert annotators, inter-annotator
κ = 0.88. It yields 14 failure modes in three categories:

| Category | Share of failures | What it means for a schema |
|---|---|---|
| **Specification** | ~42% | Unclear role boundary, missing stop condition |
| **Inter-agent misalignment** | ~37% | Undeclared handoff contract |
| **Task verification** | ~21% | No termination/verification criterion |

This is the only failure taxonomy in the literature derived from **annotated real traces**
rather than from what worked in a demo, and roughly **42% of failures trace to bad
specification** — which is the empirical justification for having a schema at all.

**Design consequence — and its limit.** MAST is the best evidence that *specification
quality matters*. It is **not** evidence that any particular template fixes it:

- **MAST tested no intervention.** It annotated failures; it did not measure whether a
  mandated section list reduces them. "42% specification" licenses concern, not a cure —
  and it is entirely compatible with Gloaguen (§2), where *more* instruction made things
  worse.
- **Its denominator is multi-agent systems.** abcd's agents are one-shot, host-delegated
  reviewers — "subagents read and analyse; one thread writes" (§7). The 37%
  "inter-agent misalignment" category is largely inapplicable by construction: our
  "handoff" is a JSON blob consumed by a Go ingest, which is serialisation, not
  inter-agent negotiation.

Using a multi-agent failure taxonomy as the skeleton of a single-agent file template
would be evidence laundering — citing a paper that establishes a *problem* as though it
endorsed a *remedy*. **Take MAST as a reason to state a role boundary, a stop condition
and a verification criterion where they are load-bearing. Do not take it as a mandate for
a fixed section list.**

The same caution applies to **TICK** (§5): TICK generates an **instruction-specific
checklist per input, at evaluation time**, and its outcome measure is *agreement with
human preference*, not *finds more real bugs*. A **static, human-authored checklist baked
into a prompt** is a different intervention — closer to the rubric baseline TICK beats.
Cite TICK for reviewer rubrics; do not cite it to mandate checklists in every agent.

## 7. Single agent vs committee

Held for completeness, because it bounds what the schema should encourage.

- **Tran & Kiela (Stanford), arXiv 2604.02460**: across FRAMES and 4-hop MuSiQue, on
  three model families, **single-agent matched or beat every multi-agent variant** once
  *thinking tokens* were held equal. Multi-agent only won when the single agent's context
  was deliberately degraded — which is the *context-isolation* argument stated as a
  negative.
- **Zhang et al., arXiv 2502.08788**: multi-agent debate often fails to beat plain CoT +
  self-consistency despite far more compute. What helps is **model heterogeneity**, not
  agent count.
- **Anthropic's "90.2% multi-agent uplift"** must be retired as a justification: it is an
  internal eval with no published methodology, costs **~15× tokens**, the same post
  concedes **token spend alone explains 80% of the variance**, and Anthropic explicitly
  exclude coding domains. **[ANECDOTE/MARKETING]**
- **Cognition** and the Berkeley MAST data converge from opposite directions on the same
  rule.

**Design consequence.** **Subagents read and analyse; one thread writes.** Delegate for
context isolation and parallel I/O, never for "more perspectives". This is already
abcd's shape — the schema should encode it, not relax it.

## 8. Security findings that apply directly to our files

- **The prompt is the wrong enforcement layer.** `intent-fidelity-reviewer`'s body says
  *"You are read-only: you never edit files"* while its frontmatter declares **no
  `tools:`** — and Claude Code **inherits every tool when `tools` is omitted**. The
  sentence is decorative; the agent has write access. Prose is not a permission system.
- **Untrusted input is undeclared.** `sota-researcher` is granted `WebFetch`
  (attacker-controlled pages) with no injection defence and no `reads_untrusted_input`
  flag. Reviewers reading diffs and commit messages are in the same class.
- **Plugin agents silently drop three fields.** Claude Code **ignores `hooks`,
  `mcpServers`, and `permissionMode` for plugin subagents, explicitly for security
  reasons.** abcd *is* a distributed plugin. That exclusion list is a free,
  battle-tested threat model: a distributable agent definition that can mutate
  permissions or spawn servers is a supply-chain hole. Do not build on those three fields.
- **A stale command name.** `docs-currency-reviewer` instructs the agent to run
  `docs-currency-lint` "(on PATH)". The **tool exists** — it is `abcd docs lint`
  (`Makefile:55`, the itd-60 deterministic docs gate). Only the *name* is stale. The fix
  is a rename; deleting the instruction would remove a working deterministic pre-pass
  from the release path and force the agent to re-derive by hand what Go already checks.

## 9. What not to adopt

- **`use PROACTIVELY`** — cargo cult; not current documented guidance (§3.2).
- **Persona openers on objective tasks** — no benefit, documented safety perturbation (§4.1).
- **Reflection with no external signal** — net-negative (§4.3).
- **JSON-wrapped reasoning** — both sides of the dispute agree on reason-then-emit (§4.2).
- **Debate/committee as a default reviewer pattern** — loses to single-agent at matched
  token budget (§7).
- **Anthropic's 90.2% figure as a design justification** — unpublished methodology, 15×
  cost, coding excluded (§7).
- **A fixed "max N tools" rule** — no source supports a number. OpenAI explicitly says
  count is the wrong variable; **overlap** is ("some implementations successfully manage
  more than 15 well-defined, distinct tools while others struggle with fewer than 10
  overlapping tools").
- **crewAI's `role`/`goal`/`backstory` triad** — three required prompt fields, no routing
  separation, and `backstory` is pure persona flavour.
- **Bedrock Agents Classic** — closed to new customers 2026-07-30; its agent/action-group
  split is a dead end.
- **A rich mandated prose template** — the schema's own principal risk (§2).

## 10. Open questions

- **Semantic stop conditions are unsolved everywhere.** Six of twelve formats have a stop
  condition and **all six are a scalar cap** (`maxTurns`, `max_iter`, `max-iterations`).
  Nobody has a "you are done when X" predicate. abcd's own record wants verification
  criteria; we would be inventing this, not adopting it.
- **Output contracts are absent from every markdown format.** All four SDKs have one
  (`output_type`, `output_schema`, `response_format`, `output_content_type`); **zero
  file-based formats do.** abcd's hybrid contract is therefore ahead of the field's
  file formats and consistent with its SDKs — but it is not a convention we can copy.
- **Distribution provenance.** A2A requires `version` and JWS `signatures`; every local
  format assumes you trust the file because it is in your repo. That assumption breaks
  when definitions travel through a marketplace — which is exactly abcd's model.

## Related documentation

- [`prompting/01-general-best-practices.md`](../prompting/01-general-best-practices.md) — the
  standing prompting baseline this research updates.
- [`../brief/05-internals/01-agents.md`](../../brief/05-internals/01-agents.md) — the agent
  catalog and the frontmatter the record declares.
- itd-5 — the intent that owns the agent-file schema.
