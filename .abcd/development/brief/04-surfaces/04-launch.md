# `/abcd:launch` — Curated Release

> **Phase ownership** ([adr-33](../../decisions/adrs/0033-launch-phase-ownership-tiered.md)): the curated-release cut — packaging with `.abcd/**` excluded plus the secret/PII scan — ships in [Phase 1](../../roadmap/phases/phase-1-ahoy.md). The full pre-flight gate suite and release automation below are separately scheduled intents (itd-65 gate suite, itd-66 render parity, itd-70 retention, itd-72 tier-b publishing, itd-73 derived versioning).

## Sub-verbs

The shipped verb surface is the `--dry-run` flag on `abcd launch` — a read-only preview of the bundle and gates. Bare `abcd launch` never mutates state: it refuses (exit 1) with a hint to pass `--dry-run`; a bare-as-status render is a design target, unshipped. The sub-verb design:

- **`/abcd:launch ship`** — **design target (itd-65 gate suite + itd-72 publishing); no `ship` verb is on any shipped surface** — `internal/core/launch` carries the `Ship` engine, unwired. The design: cut a curated release artefact from the one repo: run pre-flight gates, filter the artefact (default-deny, `.abcd/**` excluded by packaging), stamp the version, and on a `v*` tag publish a GitHub Release ([adr-28](../../decisions/adrs/0028-single-repo-curated-release.md)). The flow described in §§ 1–6 below is this sub-verb's behaviour. Flag-shaped modifiers `--allow-dirty` and `--allow-doc-warnings` belong to this sub-verb's design; the shipped verb accepts only `--dry-run` and the global `--json`. There is no version flag — the version is derived, never authored ([adr-31](../../decisions/adrs/0031-derived-versioning-from-intents.md), see [§ 3](#3-versioning--marketplace)).
- **`/abcd:launch dry-run`** — shipped as the `--dry-run` flag (the plugin command `commands/abcd/launch.md` maps its `[dry-run]` argument hint onto `abcd launch --dry-run`; the binary has no `dry-run` subcommand). **Report-only preview, always exit-0** (a preview never blocks). It runs the parts of the pre-flight suite that exist today: as of spc-64 the **secret + PII scan gate** (the native scanners, see [§ 1](#1-pre-flight-gates)) runs for real in report-only mode and prints what it *would* refuse on (a finding, or a fail-closed reason such as "scanner unavailable"); the remaining gates (marker-block, `plugin.json` parse, documentation-auditor) are the gate-suite intent's (itd-65) and render as "(not yet implemented)". It also produces the would-be artefact manifest, without writing the release artefact. dry-run is **not** "ship minus publish": running the *full* gate suite and **hard-failing** on a finding (exit non-zero) is the full `ship` verb's behaviour (itd-65 + itd-72), not dry-run's.

## 1. Pre-flight gates

- **Secret scan** — a **native Go scanner** is the default, hard-fail (absent/fail-closed, never a silent skip). **gitleaks** is an opt-in deeper scanner (the spc-64 ship gate pins `gitleaks >= 8.18.0` when wired; absent/older = fail-closed, never a regex fallback).
- **PII** scan (real names, emails) via the **native Go PII engine** (`scan_text` + the merged Config + the non-overridable secret/identity severity floor) — hard-fail.
- **Custom regex** layer — home dirs (`~/...`) and local usernames — hard-fail; GitHub usernames from git config — warn (legitimate in repo-URL contexts). The per-kind severity floor is config-raisable, never lowerable (`internal/adapter/scanner/identity.go`).
- **TruffleHog** — opt-in deep scan when `scan.deep=true` — hard-fail (live credential verification)
- **Hook compliance** check — warn-fail
- **Marker block sanity** — hard-fail on malformed
- **`plugin.json` parse + `marketplace.json` references** — hard-fail
- **Dirty tree** — refuse unless `--allow-dirty`
- **OWASP / vulnerability check** (folded into the pre-flight suite) — warn-fail
- **Documentation auditor** (subagent) — runs over `docs/` to verify user-facing documentation is well-formed before release — warn-fail

Pre-flight report written to `.abcd/logbook/launch/<timestamp>/preflight.{json,md}` — **full-`ship` behaviour (itd-65)**. The spc-64 secret/PII gate is itself side-effect-free w.r.t. the repo (its only writes are to a private temp tree it removes), and `dry-run` renders the gate result inline rather than writing a report file.

## 2. Curated release artefact (default-deny)

- **Include:** the shipped include list, pinned in `.abcd/config/launch-payload.json`: `.claude-plugin/` (holds both `plugin.json` and the ONE canonical `marketplace.json` — there is no root-level `marketplace.json`), `commands/`, `scripts/`, `docs/` (user-facing only), `README.md`, `LICENSE`, `.gitignore`. `skills/`, `agents/`, and `hooks/` are design-target additions absent from the shipped list (`skills/` exists in the tree but is not packaged; `agents/` and `hooks/` have no tree presence).
- **Exclude:** `.abcd/` (entire namespace — `development/` (brief, roadmap, research, voyage, personas), `memory/`, `lifeboat/`, `logbook/`), and patterns from `.gitignore`. Per [adr-28](../../decisions/adrs/0028-single-repo-curated-release.md) the wholesale `.abcd/` exclusion (incl. `.abcd/memory/**`) is a **packaging filter over the one tree**, not a copy between two repos: the release artefact carries plugin code, never the project's design record or knowledge store. The spc-38 restrictive-licence gate is NOT this artefact's gate — its real consumer is the lifeboat (`/abcd:disembark`), the surface that publishes curated project memory/provenance (adr-4). At launch the gate is future/inert; the shipped `dry-run` renders no licence verdicts (a diagnostic preview of the gate's verdicts in `dry-run` is part of the gate's own design, unshipped).
- **Override:** the include list in `.abcd/config/launch-payload.json` is the packaging override (the only mechanism that can put a path *into* the release artefact). The deny is **structural** (`internal/core/launch/bundle.go`, per adr-28): no include entry can promote a denied namespace — a `.abcd/**` line is never promoted — so nothing can re-include `.abcd/memory/**`. The spc-38 gate's own evaluation-input allowlist belongs to that gate's design (future/inert, see above) and is documented-distinct from packaging: it re-includes files into the gate's *own evaluation input*, never into the release artefact — two mechanisms, never one name.

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
first stable `/abcd:*` surface. The tag drives identification: what is installed
vs what is available is compared through the tagged, released artefact — the
working tree carries no version at all (adr-19, adr-28).

### Bump-tier rule

The version is **derived, never authored**
([adr-31](../../decisions/adrs/0031-derived-versioning-from-intents.md)):
`launch ship` selects the bump tier from the intents shipped since the previous
release; the tier and its reason are recorded in the launch report and the
commit message so every published version is traceable to *why* it bumped.

| Tier | Trigger |
|---|---|
| **Major** (`vx.0.0`) | Any shipped intent since the last release carries `impact: breaking`. |
| **Minor** (`v0.x.0`) | No `breaking`, at least one `impact: additive`. |
| **Patch** (`v0.0.x`) | Only `impact: fix` intents (or a release with no intent-tied change). |

**Impact derivation.** Every intent carries `impact: additive | breaking | fix`,
set when the intent is shaped and enforced by `internal/core/lint` (adr-31, tracked
by itd-73). At release, `launch ship` gathers the intents shipped since the
previous release and takes the highest-severity impact. A change not tied to any
intent falls back to conventional-commit derivation (`feat:` / `fix:` /
`feat!:` prefixes since the last tag).

**Surface-diff guardrail.** `launch ship` snapshots the `/abcd:*` command, flag,
and manifest surface and compares it to the previous release. A removed or
altered surface with no `breaking` intent in the release **fails the launch** —
a mislabelled impact cannot ship a compatibility lie.

`launch ship` is responsible for writing the version into **the selected
version location**, never a hard-coded `plugin.json`. That location is read from
the spc-77.1 decision artifact
(`.abcd/config/version-location.json`)
as `manifest_path` + `json_pointer` (see
[adr-19](../../decisions/adrs/0019-plugin-json-version-carve-out.md)). The
artifact is absent from the tree, and the shipped lockstep checker fails closed
on it — `dry-run` reports the lockstep contract unreadable — until spc-77.1
records the decision; a
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
— e.g. `chore(release): launch abcd v0.3.0 (minor: additive itd-40 shipped) from a1b2c3d`.

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

The first release cut of abcd itself is a manual `v*` tag + GitHub Release. Documenting the exception in `commands/abcd/launch.md` lands with publishing (itd-72); the shipped command file covers only the read-only dry-run preview.

## 6. Acceptance

- **Given** any abcd-aware terminal, **when** the user runs bare `/abcd:launch`, **then** the dispatcher shows current launch readiness (pre-flight gate state, last launch attempt timestamp), the available sub-verbs (`ship`, `dry-run`), and suggested next actions — bare invocation never mutates state. **Design target:** the shipped bare `abcd launch` refuses (exit 1) with a hint to pass `--dry-run`, and the shipped plugin command runs the dry-run preview directly rather than a status+help render.
- **Given** a clean tree with a deliberate PII fixture (e.g., a real email in a comment) inside the resolved artefact, **when** `/abcd:launch dry-run` runs, **then** the report-only gate (spc-64) PRINTS that it *would* refuse on that finding (the offending file/line in the gate result), still **exits 0**, and writes no artefact. (The **hard-fail** on that finding — exit non-zero plus a `preflight.{json,md}` report under `.abcd/logbook/launch/<timestamp>/` — is the full `ship` verb's behaviour (itd-65), not dry-run's.)
- **Given** a clean tree, **when** `/abcd:launch dry-run` runs, **then** the report lists exactly the include/exclude artefact manifest in [§ 2](#2-curated-release-artefact-default-deny) with no surprises and no artefact is written.
- **Given** only `impact: fix` intents (or no intent-tied change) shipped since the last release, **when** `launch ship` runs, **then** the bump tier is **patch** (`v0.0.x`) and the next patch version is written into the **selected version location** (from `.abcd/config/version-location.json`, per [§ 3](#3-versioning--marketplace)) in the **release artefact** only — the working-tree manifests stay unversioned (adr-19, adr-28) — plus the canonical `.claude-plugin/marketplace.json`, never a hard-coded `plugin.json`.
- **Given** at least one `impact: additive` intent and no `breaking` intent shipped since the last release, **when** `launch ship` runs, **then** the bump tier is **minor** (`v0.x.0`) and the launch report names the intents that drove it.
- **Given** any `impact: breaking` intent shipped since the last release, **when** `launch ship` runs, **then** the bump tier is **major** (`vx.0.0`) and the launch report names the breaking intent(s).
- **Given** a command, flag, or manifest surface removed or altered since the previous release with no `breaking` intent in the release, **when** `launch ship` runs, **then** the surface-diff guardrail **fails the launch** (adr-31) — the mislabel is reported, nothing is published.
- **Given** any `launch ship` run, **when** the release commit is written, **then** the commit message records the bump tier and its reason (e.g. `(minor: additive itd-40 shipped)`).
- **Given** a documentation-auditor warn-fail, **when** `launch ship` runs without `--allow-doc-warnings`, **then** the user is shown the warnings and asked transparently whether to proceed.
- **Given** a prior release of the same line (`vX.Y.(Z-1)`) exists, **when** `launch ship` publishes `vX.Y.Z`, **then** the superseded release's tag and GitHub Release + assets are removed, the removal is named in the launch report, and the last release of every *other* line is untouched.
- **Given** a release newer than the just-published version already exists, **when** the retention step runs, **then** it refuses to prune anything and the launch report records the refusal reason.
