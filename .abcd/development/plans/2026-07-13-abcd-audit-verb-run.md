# Run plan — `abcd audit` P1+P2 (itd-85 / iss-86)

**Status:** the run wrapper for the agreed design plan
[`2026-07-13-abcd-audit-verb.md`](2026-07-13-abcd-audit-verb.md). That plan holds the
design, the SOTA declaration, and the four locked decisions; this file holds only the
**run contract** the `/abcd:run` protocol needs. Backlog is `milestones` (this is one
intent decomposed, not a ledger sweep).

Invoke:

```
/loop Follow the protocol in .abcd/development/plans/2026-07-12-abcd-run-protocol.md
      for the run plan at .abcd/development/plans/2026-07-13-abcd-audit-verb-run.md.
```

## Run contract

```
backlog: milestones                   # M1..M7 below; ready = all depends_on merged-or-pushed
gate: make preflight                  # lint-reviews + record-lint + docs-lint + build + vet + test + race
definition_of_done: per milestone — TDD. One test watched FAIL before the change and
  PASS after, for every new behaviour; each of the five v1 rules gets its own. A green
  gate is the floor. The intent's Acceptance Criteria (itd-85) are the bar: the run is
  done when all five AC bullets are demonstrably met by a test.
ledger_verb: abcd capture             # fold `abcd capture resolve iss-86` into the M6 branch
reviewers:
  correctness: ruthless-reviewer      # every non-trivial diff
  security: security-reviewer          # M1 (git subprocess + path handling) and M3
                                       # (privacy-hygiene rule reads committed files) are
                                       # trust-boundary; default-on when unsure
branch_policy: feat/itd-85-<slug> from main; chained (M(n+1) branches from M(n) when
  dependent); merge-commit-only; never force-push; never commit to main
commit_trailer: Assisted-by: Claude:claude-opus-4-8
pr_policy: one PR per milestone; push and open the PR; DO NOT wait for merge and DO NOT
  merge; state the chain in the body. Continue straight to the next milestone.
  `Closes` only on the M6 PR (the wiring is what closes iss-86).
budget: dynamic pacing | 6 worker-agents/burst | write NEXT.md + run-journal.json continuously
strike_limit: 3                       # failed + died entries combined, keyed on M<n>
learnings: MANDATORY — record what did NOT work, not just what shipped. Every rejected
  approach, dead end, wrong assumption, and surprise gets written down at the moment it
  is learned, never reconstructed at burst end:
  - `run-journal.json` — the `note` field on any `failed` entry states the ROOT CAUSE and
    what was ruled out, so a later burst never retries it. Never repeat a recorded-failed
    approach.
  - `.abcd/work/DECISIONS.md` — one dated line for any decision a future session would
    otherwise re-litigate, INCLUDING rejected alternatives ("chose X over Y because Y
    turned out to Z"). A negative result is a decision.
  - `abcd capture` — any abcd defect or friction hit while dogfooding, on sight, every
    time (a workaround is fine; a silent one is not).
  - Divergences from the design plan get written back into it as a revised mechanic
    (never a revised premise — a false premise is a STOP).
scope_fence: ONLY M1..M7 below = P1 (engine + five rules + CLI) and P2 (plugin surface +
  prepare-this-repo wiring). P3 (SARIF `--format sarif`) is OUT — the serializer seam is
  built, the SARIF serializer is not. Do NOT fold in iss-88 (managed-repo check stays an
  ahoy/detection concern, per locked decision 2). Do NOT open repo-level rule
  override/config (explicitly deferred in itd-85). When M7 is pushed, end the loop.
stop_conditions:
  - A new dependency is needed (any new go.mod entry, incl. test) — hard stop, zero-dep
    is the load-bearing constraint of the whole SOTA verdict.
  - The rule-loader seam or the output-serializer seam cannot be built (bespoke-no-seam
    trips the sota-per-intent hard stop).
  - The design plan's PREMISE turns out false (revise mechanics freely; never the premise).
  - A milestone needs a CI-config change or touches persisted-state format / schema_version.
  - Strike limit (3) on one milestone.
  - Chain fragility (base PR closed-not-merged, unpushed base, squash-merge re-enabled).
irreversible:
  - none expected. `audit` is strictly read-only — it MUST NOT write, mutate, or fix
    anything. If a milestone would introduce a write path, that is a scope error, not a
    feature.
```

