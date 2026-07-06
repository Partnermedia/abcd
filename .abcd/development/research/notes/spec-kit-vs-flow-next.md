# GitHub Spec Kit vs flow-next — could Spec Kit replace flow-next in abcd?

Research deliverable (review session 2026-06-02). Question: could GitHub Spec
Kit replace flow-next as abcd's implementation backend, to make abcd
**implementation-tool-agnostic**? flow-next/Ralph are slated for replacement.

Relates to **itd-23** (spec-kit-interop draft) — which already frames Spec Kit
as an interop/round-trip *peer format*, NOT a replacement backend. This research
confirms that instinct and sharpens the architecture.

---

## TL;DR

**No — Spec Kit is not a drop-in replacement, and the gap is categorical, not
incidental.** Spec Kit and flow-next solve *different halves*:

- **Spec Kit = a human-in-the-loop, prompt-and-template scaffold for spec
  authoring** that stops at "ask your agent to `/implement`." Markdown-only
  artifacts. Headline strength: ~30 AI agents supported, agent-agnostic by
  design. MIT, v0.9.0 (2026-06-01), moving very fast.
- **flow-next = a structured state machine + CLI + gates + autonomous-loop
  substrate + memory store.** JSON state, `flowctl next`, plan/work/completion
  gates, the Ralph unattended loop, `reviews/`/`review-queue/`/`memory/`.

They are **complementary, not competitive.** Spec Kit owns the *authoring*
upstream; flow-next owns the *execution/state* layer abcd's loop runs on.

## What Spec Kit IS (2026, verified)

- Methodology (Spec-Driven Development) + `specify` CLI **bootstrapper** +
  prompts/templates/helper-scripts installed into your repo and agent.
- Commands: `/speckit.constitution → specify → clarify → checklist → plan →
  tasks → analyze → taskstoissues → implement`.
- On disk: `.specify/{memory/constitution.md, scripts, templates, presets,
  extensions}` + `specs/<feature>/{spec,plan,tasks,research,data-model}.md` +
  `contracts/`. Agent command files written into `.claude/`, `.github/`, etc.
- `specify` CLI = init/integration/extension/preset only; **all real work runs
  inside the AI agent via slash commands — the CLI never executes the workflow.**
- **~30 agent integrations** (Claude Code, Copilot, Cursor, Gemini, Windsurf,
  Codex, opencode, …) + generic fallback. Agent-agnosticism is the design goal.

## What Spec Kit is NOT (the crux, verified across sources)

- ❌ **No machine-readable state store** — artifacts are plain markdown; task
  state is checkbox `- [ ] T001 [P] [US1] …` only. No status enum, no
  `plan_review_status`/`completion_review_status`, no JSON.
- ❌ **No real task DAG / next-task selector** — ordering is implicit, `[P]` is
  a planning hint; "no algorithm for autonomous next-task selection."
- ❌ **No execution-time review gates** — `/clarify`/`/checklist`/`/analyze` are
  human authoring-quality gates, not loop-blocking plan/work/completion gates.
- ❌ **No autonomous unattended loop** — `/implement` is interactive, no attempt
  budget, no infra resilience. The community builds *separate* harnesses
  (e.g. `cc-sdd`) to add the loop Spec Kit lacks.
- ❌ **No review queue / reviews store, no memory/learnings substrate.**

→ Spec Kit produces excellent *spec inputs*; it does not own *execution state*
or *autonomy*.

## Capability matrix

| Capability | Spec Kit | flow-next |
|---|---|---|
| Primary role | spec-authoring scaffold | execution state machine + harness |
| Artifact format | markdown | JSON state + markdown bodies |
| Machine-readable status | ❌ checkbox only | ✅ status + review-status fields |
| Task IDs | ✅ T001 (markdown) | ✅ numbered (JSON) |
| Task DAG / deps | ⚠️ implicit + `[P]` | ✅ `depends_on_epics`, dep-aware |
| Next-task selection | ❌ | ✅ `flowctl next` |
| Review gates (plan/work/completion) | ❌ built-in | ✅ gated loop |
| Autonomous unattended loop | ❌ (3rd-party cc-sdd) | ✅ Ralph + budgets + resilience |
| Review queue / reviews store | ❌ | ✅ `reviews/`,`review-queue/` |
| Memory / learnings | ❌ (constitution ≈ static) | ✅ `.flow/memory/` |
| Multi-agent / tool-agnostic | ✅✅ ~30 agents | ❌ Claude-Code/Ralph-coupled |
| CLI | bootstrapper only | full state/loop driver |
| Maturity | v0.9.0 pre-1.0, ~10 rel/2wk | vendored, stable contract |

## Could it replace flow-next?

- **(a) Drop-in? No.** abcd reads flow-next's `.flow/` JSON contract directly
  (~38 files) and calls `flowctl` (~54 files; ~250 refs). Spec Kit has no JSON
  state, no `flowctl next`, no status fields, no gates, no loop — nothing to bind
  to.
- **(b) Missing capabilities abcd's loop depends on:** queryable state machine;
  `flowctl next`; gates; the unattended loop; reviews+memory stores.
- **(c) Spec Kit as the *authoring front-end* while abcd keeps execution/state?
  Yes — the natural fit.** A `spec/plan/tasks.md` bundle is a clean import source
  abcd can lower into its state layer. This is essentially itd-23.
