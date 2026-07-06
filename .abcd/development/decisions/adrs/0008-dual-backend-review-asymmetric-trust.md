---
id: adr-8
slug: dual-backend-review-asymmetric-trust
status: superseded
date: 2026-05-16
supersedes: null
superseded_by: adr-25
related_intents: [itd-17, itd-28]
related_rfcs: []
related_adrs: [adr-6]
---

# ADR-8: Dual-Backend Review with Asymmetric Trust

> Superseded by [ADR-25](0025-host-delegated-llm-default.md) — the fixed
> RP→codex cascade is replaced by a host-delegated LLM with opt-in oracle
> adapters; the scoped/broad asymmetric-trust principle survives as adapter
> guidance.

> **Terminology note.** The *how* layer is named the **spec**. This ADR's prose
> was updated by the spec-terminology-rename ADR
> ([adr-11](0011-spec-terminology-rename.md)).

## Context

abcd's review surfaces (`/flow-next:plan-review`, `/flow-next:impl-review`,
`/flow-next:completion-review`, and the RP-review capture in `itd-28` / `adr-6`) can route a
review to one of two backends: **RepoPrompt** (a chat in a workspace-bound RP window,
seeing only the explicitly-selected files) and **Codex CLI** (`gpt-5.2` at high
reasoning effort, running read-only *inside the repo* with shell access).

The `fn-5` spec plan review was run dual-backend across twelve rounds (16–27). That
exercise produced enough evidence about the two backends' behaviour to settle three
questions that were previously informal:

1. **One backend or both** for a high-stakes review?
2. **Which backend's verdict tag is authoritative** when they disagree or when a
   `SHIP` tag co-exists with unresolved findings?
3. **Does the review loop have a defined stopping rule**, or does it run until a
   reviewer stops finding things?

Each is hard to reverse once the review skills and downstream automation depend on it,
each is surprising without the `fn-5` evidence, and each is a real trade-off — so each
earns a decision here.

## Decisions

### Decision 1 — High-stakes reviews run BOTH backends, in parallel

A high-stakes review (spec plan review, completion-review, or any review whose verdict gates
other specs) runs RepoPrompt **and** Codex CLI, concurrently. Low-stakes reviews may
use a single configured backend.

The two backends have **asymmetric, complementary blind spots**, demonstrated repeatedly
in the `fn-5` loop:

- **Codex CLI scans the actual repository.** Its two `fn-5` Critical findings — the
  `.4` stale-string sweep matching tracked `skills/abcd-intent-grill/**` content (which
  would have broken `pytest`), and a startup-gating concurrency race — both came from
  reasoning over real tracked files. A reviewer seeing only selected spec text *cannot*
  find these. Codex also reliably caught cross-file / cross-document drift (spec vs
  `.python-version`, spec vs `docs/CONTRIBUTING.md`, `.2` vs `.3` inconsistency).
- **RepoPrompt sees only the selected files**, so it never raises a repo-wide issue —
  but within its window it is the stronger *consistency / regression* reviewer. It
  caught, precisely and correctly, contradictions introduced by the loop's own prior
  fixes (a stale queue-tuple shape; two thread-safety regressions in a finalisation
  routine).

Running only one backend would have shipped a plan with at least one `pytest`-breaking
Critical undetected.

**Rejected alternative:** single-backend review with backend chosen per task. Rejected
because the blind spots are structural (one scans the repo, one scans a selection), not
quality differences — no single backend covers both axes.

**Three-clause test:**
- Hard to reverse? **Yes** — once review skills and CI gate on a dual-backend receipt
  shape, dropping to single-backend means re-auditing what the missing backend would
  have caught.
- Surprising? **Yes** — most pipelines use one reviewer; paying for two parallel LLM
  reviews per round is counter-intuitive.
- Real trade-off? **Yes** — dual-backend trades cost and wall-clock (two reviews per
  round, Codex 8–13 min each) for coverage of two genuinely different failure classes.

### Decision 2 — The scoped reviewer's verdict gates; the broad reviewer is mined for findings

