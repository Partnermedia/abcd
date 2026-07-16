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

Delivered across the plan's milestones M0–M3b:

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
  rescue spine, and a pinned manifest hash — writing nothing.
- **M3b — the write path** (adr-35): `disembark pack <repo> <dest>` writes that
  file set out-of-tree, behind a destination safety gate, staged-then-renamed and
  contained to the destination via `os.Root`, secret-scanned before any write,
  with an append-only voyage ledger — and the `/abcd:disembark` command surface.
- **M4 — the graveyard** (adr-35): three strictly-ordered layers. Deterministic
  git archaeology and recorded abandonment are packed into every lifeboat as
  `graveyard/archaeology.json` and `graveyard/abandoned.json`; host-delegated
  interpretation is ingested afterwards by `disembark graveyard --lessons-json`
  under a cite-or-be-dropped Go validator.
- **M5 — embark and the round-trip**: `embark probe` (read-only: provenance
  schema gate, manifest verification, bulk conflict detection) and
  `embark from` (the contained write path), the packer carrying specs, and the
  round-trip properties — through-the-stores equality, the record-derived
  sub-manifest closure (P1), and literal self-closure (P2).
- **M6 — synthesis over the record**: three post-pack synthesis verbs
  (`disembark principles`/`press-release`/`oracle`) as injected host-delegated
  seams with deterministic fallbacks, the first Go home of the registered
  {SHIP, NEEDS_WORK, MAJOR_RETHINK} verdict, and the itd-5 prompt
  infrastructure (prompt_version, reads_untrusted_input, injection canaries)
  for the four synthesis agents.

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

## Approach — the pack (M3b)

`abcd disembark pack <repo> <dest>` writes the plan to `<dest>`. It is `Plan`
plus a contained write — the same file set the dry-run showed, so the two cannot
diverge. Everything that stops a pack destroying real work is here:

- **Destination safety gate.** A pack refuses unless `<dest>` is absent, an empty
  real directory, or an existing directory carrying a parseable `_provenance.json`
  — never a directory abcd did not produce. It also refuses a symlinked
  destination, one inside a `.git/` directory, and any destination that overlaps
  the source tree (equal to, an ancestor of, or inside it), which would mutate the
  source. The `.git` and overlap checks run on **symlink-resolved** paths (the
  deepest existing prefix is resolved and the remainder rejoined), so a
  destination reached through a symlinked parent that points into the source is
  caught — an adversarial-review finding — without refusing an ordinary symlinked
  ancestor like macOS's `/tmp`. It fails closed on any stat it cannot read as
  "absent".
- **Secret-scan before write.** The planned bytes are scanned in memory; a
  hard-fail refuses the whole pack. A secret is fixed at source, never redacted
  into the artefact. The scan is injected (`SecretScan`), so the lifeboat core
  stays free of the scanner adapter, and the scanner's fail-closed `Unavailable`
  state refuses rather than shipping under a weakened ruleset.
- **Contained, atomic write.** Every file is written into a fresh staging
  directory through `os.Root` (no crafted path or symlink escapes it), then the
  staging directory is renamed into place — a crash leaves staging, never a
  half-lifeboat. When `<dest>` already holds a prior lifeboat, it is renamed
  *aside* to a sibling backup before the swap and removed only on success, so a
  rename *error* (not just a crash) restores the prior lifeboat rather than
  destroying it — an adversarial-review finding. `_provenance.json` is written
  last: it is the commit marker and the gate key for a later re-pack. Every
  planned path is re-validated (relative, cleaned, no `..`, no control characters)
  before it is written.
- **The source is never touched.** `Plan` reads read-only; a test hashes the
  source tree before and after a pack.
- **Marker hygiene.** A verbatim record carrying an abcd marker block has it
  stripped (`ahoy.StripMarkerBlock`) before it travels, so embarking the lifeboat
  cannot plant a stale rules-loader. The strip happens inside `Plan`, so the
  manifest hash covers the bytes a pack actually writes.
- **The voyage ledger.** Each pack appends one line to
  `~/.abcd/voyage/<source-root-sha>/disembark/history.jsonl` — genuinely
  append-only, keyed on the source's root-commit SHA (a source without one is not
  logged rather than logged under a forged key), carrying the manifest hash that
  ties the line to the lifeboat's own provenance. Its top directories are verified
  real (not symlinks) before any is created, and the ledger is written through an
  `os.Root` anchored at the verified base so a swapped `<root-sha>`/`disembark`
  component cannot redirect the append outside it — adversarial-review findings. A
  voyage failure is reported, not fatal: the written lifeboat's `_provenance.json`
  is authoritative.

## Approach — the graveyard (M4)

"Extract what failed, so it isn't tried again." Three layers, strictly ordered so
interpretation can never float free of evidence:

