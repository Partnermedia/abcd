# Related Work

abcd shares DNA with several Claude Code frameworks. We arrived at similar conclusions independently — and in places the convergence is striking — but abcd's scope (lifeboat-shaped pack/unpack across repos, intent-as-press-release, multi-stage development workflow) is broader than any single one. This document records what each prior-art project does, where abcd overlaps, and why abcd doesn't simply depend on them.

## TL;DR

| Project | One-line | abcd overlap | Status as dependency |
|---|---|---|---|
| [PAUL][paul] | Project orchestration framework — Plan / Apply / Unify loop with mandatory closure and BDD acceptance criteria | Loop discipline, acceptance gates, state-restore-after-break | **Pattern adopted, not depended on** (see `itd-1-acceptance-gates`) |
| [CARL][carl] | Just-in-time rule injection — domain-keyed rules loaded on prompt-keyword recall | Modular rule loading instead of monolithic CLAUDE.md | **Pattern adopted, not depended on** (see `itd-3-modular-rules-loader`) |
| [claude-skills][claude-skills-rezvani] | 232+ skill/agent packages spanning engineering, marketing, compliance | Progressive-disclosure skill format | Use upstream `SKILL.md` schema; do not bundle the catalogue |
| [everything-claude-code][everything-claude-code] | Agent-harness performance optimisation: skills, instincts, memory, security, research-first | Memory-as-knowledge, research-first prompting | Inspirational; abcd has its own memory model (`.abcd/memory/`) |
| [wshobson/agents][wshobson-agents] | Multi-agent orchestration patterns | Agent dispatch + coordination | Inspirational; abcd's 15 agents are purpose-built for the lifeboat workflow |
| [Karpathy LLM Wiki][karpathy-llm-wiki] | Self-maintaining markdown knowledge base curated by an LLM (raw sources → wiki → schema) | Pattern shape for `.abcd/memory/` (multi-upstream curated substrate per itd-36) | **Pattern adopted**; component description in `05-internals/<n>-memory.md` |
| [Naur 1985][naur-1985] | "Programming as Theory Building" — a program is a theory held by people who build it; code is a lossy projection | Recovery-humility framing (lifeboat = floor, not theory); itd-37 Modification Grammar discipline closes the Modification axis | **Pattern adopted, philosophy citation**; see mental-model § The Naurian gap |
| [Dell'Acqua et al. 2023][jagged-frontier] | "Navigating the Jagged Technological Frontier" — AI capability is jagged; users can't see when they're outside it | itd-5 `capability_scope` frontmatter (static); Frontier Awareness intent (dynamic events + pre-cascade selector, in a later phase) | **Framing source**; cited in `05-internals/01-agents.md § Agent prompt frontmatter` |

## PAUL — Plan-Apply-Unify Loop

[PAUL][paul] is a structured-development framework that enforces a three-phase loop: PLAN (scope-adaptive, with acceptance criteria) → APPLY (execute / qualify against spec) → UNIFY (reconcile plan vs. actual, log decisions, update state). Its central principle: **"every plan closes with UNIFY — no orphan plans."**

Key primitives:

- `.paul/PROJECT.md`, `ROADMAP.md`, `STATE.md`, `phases/NN-name/NN-NN-PLAN.md` + `…-SUMMARY.md`.
- BDD-format acceptance criteria (`Given … When … Then …`) defined *before* tasks — see [Given-When-Then][bdd-given-when-then].
- Escalation statuses (`DONE` / `DONE_WITH_CONCERNS` / `NEEDS_CONTEXT` / `BLOCKED`) instead of binary pass/fail.
- `/paul:resume` for state restoration across sessions.

### Where abcd overlaps

| PAUL | abcd |
|---|---|
| `.paul/PROJECT.md` (requirements & context) | `.abcd/development/brief/README.md` |
| `.paul/ROADMAP.md` | `.abcd/development/roadmap/` |
| `.paul/STATE.md` (loop position) | `.flow/` checkpoint files + `.abcd/logbook/` session logs |
| `phases/NN-PLAN.md` + `…-SUMMARY.md` | flow-next epics + tasks |
| Plan → Apply → Unify | `/flow-next:plan` → `/flow-next:work` → plan-sync + intent-fidelity-reviewer |
| BDD acceptance | **First-class** — landed by [`itd-1-acceptance-gates`](../intents/disciplines/itd-1-acceptance-gates.md) (the first intent shipped, so all later intents inherit the discipline) |
| `/paul:resume` | `/abcd:embark` (lifeboat unpack — post-disembark, not mid-flight) |

### Why not depend on PAUL

PAUL writes its own `.paul/` namespace and ships its own commands (`/paul:init`, `/paul:plan`, `/paul:apply`, `/paul:unify`). abcd already owns `.abcd/` and a five-plus-command shape (`ahoy`, `disembark`, `embark`, `launch`, `intent`, plus `capture` once itd-4 ships). Two scaffolders fighting over a project's structure would be the wrong outcome. **abcd takes the patterns** (mandatory closure, BDD acceptance, escalation statuses) into its own primitives — see `itd-1`.

## CARL — Context Augmentation & Reinforcement Layer

[CARL][carl] solves a problem abcd has acutely felt: **CLAUDE.md bloat**. Static prompts in CLAUDE.md occupy context every session, even when totally irrelevant to the current task. CARL replaces the monolith with on-demand rule injection.

Architecture in one paragraph:

- Single JSON source of truth at `~/.carl/carl.json` (mergeable with project-local `./.carl/carl.json`).
- Domains (e.g. `DEVELOPMENT`, `GLOBAL`) carry `state`, `recall` keywords, `rules`.
- Python `UserPromptSubmit` hook scans each prompt, matches against `recall`, **injects only matching domain rules** as a system message.
- Signature-based dedup avoids re-injecting identical rules; forces full refresh every N prompts.
- An MCP server exposes `carl_v2_add_rule` / `toggle_domain` / `log_decision` for runtime self-management.
- "Star-commands" (`*brief …`) bypass keyword matching for explicit activation.

### Where abcd overlaps

abcd's 978-line `~/ABCDevelopment/.claude/CLAUDE.md` is exactly the failure mode CARL addresses. Most sessions need none of those rules; every session pays for all of them. The plugin-bundled defaults + per-repo `.abcd/rules.json` model proposed in [`itd-3-modular-rules-loader`](../intents/planned/itd-3-modular-rules-loader.md) is structurally identical to CARL.

### Why not depend on CARL

CARL is general-purpose; abcd is opinionated about which domains and rules belong in an abcd project. Bundling default rules with the abcd plugin (rather than asking users to populate `~/.carl/carl.json` separately) is the right ergonomics for "the fresh start" goal. abcd takes **the mechanism** (hook + JSON + recall keywords + dedup) into its own `hooks/prompt_router_hook.py` and ships abcd-shaped defaults out of the box.

If CARL ever ships a stable hook ABI that other frameworks can plug into without depending on the full CARL runtime, this could change.

## Claude Code Skills (Anthropic-native)

abcd's skill-shaped capabilities use the standard [SKILL.md][agent-skills-overview] format — not a custom system. See [Claude Code Skills docs][claude-skills-docs].

Each skill's frontmatter (~100 tokens) loads eagerly; the body loads only when the agent decides the skill applies. This is **progressive disclosure** and is the right primitive for procedural workflows ("here's how to write a milestone retrospective").

The split abcd uses:

- **Hook-injection (CARL-style)** — declarative rules ("never commit `/Users/` paths").
- **Skills (native)** — procedural workflows.
- **Agents** — bounded tasks ("audit this PR").

Three loading mechanisms, three jobs, no overlap.

## Karpathy LLM Wiki — pattern source for `.abcd/memory/`

[Karpathy's LLM Wiki gist][karpathy-llm-wiki] (April 2026) describes a three-layer pattern: **raw sources** (immutable inputs the LLM reads but never modifies) → **the wiki** (LLM-generated markdown organised by entity / concept / synthesis, owned entirely by the LLM) → **the schema** (config doc defining structure, conventions, workflows). Karpathy's claim vs RAG: *"the LLM is rediscovering knowledge from scratch on every question. There's no accumulation."* Wiki pattern compiles synthesis at ingest time, not query time.

abcd adopts the pattern as the organising shape of `.abcd/memory/` (per itd-36). Multiple upstream pipelines feed in (vendor session memory + external sources + oracle reviews + work notes + dredge synthesis); `principle-distiller` is the curator. Schema extension to existing flat naming (`<type>_<domain>_<slug>.md`): adds `index.md`, `log.md`, `contradictions.md`, plus typed `source:` frontmatter with curator semantics. Karpathy's scope caveat (*"small-to-medium, slow-moving, human-curated"*) carries forward into abcd's per-project framing: cross-project sharing comes in a later phase.

**Default-no-originals + bounded quotation.** abcd diverges from Karpathy's "user keeps the source corpus alongside the wiki" framing because of copyright/licence exposure: `/abcd:memory ingest` reads sources, distils into typed entity/topic pages with citation frontmatter, and discards the original by default. `--keep-original` flag opts in; launch-gate refuses to publish under `.abcd/memory/sources/` unless allowlisted. Quotation budgets per-page (≤5%, no contiguous span >150 words) plus cumulative-source-coverage lint (≤25% deduplicated coverage across all pages citing one source) close the laundering vector.

## Naur 1985 — Programming as Theory Building

Peter Naur's 1985 paper "[Programming as Theory Building][naur-1985]" (in *Microprocessing and Microprogramming* 15(5)) is the foundational philosophical citation for abcd's recovery semantics. Naur's thesis: **a program is not its source code** — it is a *theory* held by the people who built it. Code, docs, and tests are lossy projections of that theory. Three tacit-knowledge areas the theory captures:

1. **Mapping** — how the world maps to the program; what's deliberately not represented.
2. **Justification** — why each load-bearing decision was made; alternatives rejected.
3. **Modification** — what extends cleanly, what breaks the design, the structural rules.

When all theory-holders leave, the program enters Naur's *"dead"* state: still running, but no longer intelligently modifiable from artefacts alone.

abcd adopts Naur's framing in three places:

- **Recovery humility** on `/abcd:disembark` and `/abcd:embark` surfaces (the lifeboat is the floor of recoverable theory, not theory itself; receivers should hunt the originating session, not trust the lifeboat blindly).
- **itd-37 Modification Grammar discipline** closes the Modification axis — every epic carries a `## Modification Grammar` section with three sub-headings (`Extends cleanly` / `Breaks the design` / `Why`), extracted by `principle-distiller` at epic completion into typed memory pages.
- **The Naurian gap** sub-section in `01-product/03-mental-model.md` names the three areas explicitly and identifies Modification as the one abcd is genuinely closing (Mapping and Justification are partially captured by press releases + audit notes already).

abcd does NOT claim to recover full theory; per Naur the claim would be incoherent. The lifeboat is *a* floor, not *the* theory.

## Dell'Acqua et al. 2023 — Jagged Frontier

The [Dell'Acqua / McFowland / Mollick et al. paper][jagged-frontier] (HBS Working Paper 24-013, Sep 2023; published *Organization Science* 2025) is a field experiment with 758 BCG consultants. Headline finding: **AI's capability frontier is jagged** — tasks that look superficially similar can sit on opposite sides of "AI helps" vs "AI hurts." Inside the frontier: 12.2% more tasks, 25.1% faster, 40%+ higher quality. **Outside** the frontier: AI users perform 19 percentage points worse than the no-AI control. Workers fail to see when they're outside the frontier; over-reliance is the failure mode.

abcd adopts the framing in two places:

- **itd-5 extension** (per idea-4 jagged-frontier review): every agent's frontmatter declares `capability_scope` (the task classes the agent is designed for; closed enum + free-text designed_for). Same artefact class as `prompt_version`. Cheap (~5 min/agent at v1.0.0 lock). `intent_lint.py` validates set-membership at plan-time; semantic judgement ("is THIS task within agent X's frontier?") comes in a later phase.
- **Frontier Awareness intent** (in a later phase; no ID reserved per the idea-3 release-don't-reserve precedent): owns the dynamic half — `known_failure_modes` events appended from canary/golden-test/operator-tagged failures (NOT epic-level criterion verdicts; verdict-family disjoint per itd-1); plan-time semantic check at `/flow-next:plan-review` via Role 2 sub-check; capability-aware **pre-cascade selector** (a layer ABOVE the oracle cascade defined in itd-2 + itd-6 — does NOT modify the cascade contract); `/abcd:frontier` command (bare = render current frontier-map; sub-verbs `flag`/`history`/`explain` for actions bare can't do).

**Failure-mode tag enum** (closed, PR-to-extend; vocabulary for that later phase): `{hallucination, scope_drift, stale_context, under_specification_blindness, format_violation}`. Deliberately drops `truncation` (transport failure, oracle observability concern) and `over_confidence` (subsumed by `under_specification_blindness` and `hallucination`).

## Other prior-art

- **[claude-skills][claude-skills-rezvani]** — 232+ packaged skills. Useful as a catalogue and as `SKILL.md` schema validation; abcd doesn't bundle them. If a user has them installed alongside abcd, both should compose without conflict (different namespaces).
- **[everything-claude-code][everything-claude-code]** — agent-harness performance system. Notable for its memory-as-knowledge framing and research-first development. abcd has its own memory model (`.abcd/memory/`, vendor-agnostic dispatcher in `adapters/memory.py`).
- **[wshobson/agents][wshobson-agents]** — multi-agent orchestration patterns. abcd's 15 agents are purpose-built for the lifeboat workflow rather than general-purpose orchestration.

## Naming and credit

abcd commands borrow the **press-release-first** framing from [Amazon Working Backwards][amazon-working-backwards] (intents written as press releases before implementation), which predates and is independent of PAUL/CARL. The convergence with PAUL on acceptance-criteria-first development is independent.

When the public README is generated by `/abcd:launch`, it should include a "Compare to" section that links PAUL, CARL, and claude-skills directly so external readers understand the lineage.

## References

[paul]: https://github.com/ChristopherKahler/paul "PAUL — Plan-Apply-Unify Loop, project orchestration framework for Claude Code (Kahler)"
[carl]: https://github.com/ChristopherKahler/carl "CARL — Context Augmentation & Reinforcement Layer, just-in-time rule injection for Claude Code (Kahler)"
[claude-skills-rezvani]: https://github.com/alirezarezvani/claude-skills "claude-skills — large skills/agents collection for Claude Code and other harnesses (Rezvani)"
[everything-claude-code]: https://github.com/affaan-m/everything-claude-code "everything-claude-code — agent harness performance optimisation system (Mahmood)"
[wshobson-agents]: https://github.com/wshobson/agents "wshobson/agents — multi-agent orchestration for Claude Code"
[claude-skills-docs]: https://code.claude.com/docs/en/skills "Claude Code Skills (Anthropic docs)"
[agent-skills-overview]: https://platform.claude.com/docs/en/agents-and-tools/agent-skills/overview "Agent Skills overview (Anthropic platform docs)"
[bdd-given-when-then]: https://martinfowler.com/bliki/GivenWhenThen.html "Given-When-Then (Fowler) — BDD acceptance-criteria pattern"
[amazon-working-backwards]: https://www.allthingsdistributed.com/2006/11/working_backwards.html "Working Backwards (Vogels) — Amazon press-release-first product design"
[karpathy-llm-wiki]: https://gist.github.com/karpathy/442a6bf555914893e9891c11519de94f "Karpathy LLM Wiki gist (April 2026) — self-maintaining markdown knowledge base curated by an LLM (raw sources / wiki / schema three-layer pattern)"
[naur-1985]: https://gwern.net/doc/cs/algorithm/1985-naur.pdf "Naur (1985) — Programming as Theory Building, Microprocessing and Microprogramming 15(5):253-261"
[jagged-frontier]: https://www.hbs.edu/faculty/Pages/item.aspx?num=64700 "Dell'Acqua, McFowland, Mollick et al. (2023) — Navigating the Jagged Technological Frontier, HBS Working Paper 24-013; published Organization Science 2025"
