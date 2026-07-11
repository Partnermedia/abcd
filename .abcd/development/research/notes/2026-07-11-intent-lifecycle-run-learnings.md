# Intent-lifecycle build — run learnings

Observations captured during the autonomous itd-80 build (intent-lifecycle
automation, slice 1). A running log; graduates into intents/ADRs where a learning
is load-bearing.

## Workflow gaps (candidate capabilities)

- **Auto-pick-up of manually-merged PRs (maintainer-flagged).** The autonomous
  loop opens a PR per wired section and does not self-merge `feat`. Today, when the
  maintainer merges one manually, the loop does not react: it should detect the
  merge, prune the merged branch (local + confirm remote auto-delete), pull the new
  `main`, **rebase the stacked follow-on branch** onto it, and **re-run the gates**
  — then continue. Currently each of these is a manual step the main thread does by
  hand (`fetch --prune`, `branch -D` of the squash-merged branch, `rebase
  origin/main`). Natural home: [itd-29 autonomous-run-resilience]; worth a dedicated
  `abcd` verb (e.g. an `abcd run reconcile-branches` / a post-merge hook) so a long
  autonomous run stays in lockstep with the maintainer's async merges without
  human-in-the-loop git bookkeeping. Includes: detect when a stacked PR's base has
  merged and re-target/rebase it; re-run `make preflight` after every rebase.

- **Duplication pressure under subagent implementation.** Independent
  implementation subagents repeatedly re-copied primitives (the frontmatter
  line-scanner, atomic-write, the trust-guarded rename helper). `one-canonical-primitive`
  had to be enforced explicitly in each phase brief, and one copy (spec's
  atomic-write) slipped through until the review pass. Signals value in a
  **duplication-detector** (a lint code, or a review-checklist item) so the rule is
  mechanical rather than a per-brief reminder — the same principle-→-discipline
  promotion path the record already uses.

## What is working

- **Deterministic gates catch real drift early.** `record-lint` blocked a
  freshly-authored intent whose Prior-Art links pointed at `drafts/` siblings that
  actually live in `planned/` — caught before commit. Directly validates the
  gate-over-prose approach, and is itself a small instance of the link-drift the
  reconcile automates.

- **Phased TDD + a review gate per phase keeps the main thread lean.** Each phase is
  a background subagent (design → implement → test) whose conclusions return to the
  main thread; every phase gets a `ruthless-reviewer` pass, and trust-boundary
  phases (the reconcile, the verdict ingest) additionally get `security-reviewer`
  before their PR. The ruthless pass on Phases 1–2 returned SHIP but surfaced two
  real hardening items (retry-safe `Plan`; the slipped atomic-write copy) that were
  fixed test-first before the PR.

- **Dogfooding earns the SOTA-per-intent principle its keep.** itd-80 is the
  pipeline's own payload, and its `## SOTA` declaration (Path 2 on every axis — a
  native floor with a seam to a mature external, no new dependency) is exactly the
  record that explains *why the build proceeded autonomously without a dependency
  hard stop*. The principle is not overhead here; it is the artifact that makes the
  no-stop decision auditable after the fact.