- **Layer 1 — archaeology** (`graveyard/archaeology.json`): Tier-0 git only,
  deterministic, evidence only. Five signals, each with a stable namespaced id a
  later lesson can cite: reverted commits (`rev-<sha12>`), branches never merged
  into the default branch ranked by divergence age (`branch-<name>`), paths
  deleted after substantial history — sustained investment retired, not a scratch
  file swept — and absent at HEAD (`del-<path>`), dependencies present in a
  manifest's first revision but gone at HEAD (`dep-<manifest>`), and
  wholesale-rewrite commits that replace a large fraction of the tree in one
  non-merge commit (`rewrite-<sha12>`). Thresholds are named constants carrying
  their rationale; every signal is bounded and every human string sanitised.
- **Layer 2 — recorded abandonment** (`graveyard/abandoned.json`): what the
  project itself declared dead, keyed by the records' own ids so a lesson cites
  exactly what a reader sees — superseded intents (the `superseded/` bucket),
  superseded ADRs (frontmatter status or a non-null `superseded_by`, across the
  native and conventional ADR homes), each ADR's Alternatives-Considered options
  (`adr-N-alt`), wontfix issues with their recorded reasons, and decision-log
  lines naming a rejected option (`dec-L<line>`), matched by a deliberately
  narrow verb list so the signal stays conservative.
- **Layer 3 — interpretation** (`graveyard/lessons.json`): host-delegated, never
  packed. `abcd disembark graveyard <lifeboat-dir> --lessons-json <file|->`
  validates untrusted interpreter output against the packed layers behind the
  same trust guards as an intent verdict (size cap, symlink refusal, unknown
  fields refused, schema version gated) and enforces **cite-or-be-dropped**: a
  lesson whose evidence hits no live layer-1/2 id is dropped and the drop
  reported, never fatal; `confidence: low` lessons are quarantined under
  `graveyard/low-confidence/<id>.json` (the id shape itself is the
  path-traversal defence). The validator, not the model's good intentions, is
  the difference between a graveyard and a séance.

Both packed layers are always emitted — an empty `findings` array is the honest,
first-class statement that history or the record declares nothing dead — so the
lifeboat's file set and its pinned `manifest_sha256` stay deterministic. The
lessons files are deliberately **outside** `manifest_sha256`: the seal is pinned
at pack time over the deterministic extraction, and interpretation is a later,
mutable layer whose integrity is the per-entry citation rule, not the manifest.

## Approach — embark and the round-trip (M5)

`abcd embark probe <lifeboat-dir> [target]` and `abcd embark from <lifeboat-dir>
[target]` (target defaults to the working directory) unpack a lifeboat's record
into a repository. The lifeboat directory is untrusted input:

- **Gates before any read of substance:** the `_provenance.json` marker must
  parse; a provenance `schema_version` greater than this binary knows is refused
  with an upgrade message (itd-9's irreversible half); `VerifyManifest` re-hashes
  every file and compares against the pinned `manifest_sha256` (the post-pack
  layer-3 lessons files and `_provenance.json` itself are excluded by
  definition) — a tampered, missing, or extra file is fatal; every archived path
  is re-validated; reads are guarded (size caps, symlink refusal).
- **Only the record families embark.** The inverse mapping is table-driven:
  `docs/adrs/` → the native ADR home, `activity/issues/<state>/` → the issue
  ledger, `rescue/intents/<bucket>/` → the intent corpus (a bucket-less intent
  lands in `drafts/` — the one default that fabricates no lifecycle state), and
  `rescue/specs/<bucket>/` → the spec store. Identity- and git-derived files
  (`coverage.*`, `brief/`, `graveyard/`, `rescue/spine.md`) inform the report and
  are never written; an unknown lifeboat file is reported unmapped, never
  written.
- **Conflicts are one bulk report, and core writes nothing on refusal.** A
  target file with different bytes is a conflict; identical bytes are an
  idempotent skip, not a conflict. Any conflict refuses the whole embark before
  a single write — the core returns the conflict set and the surface renders it
  (transport-agnostic core; the design chapter's refusal-writes-a-report is
  corrected here).
- **Contained writes, two layers.** `os.Root` containment over the target plus
  independent lexical path validation, with durable bytes through the canonical
  atomic write — a bug in one layer is not an escape.
- **The marker, never the text.** Embark never copies lifeboat prose into
  `CLAUDE.md`; it re-injects the *current* ahoy marker block
  (`ahoy.EnsureMarker`), the boundary between restoring a record and planting an
  instruction implant.
- **The coverage report is the handoff.** The rendered result surfaces the
  lifeboat's blanks and their questions first — what a product thinker must
  answer — before the write summary.

The round-trip gate, all property-tested through the real stores: pack a
record-bearing source, embark into a fresh target, and `intent.Load`,
`spec.Load`, and `capture.List` on the target equal the source, ADRs are
byte-identical, and the target's `CLAUDE.md` carries the current marker block —
while the source tree is untouched end to end. Closure is two pinned
properties: **P1**, the record-derived sub-manifest — `RecordManifestSHA256`
over the ADR/issue/intent/spec/abandoned families, recorded in provenance as
`record_manifest_sha256` — is byte-identical across pack → embark → re-pack
(identity- and git-derived families are excluded *by design*: a fresh target
has a different name, root SHA, and history, and pretending otherwise would be
fiction); and **P2**, embarking a lifeboat into a byte-copy of its own source
changes nothing and a re-pack reproduces the exact original `manifest_sha256`.
The plan's original literal closure wording is amended by the 2026-07-16
decision-log entry.

## Approach — synthesis over the record (M6)

Interpretation over a packed lifeboat is host-delegated behind injected seams,
never executed by the binary — the shipped `memory.Distiller` and
`intent review ingest` disciplines, generalised to three post-pack verbs that
each run in one of two self-recorded modes:

- **`disembark principles <lifeboat>`** — deterministic mode distils
  `principles.json` from the packed ADRs' Decision/Consequences bullets,
  evidence-only, no interpretation; `--principles-json <file|->` ingests a
  `principle-distiller` agent's output under per-entry cite-or-be-dropped
  against the live record ids, graveyard finding ids, and packed paths.
- **`disembark press-release <lifeboat>`** — deterministic mode composes from
  the packed brief's own press-release page (or the spine, or an honest
  placeholder); a delegated composition citing nothing resolvable is a
  whole-document refusal, mirroring memory ingest's unattributable-page rule.
