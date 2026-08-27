# Security-advisory and issue handling — experiment record (2026-08-27)

A hand-run of an inbound-security-handling protocol, executed per the
script-first rule: run the smallest documented protocol by hand, grade where it
creaks, and only then file the automation intent that cites this note. It is the
sibling of the forge-mirror pilot
([`2026-08-19-forge-mirror-pilot.md`](2026-08-19-forge-mirror-pilot.md)) — same
ladder, different corpus.

This note is method-focused by design: it records the *procedure* and its
grading, not per-advisory exploit detail. The advisories are unpublished while
they are triaged and fixed, so their specifics stay out of this committed record
until they publish.

## Hypothesis

The end-to-end handling of inbound security reports — private advisories, public
issues, and the working-tier ledger together — is runnable by hand with the
existing surfaces (`gh`, the host agent, `abcd capture`, worktrees), and a
bounded run surfaces where the protocol creaks before a script hard-codes the
creaks. The target end state is a deterministic, SOTA-aligned automation that
works for every abcd-managed repo.

## Protocol as run

1. **Ingest.** Enumerate the three trackers: private advisories (`gh api
   repos/{o}/{r}/security-advisories`), public issues (`gh issue list`), and the
   `.abcd/work/issues/` ledger.
2. **Verify before fixing.** One independent reviewer per advisory (host
   security-reviewer), prompted to *refute* the claim against current `HEAD` —
   CONFIRMED / ALREADY-FIXED / REFUTED — with a file:line evidence pointer and an
   explicit scope correction where the report over- or under-claims. Evaluator
   outside the loop.
3. **Dedup the trackers.** Match public issues against the ledger by defect
   (file + root cause), not wording. Close stale mirrors already resolved in the
   ledger; backfill genuine ledger gaps; leave overlaps with private advisories
   out of the public ledger.
4. **Triage by exploitability, not reporter severity.** Rank by who can pull the
   trigger (a hostile clone attacking the abcd user outranks the operator
   polluting their own release), then by blast radius.
5. **Fix.** One worktree-isolated agent per fix-family, test-first (watch the
   regression test fail on `HEAD`, pass after), the canonical primitive reused
   (never a third copy), and the BUGHUNT sibling sweep (grep the pattern, not the
   instance). Agents never touch the ledger — the parent owns records.
6. **Review before the gate.** An independent adversarial pass per fix branch
   (evaluator outside the loop) confirms completeness (all siblings swept),
   correctness, and no over-reach, before any push.
7. **Gate.** The human adopts; verifier selects, gates decide. Push + PRs, with
   any unpublished-advisory id genericised out of public commit/PR trailers.
8. **Credit, differentiated by channel.** A security advisory's reporter is
   credited via the advisory `credits` field (type finder/reporter) — opt-in,
   they accept, and only on acceptance does public credit show; never
   `ACKNOWLEDGEMENTS.md`, and no public commit/PR trailer names them before they
   consent. A merged PR credits its author natively through the contributor graph
   and commit authorship (a contributor, not a write-access collaborator). A
   public issue's external reporter is credited by citing them in the fixing
   commit — `Reported-by: @handle` plus `Fixes #N` — stamped only when the issue
   author is not the operator's own account, never a self-credit. Post a
   name-free triage acknowledgement in each advisory thread within the
   SECURITY.md week. `Assisted-by` is stamped from the actual per-commit authoring
   model, never a guessed or intended one (F-G).
9. **Record & automate.** This note grades the run; the automation intent cites
   it.

## Findings (living — graded as the run proceeds)

- **F-A — recall-before-record must be explicit, not assumed.** A capture made
  mid-run duplicated an existing drafted intent and was caught only *after* the
  write. Once the agent adopted an explicit ledger recall-check before every
  capture, two subsequent captures were correctly identified as already-tracked
  *before* recording. Until the capture-time validator ships, the documented
  protocol (recall-check first) is the gate. Product fix captured
  (`iss-2608270550201935`: a non-blocking advisory-dedup echo in `capture`).
- **F-B — verification catches reporter drift both ways.** Independent
  refutation surfaced scope corrections a fix would otherwise have over-applied:
  one advisory had only one of two cited sites live-exploitable (the other dodged
  by a flag); one carried a delivery caveat (a plain clone does not import the
  remote config the exploit needs); one public issue was the same fix as a
  private advisory. Fixing to the *verified* claim, not the reported one, is
  load-bearing.
- **F-C — worktree isolation + test-first scales the fix stage.** Independent
  fixes ran in parallel, each with its own watched-fail regression and package
  race/vet/gofmt gate, without interfering.
