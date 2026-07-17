# Run plan — planned-intent drain (record-catch-up + small self-contained cuts)

**Status:** run plan for an autonomous pass over `intents/planned/`, prepared
2026-07-16 after a full review of the brief, phase docs, and all 27 planned
intents. Purpose: drain the intents whose remaining work is self-contained and
requires no maintainer decision — largely **record catch-up** (the binary is
ahead of the design record) plus two small implementation cuts. Consumed by the
generic protocol:

```
/loop 90m Follow the protocol in .abcd/development/plans/2026-07-12-abcd-run-protocol.md
          for the run plan at .abcd/development/plans/2026-07-16-planned-intent-drain-run.md.
```

## Concurrent-agent notice (read first, every burst)

Another agent is **hunting bugs in this repo concurrently** (recent output:
iss-101, iss-102, iss-103). Coordination rules for this run:

- **Never** resolve, edit, or fix ledger issues — the ledger is the hunter's
  backlog. Bugs discovered during this run are **captured**
  (`abcd capture "<text>"`), never fixed inline.
- **Never** touch a branch you did not create this run (`fix/*` and any branch
  not matching `auto/itd-*` is out of bounds).
- At every burst start, `git fetch` and re-branch new items from the updated
  default branch; if a file mid-item changes on `origin/main`, finish the item
  on its branch (merge-commit-only; never rebase) and note the overlap in the
  PR body.
- Known overlap hotspots with the hunter's current findings:
  `internal/core/capture` (iss-102), the history index (iss-101),
  `internal/core/gitutil` (iss-103). Touch these only as each item's scope
  strictly requires; do not "drive-by" fix them.

## Run contract

```
backlog: milestones                    # the ordered items below; ready = depends_on merged
gate: make preflight                   # build + gofmt + vet + test + race (internal)
definition_of_done: per item — the intent's own ## Acceptance Criteria is the bar,
  verified AGAINST THE BINARY (CONTEXT.md: where record and binary disagree, the
  binary is checked first). Every AC gets an executable test or a recorded
  adjudication note; a green gate is only the floor. TDD watched-fail for any new
  behaviour. Record work (spec body, links, DECISIONS.md) lands in the same PR.
ledger_verb: abcd capture              # capture-only in this run; never resolve/wontfix
reviewers:
  correctness: ruthless-reviewer       # every non-trivial diff
  security: security-reviewer          # trust-boundary diffs (hooks, fs writes, parsing)
branch_policy: auto/itd-<N>-<slug> from main; merge-commit-only; never force-push;
  never commit to main; delete branch after merge (via gh api)
commit_trailer: Assisted-by: Claude:claude-fable-5
pr_policy: one PR per item; do not merge; do not enable auto-merge; PR body ends
  with the same Assisted-by trailer and states any overlap with hunter-touched files
budget: 45m wall per burst | 6 worker-agents/burst | NEXT.md + run-journal.json continuous
strike_limit: 3
stop_conditions:                       # in addition to every protocol STOP
  - Any item turns out to need a maintainer decision not pre-adjudicated below
    (new AC interpretation, kind change, phase re-scoping) — SKIP and record.
  - Anything touching launch/publishing (itd-65/66/67/69/72 chain) or an external
    account/app — out of scope by construction.
  - A change to any persisted schema_version (ledger/spec/memory/history) —
    irreversible class, human checkpoint.
  - Direct collision: an open PR or fresh commit by the other agent touches the
    same function this item must edit — skip the item this burst, journal it.
irreversible:
  - abcd spec close / lifecycle moves are git-reversible but public-record-shaping:
    close a spec ONLY after every AC is verified green or explicitly adjudicated
    in the spec body; when in doubt, leave open and record.
```

## Ready backlog (verified against the binary, 2026-07-16)

Ordered smallest-risk first. Each entry lists what was **verified already
built** vs the **remaining cut**.

### M1 — itd-89 `start-the-transcript-clock`: spc-4 write-up + closure (S)

- **Built:** `abcd hook session-end` (`internal/surface/cli/cli.go:858`), tests
  (`hook_session_end_test.go`), `hooks/hooks.json` SessionEnd entry, redacting
  `history.Capture` core. The implementation shipped; **spc-4 is a one-line stub**.
- **Remaining:** author spc-4's body (approach, milestones as delivered,
  AC-satisfaction mapping incl. the documented SessionEnd-vs-Stop deviation and
  the 64 MiB cap default); confirm each AC has a covering test (add any missing,
  watched-fail where new); then `abcd spec close spc-4` and run the fidelity
  review path (`abcd intent review itd-89` → `abcd intent ingest`).
- **Guard:** closure moves itd-89 planned→shipped via the lifecycle hook —
  verify record lint green before and after.

### M2 — itd-40 `folder-classification`: verify, gap-fill, record (S)

