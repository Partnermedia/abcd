---
id: spc-3
slug: lifeboat-coverage-experiment
intent: itd-88
---
# lifeboat-coverage-experiment

## Summary

spc-3 delivers the read-only measurement half of the lifeboat for itd-88: the
`disembark probe` and `disembark coverage` verbs and the tiered source adapters
behind them. No lifeboat is packed and no source repository is mutated — the
packer, embark, and the round-trip are later milestones that build on the
adapters this spec lands. The experiment's readout is the cross-repo coverage
aggregate: probe a record-rich repo and a git-only repo, and the delta in
brief-section coverage is what the record is worth.

Delivered across the plan's milestones M0–M2:

- **M0 — the hypothesis.** `internal/core/lifeboat/mapping.go` holds the
  brief↔lifeboat contract as a table (23 brief sections × the best status each
  source tier could ground), rendered into the brief's `00-meta.md`. It is a
  *prediction*; the probe measures the same sections and the two are directly
  comparable.
- **M1 — the transcript clock** (itd-89, adr-29): a `SessionEnd` hook so the
  transcript corpus starts accruing during the rest of the build. The only
  irreversible item, so it ships first.
- **M2 — the probe and the aggregate** (this spec's core).

## Approach — the probe

`abcd disembark probe <repo> [--json]` walks a repository read-only and reports,
per brief section, whether a lifeboat could ground it, at what tier and
confidence, citing the evidence — and, for a blank, what was searched and the
question a human must answer.

- **`SourceContext`** is the read-only material every adapter probes, built once
  per repository. Every file read is contained to the repo root via `os.Root`
  (no symlinked component can redirect a read outside the repo), bounded
  (`maxProbeReadBytes`), and non-blocking (`O_NONBLOCK|O_NOFOLLOW`), so probing a
  hostile or archived tree cannot escape it, hang on a FIFO, or exhaust memory —
  reusing the hardening the privacy audit adopted. Git is queried through an
  env-isolated, cached runner (`gitutil.Run`).
- **`Source`** is one tiered adapter — `Section() / Tier() / Probe(*SourceContext)
  Evidence` — reading a single brief section at a single tier. The three tier
  constructors live in `sources_git.go` (Tier 0), `sources_conventions.go`
  (Tier 1), and `sources_native.go` (Tier 2).
- **The orchestrator** (`Probe`) runs every adapter whose tier is present
  concurrently, then reduces to the best evidence per section — highest status
  wins, a richer tier breaks a tie — and falls back to an honest blank (carrying
  what was searched and the question a human must answer) for every section no
  adapter grounded. Every one of the mapping's sections appears exactly once and
  the result is deterministic.
- **The Evidence contract** is the anti-fiction rule made structural: a
  `grounded` or `partial` result must cite non-empty evidence; a `blank` must
  carry a question. A blank is a first-class result, not a failure.

## Approach — the aggregate

`abcd disembark coverage <report.json>...` reduces several probe reports to the
cross-repo table (section × repo), with an always-blank verdict per section (a
section no probed repo could ground). That table is the artefact the M2 gate
reads to decide which brief sections survive. The coverage schema carries
`schema_version`; a report from a newer schema is refused with an upgrade
message rather than silently misread.

## The tiers

| Tier | Reads | Present in |
|---|---|---|
| 0 — git | commit history, reverts, deleted files, tags, dependency churn | every git repo |
| 1 — conventions | README, docs/, CHANGELOG, LICENSE, CONTRIBUTING, ADRs wherever they live, manifests, CI | most repos |
| 2 — abcd-native | `.abcd/` brief, decisions, intents, specs, roadmap, work issues/reviews, DECISIONS.md | an abcd-managed repo |

The `graveyard` section grounds from Tier 0 alone — what a project abandoned is
in its git history (reverts, unmerged branches, files deleted after substantial
history, dependencies added then removed) whether or not anyone wrote it down.

## Surface

`disembark probe`/`coverage` are CLI-only operator tooling, like `spec` and
`rules` — not a `/abcd:` command surface. The `disembark` brief row stays
`staged` until M3 ships the packer, so no `commands/abcd/disembark.md` exists and
the `surface_coverage` record-lint rule is satisfied.

## How it satisfies the Acceptance Criteria

- *Poor repo still reports* — `Probe` never fails on a record-poor repo; every
  section it cannot ground returns a blank with searched + question.
- *Byte-identical source* — a test hashes a fixture tree before and after a probe;
  reads are contained and read-only by construction.
- *Grounded cites its source* — the Evidence contract requires non-empty evidence
  for any non-blank status; a test asserts no non-blank section cites nothing.
- *Cross-repo delta legible as a number* — `Aggregate` renders section × repo with
  a per-section verdict; the rich-vs-git-only delta is the summary counts.
- *Deterministic* — a test probes the same repo twice and compares JSON byte for
  byte.
- *Graveyard from git alone* — a git-only fixture with a reverted commit grounds
  `graveyard` at Tier 0.
- *cite-or-be-dropped / dry-run cannot lie* — the graveyard interpreter's validator
  and the shared plan/pack code path are M4 and M3 respectively; this spec lands
  the adapters and the read-only `Probe` they share.

## Out of scope for this spec

The packer (`disembark to`), embark, the round-trip, the graveyard's
interpretation layer, and synthesis over the record — all later milestones of
itd-88, tracked in the plan.