- **F-D — collision management is the open cost.** Parallel agents on shared
  files (one secret-pattern file, one guard matcher, one launch bundler) must be
  grouped by family or sequenced, and *landing many branches into one cut is the
  unsolved integration step* — the automation needs a defined merge strategy
  (integration branch vs ordered merges) rather than N independent PRs.
- **F-E — three trackers, no id linkage.** The dedup pass found the public-issue
  and ledger populations almost entirely disjoint, and the private advisories are
  a third, unlinked set. A deterministic automation needs one unified view keyed
  on the defect, or the same bug is worked twice.
- **F-F — agents stay out of the ledger.** Keeping record/credit writes with the
  parent avoids parallel id-mint races and keeps resolution riding the fixing
  commit (RS001) under the parent's control.
- **F-G — attribution must read the running model, not any self-report, and
  credit must name the true reporter.** Two concrete failures this run. (1) The
  session stamped a commit from its templated identity blurb (Fable) and had to
  be corrected against the operator's actual model indicator (Opus 4.8). (2)
  Fix subagents launched with a `model: opus` override *intending* Opus 5 in fact
  ran on `claude-opus-4-8` — the transcripts' own `"model"` field is the proof —
  yet each wrote `Assisted-by: Claude:claude-opus-5` into its commit because the
  brief said so. The override intent, the brief, and the session blurb were all
  unreliable; only the per-agent transcript model field was authoritative. The
  automation must therefore stamp `Assisted-by` from the **actual per-commit
  authoring model**, and credit the reporter correctly by channel: advisory
  reporters via the GHSA `credits` field (opt-in), a public issue's external
  reporter via `Reported-by: @handle` + `Fixes #N`, the operator's own routine
  via `Fixes #N` alone — never a self-credit, never a guessed model.
- **F-H — external contributions need provenance-aware intake, not a shortcut.**
  Inbound issues (and PRs) split by author: the operator's own bug-hunt routine
  arrives already classified and labelled, while external contributors' issues
  arrive with *no* classifiers (no severity, no ledger record) and from a person,
  not a routine. Respect for a contributor means their issues, PRs, and
  advisories are **prioritised, acknowledged, and credited** — it does *not* mean
  verification is skipped: a known, trusted contributor's report still gets the
  same independent verify-against-`HEAD` as any other, and nothing is rushed
  (trust sets priority and courtesy, not the evidence bar). The intake step is
  therefore provenance-aware: (a) acknowledge + prioritise, (b) verify the claim
  against `HEAD`, (c) assign the missing classifier — severity by exploitability,
  (d) dedup against the ledger and record it, marking external provenance, (e)
  credit the contributor (issue/PR credit, advisory `credits` field — opt-in,
  never `ACKNOWLEDGEMENTS.md`). Two gaps this run exposed and that the automation
  must close: the ledger `source` enum has no external-contributor value, so
  provenance is currently unrecorded; and nothing links an external issue → its
  ledger record → the contributor's credit.