- **Built:** bare `abcd ahoy` is read-only classification (`ahoy.DryRun` →
  `FolderKind`); install bootstraps `~/.abcd/history/index.json`.
- **Remaining:** AC-by-AC verification against the binary (all five bullets);
  fill genuine gaps only (e.g. the unmanaged-repo adopt hint naming
  `/abcd:ahoy install`, index-resolved host location); create/link a spec
  (`abcd intent plan itd-40` path or `spec create` + `intent link`), write its
  body, close if all AC green.
- **Guard:** iss-101 (index lost-update race) is the hunter's — do not fix it
  here; if an AC requires touching that code path, test around it and journal.

### M3 — itd-4 `issue-capture`: adjudicated record catch-up (M)

- **Built:** `abcd capture [text] | list | resolve | wontfix` all live and in
  daily use; `related_specs`/`related_issues` schema live.
- **Pre-adjudicated for this run** (record in the spec body + one DECISIONS.md
  line, not silently): (a) AC-promote is satisfied at the skill layer by design —
  `cli.go:1540` records "promote is skill-orchestrated, never a CLI sub-verb
  (brief 04-surfaces/06)"; verify the plugin-surface promote path exists and is
  wired, else that AC is a genuine gap to implement at the skill surface.
  (b) The migration AC's source (`.abcd/.work.local/issues.md`) no longer
  exists — the migration already happened; record as satisfied-by-history with
  the evidence (resolved ledger entries), do not build dead migration code.
- **Remaining:** AC verification with tests, spec create/link/body, close if
  green.
- **Guard:** iss-102 (orphan-sweep commit race) lives in this package — no
  drive-by fixes.

### M4 — itd-46 `intent-quoted-text-create-symmetric`: small implementation (S/M)

- **Verified missing:** `abcd intent` has plan/link/review/ingest but **no
  create-from-text path**; bare `intent` is status.
- **Remaining:** implement `abcd intent "<text>"` create → `intents/drafts/`
  seeded from the quoted text, mirroring `capture [text]` routing; keep bare
  invocation status-only; decision leans stated in the AC are taken as written
  (transition alias per lean (a)); help text carries the one-line
  capture-vs-intent decision rule. Then spec + record as above.
- **Sequencing:** independent of M3, but land after it so the shared help-text
  and skill-surface conventions are settled once.

### M5 — itd-43 `epic→spec` remainder: advance, do NOT close (M)

- **Built:** Go tree is clean of `epic`; `related_specs` shipped;
  `glossary/core/spec.md` exists with `epic` in `forbidden_synonyms`.
- **Remaining (implementable):** the forbidden-synonym lint gate (AC #5 — no
  GL-family rule exists in `internal/core/lint` yet; build it detector-first
  against the glossary's `forbidden_synonyms`) and the live-prose sweep (AC #1),
  excluding historical records/research (the AC exempts them).
- **Blocked remainder (record, don't force):** AC #3 needs the native reviews
  subsystem (`spec-review` token), which is itd-28 — NEEDS-MAINTAINER (new
  gitleaks dependency). Land the implementable subset, record the blocked AC in
  the spec, **leave intent and spec open.**

## Explicit skips (verdicts from the 2026-07-16 review)

- **Maintainer decisions / human judgment:** itd-2 (wire-protocol undecided),
  itd-27 (large; spec-id mismatch to resolve; new third-party patterns), itd-28
  (new gitleaks dependency — sign-off), itd-34 (corpus-reshaping kind model),
  itd-36 (copyright judgment + human-run adversarial examples), **itd-88
  readout** (the coverage-experiment verdict is the maintainer's; its probe
  implementation is already merged — do not close spc-3).
- **External apps/accounts:** itd-6, itd-7 (RepoPrompt on a live Mac), itd-66,
  itd-67, itd-72 (published release / marketplace identity / tag pushes).
- **Blocked by absent substrate:** itd-20, itd-24, itd-63, itd-69 (spc-83
  operator-surfaces bundle not in the open store), itd-65 (blocked_by itd-66),
  itd-42 (blocked_by itd-27), itd-48/itd-53/itd-50/itd-58 (itd-47 headless
  oracle + seam enforcement wiring unshipped), itd-29 (own text defers until
  substrate + revisit triggers).
- **The ledger** — belongs to the concurrent bug-hunting agent this run.

## Why this is a safe run

Every ready item is either record catch-up over already-merged, already-tested
behaviour (M1–M3) or a small additive surface change with concrete
Given/When/Then AC and stated decision leans (M4–M5). The repo has a strong
deterministic gate (`make preflight`), merge-commit-only history, and the two
genuinely judgment-shaped calls in the backlog are pre-adjudicated above in
writing rather than left to mid-run improvisation. Nothing touches launch,
publishing, external apps, schema versions, or the bug hunter's files.
