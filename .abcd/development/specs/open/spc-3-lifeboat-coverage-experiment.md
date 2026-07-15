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

Delivered across the plan's milestones M0–M3a:

- **M0 — the hypothesis.** `internal/core/lifeboat/mapping.go` holds the
  brief↔lifeboat contract as a table (23 brief sections × the best status each
  source tier could ground), rendered into the brief's `00-meta.md`. It is a
  *prediction*; the probe measures the same sections and the two are directly
  comparable.
- **M1 — the transcript clock** (itd-89, adr-29): a `SessionEnd` hook so the
  transcript corpus starts accruing during the rest of the build. The only
  irreversible item, so it ships first.
- **M2 — the probe and the aggregate** (this spec's core).
- **M3a — the read-only spine** (adr-35, adr-36): the plan orchestrator and the
  `disembark plan` dry-run verb. It assembles the full lifeboat file set *in
  memory* — brief citation maps, the coverage report, verbatim record copies, the
  rescue spine, and a pinned manifest hash — writing nothing. The destination
  write path is M3b.

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

## Approach — the plan (M3a)

`abcd disembark plan <repo> [--json]` shows the complete file set a pack would
write, without writing anything. It is the read-only spine on which M3b's write
path bolts: **one** code path — `lifeboat.Plan` — produces the file set, so a
dry-run cannot describe a pack that a real pack would not perform. The plan
carries:

- **Brief citation maps** for grounded and partial sections only. Each is an
  honest map back to the evidence the probe cited, not synthesised prose — a
  blank section gets no brief page, and sections whose lifeboat home is outside
  `brief/` (graveyard, docs/adrs, activity/issues, rescue) are materialised at
  top level, not as brief stubs.
- **The coverage report**, `coverage.json` and `coverage.md`, first-class.
- **Verbatim record copies**: the ADRs (from the native decisions dir and every
  conventional ADR home) and the issue ledger, byte for byte.
- **The rescue spine**: the intent corpus copied verbatim where one exists, a
  single git-derived summary where it does not.
- **`_provenance.json`**, written last, carrying `manifest_sha256` — SHA-256 over
  the sorted `<sha256>  <path>\n` lines of every *other* file (adr-35's pinned
  definition), so the lifeboat's integrity is verifiable and a re-plan of an
  unchanged source is byte-identical (the provenance carries no timestamp).

**Schema v2** (adr-36): a `SectionCoverage` gains a `Kind` (`extractable` vs
`human-owned` — the durable form of the M2 gate decision), and a blank gains a
`Resolution` and, once answered, an authored `Answer` whose provenance is a
person and a date, never a file.

The plan is assembled through a builder holding three invariants an adversarial
review demanded, because the source tree is treated as untrusted:

- **No duplicate destinations.** A real pack writes one file per path, so the
  plan must never list a path twice. The same ADR basename in two source homes
  (a migrated ADR left in both `docs/adr` and `docs/adrs`) resolves to one dest
  by first-writer-wins in sorted-source order — otherwise the manifest would
  describe a two-file write a pack cannot perform.
- **A bounded plan.** Per-file and per-directory caps bound one read; a whole-plan
  file/byte ceiling bounds the product, so a pathological tree cannot exhaust the
  operator's memory by multiplying many bounded reads.
- **No silent gaps.** A record too large to read, unreadable, or dropped at the
  ceiling is recorded as an `Omission` in `_provenance.json` and the dry-run —
  the lifeboat declares what it left out rather than losing it.

Destination filenames — and directory components — are derived through a
`safeLeaf` guard that *rejects* (drops the file), never normalises, any name that
is not already a clean single component, so a hostile source filename or
directory name cannot steer where a file lands.

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

`disembark probe`/`coverage`/`plan` are CLI-only operator tooling, like `spec`
and `rules` — not a `/abcd:` command surface. `plan` is a dry run: it writes
nothing, so it too stays operator-internal. The `disembark` brief row stays
`staged` until M3b ships the packer's write path, so no `commands/abcd/disembark.md`
exists and the `surface_coverage` record-lint rule is satisfied.

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
  is M4; the shared plan/pack code path lands here in M3a as `lifeboat.Plan`, the
  single producer both `disembark plan` (now) and the packer (M3b) run, so a
  dry-run cannot describe a pack a real pack would not perform.

## Out of scope for this spec

The packer's **write path** (M3b: `os.Root` write containment, the destination
safety gate, staging-then-rename, secret-scan-before-write, the voyage ledger,
and the `/abcd:disembark` command surface), embark, the round-trip, the
graveyard's interpretation layer, and synthesis over the record — all later
milestones of itd-88, tracked in the plan.
