# Run plan — abcd-cli ledger drain #2 (follow-up fixables)

**Status:** the second C-phase `/abcd:run` drain against abcd-cli's ledger, scoped
to the two self-contained fixables surfaced by drain #1
([`2026-07-12-abcd-cli-ledger-drain-run.md`](2026-07-12-abcd-cli-ledger-drain-run.md)).
This is a deliberately small, safe batch — the hand-authored precursor to
`abcd drain`'s triage ([itd-82](../intents/drafts/itd-82-drain-ledger-triage.md)).

Invoke (30m work / 30m pause, dynamic ScheduleWakeup 1800s):

```
/loop 30m Follow the protocol in .abcd/development/plans/2026-07-12-abcd-run-protocol.md
          for the run plan at .abcd/development/plans/2026-07-12-abcd-cli-drain-2.md.
```

## Run contract

```
backlog: ledger                       # abcd capture list --open --json; ready = blocked_by_open empty
gate: make preflight                  # build + vet + test + race + record-lint + docs-lint + lint-reviews
definition_of_done: per issue — detector-first (fix-the-detector). Arm the issue's
  named detector against its stated acceptance corpus, watch it flag, then drain the
  fix behind it. A green gate is the floor; the issue's own detector is the bar.
  Fold `abcd capture resolve <iss-N>` into the fix branch.
ledger_verb: abcd capture
reviewers:
  correctness: ruthless-reviewer      # every non-trivial diff
  security: security-reviewer          # trust-boundary diffs (default-on; both items qualify)
branch_policy: fix/<iss-N>-<slug> from main; merge-commit-only; never force-push;
  never commit to main; delete branch after merge (via gh api, not a gate-tripping push)
commit_trailer: Assisted-by: Claude:claude-opus-4-8
pr_policy: one PR per issue; Closes #<gh-issue> only if the ledger issue maps to one;
  do not merge; do not enable auto-merge
budget: 30m wall | 6 worker-agents/burst | write NEXT.md + run-journal.json continuously
strike_limit: 3                        # failed + died entries combined, keyed on iss-N
scope_fence: ONLY the two issues in the Ready backlog below. This is a curated batch,
  NOT an open-ledger sweep — do NOT auto-continue into other open issues (the older
  cluster iss-33/34/37-49 is un-triaged; the new follow-ups iss-76/79 are the vetted
  set). When both are done (or STOP), end the loop and report.
stop_conditions:
  - An issue is a maintainer DESIGN-STOP / adjudication, not a self-contained fix — SKIP.
  - A fix would need a new dependency, a bespoke no-seam solution, a DB migration, or
    a CI-config change not already in scope.
  - A fix turns out to be a feature/contract change rather than the stated bug.
  - Both Ready-backlog issues are resolved (or hit a STOP) — the curated batch is drained.
irreversible:
  - none expected; if a fix touches persisted-state format or a schema_version, treat
    as irreversible (human checkpoint + rollback), even though not named here.
```

## Ready backlog (the ONLY two issues in scope)

Both self-contained, detector-first, trust-boundary (security reviewer applies):

- **iss-76** `json-error-abspath-leak` (minor, bug) — `cli.Run` routes every command
  error through the `--json` envelope, so any verb returning a bare `*fs.PathError`
  emits an absolute local path into machine JSON (the systemic version of the leak
  drain #1 fixed for docs-lint). Detector: a table test running each verb's known
  filesystem-error path under `--json`, asserting the envelope carries no absolute
  path. Fix: sanitise PathError/LinkError-bearing errors at the `Run()` boundary (or
  audit which errors reach the envelope). Composes with the iss-29 fix already on main.
- **iss-79** `storeoriginal-inline-atomic-write` (minor, tech-debt) — `memory/ingest.go`
  `storeOriginal` (~:806) is a fifth inline temp+`O_EXCL`+rename durable write left
  untouched by the iss-32 consolidation; it lacks parent-dir fsync + explicit chmod.
  Detector: extend `TestNoNonCanonicalAtomicWritePrimitives` to catch inline
  temp+rename sequences (or a targeted test), then route storeOriginal through
  `fsutil.WriteFileAtomic(target, material.rawBytes, 0o644)`, keeping the existing
  sources-dir symlink guard. Watch the security-relevant symlink guard is preserved.

## Explicit skips (do NOT work these)

- **iss-28** `hermetic-git-test-env` — SKIP (assessed drain #1: cross-repo scaffolding
  feature, design-STOP; awaits maintainer design / promotion to intent).
- **iss-35** `brief-surface-reconciliation` — SKIP (maintainer design-STOP).
- **iss-77** `launch-payload-omits-agents-hooks` — SKIP (design-shaped: the public
  includes set is a maintainer call).
- **iss-78** `launch-dryrun-needs-version-location` — SKIP (ops/setup, not a code fix).
- **The older review cluster (iss-33/34/37-49)** — SKIP: un-triaged this batch. They
  need a triage pass (the `abcd drain` classifier, itd-82) before they are drain-safe.

## Why this is a safe batch

Two minor, self-contained defects, each with a pre-stated detector and acceptance
corpus, both composing cleanly with fixes already merged (iss-29, iss-32). No
irreversible actions. The scope fence + explicit skips keep the loop from wandering
into un-triaged or design-shaped work. It also exercises two detector shapes drain #1
did not: a per-verb `--json` error-shape table test (iss-76) and extending the
canonical-primitive detector to inline sequences (iss-79).
