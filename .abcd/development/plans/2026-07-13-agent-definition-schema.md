# Plan: agent-definition schema and its enforcement

Companion to the research record:
[`../research/notes/2026-07-13-agent-definition-schema.md`](../research/notes/2026-07-13-agent-definition-schema.md).

**Revision note.** Two earlier drafts were rejected by three independent adversarial
reviews (Claude and GPT-5.5, neither the author). They were right, and § 7 records what
they killed — a future session will otherwise re-make those errors.

## Recheck against `main` @ `61cf985` (2026-07-14)

All load-bearing claims re-verified after itd-85 merged (#59–#67). **Every blocker holds:**
release gate still reads `PROMOTE` (`lint.go:677`); `agents/` still in no lint root;
`frontmatter` still top-level-keys-only; `checkSurfaceCoverage` still present;
`internal/core/schema` still absent; `intent-fidelity-reviewer` still declares no `tools:`;
itd-81 still strikes the length tiebreak.

**Two unmerged branches interact with this work — land them before Phase 6:**

- **`auto/iss-69-frontmatter`** — routes record-lint's frontmatter scan through
  `internal/core/frontmatter`. It does **not** add nested-key support, and it *reaffirms*
  the package as *"deliberately a line scanner, not a YAML parser"*, now naming
  "record-lint's top-level frontmatter checks" as a consumer. **This strengthens the defer
  decision**: adding nested YAML would now contradict the primitive's stated design intent,
  making it a real architectural call, not a convenience. It also means the PQ rules must be
  built on `internal/core/frontmatter` (`one-canonical-primitive`), not a private replica.
- **`auto/iss-67-intent-fidelity`** — hardens the **zero-criteria** path: the Plan AC gate
  now requires bullets, so an intent whose criteria the reviewer cannot enumerate can no
  longer be planned. **It does NOT close the ⚠ trap in Phase 1**, which is the *K*-criteria
  case. Confirmed still live: in `06-lint.md`, `RC003` (rollup `INCONCLUSIVE`) is severity
  **`info`**, gate is **REPORT-not-block**, `info → exit 0`. An all-`INCONCLUSIVE` fidelity
  review still exits **green**.

## Context

abcd ships five agents written ad hoc, with no normative template. The task was framed as
"the agents are stubs; rewrite them to a schema". That framing is wrong in an instructive
way: two of the five are good prompts. The defect is underneath them — **abcd already
specifies an agent schema, documents its lint as "delivered", and none of it exists.**

- `06-lint.md` claims `PQ001`–`PQ006` are "**delivered**". There is no `PromptLinter`, no
  `PQ*` rule, no `task_classes` enum in Go.
- `01-agents.md` requires `capability_scope` and `reads_untrusted_input`. **Zero of five**
  files carry either.
- The catalog declares 16 agents; disk holds 5; **one name is common to both.**

iss-11 was closed by editing prose until the documents agreed with *each other*, never
against `ls agents/`.

**But there is one genuinely live bug, and it is a security hole:**
`intent-fidelity-reviewer` declares **no `tools:`** while its body says *"You are
read-only: you never edit files."* Claude Code **inherits every tool when `tools` is
omitted**. Prose is not a permission system.

## The two constraints that shape this plan

### 1. Nothing here is "behaviour-neutral"

An earlier draft split the work into "behaviour-neutral frontmatter fixes (safe now)" and
"body rewrites (blocked)". **That axis is wrong.** The tool grant is the agent's *action
space*; `model:` decides *which model judges*; `description:` **is the routing rule**
(~79% activation reliability, research § 4.4). An agent dispatched less often, on a
different model, with fewer tools **is a behaviourally different agent with an identical
body.** "Bodies are byte-identical" controls the wrong variable.

The causally correct gate is narrower and statable:

| | Gated on calibration? |
|---|---|
| Changes how a judge **scores a well-formed input** — rubric, checklist, severity scale, verdict enum, evidence bar, model, tool grant | **Yes** |
| Injection defence, stale tool names, dead output formats, non-judge agents | **No** |

"No corpus ⇒ no body change" proves too much: it would freeze **security** fixes to judge
prompts indefinitely on a **quality** discipline.

### 2. `sota-researcher` is not a judge

itd-81 gates *judges*. `sota-researcher` has no verdict enum, no gate, no consumer. Its
body is **not** blocked. The previous draft blocked four agents and silently dropped the
fifth.

## Phase 0 — arm the detectors before the fix

Per the standing rule (*never fix ahead of an armed detector*). **Zero code:**
`Use PROACTIVELY` and persona openers (`You are a senior…`) become `banned_tokens` entries
in `.abcd/record-lint.json`, scoped to `agents/`. They then fail loudly until fixed, and
cannot silently return.

Evidence: `Use PROACTIVELY` is cargo cult — the current Claude Code docs contain no such
guidance (its only appearance is lowercase, in a JSON example); the documented rule is
trigger-condition specificity. Personas do not beat a no-persona control (arXiv 2311.10054)
and perturb safety boundaries (2509.08075).

## Phase 1 — the security fix (⚠ has a hard precondition)

Give `intent-fidelity-reviewer` an explicit read-only grant.

**PRECONDITION — verify first, or this manufactures a false green.**
`intent-fidelity-reviewer` is **the only agent whose output Go parses**
(`internal/core/intent/review.go`, `abcd/intent-fidelity-verdict/v1`). It currently
inherits `Bash`, so it may be **fetching the diff itself**. Take that away and:

> it cannot resolve the diff → its own rubric correctly says *"missing evidence ⇒
> `INCONCLUSIVE`"* → it emits `INCONCLUSIVE` for every criterion → the ingest **validates
> that** (the rollup sums to K) → `RC003` is severity **`info`** → **the gate reports
> green.** The fidelity audit silently degrades to a no-op.

That is a manufactured false green in the one place abcd has a real verdict — a
`loud-staging` violation. **So: confirm the host dispatch actually supplies `delivered`
(a diff / commit range) before narrowing the grant.** If it does not, the grant must
retain the means to fetch it.

**The read-only invariant is not "`tools:` is present."** `security-reviewer`,
`ruthless-reviewer` and `docs-currency-reviewer` all declare `tools: …, Bash` — and
**`Bash` writes.** A presence check would bless `Bash(rm -rf …)` as read-only. The
enforceable invariant: **no `Edit`/`Write`/`NotebookEdit`, and any `Bash` is
command-restricted** to what the agent actually runs (`Bash(git diff:*)`,
`Bash(go test:*)`, …). Tightening a grant is a **judge change** (§ 1) — it goes through
Phase 3's corpus, except on `intent-fidelity-reviewer`, whose hole is a security defect
fixed now and scored immediately after.

## Phase 2 — stop lying in the record

- **`06-lint.md` is fiction end to end on the PQ rules** — fix the whole section, not the
  one line that names them.
- `01-agents.md`: reconcile the catalog against `ls agents/`. The four non-catalog agents
  are `ruthless-reviewer`, `security-reviewer`, `docs-currency-reviewer`,
  `sota-researcher` (`intent-fidelity-reviewer` **is** already a row).
- Correct the stale command: `docs-currency-reviewer:14` says to run `docs-currency-lint`
  "(on PATH)". The tool **exists** — it is `abcd docs lint` (`Makefile:55`, the itd-60
  deterministic gate). **Rename it; do not delete the instruction** — deleting it strips a
  working deterministic pre-pass off the release path. (Not a judge-scoring change.)
- Record in `DECISIONS.md` that a prompt-length lint is **rejected**, citing itd-81 — this
  plan's first draft proposed one.

## Phase 3 — make itd-81 real: the calibration corpus (user decision)

**This is the gate on every judge change, and it does not exist today** (itd-81 is
`spec_id: null`; there is no `corpus/` anywhere).

Per judge (`ruthless-reviewer`, `security-reviewer`, `docs-currency-reviewer`,
`intent-fidelity-reviewer`): a small **labelled corpus** — known-defective inputs (a diff
with a real bug; a doc contradicting code) and **known-clean inputs**, which are the ones
that matter, because the failure mode is a judge that cries wolf.

- **Baseline the CURRENT prompts first.** Score recall and **true-negative rate** before
  any edit. A judge with high recall and an unmeasured TNR is an unmeasured judge.
- Judges **over-accept AI-written work by up to 1.91×** (research § 5), and every abcd
  reviewer judges AI-written code all day. The corpus must therefore contain
  **machine-written** samples, not just human ones.
- Verdicts stay in their own enums (§ 6). Scoring is per-judge.

## Phase 4 — rewrite the judges, measured

Only now. Each change is scored against Phase 3's baseline **before and after**; a drop in
TNR blocks it.

**Diff the wishlist against reality first.** A *required quoted evidence span per finding*
and a *self-refutation step* are **already in** `ruthless-reviewer` and `security-reviewer`.
Phase 4 is smaller than the first draft advertised. What it adds where missing: a
decomposed checklist (for reviewers only — TICK does **not** license checklists in every
agent, research § 6), an explicit `INCONCLUSIVE`, and *absent evidence is never an inferred
pass*.

## Phase 5 — rewrite `sota-researcher` (unblocked)

Not a judge; no corpus needed. Highest-value unblocked body edit. It holds `WebFetch`
(attacker-controlled pages) with **no injection defence** and no `reads_untrusted_input`.
Add both — copy the injection-resistance section that `intent-fidelity-reviewer` already
has. Declaring `reads_untrusted_input: true` and adding injection defence applies to the
four reviewers too (they read diffs and commit messages): **injection defence is not a
judge-scoring change** and is therefore not gated.

## Phase 6 — the lint, with an execution path

**`agents/` is in no lint root.** `Lint()` walks `cfg.Roots`, and `record-lint.json` has
`"roots": [".abcd/development"]`. Rules that live outside the roots (`stray_root_docs`,
`surface_coverage`, `receipt_gate`, …) are each **explicitly hoisted out** of the per-root
loop with repo-root-scoped readers. **The PQ rules need the same treatment.** Without it:

> the checks never fire, `make preflight` is green, and *"preflight clean"* is satisfied by
> **a lint that never ran** — the plan could not distinguish "clean" from "did nothing".

Naively adding `"agents"` to `roots` instead fires `directory_coverage` on `agents/`
immediately. Hoist; do not widen the roots.

**Rules that can be built on the parser that exists** (`frontmatter.Fields` reads
**top-level keys at column 0 only**):

| Rule | Check |
|---|---|
| `PQ001` | `prompt_version` present, valid semver |
| `PQ002` | `agents/CHANGELOG.md` entry for the current version — **the file must be created; it does not exist** |
| `PQ007` | **write-capability** — read-only agents hold no `Edit`/`Write`/`NotebookEdit`; any `Bash` is command-restricted. *The only rule that closes a real security hole.* |
| `PQ008` | `hooks`/`mcpServers`/`permissionMode` absent — Claude Code **ignores these for plugin subagents for security reasons**, and abcd *is* a distributed plugin |

**Roster drift — extend the canonical primitive.** `checkSurfaceCoverage` (`lint.go:374`)
**already is** a status-aware registry↔reality bijection (`shipped` row ⇒ artefact exists;
`staged` row ⇒ it does not; artefact with no row ⇒ finding). Point it at `agents/*.md` ↔
`01-agents.md`'s **Status** column. A bespoke `PQ009` would be a weaker copy of a shipped,
tested rule — a **`one-canonical-primitive` violation** — and, being one-directional, could
not even detect iss-11's own incident (delete an agent file; every remaining file still has
a row; lint stays green).

