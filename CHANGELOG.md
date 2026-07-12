# Changelog

All notable changes to abcd are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and abcd
uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html) with a
leading `v`.

Before v1.0.0, minor releases may make breaking changes; each one is
called out in a **Breaking** section.

## [Unreleased]

### Added

- **Intent-fidelity review** (itd-80): the ship move now emits a report-only
  fidelity-review receipt, and `abcd intent review ingest --verdict-json <path>`
  applies the host-produced verdict back onto the record. When `abcd spec close`
  ships a linked intent (`planned/ → shipped/`), it parks a deterministic OWED
  receipt marker in the intent's `## Audit Notes` and writes an ephemeral review
  request under `.abcd/.work.local/reviews/` (gitignored); the emit is
  non-fatal, so a failure never un-ships the intent. `abcd intent review ingest`
  validates an untrusted intent-fidelity verdict JSON fail-closed (schema,
  in-enum verdicts, cited evidence, and each `criterion_id` bound to an actual
  Acceptance-Criteria bullet), then either replaces the OWED stub with the
  rendered per-criterion verdicts and honoured/diverged/missing audit
  (`INGESTED`, idempotent — a re-ingest is a no-op) or quarantines a bad payload
  (`DEAD_LETTER`: all criteria `INCONCLUSIVE`, raw payload retained) — never a
  partial application. Bare `abcd intent review <itd-N>` re-emits a shipped
  intent's request. The single source of truth is the intent file's Audit Notes;
  there is no side receipt store.

- The **intent lifecycle** verbs `abcd intent` and `abcd spec` (itd-80), the
  front doors onto the native intent store (`internal/core/intent`). Bare
  `abcd intent` renders a read-only lifecycle summary (intent counts by bucket,
  spec counts by status, and the linked intent↔spec pairs); `abcd intent plan
  <itd-N>` mints a native spec for a draft intent that carries a non-empty
  `## Acceptance Criteria` section (the itd-1 gate), writes both sides of the
  bidirectional link (the spec's `intent: itd-N` and the intent's
  `spec_id: spc-N` plus a default `kind: standalone`), and moves the intent
  `drafts/ → planned/` — fail-closed, so every intermediate on-disk state stays
  valid under the `intent_lifecycle` record-lint rule. `abcd intent link <itd-N>
  <spc-N>` retroactively links a planned intent to an existing spec, refusing a
  spec that realises a different intent. Bare `abcd spec` renders the spec-store
  status; `abcd spec close <spc-N>` moves a spec `open/ → closed/` (the
  lifecycle reconcile that trails a close lands in a later phase). The
  frontmatter line-scanner shared by these stores now lives in
  `internal/core/frontmatter`.
- The **modular rules loader** core and its `abcd rules [domain]` verb (itd-3,
  phases 1 + 3). `internal/core/rules` holds binary-bundled default rule domains
  (COMMITTING, DOCUMENTATION, ROADMAP, ISSUES, INTENTS, LIFEBOAT, PII, and
  OPINIONS — whose rules point at the canonical conventions under
  `.abcd/development/principles/` rather than copying them) merged
  with an optional per-repo `.abcd/rules.json` override (per-field domain
  override, sticky kill switch), with word-bounded recall matching (including a
  conservative suffix stemmer so `commits`/`issues` recall their keyword),
  `*<DOMAIN>` star-commands, and per-domain dedup signatures. Bare `abcd rules` renders the
  active rule set; a positional `DOMAIN` (case-insensitive) scopes to one; a
  malformed `rules.json` fails closed. A Claude Code prompt-router hook
  (`abcd hook prompt-router` / `prompt-router-reset`, operator-internal) injects
  the matched rules just-in-time on `UserPromptSubmit` with per-session
  signature dedup, clears the ledger on a `SessionStart`/`PreCompact` reset
  (event-driven refresh; a large fixed-N counter is only a backstop), and is
  fail-closed and non-blocking — a malformed payload, unreadable `rules.json`,
  or state error injects nothing and logs out-of-band, never wedging a session.
  The `hooks/hooks.json` manifest wiring lands with ahoy in the next phase.
- A `surface_coverage` record-lint rule (iss-35): the deterministic half of the
  brief↔surface cross-check. It reads the plugin surface
  (`rules.surface_coverage.commands_dir`, `skills_dir` — outside the lint roots)
  and the brief's surface registry table (`rules.surface_coverage.registry`, by
  convention `.abcd/development/brief/04-surfaces/README.md`), and asserts three
  invariants: every real surface has a registry row; every row marked `shipped`
  in the registry's **Status** column has a backing surface while every `staged`
  row (a design target) has none; and every row's status is `shipped` or
  `staged`. The bare `/abcd` top-level is binary-backed and exempt from the file
  check. Chapter-link resolution stays with `links_resolve`; the semantic half —
  each row's prose vs. binary behaviour — stays a release-gate agent check.
- A managed-repo **git-identity gate** (iss-62): a repo can pin its expected
  commit identity in `.abcd/config/identity.json`, and every commit is checked
  against it. `ahoy doctor` reports a divergence (a repo-local override that
  differs from the pin, or an unset identity) or an un-pinned repo; `ahoy
  install` adopts the gate by pinning the current git identity; `ahoy
  identity-check` exits non-zero on a mismatch; and the `pre-commit` hook
  fail-closes so a stray identity (e.g. a sandbox default) is caught at commit
  time rather than discovered later. A repo with no pin is unaffected.
