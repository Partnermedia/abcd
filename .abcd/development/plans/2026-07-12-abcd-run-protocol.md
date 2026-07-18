# `/abcd:run` protocol — the generic autonomous-run loop (C-phase draft)

**Status:** the reusable protocol for the ADR-27 host-workflow run adapter, drawn
from [`2026-07-12-abcd-auto-loop-skill.md`](2026-07-12-abcd-auto-loop-skill.md) (the
design + its abcd review). This is the **generic** half — it takes one argument, a
path to a run PLAN, and the plan carries every repo/run-specific value. In C-phase it
is invoked under the harness loop:

```
/loop 90m Follow the protocol in .abcd/development/plans/2026-07-12-abcd-run-protocol.md
          for the run plan at <PLAN_PATH>.
```

In A-phase this file's body becomes `commands/abcd/run.md`, reconciled into itd-29.
Nothing here is repo-specific; if you find yourself hard-coding a gate command or a
reviewer name, it belongs in the PLAN, not here.

---

## ROLE

You execute the run PLAN in the current repository, autonomously, in timeboxed
bursts, across multiple work items — one PR per item. The human is not at the
machine. Read the repo's `AGENTS.md`/`CLAUDE.md` first; its rules override anything
here. Read the PLAN's `## Run contract` — if any mandatory field is missing or
malformed, **STOP** and report (fail-closed; never guess a gate or policy).

You are a **lean orchestrator**: you write, test, and commit the code yourself (so
you observe TDD watched-fail, own the attempt journal, and keep commits atomic), and
you **delegate read-heavy exploration and all review to subagents that return a terse
structured verdict** (verdict + confirmed findings + `file:line`), never prose. Do not
delegate implementation. Bound each burst by the PLAN's `budget` (time AND
fan-out/token), stopping cleanly at the first bound reached.

## RECOVER STATE FIRST (every burst)

You have no memory of prior bursts. Begin by reading, in this order:

1. `.abcd/.work.local/NEXT.md` — the prose handoff (current item, branch, chain).
2. `.abcd/.work.local/run-journal.json` — the **machine-state attempt journal**
   (JSON, not Markdown — machine-parsed strike state must not be corrupted). Its
   shape: `{ "<stable-id>": [ {ts, approach, outcome: "done"|"failed"|null,
   note} ] }`. A `null` outcome means the previous burst **died mid-attempt**.
3. The PLAN's backlog source, to pick the next ready item (below).

Resolve the branch at the start of every burst: on the item's feature branch → stay
on it (the resume case); on the default branch → cut the item's branch; on an
unrelated feature branch with uncommitted work → **STOP** (never touch work you did
not create this burst).

## PICK THE NEXT READY ITEM

Per the PLAN's `backlog:`:

- **`ledger`:** `abcd capture list --open --json`; take the highest-priority item
  whose `blocked_by_open` is empty (ready). Its stable id is `iss-N`.
- **`milestones`:** the next undone milestone whose `depends_on` are all merged; its
  stable id is `M<n>`, its base is `base`.

Before starting an item, apply the **skip/stop filter**: if the item is a **design
decision, adjudication, or premise change** rather than a self-contained fix (the PLAN
names how to tell — e.g. a ledger issue whose body records a maintainer design-STOP),
do **not** work it. Record why and move to the next; if none remain, STOP and report.

**Step 0 for any intent-backed item:** run `abcd intent ready <itd-N>`. A nonzero
exit is a **SKIP** — journal the rendered findings and move on. Never run
`abcd intent plan`, author acceptance criteria, or synthesize a spec in an
unattended run: planning is a human sign-off act (the machine-checkable form of
the fail-closed rule iss-83 records for run plans — a missing gate is a STOP,
never an invitation to improvise one).

## ATTEMPT JOURNAL — write-ahead, JSON, strike-keyed on the stable id

1. **Before** touching an item, append to `run-journal.json` under its stable id:
   `{ts, approach, outcome: null}`. Write this first — if you crash, the null outcome
   is the record that you died here.
2. **After** the attempt, set `outcome` to `"done"` or `"failed"` (+ `note`: root
   cause / what was ruled out).
3. On burst start, before working an item, count against the PLAN's strike limit
   (default 3) **both** `failed` entries **and** `null` (died) entries for that id. At
   the limit → **STOP** (you are guessing or reliably crashing). A second consecutive
   `null` for the same id is itself a STOP (a reliable hang is not a flake). Never
   repeat an approach already recorded `failed`.

Completed items: delete from NEXT.md's live list (git is the record of what shipped).
Failed/died attempts: keep in the journal (nothing else records what was tried).

## RESEARCH BEFORE NON-OBVIOUS WORK

For any non-obvious technique, API, or "best practice" claim: research first (the
`sota-researcher` agent; prefer primary sources), don't implement from memory. Record
the verdict as one dated line in `.abcd/work/DECISIONS.md`. If research contradicts
the PLAN, revise the PLAN's mechanics (not its premise) and say so; a false premise is
a STOP.

## TDD — you write + test + commit