## Milestones

| id | depends_on | scope |
|---|---|---|
| **M1** | — | **Path + gitignore primitives.** Promote a presence/absence helper into `internal/fsutil` (three private `exists` copies today: `core/core.go:73`, `core/ahoy/fsutil.go:12`, `core/lint/lint.go:892` — consolidate, don't add a fourth). Add dir-has-≥1-entry. Export a gitignore predicate by promoting `launch.checkIgnored` (`internal/core/launch/bundle.go:771`) — it already hardens the git shellout (`GIT_CONFIG_GLOBAL=/dev/null`, negation-pattern handling, fails open when git is absent). Behaviour-preserving promotion + its own tests; callers switched in a separate commit from any fix. |
| **M2** | M1 | **Rule schema + loader + evaluator seam** in a new `internal/core/audit`. The existing `lint` engine has **no registry** — rules are hand-coded branches in a dispatcher (`lint.go:139-212`) and severity is `blocker`/`warn` strings. So: define the declarative rule object (`id`, `severity`, `where`, `fix`, `policyInfo`), a loader over bundled in-binary defaults, and a `Rule` interface the evaluator ranges over (the rule-loader seam). Result envelope `AuditResult{Findings, Blockers, ExitCode}` following `memory.LintResult` (`core/memory/lint.go:443`). Output behind a serializer seam so P3 SARIF is a thin add-on. **Reconcile the severity vocabulary** (`error|warn|off` in the plan vs `blocker|warn` in `lint`) and record the call in DECISIONS.md. |
| **M3** | M2 | **The five v1 rules**, each with a watched-fail→pass test: `three-tier-layout` (error), `conventions-router` (error), `decision-durability` (warn), `docs-currency` (warn — reuse the existing docs-lint primitives, do not reimplement), `privacy-hygiene` (error — honour the existing line-scoped `allow_context` waiver, `lint.go:931`). `where`-conditional enablement proven by the "docs/ absent ⇒ docs-currency skipped, not failed" AC. |
| **M4** | M3 | **CLI verb.** `newAuditCommand(&asJSON)` in `internal/surface/cli/cli.go`, copying `newDocsCommand`'s tri-state contract (`cli.go:157-221`): 0 clean / 1 warnings-only / 2 any error. Human doctor-style render (grouped, severity glyph, inline fix), diagnostics to stderr so `--json` stays clean. Read-only — no write path. |
| **M5** | M4 | **Plugin surface + registry.** `commands/abcd/audit.md` (follow `ahoy.md`'s shape; state "performs zero writes") **plus** a `16-audit.md` detail file **plus** a `shipped` row in `.abcd/development/brief/04-surfaces/README.md` — the `surface_coverage` record-lint rule (`lint.go:373`) fails preflight if the row and the file disagree, so these land together or not at all. |
| **M6** | M5 | **Wire onboarding (this is what closes iss-86).** `prepare-this-repo` Phase 2 consumes `abcd audit --json` and renders it, replacing the hand-produced gap report. Wired-or-it-isn't-done: the engine must actually be what onboarding calls. Fold `abcd capture resolve iss-86` into this branch. |
| **M7** | M6 | **Record.** `ACKNOWLEDGEMENTS.md` entries for the three adapted patterns (repolinter rule schema, Conftest severity/exit semantics, SARIF-as-future-export) — same change that lands them, never retroactively. CHANGELOG entry (user-facing verb). Any user-facing doc. |

## Known sharp edges (from recon — do not rediscover)

- `internal/core/lint` has **no rule registry**; adding a rule there today means editing a
  dispatcher. That is precisely why M2 builds the seam rather than adding a sixth branch.
- Severity vocabulary diverges between the plan (`error|warn|off`) and the engine
  (`blocker|warn`). Decide once, in M2, and record it.
- `launch.checkIgnored` is the **only** `git check-ignore` implementation in the tree, and
  it is unexported and hardened. Promote it; do not write a second one.
- `make smoke` self-discovers the Cobra tree and exercises every read-only verb — a new
  `audit` verb is auto-covered there for free (it is not in `preflight`; run it anyway).
- `abcd intent plan itd-85` is the **maintainer adoption gate** and has not run (itd-85 is
  still in `intents/drafts/`). It is the maintainer's to run — flag it in the M7 handoff
  as the pre-ship gate; do not cross it autonomously.
