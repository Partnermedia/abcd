# Prompt Quality Infrastructure

15 agents = 15 prompt `.md` files. Static prompts rot as models evolve and prompting best practices shift. abcd ships three layers (B+C+D) plus the itd-5 additions:

**B — Per-agent golden-test fixtures.** Each agent spec ships 2–3 fixture inputs with expected output structure (JSON schema validation + oracle-judged "is this output good enough"). Fixtures live in `agents/<name>/fixtures/`. **Shipped so far:** fixtures for `intent-fidelity-reviewer` only (the one agent that exists). The generic `scripts/abcd/test_prompts.py` harness that runs fixtures in CI is **design target** — it lands with the first Pass-A agent spec (Phase 4), the point at which a second agent exists to generalize the runner over. Catches regressions when models change.

**C — Prompt linter.** `scripts/abcd/lint_prompts.py` — static analysis of agent prompt files (`agents/<name>.md`, excluding `CHANGELOG.md`, `README.md`, and `*_template.md` / `*.template` files, which are not prompts).

The linter is delivered in two stages. **Shipped by fn-8 — the mechanical itd-5 floor (`PQ001`–`PQ006`):** `prompt_version` present and valid semver; a matching `### <agent> <version>` heading exists in `agents/CHANGELOG.md`; `capability_scope` present and well-formed; `capability_scope.task_classes` strict set-membership against `task_classes.json`; canary-fixture *presence* when `reads_untrusted_input: true` is declared.

**Deferred to the prompt-test spec — the broader structural checks:** missing role definitions, missing output schemas, vague instructions, missing example I/O, prompt-length outliers, missing `## Research basis: <path>` / `## Last SOTA Audit: <date>` footers. These role/schema/example-IO structural checks are not in fn-8's scope; they ship alongside the B-layer `test_prompts.py` golden-test harness. The linter runs pre-commit. It catches structural issues; it does not catch semantic quality.

**D — Periodic SOTA audit (research-gated).** `oracle-prompt-audit.md` template — once per minor release, run an oracle audit (RP MCP → Codex CLI → in-session subagent per [`04-universal-patterns.md § 2`](04-universal-patterns.md#2-mcp-preferred-structural-fallback)) over all agent prompts. **Reference is the agent's research file** (see "Research-driven prompts" below), not "general knowledge": "Does this prompt align with the recommendations in `.abcd/development/research/prompting/agents/<name>.md`? Where does it diverge? Is the divergence justified?" Findings go to `.abcd/logbook/sota-audits/<date>.md`. Treated as RFC input, not auto-applied.

**itd-5 prompt-quality additions on top of B+C+D:**

- **`prompt_version` frontmatter** — every `agents/*.md` carries `prompt_version: <semver>` alongside existing `name`, `description`, `tools`, `model`. Initial value `1.0.0`. A consolidated `agents/CHANGELOG.md` records each version bump with: agent name, old → new version, one-line rationale, eval delta (golden-test pass/fail count change). Bump rules (semver-adapted): MAJOR for behaviour-breaking output schema change; MINOR for behaviour change preserving schema; PATCH for typo / non-behavioural edit.
- **One-shot oracle self-improvement pre-flight** — before each agent's prompt is locked at `1.0.0`, the author runs the pre-flight: submit candidate prompt to `lifeboat-oracle` with rewrite-for-clarity directive; run all golden-test fixtures against both candidate and oracle-rewritten variants; if oracle variant ≥ candidate on goldens AND shorter by >10%, accept oracle variant; otherwise keep candidate. Log decision + diff in `agents/CHANGELOG.md` as the agent's first entry. One-time gate per agent at v1.0.0 lock-time, not recurring.
- **Injection-canary fixtures** — every agent that reads untrusted input (transcripts, lifeboat content, GitHub issues, commit messages, model-emitted reviews) MUST have at least one fixture with a prompt-injection payload. Agents in scope: `chat-distiller`, `embark-scaffolder`, `issue-scout`, `code-rescuer`, `decision-archaeologist`, `review-collator`. Failing the canary fixture blocks the agent's spec from closing.

**Research-driven prompts.** Every agent has a research file at `.abcd/development/research/prompting/agents/<name>.md` — current SOTA best practices for that agent's role (e.g., "best practices for prompting a content-fidelity auditor", "best practices for prompting a product-thinker critic"). One general baseline at `.abcd/development/research/prompting/01-general-best-practices.md` covers cross-cutting prompting SOTA (structure, role definition, output formats, extended thinking, examples, etc.).

- **General baseline** is its own early flow-next spec, sequenced before any agent spec in Phase 3 onwards
- **Per-agent research** is task #1 of each agent's spec (research before prompt drafting; blocks subsequent prompt+fixture tasks)
- **Coupling philosophy:** research is **gate** (validation criterion), not **source** (template generator) — author writes the prompt informed by research, oracle audit checks alignment after the fact. Author retains freedom; auditor has ammunition
- Research files committed to `.abcd/development/research/` per the abcd CLAUDE.md doc structure; survive the prompt and outlive specific model versions

**Per-agent spec acceptance includes:**
- Research file exists at `.abcd/development/research/prompting/agents/<name>.md`, oracle reviews it as "non-trivial" (specific findings, not empty bullets)
- Prompt passes the linter (including `prompt_version` frontmatter present per itd-5)
- Prompt cites research file in `## Research basis: .abcd/development/research/prompting/agents/<name>.md` footer
- ≥2 golden-test fixtures pass
- Injection-canary fixture (per itd-5 Add 3) passes for agents reading untrusted input
- One-shot self-improvement pre-flight outcome recorded in `agents/CHANGELOG.md`
- Prompt has a `## Last SOTA Audit: <date>` footer line (initial value: spec completion date)

**In a later phase (recorded as intents):**
- **itd-14 — Prompt registry + versioning** — full diff-on-update workflow, treated like code (heavier rigour layer than itd-5's `prompt_version` field)
- **itd-15 — Self-dogfooded SOTA audit** — abcd's own disembark of abcdDev runs the prompt audit as part of Pass C (eat-your-own-dogfood)
