# `/abcd:launch` — Curated Release

> **Status:** design target — builds in Phase 5 (round-trip and ship). Today only the probe / dry-run stubs ship (fn-17); the full pre-flight gate suite and release flow below are not yet built.

## Sub-verbs

Bare `/abcd:launch` shows status + help only — never mutates state. Current sub-verbs:

- **`/abcd:launch ship`** — cut a curated release artefact from the one repo: run pre-flight gates, filter the artefact (default-deny, `.abcd/**` excluded by packaging), stamp the version, and on a `v*` tag publish a GitHub Release ([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)). The flow described in §§ 1–6 below is this sub-verb's behaviour. Flag-shaped modifiers: `--version <x.y.z>`, `--allow-dirty`, `--allow-doc-warnings`.
- **`/abcd:launch dry-run`** — **report-only preview, always exit-0** (a preview never blocks). It runs the parts of the pre-flight suite that exist today: as of fn-64 the **secret + PII scan gate** (the native scanners, see [§ 1](#1-pre-flight-gates)) runs for real in report-only mode and prints what it *would* refuse on (a finding, or a fail-closed reason such as "scanner unavailable"); the remaining gates (marker-block, `plugin.json` parse, documentation-auditor) are Phase-5 and render as "(not yet implemented)". It also produces the would-be artefact manifest, without writing the release artefact. dry-run is **not** "ship minus publish": running the *full* gate suite and **hard-failing** on a finding (exit non-zero) is the Phase-5 `ship` verb's behaviour, not dry-run's.

## 1. Pre-flight gates

- **Secret scan** — a **native Go scanner** is the default, hard-fail (absent/fail-closed, never a silent skip). **gitleaks** is an opt-in deeper scanner (the fn-64 ship gate pins `gitleaks >= 8.18.0` when wired; absent/older = fail-closed, never a regex fallback).
- **PII** scan (real names, emails) via the **native Go PII engine** (`scan_text` + the merged Config + the non-overridable secret/identity severity floor) — hard-fail.
- **Custom regex** layer — home dirs (`~/...`), GitHub usernames from git config — hard-fail
- **TruffleHog** — opt-in deep scan when `scan.deep=true` — hard-fail (live credential verification)
- **Hook compliance** check — warn-fail
- **Marker block sanity** — hard-fail on malformed
- **`plugin.json` parse + `marketplace.json` references** — hard-fail
- **Dirty tree** — refuse unless `--allow-dirty`
- **OWASP / vulnerability check** (folded into the pre-flight suite) — warn-fail
- **Documentation auditor** (subagent) — runs over `docs/` to verify user-facing documentation is well-formed before release — warn-fail

Pre-flight report written to `.abcd/logbook/launch/<timestamp>/preflight.{json,md}` — **Phase-5 `ship` behaviour**. The fn-64 secret/PII gate is itself side-effect-free w.r.t. the repo (its only writes are to a private temp tree it removes), and `dry-run` renders the gate result inline rather than writing a report file.

## 2. Curated release artefact (default-deny)

- **Include:** `.claude-plugin/` (holds both `plugin.json` and the ONE canonical `marketplace.json` — there is no root-level `marketplace.json`), `commands/`, `skills/`, `agents/`, `scripts/`, `hooks/`, `README.md`, `LICENSE`, `.gitignore`, `docs/` (user-facing only)
- **Exclude:** `.work/`, `.abcd/` (entire namespace — `development/` (brief, roadmap, research, activity, voyage, personas), `memory/`, `lifeboat/`, `logbook/`), and patterns from `.gitignore`. Per [adr-28](../../decisions/adrs/0028-single-repo-curated-release.md) the wholesale `.abcd/` exclusion (incl. `.abcd/memory/**`) is a **packaging filter over the one tree**, not a copy between two repos: the release artefact carries plugin code, never the project's design record or knowledge store. The fn-38 restrictive-licence gate is NOT this artefact's gate — its real consumer is the lifeboat (`/abcd:disembark`), the surface that publishes curated project memory/provenance (adr-4). At launch the gate is future/inert; `/abcd:launch dry-run` renders its verdicts only as a diagnostic preview.
- **Override:** `.abcd/launch.allow` allowlist — the packaging override (the only mechanism that can put a path *into* the release artefact). Per adr-28 it must **never** promote any `.abcd/**` path (a `.abcd/**` line is refused / never promoted), so it cannot re-include `.abcd/memory/**`. This is a **distinct** mechanism from the gate's JSON `.abcd/launch-allowlist.json` (`_ALLOWLIST_REL`), which only re-includes files into the fn-38 gate's *own evaluation input*, never into the release artefact — the two are documented-distinct, never one name.

## 3. Versioning + marketplace

The curated release artefact is the only abcd artefact that carries a semantic
version. Versions are an *output* of cutting a release, never a sequencing input
on the design record — the repo organises work by **phase** (see
[adr-9](../../decisions/adrs/0009-phase-as-product-layer.md)), and a release
number is what falls out when a stretch of that work is published. The brief,
intents, and roadmap carry no version label.

Versioning is **strict SemVer**: the version string is `MAJOR.MINOR.PATCH` (no
leading `v`) at the selected version location, and the release's git tag is
`v<version>`. While the version is `0.y.z` the operator surface may still change
between minor versions (pre-1.0 = not yet surface-stable); `1.0.0` marks the
first stable `/abcd:*` surface and is a manual, maintainer-declared bump like
every other major. The tag drives identification: what is installed vs what is
available is compared through the tagged, released artefact — the working tree
carries no version at all (adr-19, adr-28).

### Bump-tier rule

`launch ship` selects the bump tier automatically; the tier is recorded in the
launch report and the commit message so every published version is traceable
to *why* it bumped.

| Tier | Trigger | Selected by |
|---|---|---|
| **Patch** (`v0.0.x`) | Any release cut that is not a phase completion — a routine snapshot of in-progress work. | Default — every `ship` that isn't one of the rows below. |
| **Minor** (`v0.x.0`) | A phase completed since the last release: every spec carrying that phase's `phase:` anchor is closed, and the phase's `## Phase Acceptance` has been reviewed. | Auto-detected (see below). |
| **Major** (`vx.0.0`) | A deliberate milestone the maintainer declares. | **Manual only** — `launch ship --version <x.0.0>`. Never auto-selected. |

**Phase-completion detection.** Before choosing a tier, `launch ship` reads
`.abcd/development/roadmap/phases/` and, for each phase, checks whether every
spec anchored to it (`phase:` frontmatter, per adr-9) is closed in the native
spec store. The `phase:` anchor is deferred today (see [`roadmap/phases/README.md`](../../roadmap/phases/README.md));
it activates with this launch surface, so until then the phase set this
detection reads is the editorial `## Scope` membership, not a frontmatter field. If
a phase became fully-closed since the version recorded in the last
`launch-report.json`, the bump is **minor** and the report names the completed
phase. If more than one phase completed since the last launch, it is still a
single minor bump; the report lists all of them. If none did, the bump is
**patch**.

`--version <x.y.z>` overrides the auto-selected tier entirely (the only way to
trigger a major bump, and an escape hatch if detection is wrong).

`launch ship` is responsible for writing the version into **the selected
version location**, never a hard-coded `plugin.json`. That location is recorded
by the fn-77.1 decision artifact
(`.abcd/config/version-location.json`)
as `manifest_path` + `json_pointer` (see
[adr-19](../../decisions/adrs/0019-plugin-json-version-carve-out.md)); a
`blocked: true` decision has no schema-valid location, so version-writing
refuses and the escalation stands. Concretely, `ship`:

1. Stamps the bumped version into the **release artefact** at the selected
   `manifest_path` + `json_pointer` — the manifest renderer reads the decision
   artifact and never parses a location string.
2. Leaves the **working-tree** manifests UNVERSIONED. Per
   [adr-19](../../decisions/adrs/0019-plugin-json-version-carve-out.md) and
   [adr-28](../../decisions/adrs/0028-single-repo-curated-release.md) the version
   is single-sourced in the *cut artefact*; the repo's committed manifests carry
   no version, so there is nothing to keep in sync on the working-tree side. The
   renderer stamps the version into the artefact content only — it never mutates
   the working-tree manifests.
3. Records the version + changelog entry in the marketplace metadata at the ONE
   canonical `.claude-plugin/marketplace.json` (never a root-level copy). The
   changelog entry conforms to
   `changelog-entry.schema.json`
   (validated programmatically by this bump step, per
   [adr-20](../../decisions/adrs/0020-manifest-version-lockstep.md)).
4. Refreshes any other version references generated from the config slug.

**Anti-drift (present state).** The two manifests in the artefact describe one
release, so they must stay version-consistent: the version at the selected
location and the marketplace entry's version + changelog must agree. A read-only
lockstep checker proves this over the pinned path list
[adr-20](../../decisions/adrs/0020-manifest-version-lockstep.md) records; a
half-state (a version in one manifest and not the other) is drift. `ship`'s bump
step runs it against the staged release artefact and refuses to publish on drift.
The checker has no bypass flag, and adr-20 records that `--allow-dirty` must not
bypass manifest consistency (wiring policy, enforced by the pre-flight suite).

Commit message: `chore(release): launch abcd v<version> (<tier>: <reason>) from <source-sha>`
— e.g. `chore(release): launch abcd v0.3.0 (minor: phase-2-ahoy complete) from a1b2c3d`.

### Release cut + retention

Every `launch ship` **cuts a release**: the release commit, the `v<version>`
git tag on it, the marketplace changelog entry, and — on the `v*` tag — a
published GitHub Release with **SLSA provenance** attached to the artefact
([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)) all describe
one released snapshot. The version lives only on the tag and in the cut artefact;
the working tree is never versioned (adr-19, adr-28).

Retention is **newest-per-line**: each release line (`MAJOR.MINOR`) keeps only
its newest release. Shipping `v0.1.2` removes the superseded `v0.1.1` — its git
tag and the GitHub Release and its assets — while shipping `v0.2.0` keeps the
last `v0.1.x` alongside it (the last release of every previous line survives as
that line's terminal snapshot). Three safety rules bound the prune:

- The release just published is **never** pruned.
- Pruning **refuses** if a release newer than the one just published already
  exists (out-of-order ship — resolve manually, never auto-delete forward).
- Retention prunes **release tags and Releases only**; git history is untouched.
  The launch report under `.abcd/logbook/launch/<timestamp>/` is the durable
  record of every launch, including pruned ones — deleting a release tag never
  deletes the evidence a launch happened.

A prune is a destructive, outward-visible act, so `ship` reports exactly which
release it removed (or why it refused) in the launch report and the ship
transcript; a `--dry-run`-shaped preview of the prune decision is part of the
`dry-run` artefact preview.

## 4. Reports

`launch-report.{json,md}` in the repo's `.abcd/logbook/launch/<timestamp>/`.

## 5. Bootstrap exception

The first release cut of abcd itself is a manual `v*` tag + GitHub Release; document in `commands/abcd/launch.md`.

## 6. Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:launch`, **then** the dispatcher shows current launch readiness (pre-flight gate state, last launch attempt timestamp), the available sub-verbs (`ship`, `dry-run`), and suggested next actions — bare invocation never mutates state.
- **Given** a clean tree with a deliberate PII fixture (e.g., a real email in a comment) inside the resolved artefact, **when** `/abcd:launch dry-run` runs, **then** the report-only gate (fn-64) PRINTS that it *would* refuse on that finding (the offending file/line in the gate result), still **exits 0**, and writes no artefact. (The **hard-fail** on that finding — exit non-zero plus a `preflight.{json,md}` report under `.abcd/logbook/launch/<timestamp>/` — is the Phase-5 `ship` verb's behaviour, not dry-run's.)
- **Given** a clean tree, **when** `/abcd:launch dry-run` runs, **then** the report lists exactly the include/exclude artefact manifest in [§ 2](#2-curated-release-artefact-default-deny) with no surprises and no artefact is written.
- **Given** no phase completed since the last launch, **when** `launch ship` runs without `--version`, **then** the bump tier is **patch** (`v0.0.x`) and the next patch version is written into the **selected version location** (from `.abcd/config/version-location.json`, per [§ 3](#3-versioning--marketplace)) in the **release artefact** only — the working-tree manifests stay unversioned (adr-19, adr-28) — plus the canonical `.claude-plugin/marketplace.json`, never a hard-coded `plugin.json`.
- **Given** every spec anchored to a phase (`phase:` frontmatter) is closed and that phase was not yet complete at the last recorded launch, **when** `launch ship` runs without `--version`, **then** the bump tier is **minor** (`v0.x.0`) and the launch report names the completed phase.
- **Given** two or more phases completed since the last launch, **when** `launch ship` runs without `--version`, **then** a single **minor** bump is applied and the report lists every completed phase.
- **Given** the user passes `launch ship --version <x.y.z>`, **when** the command runs, **then** the explicit version overrides the auto-selected tier — this is the only path to a major (`vx.0.0`) bump.
- **Given** any `launch ship` run, **when** the release commit is written, **then** the commit message records the bump tier and its reason (e.g. `(minor: phase-2-ahoy complete)`).
- **Given** a documentation-auditor warn-fail, **when** `launch ship` runs without `--allow-doc-warnings`, **then** the user is shown the warnings and asked transparently whether to proceed.
- **Given** a prior release of the same line (`vX.Y.(Z-1)`) exists, **when** `launch ship` publishes `vX.Y.Z`, **then** the superseded release's tag and GitHub Release + assets are removed, the removal is named in the launch report, and the last release of every *other* line is untouched.
- **Given** a release newer than the just-published version already exists, **when** the retention step runs, **then** it refuses to prune anything and the launch report records the refusal reason.