- **`disembark oracle <lifeboat> <source>`** — the audit verdict is the first
  Go home of the registered `{SHIP, NEEDS_WORK, MAJOR_RETHINK}` vocabulary.
  Deterministic mode scores mechanically — manifest verification failure is a
  `MAJOR_RETHINK` verdict input (never a fatal error), absent or degraded
  coverage or more blanks than grounded sections is `NEEDS_WORK`, else `SHIP` —
  and the binary stamps the attestation fields itself, so a delegated payload
  can never fabricate a manifest hash. The audit lands at
  `audit/oracle-<manifest12>.json`, keyed by the lifeboat's own manifest hash —
  no wall-clock ever enters a lifeboat artifact (amending the plan's `<ts>`
  wording; 2026-07-16 decision log).

All synthesis artifacts are the post-pack mutable layer: excluded from
`manifest_sha256` like the graveyard lessons, fully replaced by each re-run,
and each self-records its `mode` — `_provenance.json` is never mutated after
the pack (the second 2026-07-16 amendment: the commit marker stays immutable).
Untrusted payloads are read behind the intent-verdict guards (size cap, symlink
refusal, unknown fields refused, schema and mode gates), every rendered string
is sanitised, and the four agents ship itd-5's prompt infrastructure from zero:
`prompt_version` in the 0.x calibration band, `reads_untrusted_input: true`, a
registered `capability_scope`, and one injection-canary fixture each under
`agents/` — the canary text must survive only as inert quoted data.

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
and `rules` — read-only, no `/abcd:` command surface. `pack` is the user-facing
verb: it ships the `/abcd:disembark` command surface (`commands/abcd/disembark.md`)
and the surface-registry row flips from `staged` to `shipped` in the same change,
satisfying the `surface_coverage` record-lint rule (a `shipped` row without a
backing command file — or a `staged` row that has one — is a blocker).
`disembark graveyard` (M4) rides the same shipped surface: the command file's
graveyard section tells the host how to run the interpreter and hand its output
to the verb.

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
- *cite-or-be-dropped / dry-run cannot lie* — the graveyard validator (M4) drops
  any lesson citing no live layer-1/2 finding id, and a test asserts a
  zero-valid-refs lesson is dropped while its batch survives; the shared
  plan/pack code path is `lifeboat.Plan` (M3a), the single producer both
  `disembark plan` and `disembark pack` (M3b) run, so a dry-run cannot describe
  a pack a real pack would not perform.
- *Byte-identical source through a pack* — `Pack` reads the source only through
  the read-only `Plan`; a test hashes the source tree before and after a pack.
- *Never overwrite what abcd did not produce* — the destination safety gate; a
  test packs over a non-empty non-lifeboat directory and requires refusal, and
  packs over an existing lifeboat and requires success.

## Out of scope for this spec

The deliberately deferred items the plan lists — the hostile-lifeboat battery,
itd-35's Merkle chain, `--with-code` (itd-8), schema migrators beyond the
version stamp (itd-9), and backgrounded execution. All itd-88 milestones
(M0–M6) are delivered by this spec. The multi-agent oracle passes and the aspirational
output tree in the older [`02-disembark.md`](../../brief/04-surfaces/02-disembark.md)
chapter are superseded by adr-35; that chapter's full rewrite is a follow-up.