## Deferred, explicitly (not silently)

- **`PQ003`/`PQ004`/`PQ005` (`capability_scope`, `task_classes`)** — *user decision: defer.*
  Two hard blockers: `capability_scope` is **nested** YAML and the parser sees top-level
  keys only, so `task_classes` is invisible; and the closed enum's declared home,
  **`internal/core/schema`, does not exist**. Correct the record to stop claiming
  enforcement. A hand-rolled YAML *subset* would also make `PQ008` unsound — our parser and
  Claude Code's real one would disagree about what a file says, and `PQ008` is a security
  rule. A YAML dependency needs **explicit sign-off before `go get`**.
- **`PQ006` canary-fixture presence — cut.** It is `len(ReadDir(dir)) > 0`: an empty file
  passes. It proves a byte exists on disk. There is no canary executor, no phase builds
  one, and `06-lint.md` already defers execution. Demanding canaries "fail, never skip"
  with no executor would commit the exact sin this plan indicts. Injection resistance is
  not testable offline — the defence moves to the capability grant (`PQ007`), which *is*
  testable.
- **Verdict-enum unification — dropped.** `lint.go:677` reads
  `if r.VerificationResult != "PROMOTE"` — the release gate keys on the **value**, and
  `PROMOTE`/`HOLD` is a **third** family. The unification is lossy:
  `security-reviewer`'s `NEEDS-INPUT` ("cannot decide") has no image in the target enum;
  `BLOCK` is load-bearing outside the file (`CLAUDE.md`: *"a BLOCK verdict stops the
  change"*); collapsing `STALE` and `INCOMPLETE` destroys the distinction the release gate
  exists to know. **Leave the enums alone.**
- **JSON output contract** — mandated **only where a consumer exists**: today,
  `intent-fidelity-reviewer` alone. Nothing in `internal/` parses a verdict from the other
  four; mandating JSON for them is tokens emitted for no reader — dead scaffolding, which
  `AGENTS.md` forbids. The *principle* (reason in prose, then emit — arXiv 2606.09410:
  Opus 4.7 drops 96.2% → 91.0% on AIME answering *in* JSON) applies when a consumer lands.
- The 15 unshipped roster agents. itd-2 dispatch.

## Verification

- **The lint actually ran.** A deliberately-broken agent fixture must FAIL. "preflight
  green" is not evidence until a negative control has been seen to fail.
- **Negative controls**: `tools` omitted; unrestricted `Bash` on a read-only agent;
  `hooks:` present; an agent file with no catalog row.
- **Judge regression**: recall and TNR per judge, before and after every Phase-4 edit. A
  TNR drop blocks the change. *This is the whole point of Phase 3.*
- **F5 guard**: after Phase 1, `intent-fidelity-reviewer` must still resolve a real diff and
  emit non-`INCONCLUSIVE` verdicts on a known-good intent. An all-`INCONCLUSIVE` run is a
  **failure**, not a pass, however green the gate reports.
- `make preflight` clean, offline. Release gate still accepts a `PROMOTE`
  `docs-currency-reviewer` receipt.

## STOP conditions

- Phase 1 precondition unmet (host does not supply the diff) → **stop.** Do not narrow the
  grant into a rubber stamp.
- Any judge edit without a corpus baseline → **stop.** That is Phase 4 and it needs Phase 3.
- Nested-YAML parsing needs a new dependency → **stop and get sign-off.** Do not `go get`.

## § 7 — What the rejected drafts got wrong (kept deliberately)

All caught by adversarial review, none by the author.

1. **A prompt-length lint (`PQ010`)** — reinstating a rule **itd-81 struck on 2026-07-12**,
   citing Agentic Context Engineering (arXiv 2510.04618) on **brevity bias**. The draft
   never cited itd-81 because its research never read `intents/disciplines/`. *An external
   SOTA finding does not override a local decision that already weighed it.*
2. **"The release gate ignores verdict values."** It does not (`lint.go:677`). That false
   claim was used to wave a change onto the release path.
3. **"`docs-currency-lint` does not exist."** It does — as `abcd docs lint`. The prescribed
   fix would have deleted a working gate.
4. **Miscounted the roster** (11 unshipped; actually 15) — in a plan whose thesis is that
   iss-11 was closed without running the intersection. The draft did not run it either.
5. **Reinvented `checkSurfaceCoverage`** as a weaker `PQ009`, without citing the shipped
   rule in the same package.
6. **Specified `PQ003`–`PQ005` against a parser that cannot see nested keys**, and against
   `internal/core/schema` — **a package that does not exist**: inheriting the record's
   declared-not-delivered claim and repeating it, in the plan indicting that sin.
7. **Kept `PQ006` presence-checking as "enforcement"** while demanding canaries "fail,
   never skip" with **no executor in any phase**.
8. **Laundered MAST and TICK** — citing papers that establish *problems* as though they
   endorsed *remedies*. MAST tested no intervention; its denominator is multi-agent systems,
   which the draft's own research argues abcd is not.
9. **Called Phase A "behaviour-neutral"** while changing tool grants, `model:` and
   `description:` — and proposed "byte-identical bodies" as the regression control, which
   measures the wrong variable.
10. **Blocked four judges and silently dropped the fifth** (`sota-researcher` is not a judge
    and was never gated).