- A `context_status_free` record-lint rule: the shared orientation file
  (`rules.context_status_free.target`, by convention `.abcd/work/CONTEXT.md`)
  must carry no phase/status claims — status is read live from the CLI and
  the ledger, never hand-written into orientation docs. Patterns are
  configurable (`rules.context_status_free.patterns`) with sensible defaults;
  lines matching inside fenced code blocks are skipped.

- A `/abcd:prepare-this-repo` command — audits the current repository against
  the abcd record and adopts the three-tier `.abcd/` layout, a marked
  working-conventions section in `AGENTS.md`, and the commit gates; an interim
  bridge until repos are managed directly. Owned repos only (it refuses
  elsewhere), and it migrates the older root-level `.work/` scaffold layout
  with explicit sign-off.
- `/abcd:consult` and `/abcd:ingest` commands — consult the user-level sources
  corpus (confidential entries are never cited or named in public artifacts)
  and ingest a URL or document into it with extracted reference metadata,
  keywords, and a text-quality check. Both are thin fronts on the corpus's own
  tooling and stop gracefully when no corpus exists.
- A `persona_registry` record-lint rule: press-release quote attributions
  (`said <Name>,`) must name a persona from the registry file the rule's
  `registry` key points at; unknown names are blocker findings. Configured
  per repo in `record-lint.json`; the historical record is skipped via the
  standard content-drift exemptions.
- `abcd capture --blocked-by <iss-N,…>` records typed dependency edges on a new
  issue, and `capture list` / the status board now render a derived-priority
  view: unblocked issues first, then by severity, with blocked rows annotated
  `[blocked-by iss-N,…]`. There is no stored priority — the ordering is a
  read-time projection, so resolving a blocker re-prioritises its dependents
  automatically.
- A store-contract README for the issue ledger (`.abcd/work/issues/README.md`).

### Changed

- `abcd spec close <spc-N>` now reconciles the linked intent (itd-80): it moves
  the intent `planned/ → shipped/` and then closes the spec, so one command
  completes the lifecycle transition. It is fail-closed (a missing/empty intent
  link, a non-existent or ambiguously-linked intent, bidirectional drift, or an
  intent in an unexpected bucket refuses with no partial move) and idempotent (a
  re-run on an already-shipped intent / already-closed spec is a clean no-op).
  The intent's `## Audit Notes` are left untouched. A new `spec_lifecycle`
  record-lint rule mirrors `intent_lifecycle` on the spec side: every spec under
  `specs/{open,closed}/` must carry a well-formed `id`/`slug`/`intent` link whose
  named intent EXISTS and points back at this spec (bidirectional agreement).
- The issue ledger moved from `.abcd/development/activity/issues` to
  `.abcd/work/issues` (the committed shared-working tier).

### Removed

- The `created` and `updated` frontmatter fields on issues. Git is the canonical
  source of an issue's timeline; the ledger no longer duplicates it.

### Fixed

- `abcd` status now reports `IsGitRepo` correctly in a linked git worktree or a
  submodule, where `.git` is a regular gitfile rather than a directory (iss-72).
- `abcd intent plan` now refuses an `## Acceptance Criteria` section with no
  top-level `-`/`*` bullet, matching the ingest gate — an intent can no longer be
  planned into a state where every fidelity verdict dead-letters for having zero
  positional criteria. The intent template's Audit Notes placeholder is cleared
  when the first review block lands, so a populated audit carries no stale "Empty"
  claim (iss-67).

### Security

- **Release receipt-gate hardening** (iss-70). The `receipt_gate` record-lint
  rule now binds each semantic-pass receipt to the gate it attests: a receipt
  satisfies a required gate only when its `policy.detector` equals that gate name,
  not merely when a `<gate>.json` file exists. This closes a hole where one
  genuine PROMOTE receipt copied across every gate's path satisfied them all.
  Arming (`record-lint --release-gate`) now treats the caller's required-gate list
  as authoritative even when empty — an argless arming clears the gates and fails
  closed rather than inheriting the committer-editable in-tree list. The
  `gate_lockstep` workflow parser no longer mistakes a nested `with: name:` for a
  step name. (The receipt-gate remains disabled outside release time.)
- **Secret-scanner serialisation hardening** (iss-65). A serialized scan finding's
  snippet now masks *every* secret on its source line, not only the finding's own
  token — two secrets sharing a line (a minified `.env`, collapsed JSON) no longer
  leak each other into the `abcd launch --json` report. The content sniff no longer
  misclassifies a valid UTF-8 file as binary when a multibyte rune straddles the
  8 KB boundary (which would have skipped scanning it), and a bundle file that
  cannot be read is now surfaced in `unscanned` rather than silently dropped.
- **Issue-ledger transition hardening** (iss-71). `abcd capture resolve`/`wontfix`
  now run their find→move under the same ledger lock id allocation uses, so two
  concurrent conflicting transitions on one issue can no longer land it in two
  status directories at once. A migrator-supplied `ForceID` is validated against
  the `iss-N` shape before any path is built, so a traversal id cannot touch the
  filesystem outside the ledger.
- **Rules-loader trust hardening** (iss-66). The per-repo `.abcd/rules.json` is now
  opened once with `O_NOFOLLOW` and validated on that file descriptor, closing a
  Lstat-then-read window where the file could be swapped for a symlink. The
  prompt-router's per-session dedup state moved off the world-writable shared temp
  dir to the per-user cache dir (`ABCD_RULES_STATE_DIR` still overrides), so a local
  co-tenant can no longer pre-create the predictable state path to suppress rule
  injection.

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