- **(d) "Implementation-tool-agnostic" requires a thin backend interface**, with
  flow-next as backend #1 and Spec Kit (or others) behind it:

```python
class SpecStore(Protocol):
    def list_specs() -> list[SpecRef]
    def get_spec(id) -> Spec   # normalized: id, status, plan/completion_review_status, deps, branch
    def create_spec(intent) -> SpecRef
    def set_status(id, field, value)
class TaskStore(Protocol):
    def tasks_for(spec_id) -> list[Task]
    def next_task(filter) -> Task | None   # flow-next: flowctl next; Spec Kit: synthesize
    def set_task_status(task_id, status)
class GateRunner(Protocol):
    def run_gate(kind: Literal["plan","work","completion"], ref) -> GateResult
class Loop(Protocol):
    def select_and_run(budget, resilience) -> RunReport
```

`FlowNextBackend` implements all four (already exposes them). `SpecKitBackend`
can back `SpecStore`/`TaskStore` read/write but **must synthesize `next_task`,
`GateRunner`, `Loop` (abcd-owned) or delegate to a harness like cc-sdd.** That
asymmetry is the whole story: Spec Kit backs the *authoring* half; it cannot back
the *execution* half without abcd supplying the missing machinery.

## Already in motion: spc-34 is the first brick

`scripts/abcd/tools/flowctl_loader.py` (the spc-34 work) is abcd's **flow-next
contract loader** — it resolves and importlib-loads the *installed* upstream
flowctl by path (TRACK LATEST) — never the repo fork — and AST-extracts the
upstream verb inventory so abcd can route `flowctl <verb>` between exec-through
(the unchanged vendored fork) and abcd extensions. That is precisely the *beginning* of the SpecStore/TaskStore
seam — abcd is already decoupling from a hard-vendored fork toward a
loaded-contract model. The backend-interface recommendation extends this
existing direction rather than starting fresh.

## Strategic recommendation

1. **Adopt Spec Kit wholesale → REJECT.** Loses the loop, state machine, gates,
   memory — abcd's entire differentiator.
2. **Keep flow-next + Spec Kit interop only (status-quo itd-23) → correct &
   shippable, but insufficient** — leaves abcd hard-coupled at ~250 sites; does
   not deliver tool-agnosticism. Keep itd-23 as a *feature*, not the architecture.
3. **Thin abcd backend interface; flow-next + Spec Kit pluggable → RECOMMENDED
   primary direction.** Only option that delivers tool-agnosticism AND honours
   "flow-next/Ralph replaced later." Extends spc-34's loader seam. Cost: abstract
   ~38 `.flow/`-readers + ~54 `flowctl` sites behind the four protocols,
   incrementally (route flow-next through them first, no behaviour change).
4. **abcd grows its own minimal state+loop; Spec Kit authoring/import only →
   likely long-term endpoint**, but premature before the interface exists.

**Sequence:** ship itd-23 import as `SpecKitBackend.read → normalized model →
flow-next create` (forces the normalized model into existence — first brick) →
extract the four protocols and route flow-next through them (pure decoupling) →
later, replace the `Loop`/`GateRunner` backend (abcd-native or cc-sdd-style)
without touching abcd's command/skill surface.

**One sentence:** Spec Kit is a superb spec-authoring + multi-agent front-end and
a worthy import/export peer (vindicating itd-23), but owns none of the
execution-state / gating / next-task / autonomous-loop machinery abcd's Ralph
harness runs on — so the road to tool-agnosticism is a thin abcd-owned backend
interface (option 3), Spec Kit plugged into the *authoring* half, flow-next (then
later something else) behind the *execution* half.

## Confidence & caveats
- **High (multi-source):** Spec Kit command set, markdown-only artifacts, ~30
  agents, MIT, v0.9.0/2026-06-01, and absence of state machine / loop / gates /
  memory (corroborated by official `spec-driven.md`, MS Developer blog,
  independent reviews, and the existence of 3rd-party loop harnesses).
- **Inference (flagged):** flow-next's internal contract — reconstructed from
  abcd's observable coupling, not flow-next source.
- **Fast-moving:** Spec Kit ships multiple releases/week; re-check CLI specifics
  at adoption time. The *architecture* (authoring scaffold, no execution state)
  is stable.

## Sources
- github/spec-kit (README) https://github.com/github/spec-kit
- docs https://github.github.com/spec-kit/ · spec-driven.md https://github.com/github/spec-kit/blob/main/spec-driven.md
- tasks template https://github.com/github/spec-kit/blob/main/templates/commands/tasks.md
- releases https://github.com/github/spec-kit/releases
- MS Developer: SDD with Spec Kit https://developer.microsoft.com/blog/spec-driven-development-spec-kit
- MarkTechPost (2026-05-08) https://www.marktechpost.com/2026/05/08/meet-github-spec-kit-an-open-source-toolkit-for-spec-driven-development-with-ai-coding-agents/
- Scott Logic review (2025-11) https://blog.scottlogic.com/2025/11/26/putting-spec-kit-through-its-paces-radical-idea-or-reinvented-waterfall.html
- gotalab/cc-sdd (autonomous SDD harness on spec-kit) https://github.com/gotalab/cc-sdd
- in-repo: itd-23-spec-kit-interop.md; spc-34 tools/flowctl_loader.py
