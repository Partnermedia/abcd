# 2026-08-16 — Rehearsal: the CI inert-path skip gate

This note is the probe for the skip gate that PR #244 (iss-234) lands: it
is the only file its pull request changes, and it sits inside the inert
allowlist (`.abcd/development/**`), so the pull request that carries it is
a genuinely inert change classified by the very gate under test. The
rehearsal is self-demonstrating: if this note is on `main` via its own
auto-merged pull request, the gate works.

## The claim under test

The `changes` classifier job reports `inert=true`, the gated lanes stand
down, and a job skipped by a job-level `if:` reports a `skipped`
conclusion that the ruleset accepts as satisfying its required status
checks — so an armed auto-merge completes rather than wedging behind a
never-reporting context. This behaviour is documented by the host but
deliberately unverified here until this rehearsal (loud staging: no
assumed green).

## Pass criteria

- The `changes` job logs `verdict: inert=true` for this pull request.
- `smoke` and `zizmor` conclude `skipped`; `check (macos-latest)`
  concludes `success` with its heavy steps skipped (the leg is
  step-gated, not dropped from the matrix).
- `check (ubuntu-latest)` runs the unit lane and both record gates;
  `gitleaks`, `record-lint`, and `attribution` run as always.
- The ruleset reports every required check satisfied and the armed
  auto-merge completes without manual intervention.

## On failure

Any required check stuck pending, or auto-merge wedged, executes the
rehearsed rollback: a revert of PR #244's merge on a `ci:` branch and
pull request, with iss-234 left open and the findings appended to its
capture. Until this rehearsal passes, nothing relies on the skip path —
iss-234 stays open in the ledger.
