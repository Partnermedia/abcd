# Changelog

All notable changes to abcd are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and abcd
uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html) with a
leading `v`.

Before v1.0.0, minor releases may make breaking changes; each one is
called out in a **Breaking** section.

## [Unreleased]

### Added

- `abcd capture --blocked-by <iss-N,…>` records typed dependency edges on a new
  issue, and `capture list` / the status board now render a derived-priority
  view: unblocked issues first, then by severity, with blocked rows annotated
  `[blocked-by iss-N,…]`. There is no stored priority — the ordering is a
  read-time projection, so resolving a blocker re-prioritises its dependents
  automatically.
- A store-contract README for the issue ledger (`.abcd/work/issues/README.md`).

### Changed

- The issue ledger moved from `.abcd/development/activity/issues` to
  `.abcd/work/issues` (the committed shared-working tier).

### Removed

- The `created` and `updated` frontmatter fields on issues. Git is the canonical
  source of an issue's timeline; the ledger no longer duplicates it.

## [v0.1.0] - 2026-07-07

First tagged milestone: the Go rebuild through Phase 2. abcd is a single,
host-agnostic Go binary that is also a plugin for compatible agent harnesses, holding all
behaviour in a transport-agnostic `internal/core` behind a Cobra CLI front door and
a markdown plugin surface that shells out to it.

### Added

- Phase 0 scaffold: Go module (`github.com/REPPL/abcd-cli`), a
  transport-agnostic `internal/core`, a Cobra CLI front door (`abcd` status
  board and `abcd version`), the plugin surface, and the design record carried
  forward as the build specification.
- Phase 1 — install and launch. `abcd ahoy` installs abcd into a repo
  (folder-kind detection, visibility-driven gitignore, idempotent marker blocks in
  CLAUDE.md/AGENTS.md). `abcd launch --dry-run` renders a curated release bundle
  that excludes `.abcd/**` by default-deny, running a native secret + PII scanner,
  strict SemVer, marketplace-lockstep anti-drift, and newest-per-line retention over
  the bundle.
- Phase 2 — native capture substrates. `abcd history` is a SHA-keyed, redacted,
  gitignored transcript store (`list`, `show`, and a fail-closed `capture` write
  path); `abcd capture` is a directory-as-status issue ledger; `abcd memory`
  provides deterministic ingest / ask / lint.
- `abcd docs lint` (itd-60 layer 1) — a deterministic docs-currency gate over
  `docs/` and the repo root: change-narration in a doc body, a broken relative
  link, or a stray top-level markdown file each fails the gate.
- `record-lint` — a deterministic drift gate for the `.abcd/development` design
  record (banned tokens, git-metadata, link resolution, intent lifecycle), wired
  blocking into CI and the pre-push preflight.
- Derived-versioning design record (intent itd-73 + ADR-31): the release version
  is derived from intents' declared impact, never hand-authored. The derivation
  itself lands in a later phase.
