# Run plan — abcd-cli open-ledger drain (first `/abcd:run` dogfood)

**Status:** the first C-phase dogfood run of the `/abcd:run` protocol
([`2026-07-12-abcd-run-protocol.md`](2026-07-12-abcd-run-protocol.md)) against
abcd-cli's own open issue ledger. Purpose: shake out the loop mechanics + handoff on
real, self-contained fixes, and generate itd-29's real-evidence (revisit trigger #5).

Invoke:

```
/loop 90m Follow the protocol in .abcd/development/plans/2026-07-12-abcd-run-protocol.md
          for the run plan at .abcd/development/plans/2026-07-12-abcd-cli-ledger-drain-run.md.
```

## Run contract

```
backlog: ledger                       # abcd capture list --open --json; ready = blocked_by_open empty
gate: make preflight                  # build + vet + test + race + record-lint + docs-lint + lint-reviews
definition_of_done: per issue — detector-first (fix-the-detector). Arm the issue's
  named detector against its stated acceptance corpus, watch it flag, then drain the
  fix behind it. The issue's own "Detector"/"Acceptance corpus" lines ARE the bar; a
  green gate is only the floor. Fold `abcd capture resolve <iss-N>` into the fix branch.
ledger_verb: abcd capture
reviewers:
  correctness: ruthless-reviewer      # every non-trivial diff
  security: security-reviewer          # trust-boundary diffs (default-on; most of this backlog qualifies)
branch_policy: fix/<iss-N>-<slug> from main; merge-commit-only; never force-push;
  never commit to main; delete branch after merge (via gh api, not a gate-tripping push)
commit_trailer: Assisted-by: Claude:claude-opus-4-8
pr_policy: one PR per issue; Closes #<gh-issue> only if the ledger issue maps to one;
  do not merge; do not enable auto-merge
budget: 30m wall | 6 worker-agents/burst | write NEXT.md + run-journal.json continuously
strike_limit: 3                        # failed + died entries combined, keyed on iss-N
stop_conditions:
  - An issue is a maintainer DESIGN-STOP / adjudication, not a self-contained fix —
    SKIP it (do not autonomously decide it). iss-35 is exactly this: its body records
    a design-STOP held for maintainer sign-off (record-lint-graduation adjudication
    items 5 & 6). Never "reconcile" or "fix" iss-35 in this run.
  - A fix would need a new dependency, a bespoke no-seam solution, a DB migration, or
    a CI-config change not already in scope.
  - A fix turns out to be a feature/contract change rather than the stated bug.
irreversible:
  - none expected in this backlog; if a fix touches persisted-state format or the
    schema_version of ledger/spec/memory records, treat as irreversible (human
    checkpoint + rollback), even though named here.
```

## Ready backlog (snapshot 2026-07-12 — the loop re-queries each burst)

Self-contained, detector-first fixes, all trust-boundary (security reviewer applies):

- **iss-29** `fail-closed-capture-surface` (major, bug) — misspelled subcommand writes
  a ledger entry; `--json` errors emit raw Go text. Detector: malformed-input +
  did-you-mean + `--json` error-shape tests (`unrecognized-input-never-writes`).
- **iss-30** `memory-ingest-boundary` (major, bug) — **partially resolved** (PR #38);
  remainder: `--keep-original` partial-failure reporting, CRLF parser-parity, URL /
  content-type / PDF coverage. Land the clear subset, keep open with remainder recorded.
- **iss-31** `launch-dogfood-gate` (major, bug) — identity scanner false-positive on
  `/dev/null`; launch payload omits `skills/`; unsynchronised `globRegexpCache` (race).
  Detector: a `launch --dry-run` dogfood gate + `-race` on the resolver.
- **iss-32** `atomic-write-consolidation` (major, tech-debt) — four divergent
  `writeFileAtomic`; `internal/fsutil` untested. Detector: a lint flagging non-canonical
  redefinitions + an fsutil crash-safety suite; consolidate behind it
  (`one-canonical-primitive`).
- **iss-28** `hermetic-git-test-env` (major, future-work-seed) — larger; a hermetic-git
  test helper. Assess scope on pickup; if it is a cross-repo scaffolding feature rather
  than a self-contained fix, it may hit the feature/contract STOP — report, don't force.

## Explicit skips

- **iss-35** `brief-surface-reconciliation` (critical) — **SKIP.** It is open *only*
  for a maintainer design-STOP (record-lint-graduation options + adjudication items 5
  & 6). Not autonomously actionable.

## Why this is a safe first run

Every ready item is a self-contained defect with a pre-stated detector + acceptance
corpus, on a repo with a strong deterministic gate (`make preflight`) and merge-commit
history. No irreversible actions expected. The risky/ambiguous items (iss-28, iss-35)
are pre-marked assess-or-skip, so the loop cannot wander into a design decision. This
is the evidence itd-29 asked for before its binary operator surface is designed.