When the two backends are used together, **RepoPrompt's verdict tag is the gate** and
**Codex's verdict tag is advisory** — its *findings body* is authoritative, its tag is
not.

In the `fn-5` loop Codex returned `<verdict>SHIP</verdict>` eight times while its own
body still listed Major (once Critical) findings. RepoPrompt's verdict, by contrast, was
honest every round: `NEEDS_WORK` meant real issues, `SHIP` meant a clean body. Codex
also hallucinated a citation once and flip-flopped on an SDK detail across rounds —
its *reasoning over the repo* is its value, its *self-assessment* is not.

The operating rule: a review is "passed" when the **scoped backend (RepoPrompt) returns
`SHIP` with a clean body**. Codex's verdict is ignored; Codex's findings are triaged —
Critical/Major fixed, Minor/Nitpick recorded as impl-time notes.

**Rejected alternative:** require both verdict tags to read `SHIP`. Rejected because
Codex's tag does not reliably reflect its body — gating on it would either block
indefinitely (it keeps finding things) or pass on a tag that contradicts the findings
beneath it.

**Three-clause test:**
- Hard to reverse? **Yes** — automation that parses verdict tags would need rewriting.
- Surprising? **Yes** — "trust one reviewer's verdict, ignore the other's verdict but
  read its findings" is a non-obvious split.
- Real trade-off? **Yes** — asymmetric trust trades a simple "all green" rule for a
  rule that matches each backend's observed reliability.

### Decision 3 — The review-fix loop has a mandatory stopping rule

A review-fix loop MUST declare a stopping rule before it starts. The default:

> Loop until the **gating backend returns `SHIP` with a clean body**. Then run **one
> confirming round**. After the confirming round, ship regardless of residual
> Minor/Nitpick findings — record them as tracked implementation-time issues
> (`.work/issues.md`). Critical/Major findings always reset the loop.

The `fn-5` loop ran twelve rounds because Codex produced a fresh tier-2 finding nearly
every round — frequently a *follow-on from the previous round's fix*. A high-effort
reviewer with repo access will almost always find *something*; without a human-set
cutoff the loop does not self-terminate.

**Rejected alternative:** loop until a backend returns zero findings of any severity.
Rejected because the `fn-5` evidence shows that state may never arrive — each fix can
surface an adjacent consideration. A stopping rule keyed on *severity* and a *confirming
round* terminates while still guaranteeing no Critical/Major ships.

**Three-clause test:**
- Hard to reverse? **Yes** — once specs ship under this rule, changing it changes what
  "reviewed" means for everything already shipped.
- Surprising? **Yes** — an explicit "stop even though the reviewer still has comments"
  rule is counter-intuitive for a quality process.
- Real trade-off? **Yes** — a severity-keyed cutoff trades theoretical completeness
  (every Nitpick fixed) for termination (the loop provably ends).

## Consequences

- The review skills should treat a dual-backend run as the default for spec-level
  reviews and surface both the gating verdict and the broad backend's findings list.
- Residual Minor/Nitpick findings are not lost — they are written to `.work/issues.md`
  as impl-time notes (see the `fn-5` round-27 residual entry as the worked example).
- This ADR records *human-observed* backend behaviour. The *systematic, ongoing*
  capture of per-backend strengths and failure modes is the domain of **itd-17**
  (frontier mapping) — itd-17 adds "review backend" as a tracked `task_class`. This
  ADR is the qualitative decision; itd-17 is the quantitative substrate.
- Operational how-to (curating RP's file selection, same-chat re-review, Codex
  `--receipt` continuity, opening a per-repo RP window) lives in agent memory, not
  here — it is procedure, not architecture.

## Related

- [adr-6](0006-rp-review-storage-and-architecture.md) — where review artifacts land
- [itd-28](../../intents/shipped/itd-28-rp-reviews-into-flow.md) — RP review capture into Flow
- [itd-17](../../intents/drafts/itd-17-model-effectiveness-tracking.md) — frontier mapping; absorbs review-backend effectiveness as tracked data
- `.flow/reviews/fn-5-rp-mcp-integration-declare/` — the twelve-round dual-backend loop this ADR generalises from