- **F-I — a fix has four roles, and a multi-maintainer repo must not conflate
  them.** Reporter (who found it), AI assistant (which model), author of record
  (the accountable human who drove it — the git author), and authoriser (the
  maintainer who reviewed and approved the merge). A single-maintainer repo
  collapses author and authoriser into one person, so three signals suffice
  (`Reported-by` / `Assisted-by` / git author). With multiple maintainers the
  authoriser is distinct: GitHub captures it natively through required PR review
  (the approving reviewer) plus `merged-by`, enforced by branch protection.
  **Decided approach for abcd:** keep required review (authorisation is enforced
  and captured), and the merge step stamps `Reviewed-by: Name <email>` from the
  approving review for git-native durability, so one commit carries author /
  `Assisted-by` / `Reported-by` / `Reviewed-by`. `Reviewed-by` is not the
  `Signed-off-by` adr-43 dropped — it attests review/authorisation, not origin,
  and a "require review from someone other than the author" branch rule makes
  self-authorisation impossible. Two implementation seams: the branch-protection
  config (require a non-author approval) and the merge automation (stamp
  `Reviewed-by` from the PR's approving review). The automation stamps all four
  roles deterministically.

- **F-J — agent-surfaced siblings need a flag-then-parent-capture intake.** A
  fix agent deep in one file routinely spots an adjacent defect out of its scope
  (an unhardened sibling call, a nitpick, a follow-up). Losing those loses the
  most valuable context: the finding seen in situ. In this orchestration the
  agents are barred from the ledger (to avoid parallel id-mint races and keep fix
  commits atomic), so they FLAG the sibling in their report and the PARENT
  captures it, serialised through one writer, recall-checked, and fixed in a
  SEPARATE change (record first, never widen the fix's scope). Six such siblings
  were captured this run (installsurface `dirTree.resolve` guard, guard
  `noglob`/`nocorrect` wrappers, the #357 block-sequence remainder, the memory
  write-time control-char path, the `.lock` skip-extension, the site lint-config
  read). The automation formalises this agent-to-parent capture channel; the
  capture-time dedup echo (iss-2608270550201935) then makes the parent capture
  safe from duplicates.

## Integration review (live log)

The step F-D named as the open problem: landing many independent fixes into one
cut. Recorded in the detail the automation needs.

**Pre-flight of the integration itself:**
1. **Peer scan before any git-state change.** `ListAgents` shows one idle peer
   session (`abcd-cli-7c`); worktree creation churns the shared checkout (the
   iss-2608261331317889 window), so the integration runs in its own worktree and
   the peer is watched.
2. **Working-tree hygiene.** The primary checkout carries ~27 uncommitted paths
   (this note, itd edits, the session's captures, peer WIP), so the integration
   branch is cut from the clean committed `main`-HEAD in a DEDICATED worktree,
   never from the dirty tree, so none of that leaks into the cut.
3. **Footprint map to merge order.** Each branch's `git diff --name-only
   main...branch` gives the conflict map. Disjoint-file branches merge first;
   overlapping branches merge consecutively so a shared file is resolved once.
   Predicted conflict files: `guard/match.go` (#318 + guard-family),
   `scanner/scanner.go` (#370 + #328) and `patterns.go` (#358),
   `launch/bundle.go` (h2gm + #328) and `render.go` (#328 + #488),
   `memory/ingest.go` (72rp + j5f5) and `ask.go` (72rp + #250/262),
   `ahoy/store.go` (xrf8 + #334), `cli.go` (#250/262 + #485). #328 is a superset
   of the launch-payload branch, so that branch is not merged separately.

**Executed:**
- **Base:** `integration/security-cut` cut from clean `main`-HEAD in a dedicated
  worktree; base confirmed clean (no uncommitted leak).
- **Merge order:** disjoint standalone branches first (4), then first-of-cluster
  and disjoint (8), then the 7-branch overlap cluster last.
- **Conflicts:** of 19 merges (the launch-payload branch folded into its #328
  superset), **18 auto-resolved** — git's 3-way merge handled non-overlapping
  hunks even in shared files (`ingest.go`, `ask.go`, `scanner.go`, `bundle.go`,
  `render.go`, `cli.go`, `store.go`). Exactly **one** textual conflict:
  `guard/match.go`, where #318's `coproc` branch and the guard-family's
  case-folded wrapper lookup edit the same line in `commandOf`. Resolved
  **keep-both** (both hardenings coexist); guard package rebuilt and tested green
  before committing the merge.
- **The payoff — a finding only the tree-wide suite could surface (this is why
  F-D matters).** `go test ./...` on the integrated tree failed
  `TestTestGitCallsAreHermetic` (iss-28): the GHSA-h2gm fix's new
  `fsmonitor_exec_test.go` deliberately spawns git with a hostile repo-local
  `core.fsmonitor`, which the hermetic-git gate forbids. It passed in the fix's
  own `go test ./internal/gitutil/...` because the meta-gate lives in a different
  package and only runs tree-wide. **No per-branch review could have caught it.**
  Resolved with a file-precise allowlist entry (matching that one file), mirroring
  the existing lifeboat exemption.
- **Gate on the candidate:** `go build ./...`, `go vet ./...`, `go test ./...`,
  `go test -race ./internal/...`, and `gofmt -l .` all clean.

**Still to do before the cut (recorded so the automation covers them):** full
`make preflight` (the lint gates: lint-reviews / lint-issues / record-lint /
docs-lint / site-render); one consolidated adversarial security review over the
whole integrated diff (evaluator outside the loop); trailer reconciliation
(correct every merged fix commit's `Assisted-by` to `claude-opus-4-8` and add
`Reported-by: @jogrun` to the four jogrun fixes — F-G/F-I); and the
changelog/resolution reconciliation (the security issues' ledger records, minted
on the unmerged backfill branch, must resolve in the cut for the derived
changelog to describe it).

**Cut — Option A (chosen 2026-08-27).** Bring the fixes and their resolved
records into one tree, then ship. Step 1 (land the backfill so the records exist
on `main`) was already done — PR #522 had merged, so `main` carries the 31
backfilled records and, crucially, NO Go code changed since the cut's base, so
the sync is conflict-free. The cut branch is synced to `main`, then each security
fix's record is resolved (`impact: fix`, a patch bump) — the ~15 backfilled ones
resolved in place, the five deliberately-unbackfilled public issues (#324, #328,
#335, #486, #487) and the private GHSAs minted-then-resolved with a hardening
description and no exploit detail. Preflight re-runs (RS001 now validates the
`Resolves:` trailers); the cut PR merges; `abcd launch ship` derives the version
and composes the changelog from the resolved records; the GHSAs then publish with
opt-in credits. Release retention (itd-70 / `retention.go` `ComputeRetention`:
newest-patch-per-`MAJOR.MINOR`, never prune the just-published) prunes the older
patches of the published line at ship time.

- **F-K — a derived release ships the whole unreleased backlog, so a "scoped"
  (security-only) cut is not achievable while other resolved work is
  unreleased.** The cut derives from every record resolved since the last tag. At
  cut time that was 95 records (70 `fix`, 24 `internal`, 1 `additive`), of which
  only 29 were this session's security fixes — the other 66 were a pre-existing
  backlog of resolved-but-unreleased work from prior sessions. abcd has no lever
  to ship a subset (`shipped_in` only excludes already-released work, which these
  are not). The lesson for the automation and the operator: a security release
  cannot be isolated from the release backlog — either the backlog ships with it,
  or the security fixes wait. Pragmatic resolution taken here: stop treating the
  cut as security-only, run the deferred docs/correctness sweep first, and ship
  one combined release. Also recorded: the derivation refuses when the running
  binary's surface snapshot mismatches the tree (a stale-binary guardrail) —
  rebuild from the tree being released before deriving.

- **F-L — proactive agent-to-agent coordination must be a default, not
  human-prompted.** Two concurrent sessions (this security cut and the
  docs/inconsistency sweep) had interdependent work heading to ONE combined
  release, yet coordination happened only after the operator prompted it twice —
  the human was acting as the relay between them. AGENTS.md's concurrency rule
  already says to "announce the mutation to any peer found", but that was not
  self-invoked: peers were repeatedly scanned (`ListAgents`) without ever being
  messaged. The lesson: when a session detects a concurrent peer on the same
  repo/release, it should proactively open a coordination channel over
  `SendMessage` — current state, constraints, the shared-release decision,
  sequencing — the moment the interdependence is visible, not when the human asks.
  The automation makes peer coordination a reflex triggered by a detected
  interdependent peer, and the convention graduates from "announce a mutation" to
  "coordinate interdependent concurrent work by default".

- **F-M — enabling a security control mid-flight caught the security work's own
  test fixtures.** Turning on GitHub secret-scanning push protection (a posture
  change made this session) immediately blocked the cut's push: the
  redaction/scanner fix TESTS embedded literal secret-shaped tokens (fake
  `ghp_`/`AKIA`/JWT) needed to exercise detection, and push protection cannot tell
  a fake from a real one. Two lessons. (1) Self-validating: the control worked on
  day one — it caught real-looking tokens the moment it was armed. (2)
  Fixture hygiene: a test that feeds a secret-shaped input to a detector must
  CONSTRUCT the value at runtime (`"ghp_" + "0123…"`, split each `eyJ`), never
  embed a literal, or it trips the very control it helped ship — and CI
  `gitleaks`. The fix was a tree-filter splitting each literal across the cut's
  history (the fix/ branches kept as a fallback); the repo's own
  `storeEntropySpecimenGolden` already models the convention. The automation's
  fix-agent brief must require runtime-constructed secret fixtures. (Also logged:
  a transient push-pack transfer corruption that a repack cleared, and 4 dangling
  corrupt objects the repack surfaced — pre-existing, unreachable, harmless, prune
  later.)

## Verdict (interim)

The protocol survives contact through verify → dedup → fix → (review pending).
The one structural gap is F-D (landing many fixes into one cut). Grade the
remaining stages (review, gate, land, credit) as they complete, then file the
automation intent — "abcd handles inbound security advisories and issues for
every managed repo" — citing F-A…F-G as acceptance criteria.

## Audit trail (this run)

- Advisories verified: 7 (all CONFIRMED on `HEAD`).
- Public criticals fixed: 3. Public security-majors fixed / in-flight: see the
  session's fix branches (`fix/ghsa-*`, `fix/gh-*`).
- Ledger dedup: 46 open public issues vs the ledger — 7 stale mirrors closed, 31
  genuine gaps backfilled (PR for the backfill raised).
