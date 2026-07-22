# Agent Catalog

The catalog below declares the **16-agent design roster** (not all are shipped yet; the **Status** column marks which of these exist today, and [`06-delivery/`](../06-delivery) carries current delivery state); the shipped agents that live outside this roster are listed under [§ Shipped agents outside the design roster](#shipped-agents-outside-the-design-roster). Ten agent prompt files ship in `agents/` today. Each catalog agent declares JSON inputs/outputs (schemas owned by the core, `internal/core/schema`). Agents are **markdown**, host-delegated reviewers the host dispatches (adr-25); markdown is rendered, not authored by agents.

`agents/` also holds two plain docs — `agents/README.md` and `agents/CHANGELOG.md` — that carry no agent frontmatter. Because `.claude-plugin/plugin.json` declares no `agents` key, the plugin loader globs the flat `agents/*.md` set and registers both as harness agents (`abcd:README`, `abcd:CHANGELOG`) alongside the real prompt files; iss-110 tracks the mis-registration.

| Agent | Pass | Status | Inputs (JSON) | Outputs (JSON) |
|---|---|---|---|---|
| `flow-essence` | A | design target — Phase 6 | native spec store (newest-first) | `spec-essence.json` |
| `decision-archaeologist` | A | design target — Phase 6 | ADRs, CLAUDE.md, git log | `decisions-timeline.json` |
| `review-collator` | A | design target — Phase 6 | `.abcd/work/reviews/oracle-review-*` | `reviews-consolidated.json` |
| `code-rescuer` | A | design target — Phase 6 | codemap adapter (when wired) or spec-window file selection | `code-principles.json` |
| `chat-distiller` | B | design target — Phase 6 | spine entry + time-windowed transcript subset (per call) | per-call: emits `rationale-fill`, `unrecorded-decision`, `pitfall` delta entries; aggregated across calls into `research/rationale-fills.json`, `research/unrecorded-decisions.json`, `research/pitfalls.json` (delta). Density measured on user-message denominator per Phase 0 Measurement Deviation; see `research/phase/0/transcript-sampling.md`. |
| `principle-distiller` | C | **shipped** — `disembark principles` | `.abcd/memory/`, ADRs, conventions, `spec-essence.json` (spine), `code-principles.json`, `candidate-pitfalls.json` (from review-collator), Pass B distiller deltas | `principles.json` (domain-grouped, with four-source pitfall dedup by topic-hash) |
| `artefact-curator` | C | design target — Phase 6 | user-docs (tutorials/guides/reference/explanation), assets (logos/charts/screenshots) | `assets/_manifest.json` with per-item classification: `keep` (copy verbatim), `adapt` (suggest adaptation, embark prompts user), `drop` (skip silently). Also writes `docs/` lifeboat copies. |
| `brief-composer` | C | design target — Phase 6 | `spec-essence.json`, `decisions-timeline.json`, `principles.json`, `code-principles.json`, `reviews-consolidated.json`, `rationale-fills.json`, `unrecorded-decisions.json`, `pitfalls.json`, `assets/_manifest.json` | `README.json` (lifeboat brief synthesising all Pass A/B/C inputs) |
| `press-release-composer` | C | **shipped** — `disembark press-release` | spec-essence, decisions, principles, source CLAUDE.md/README/metadata | `press-release.json` + invokes the oracle seam for a product-thinker audit (host-delegated by default) → `audit/press-release-oracle-<ts>.json` |
| `issue-scout` | C (opt-in) | design target — Phase 6 | `.abcd/work/issues/<slug>.md` entries | annotated entries with "Related upstream" sections; uses `gh` CLI. **Peer-preferred:** delegates to a peer github-scout over MCP when one is present; this native agent is the default. Default: disabled in `config.json`; ahoy asks. |
| `lifeboat-oracle` | C (audit) | **shipped** — `disembark oracle` | rendered MD + JSON corpus | `audit/oracle-<ts>.json` |
| `embark-scaffolder` | embark | design target — Phase 6 | lifeboat JSONs + target probe | `scaffold-plan.json` + extended responsibilities for doc-architecture at scaffold time (legacy harvest) |
| `launch-gatekeeper` | launch | design target — itd-65 (gate suite, per [adr-33](../../decisions/adrs/0033-launch-phase-ownership-tiered.md)) | scan results + payload manifest | `preflight.json` + extended responsibilities for doc updates and OWASP/security auditing at promotion time (legacy harvest) |
| `intent-fidelity-reviewer` | intent — **three roles, three verbs** (`review` / `consistency` / `shape`) | **shipped** (`review` discipline subset, spc-12); `consistency` / `shape` design targets (spc-29) | varies by role (see [`04-surfaces/05-intent.md § 7`](../04-surfaces/05-intent.md#7-the-intent-fidelity-reviewer-agent-three-roles-three-verbs)) | varies by role |
| `documentation-auditor` | subagent (pre-pack, post-scaffold, pre-promotion) | design target — Phase 6 (disembark/embark) + itd-65 (launch pre-promotion) | source `docs/` directory or lifeboat `docs/` directory | `documentation-audit-<ts>.json` — invoked by disembark (pre-pack), embark (post-scaffold), and launch (pre-promotion). Subagent-only; not user-invoked. Lifted from legacy `~/ABCDevelopment/.claude/agents/documentation-auditor` |
| `reflection-composer` | reflect | design target — spc-83 (thin V1) | selected spc-66 phase-audit receipt (`.abcd/logbook/audit/phase-<ts>/report.json`) — per-bullet acceptance verdicts + `member_specs` + `done_total` | single JSON object with the five retrospective section keys (`went_well` / `could_improve` / `lessons_learned` / `decisions_made` / `metrics`); the deterministic reflect writer (`internal/core/reflect`) renders `.abcd/retrospectives/<phase-id>/README.md`. Dispatched by `/abcd:reflect <phase-id>` (itd-24). Phase-only grain. |

## Shipped agents outside the design roster

Six further agent prompts ship in `agents/` today, outside the 16-agent design roster above — lifeboat/launch synthesis helpers and repo-workflow reviewers, each a **markdown**, host-delegated agent the host dispatches:

| Agent | Purpose |
|---|---|
| `graveyard-interpreter` | Interpret a packed lifeboat's graveyard into cited lessons (each citing the layer-1/2 finding ids it rests on) — feeds `abcd disembark graveyard <lifeboat-dir> --lessons-json`. |
| `release-changelog-composer` | Compose one release cut's changelog prose, every line citing the record id it reports so the binary can prove the cut — feeds `abcd launch ship --changelog-json`. |
| `docs-currency-reviewer` | Semantic docs-currency review — verifies every user-facing claim against the code that implements it (the release-gate docs check). |
| `ruthless-reviewer` | Demanding senior code review — correctness, resource handling, error paths, API misuse, dead code — run before presenting a non-trivial diff. |
| `security-reviewer` | Adversarial security review of a diff or design touching a trust boundary (auth, secrets, network, input parsing, subprocess). |
| `sota-researcher` | Deep state-of-the-art research — ranked recommendations with evidence tiers and source attributions. |

## `intent-fidelity-reviewer`'s three roles, three verbs

The catalog row above declares one agent with three roles, sharing the agent's prompt scaffolding, oracle backend resolution, and receipts. Each role has its own user-facing verb under `/abcd:intent` — no role-by-kind dispatch:

1. **Single-document fidelity → `/abcd:intent review <itd-N>`** (per the itd-1 discipline). **Inputs (updated by spc-3/itd-27):** shipped intent's press release + acceptance criteria + delivered reality + **`terminology/` glossary** (when present) + **frozen PRD at `.abcd/intents/<itd-N>/prd.md`** (when present). **Outputs (updated by spc-3/itd-27):** per-criterion verdicts (MET/MET_WITH_CONCERNS/NOT_MET/INCONCLUSIVE) + three-bucket prose audit (honoured/diverged/missing) + **term-drift findings** (terms used in delivered reality that have drifted from `terminology/` canonical definitions) + **PRD-fidelity findings** (delivered reality vs the frozen PRD's user stories and implementation/testing decisions — what was honoured, diverged, or missing). **Two passes, two destinations:** the itd-1 acceptance pass (a shipped intent) writes per-criterion verdicts into the intent's own `## Audit Notes` (the verdict of record) plus a per-run `audit/review-<ts>/` logbook report; the itd-37 `MG004` pass (a native spec's `## Modification Grammar`) writes its `PASS`/`FAIL` verdict to an `audit/spec-mg-<ts>/` logbook receipt (native specs have no `## Audit Notes` section). For bundles, runs per-member-intent against the same delivered reality. The verb name aligns with the native review vocabulary `plan-review` / `impl-review` / `completion-review`; `audit` is reserved for the top-level `/abcd:audit` (compliance / hash-chain). **spc-12 ships the discipline-judgement subset only** — the itd-1 per-criterion acceptance verdicts and the itd-37 `MG004` boilerplate check, with their writers and receipts; the broader prose / term-drift / PRD-fidelity outputs above are **deferred** to a later spec. spc-12 ships the **manual** `/abcd:intent review` surface. spc-28 then shipped the **on-close lifecycle hook**: when a linked native spec closes, the intent moves `planned/` → `shipped/` and a review is **queued** on that transition (spc-28 also backfilled already-shipped intents that never had a review queued). What remains deferred is **automatic firing of the reviewer** off the queue — the move and the queue entry are produced automatically, but running `intent-fidelity-reviewer` on a queued entry is still a manual `/abcd:intent review <itd-N>` step (spc-6 disowned auto-firing; no spec currently owns it).
2. **Cross-document fidelity → `/abcd:intent consistency [<itd-N>]`** (introduced by itd-48, which superseded itd-31). Inputs: brief + every intent. Outputs: five judgement-category findings (terminology drift, premise contradictions, scope leakage, sequencing impossibilities, naming conflicts) at `.abcd/logbook/audit/consistency-<ts>/report.{json,md}`. spc-29 ships the judgement half on demand via `/abcd:intent consistency` (bare = whole corpus; with `<itd-N>` = one intent vs the rest). **Deferred follow-up** (recorded in the `.abcd/work/issues/` ledger under `[spc-29 follow-up]`): the Role 2 mechanical half (schema/state contradictions, reference rot, acknowledgement gaps) and pre-commit hook scheduling that would let consistency findings block commits.
3. **Kind classification → `/abcd:intent shape [<itd-N>]`** (introduced alongside the three intent kinds in itd-34). Inputs: intent corpus (cross-references, scope sections, supersession candidates). Outputs: suggested reclassifications across the three live `suggestion_type` values (`kind_change`, `bundle`, `supersession`) at `.abcd/logbook/audit/shape-<ts>/report.{json,md}`. spc-29 ships the on-demand surface only: `/abcd:intent shape` (bare = corpus; with `<itd-N>` = one intent). Bare `/abcd:intent` (status+help) *surfaces the latest cached shape suggestions* in its summary output — bare invocation never runs a fresh scan or mutates the report. User accepts via `/abcd:intent reclassify`; declined suggestions are logged so they aren't re-surfaced. Concurrency between any future scheduled invocation and the on-demand verb is mediated by `flock(2)` on `.abcd/coordination/shape.lock` — see `04-surfaces/05-intent.md § 7` for the contract. **Deferred follow-up** (recorded in the `.abcd/work/issues/` ledger under `[spc-29 follow-up]`): pre-commit hook wiring for continuous shape scanning (`shape(...)`'s `mode="pre_commit"` parameter is preserved as a seam but no hook invokes it).