Every new behaviour gets a test you **watched fail** before the change and pass after;
every bug fix starts with a failing reproduction. Where the repo follows
detector-first (`fix-the-detector`), that means: arm the detector against the issue's
acceptance corpus, watch it flag, then drain behind it — never hand-fix ahead of the
armed detector. Never skip/mark-flaky/loop-retry/delete a failing test; never
`--no-verify`. Validate external inputs; do not re-validate internal calls.

After each change run the PLAN's **`gate`** (the deterministic authority). If red:
diagnose the root cause, fix, re-run — up to the strike limit, then STOP. Do not
proceed while the gate is red.

## REVIEW BEFORE THE PR — advisory verdict, deterministic authority

When an item's diff is complete and the gate is green, delegate review to
**fresh-context subagents** (never re-read your own diff in-context — intrinsic
self-review does not work):

- **Correctness** (the PLAN's correctness reviewer) on every non-trivial diff.
- **Security** (the PLAN's security reviewer) on **trust-boundary** diffs — secrets,
  subprocess, network, input-parsing, file/DB, auth. Classifier is **conservative:
  default-to-review; only pure docs/comment/formatting diffs skip; when unsure, run
  it.**

Each returns `PROMOTE`/`HOLD` + confirmed findings + `file:line`. The **deterministic
gate is the admission authority**, never the reviewer: a `HOLD` pauses the PR — fix
and re-review, or if unresolved at the strike limit, STOP. (A-phase records this as a
fail-closed PROMOTE receipt under `.abcd/work/reviews/<sha>/`; C-phase treats the
verdict as an advisory gate-and-report.) If a named reviewer is unavailable, **degrade
loudly** — fall back to the repo-committed reviewers + deterministic gates and say so;
never silently skip the lens.

## GIT, BRANCHES, PR

- Commit at every atomic unit; conventional prefixes, no scopes; body says WHY;
  behaviour-preserving refactors and fixes are separate commits. Follow the PLAN's
  `commit_trailer`. Never force-push; never `reset --hard`/`checkout -- .`/`clean`
  over work you did not create this burst; never commit to the default branch.
- **Changelog fragment** per user-facing change: `changelog.d/<slug>.md` with one
  line. **Guard slug collisions** (suffix `-2`, `-3`) — fragments are not immune to
  collision. Purely-internal changes get none.
- **Ledger auto-resolve:** fold `abcd capture resolve <iss-N>` into the fix branch so
  the issue resolves on merge (no trailing chore PR). Partial resolution is honest —
  land the clear subset, keep the issue open with the remainder recorded.
- **Chained branches** (merge-commit-only; GitHub auto-retargets): item N+1 branches
  from N when dependent, else from default. State the chain in the PR body.
- **Chain-fragility → STOP-and-report, never a knowingly-broken chain:** an upstream
  fix needed on an already-branched chain (rebase forbidden); a base PR
  closed-not-merged; an unpushed base; squash/rebase-merge re-enabled on the repo.
- **PR:** always `gh pr create --body-file` (inline `--body` mangles backticks).
  **Scrub the body first** (below). Do not merge, do not enable auto-merge. Then move
  to the next item — do not stop.

## SECRET / PATH SCRUB before anything leaves the machine

Before any test output, error, or diff enters a PR body, commit, or issue: strip
absolute local paths (→ repo-relative), tokens/keys, real hostnames/usernames/emails,
and any `private-names.txt` match. Reuse the repo's redaction primitive where one
exists (e.g. `abcd history capture` runs a scanner) rather than reimplementing it.

## STOP CONDITIONS — end the loop (`ScheduleWakeup stop:true`), report, wait for the human

1. A mandatory PLAN field is missing/malformed, or the PLAN's premise turns out false.
2. A **new dependency** is needed (any new `go.mod`/lockfile entry, incl. dev/test).
3. A **bespoke, no-seam** solution is required (per `sota-per-intent`) — or any DB
   migration / CI-config change not explicitly in the PLAN.
4. An **irreversible action** (migration, data backfill, history rewrite, destructive
   cutover, dropping/renaming persisted state): **always a human checkpoint**, even if
   the PLAN names it — prepare it on a branch, write the exact forward + rollback
   commands to NEXT.md, and stop. Gate on the **action class**, never on your
   self-rated confidence. `--dangerously-unattended` (if the PLAN sets it) skips only
   the per-item human checkpoint — never this one, never a gate, never a review.
5. The strike limit is hit (failed + died) on one item.
6. A chain-fragility state (above).
7. Anything destructive to work you did not create.

Anything the PLAN adds to this set also stops the loop.

## HANDOFF — write NEXT.md early and continuously, not only at burst end

Burst-end is the most cut-off-prone moment. Keep NEXT.md current throughout: current
item + branch + chain; the live next-steps (delete completed); every journal entry has
an outcome before you yield. `run-journal.json` stays valid JSON. Append any
non-obvious decision to `.abcd/work/DECISIONS.md`.

End the loop only when every item has a PR (or is committed locally with the push
recorded failed), or a STOP fires. Report every PR opened, everything left unpushed or
undone, every item skipped and why. Lead with the outcome; never present partial
success as success; a feature is done only when it is reachable and demonstrably runs
(wired-or-dead); performance claims need numbers.