The `intent-fidelity-reviewer`'s three roles still share **one** catalog entry — the count grows by user-facing responsibility, not by role. The roster reached **16** when spc-83 added `reflection-composer` (the `/abcd:reflect` retrospective composer, itd-24) — a genuinely new user-facing responsibility, not a role of an existing agent. See [`04-surfaces/05-intent.md § 7`](../04-surfaces/05-intent.md#7-the-intent-fidelity-reviewer-agent-three-roles-three-verbs) for the user-facing description of the reviewer's roles.

## Oracle backend resolution

**Scope, per adr-25:** agents that need a model reach it through the `oracle` seam ([`02-adapters.md`](02-adapters.md)). The default is **host-delegated**: abcd's core does the deterministic work and hands a **prompt** to the host's subagent dispatch (the agent harness driving abcd); the host owns model choice, credentials, and execution, and abcd consumes the structured result. The default install needs no API keys and no model config — abcd emits prompts and the host runs them.

Concrete oracle backends are **opt-in adapters** behind the same seam, selected when an operator wants abcd to reach a model directly:

- **native** — abcd calls a local model itself.
- **cli** — abcd shells to a model CLI (e.g. `codex exec`) as a non-interactive subprocess.
- **api** — abcd calls a provider API directly.
- **mcp** — abcd calls a model over MCP (e.g. RepoPrompt routing to the user's configured model, or codex over MCP).

`oracle.backend` config: `"host-delegated"` (default) `| "native" | "cli" | "api" | "mcp"`. The default needs no external tool; an explicit value selects a wired adapter, and an unreachable adapter degrades to the host-delegated default rather than blocking. Per [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching), the `oracle` seam is one interface with a native default and opt-in adapter shapes — never a fixed cascade the core imposes.

**Adapter guidance for high-stakes reviews (adr-25).** When an operator wires two oracle adapters, a **scoped** reviewer (seeing only a selection) and a **broad** reviewer (reasoning over the whole repo) have complementary blind spots and are trusted **asymmetrically** — the scoped verdict gates, the broad reviewer is mined for findings, and the review-fix loop declares its stopping rule up front. This is advice the adapter layer offers, not a pipeline the core imposes.

The same seam serves multiple purposes:
- `lifeboat-oracle` — generic content-fidelity audit ("does the lifeboat match the source?")
- `press-release-composer` — product-thinker audit on the press release ("would a product person come away with a true mental model?")
- `intent-fidelity-reviewer` — three roles share the seam: single-document fidelity (per itd-1), cross-document fidelity (per itd-48, which superseded itd-31), and shape classification (per itd-34)
- Future audits use the same `oracle` seam with their own prompt templates.

## Verdict-tag protocol

Two verdict-enum families are in play across abcd:

**1. Review verdicts** (oracle reviews of plans, implementations, completions; emitted as `<verdict>...</verdict>` tags in oracle output): `{SHIP, NEEDS_WORK, MAJOR_RETHINK}`. This is abcd's native review-verdict enum and a **shared convention** at the `spec`/`run` seam boundary (adr-24, adr-26): a review gates a **receipt** in the native run loop (adr-27) or a wired peer loop alike, so any abcd-produced review is portable across the seam without a tool-specific validator.

**2. Per-criterion intent acceptance verdicts** (per itd-1, used by `intent-fidelity-reviewer`): `{MET, MET_WITH_CONCERNS, NOT_MET, INCONCLUSIVE}`. These describe whether each acceptance bullet was honoured; rollup logic lives in [`04-surfaces/05-intent.md § 7`](../04-surfaces/05-intent.md). Different category from review verdicts — review verdicts assess a *change*, criterion verdicts assess a *promise vs reality*.

These two enums are deliberately disjoint and never mixed. Reviews emit family 1; auditors emit family 2.

## Agent prompt frontmatter

Every agent's prompt file carries declared frontmatter. Current fields:

| Field | Required | Source | Purpose |
|---|---|---|---|
| `prompt_version` | yes | itd-5 | Semver of the agent prompt; bumps on any prompt change |
| `capability_scope` | synthesis/composer prompts | itd-5 extension (idea-4 jagged-frontier) | Declared task classes the agent is designed for. Object: `{ task_classes: [<token>, ...], designed_for: "<free-text 1-line>" }`. **`task_classes` is authored as a YAML inline list** — `task_classes: [spec_planning, audit]` on one line, never a block list of `- token` items (the frontmatter parser does not support a block list nested under a nested key). Carried today by the five synthesis/composer prompts (`graveyard-interpreter`, `release-changelog-composer`, `press-release-composer`, `principle-distiller`, `lifeboat-oracle`); the reviewer prompts omit it. Set-membership lint in `internal/core/lint` against the `task_classes` closed enum (owned by `internal/core/schema`; prose counterpart in `02-constraints/04-naming.md`, PR-to-extend) is a design target — no shipped check reads the field. Cites Dell'Acqua et al. 2023 ("Navigating the Jagged Technological Frontier") as the framing source. |
| `reads_untrusted_input` | conditional (agents that read untrusted input) | itd-5 | Boolean. `true` declares the agent reads attacker-influenceable input (transcripts, lifeboats, GitHub issues, commit messages, model-emitted reviews). When `true`, the agent MUST carry at least one canary fixture under `agents/<name>/fixtures/`; a fixture-presence prompt-lint in `internal/core/lint` is a design target — no shipped check enforces it (the M6 agents ship conforming files, not the linter — see `agents/README.md`). |

**Deliberately omitted from agent frontmatter** (boundary against scope creep, per idea-4's static/dynamic split): `known_failure_modes` (runtime-appended events), per-task-class `model_id` history, plan-time capability gating output. These belong to the later-phase Frontier Awareness intent — capture-stable, no ID reserved.

**Why `capability_scope` is in itd-5 (cheap) and not its own discipline:**

- Same artefact class as `prompt_version`: agent frontmatter, versioned with the prompt, mechanical to write at v1.0.0 lock.
- ~5 min/agent at v1.0.0 lock; itd-5 stays cheap.
- Lint validation in `internal/core/lint` is strictly set-membership (every `task_classes` token must be in the `task_classes` enum). NEVER inference — the linter does not read `designed_for` prose or task descriptions to judge scope. Anything fuzzier (semantic judgement: "is THIS task within agent X's frontier?") is the later-phase Frontier Awareness Role 2 sub-check.

**Oracle seam contract preserved.** The `oracle` seam (host-delegated by default, opt-in adapters per adr-25, with the framing in [`04-universal-patterns.md § 7`](04-universal-patterns.md#7-vendor-agnostic-adapters-with-environment-branching)) is unchanged by the `capability_scope` field. Capability-aware routing — when it ships in a later phase — is a *pre-dispatch selector* layer above the seam, NOT a modification to the seam contract. Selector consumes `(task_class, agent, model_id) → preferred oracle backend`; the seam consumes `(backend) → host-delegated-or-adapter dispatch`. Thin seam.
