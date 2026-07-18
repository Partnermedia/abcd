# Changelog

All notable changes to abcd are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and abcd
uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html) with a
leading `v`.

Before v1.0.0, minor releases may make breaking changes; each one is
called out in a **Breaking** section.

## [Unreleased]

## [0.3.0] - 2026-07-18

### Security

- **The redaction scanner no longer lets a secret survive a trailing underscore.**
  Nine hard-fail token patterns (GitHub `ghp_`/`ghs_`/`gho_`/`ghu_`/`ghr_` and
  fine-grained PATs, AWS `AKIA`, Stripe `sk_live_`/`sk_test_`) used a pure
  alphanumeric charset closed by a `\b` word boundary. Because `_` is itself a
  word character, a credential immediately followed by `_` (a JSON key, a
  concatenation, `token=ghp_..._old`) had no boundary and slipped through
  unredacted into the stored transcript. The trailing `\b` is dropped (matching
  the existing Google-key fix); the leading boundary, prefix, and minimum length
  keep the match precise. The same fix extends to Slack `xox` tokens (whose
  charset also excludes `_`) and the Anthropic/OpenAI `sk-ant-`/`sk-proj-`/
  `sk-svcacct-` keys (a minimum-length key ending in `-`), so no token pattern
  relies on a trailing word boundary.
- **Untrusted file reads are guarded against symlink-follow and unbounded
  reads.** Reads of content that can originate outside the local worktree — the
  sources-index registry, packed-lifeboat layer files, and the CLI JSON operands
  (`disembark coverage`, the memory `--pages-json`/`--page-json` transport, and
  the lesson/synthesis payloads) — now route through a single guarded primitive
  (`O_NOFOLLOW` + regular-file on the open fd + size cap, in one call). Previously
  some followed a symlink to its target's content or read an endless/oversized
  file unbounded, and a symlinked registry surfaced a raw, path-leaking error;
  all are refused consistently, closing a class of `lstat`→`read` swap windows.

### Added

- **A one-line, checksum-verified installer in the README.** The command
  detects OS/architecture, downloads the binary and `checksums.txt` from the
  latest GitHub Release, verifies the binary's SHA-256 against the manifest
  fail-closed (a mismatch — or a binary the manifest does not list — refuses
  to install), and installs to `/usr/local/bin`. The README also documents
  the inspect-first manual equivalent.

## [0.2.0] - 2026-07-17

### Security

- **Two more git call sites are isolated — finishing the env-inheritance sweep.**
  An ultracode sweep-completeness pass found the two the earlier round missed.
  `identity.gitConfig` (the commit-identity gate) read `user.name`/`user.email`
  with the ambient environment, so an injected `GIT_CONFIG_*` could forge — or an
  inherited `GIT_DIR` redirect — the very identity the gate verifies; it now uses
  `gitutil.ScrubbedEnv` (keeping global config, like the identity probe).
  `capture.discoverRepoRoot` ran `git rev-parse --show-toplevel` with the ambient
  environment, so an inherited `GIT_WORK_TREE` redirected repo-root discovery — and
  thus where the issue ledger is read and written — at an attacker-chosen tree; it
  now uses `gitutil.IsolatedEnv`.
- **The embark/pack path guards every read of an arbitrary target repo.** Five
  reads on the lifeboat embark/pack path — `embark`'s target `CLAUDE.md`
  (`embarkMarker`), the target record compare (`classifyEmbark`), the coverage
  handoff (`readCoverageHandoff`), the pinned provenance (`readProvenance`), and the
  destination-gate provenance probe (`isAbcdLifeboat`) — used a plain `os.ReadFile`
  (three of them after a separate `Lstat`, a TOCTOU window; one checking the size
  only *after* reading it all). The source is an arbitrary, possibly hostile repo,
  so a symlinked or device/endless file could be followed or read unbounded. All
  now route through `fsutil.ReadGuarded` (`O_NOFOLLOW` + regular-file on the open fd
  + size cap), closing the swap window and the resource-exhaustion path in one call.
- **The identity probe and the ahoy git helper ignore inherited `GIT_DIR`/
  `GIT_WORK_TREE` and injected `GIT_CONFIG_*` — completing the `IsolatedEnv`
  sweep.** Two git call sites still ran with the ambient environment. Worse of the
  two: `scanner.ProbeIdentity` reads the caller's `user.name`/`user.email` to
  build the hard-fail identity-redaction matchers, so an injected
  `GIT_CONFIG_COUNT`/`GIT_CONFIG_KEY_*` (as a CI/agent sandbox can export) forged a
  fake identity and the caller's *real* name/email then sailed through the ship
  gate and the transcript sanitiser unredacted. It now runs with a new
  `gitutil.ScrubbedEnv` — the repo-selection and config-injection vars stripped but
  global config kept, since the caller's identity legitimately lives there and full
  isolation would blind the probe. `ahoy.runGit` (which derives the root-commit SHA
  and origin URL that key the cross-repo history registry) now uses the fully
  isolated `gitutil.IsolatedEnv`, so an inherited `GIT_DIR` can no longer register
  one repo's transcripts under another's immutable key.
- **The secret scanner detects a Google API key whose 35th character is `-`.** The
  `\bAIza…{35}\b` pattern is fixed-length and its class includes `-`, so a key
  ending in `-` had no shorter match to satisfy the trailing ASCII `\b` and the
  `hard_fail` secret slipped both the launch gate and the transcript sanitiser. The
  trailing `\b` is dropped (the `AIza` prefix and fixed length still bound it).
- **`home_path_self` redaction is case-insensitive.** The overlapping-secret and
  Unicode-boundary hunt added `(?i)` to the email/name/github identity matchers but
  left the caller's own home path case-sensitive, so on a case-folding filesystem a
  differently-cased spelling of `$HOME` — the same directory on disk — escaped the
  hard-fail `home_path_self` gate. It now folds case like its siblings.
- **Untrusted repo content is sanitised on every terminal render path, not just
  the lifeboat report.** The escape/C1 sanitiser that guarded `disembark`'s human
  report is now a shared `internal/termsafe` primitive, and the other render paths
  that printed repository-derived text raw — `abcd audit`, `docs lint`, `capture
  list` skip rows, the `intent`/`spec` boards, `memory` (bare + `ask`) — route
  through it. A crafted commit subject, file path, error string, or memory-page
  summary can no longer inject an ANSI escape to recolour or corrupt the report.
  The primitive also now masks the bidirectional-override and zero-width
  ("Trojan Source") classes, so untrusted text cannot visually reorder or hide
  characters so the rendered line differs from the bytes. JSON output is unaffected.
- **The identity pin is written atomically.** `identity.WritePin` persisted
  `.abcd/config/identity.json` with a plain in-place `os.WriteFile` — the fifth
  writer the iss-32 atomic-write consolidation missed (the guard flags only
  divergent atomic primitives, not a non-atomic one). It truncated the pin before
  rewriting, so a crash mid-write left a corrupt or empty gate config, and it
  followed a symlink at the path. It now routes through the canonical
  `fsutil.WriteFileAtomic` (temp + fchmod + fsync + rename + parent fsync).

- **The secret scanner now detects GitHub fine-grained PATs (`github_pat_…`) and
  PEM private-key headers.** Neither was in the bundled pattern set, so a
  current-generation GitHub token (GitHub's default since 2022) or a committed
  private key passed both the `abcd launch` ship gate and the history-transcript
  sanitiser unflagged.
- **Write-time transcript redaction no longer leaks raw secret bytes from
  overlapping matches.** `scanner.Redact` masked by longest-first substring
  replacement, so two partially-overlapping secret spans (e.g. an `sk-ant-` key
  running into a JWT) left the shorter token's tail verbatim, and the fail-closed
  re-scan could not catch the now-truncated remainder. Redaction is now by
  authoritative byte span (overlap bytes forced to `*`), matching the serializer.
- **History capture fails closed when the per-repo `pii.json` is broken.**
  `Scanner.ScanText`/`Redact` could not signal the degraded (`unavailable`) state
  the way `ScanBundle` does, so a malformed config silently redacted transcripts
  with a weakened pattern set and still reported success. A new `Unavailable()`
  accessor is now consulted before capture.
- **A per-repo config can no longer neuter a bundled detector by replacing its
  regex.** The severity floor clamped an override's severity but not its regex, so
  swapping a bundled pattern's regex for a never-match one disabled detection at
  full `hard_fail` severity. Bundled regexes are now immutable (a config may only
  raise severity / adjust label; new detection must use a new pattern name), and
  the config merge is all-or-nothing.
- **The launch bundle no longer ships denied namespaces or gitignored files
  reached through a symlink.** A symlink to (or into) the repo root let a
  dereferenced walk descend into `.git/**` and `.abcd/**`, and a symlink whose
  target was gitignored shipped the ignored content under the symlink's benign
  name. The structural deny is now re-applied to real paths at every level of a
  dereferenced walk, and the ignore probe covers the symlink target.
- **Command-error output no longer leaks a path equal to `$HOME` itself.** The
  redactor only rewrote paths *under* a root, so a message or `PathError` naming
  exactly the home directory slipped through — and its base segment is the
  username. `record-lint` likewise printed raw `*os.PathError` config-load paths;
  both now scrub the home/root prefix.
- **The launch bundle's gitignore exclusion fails closed on a git failure.** It
  delegated to a fail-open probe, so if git errored the exclusion pass admitted
  every gitignored file (typically `.env`-style secrets). A launch-local strict
  probe now distinguishes "nothing ignored" from a real git failure and rejects
  the affected candidates on failure; a plain non-git directory still resolves.
- **Git queries ignore inherited `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE` and
  config-injection env vars.** An inherited repo-selection variable could redirect
  abcd's isolated git queries at a different repository; those variables are now
  stripped before every invocation.
- **Transcript snippets no longer leak the head/tail of a short multi-byte
  identity value.** `sealLine` measured its fingerprint threshold in bytes while
  `maskSecret` used runes, so a short non-ASCII identity value kept visible
  characters; the two are now both rune-based.
- **`WriteFileAtomic` sets permissions on the open descriptor** (`fchmod`) rather
  than by name after close, closing a TOCTOU symlink-swap window; and
  `WriteFileAtomicPreserveMode` no longer silently widens an existing file to
  `0644` on a transient stat error (it fails closed).
- **The PII scanner detects real names, emails, and third-party home paths that
  RE2's ASCII `\b` was silently skipping.** A `hard_fail` `real_name` whose first
  or last character is non-ASCII (accented, CJK, Cyrillic) never matched — the
  name shipped unredacted; the `real_email`/`github_username` matchers were
  case-sensitive, so a case variant of the caller's own address slipped the gate;
  and `home_path_other` never fired at a realistic boundary (line start, after a
  space or `=`), so third-party home paths published verbatim. All three now use
  Unicode-aware, case-folding boundary predicates. Relatedly, a warn-level
  `home_path_other` span no longer suppresses a `hard_fail` `local_username`
  finding underneath it (which downgraded a username leak out of the ship gate).
- **The privacy-hygiene audit flags a bare home path with no trailing separator.**
  `absPathRe` required a path component *after* the username, so the leak itself —
  `HOME=/home/alice` at end of line — was never caught. The trailing separator is <!-- abcd-audit:allow -->
  now optional (matching the Windows branch).
- **The release-bundle gitignore probe ignores inherited `GIT_DIR`/`GIT_WORK_TREE`
  and config-injection env vars.** The strict probe appended to `os.Environ()`, so
  an inherited repo-selection variable could redirect it at a different repository
  and make a gitignored secret read as "not ignored" — promoting it into the
  release. It now runs with the same scrubbed environment as every other isolated
  git call (also applied to release-tag retention).
- **Terminal-report sanitisation strips the C1 control range (`0x80`–`0x9F`).**
  It masked C0 controls and DEL but let U+009B (CSI) through — an 8-bit terminal
  acts on it exactly like `ESC[`, reopening the escape-injection path from
  untrusted commit subjects/refs that masking `ESC` alone was meant to close.
- **The lifeboat pack overlap gate is case-insensitive on case-folding
  filesystems.** On macOS's default filesystem a differently-cased destination
  inside the source (`.../REPO/lifeboat` vs source `.../repo`) computed as an
  out-of-tree sibling and slipped the gate, so the pack wrote into the source
  tree.
- **The graveyard probe rejects option-like git refs from a hostile repo.** The
  lifeboat probe (whose stated threat model is hostile/archived repositories) fed
  branch names and an `origin/HEAD`-derived default branch straight to `git
  merge-base`/`branch`/`rev-list` as positional args; a crafted ref such as `-x`
  (written into `.git/refs`) parsed as a flag. Repo-derived refs beginning with
  `-` are now refused before reaching git — closing the argument-injection vector
  before a future richer subcommand makes it exploitable.
- **Memory-store reads are guarded against symlink and size attacks.** The sources
  registry (`.sources_index.json`) and each memory page were read with a bare
  `os.ReadFile` — no size cap, following symlinks — inside the repo working tree
  (a trust boundary) on every `abcd memory` verb, some under the store lock. A
  committed symlink to `/dev/zero` could OOM or hang the CLI. Both now route
  through a shared `fsutil.ReadGuarded` (`O_NOFOLLOW` + regular-file check on the
  open fd + byte cap).

### Fixed

- **`abcd install` no longer deletes a user's `.gitignore` content after an
  orphan `# BEGIN ABCD` fence.** An unbalanced BEGIN with no matching END made
  the rewriter drop every line to end-of-file, taking the user's own ignore rules
  with it. An unmatched BEGIN is now dropped alone; the content after it is
  preserved (mirroring the stray-END policy).
- **`git check-ignore` no longer inverts the answer for force-added files.**
  Dropping `--no-index` means a tracked file that matches a `.gitignore` pattern
  is correctly reported not-ignored, so the privacy audit stops flagging a
  force-added `DECISIONS.md` and the bundler stops dropping force-added files.
- **Recall keyword matching now handles inflected forms.** The stemmer could not
  round-trip e-drop (`merging`→`merge`) or doubled-consonant (`committing`→
  `commit`) inflections, and multi-word aliases bypassed stemming entirely, so
  guardrail domains like `COMMITTING` silently failed to match common phrasings.
- **Frontmatter and intent records agree on delimiter handling.** An unclosed
  frontmatter block is now treated as no-frontmatter (it previously harvested body
  prose as top-level fields), and the intent writer tolerates a trailing-space
  `---` delimiter exactly as the reader does.
- **YAML frontmatter flow-map keys are quoted.** A citation key containing a YAML
  metacharacter previously corrupted the record or broke the write→read round-trip.
- **Concurrent `abcd intent plan` runs mint distinct spec ids** (an exclusive
  advisory lock now guards the id-scan-and-write), and history capture attributes
  a byte-identical transcript from a distinct session to its own record instead of
  the first session's.
- **Several confident-but-wrong diagnostics are now guarded:** `abcd audit
  --root <missing>` and `disembark coverage <non-report>` return a usage error
  instead of fabricated findings; the privacy-hygiene rule surfaces an error
  rather than reporting clean when it cannot read the repo; brittle-reference
  linting skips fenced code blocks; and Cobra usage errors exit `2`, not `1`
  (which `abcd audit`'s tri-state reserves for "warnings only").
- **Robustness:** a malformed glob in the include config is a preflight error
  instead of a panic; an overflowing manifest pointer index is rejected instead
  of panicking; the intent identity gate compares the author git will actually
  stamp (honouring `GIT_AUTHOR_*`); inline-list items with quoted commas round-trip
  faithfully; and the lifeboat probe's tier gate matches what its adapters read.
- **The modular-rules loader resolves `.abcd/rules.json` from the repo root, not
  the current directory.** Run from any subdirectory, `abcd` looked up the
  per-repo overrides — and the kill switch — under the subdirectory, found
  nothing, and silently injected the default ruleset a repo had disabled. The
  loader now walks up to the nearest `.abcd` directory.
- **`abcd install` no longer rebuilds a malformed `.abcd/config.json` from
  scratch,** which destroyed whatever the user had. A JSON parse error is now
  respected: the file is left untouched and the install reports partial.
- **The issue-ledger reader rejects malformed records it used to accept
  silently:** a duplicate top-level (or nested) frontmatter key — where the reader
  kept the last value but a status transition rewrote the first — and a
  non-string `resolved_by` sub-value that validated clean then dropped to `""` on
  read.
- **The disembark voyage ledger logs SHA-256-format repositories.** Its root-SHA
  key accepted only a 40-char SHA-1; a 64-char SHA-256 root was rejected and the
  pack silently went unlogged.
- **The history transcript store accepts a SHA-256 root key too.** The same
  40-char-only assumption in `history.store`'s `rootSHARe` (a sibling of the voyage
  key above) made `history capture`/`list`/`read` all fail for a repo in git's
  SHA-256 object format — the ahoy layer derives the 64-char root SHA, but history
  refused it, so no session was ever stored. The key now accepts 40 or 64 hex.
- **Spec id minting cannot wrap to a negative id.** `specNum` discarded the
  `strconv.Atoi` overflow error, keeping the clamped `MaxInt64`, so an over-int64
  spec number made `NextID` compute `max+1` and mint `spc--9223372036854775808`. An
  unparseable/over-range number is now treated as no reservation.
- **Release-tag retention ignores prerelease/build tags.** `Tag()` renders the
  core `MAJOR.MINOR.PATCH` only, so a real tag `v1.2.3-rc1` surfaced in the plan
  as a phantom `v1.2.3` and collapsed against the real release; prerelease/build
  tags are now excluded (retention operates on release cores).
- **Distinct deleted paths key distinct graveyard findings.** The id cleaner
  *deleted* spaces and control characters, so two paths differing only in
  whitespace collided onto one finding id and one shadowed the other; the
  transform is now injective (percent-encoding), leaving ordinary paths unchanged.
- **The lifeboat pack destination gate treats an `ENOTDIR` stat as "absent"**
  (a prefix component being a file) rather than an uninterpretable error that
  refused a writable destination.
- **A relative PATH symlink is resolved against the symlink's own directory,**
  not the process working directory — so `abcd ahoy` no longer reports a bogus
  "foreign symlink" gap (and uninstall no longer refuses to remove a link it
  owns) for a correct relative install such as `/usr/local/bin/abcd ->
  ../lib/abcd/abcd` when run from another directory.
- **`source.classes` on a memory page is validated as a set, not an ordered
  list.** The same classes declared in a different order from their first
  appearance in `sources[]` were rejected, contradicting the schema's (and the
  error message's) set semantics.

### Changed

- **The coverage report is now schema v2.** Each brief section carries a `kind`
  (`extractable` — a source or a better adapter could ground it, so a blank is
  coverage debt — versus `human-owned` — a question only a person can answer, so a
  blank is not a failure), the durable form of the M2 cross-repo gate decision
  (adr-36). A blank additionally carries a `resolution` (`open`/`answered`/
  `deferred`) and, once answered, an authored `answer` whose provenance is a
  person and a date rather than a file it did not come from. `abcd disembark
  coverage` still refuses a report from a newer schema with an upgrade message.

### Added

- **`abcd intent "<text>"` files a new draft from quoted text — a symmetric
  create path.** Typing `abcd intent "I want users to feel X"` mints the next
  `itd-N` under the intent-store lock and writes
  `.abcd/development/intents/drafts/itd-N-<slug>.md`, seeded from the text with the
  canonical draft frontmatter and a minimal, lint-valid body skeleton — no `new`
  sub-verb required, mirroring `abcd capture "<text>"`. The old `abcd intent new
  "<text>"` still works as a backwards-compatible alias but prints a deprecation
  warning on stderr naming the quoted-text shape. Bare `abcd intent` stays
  read-only status + help and mutates nothing. Both ledgers' bare-form help now
  carry a one-line decision rule (nitpick/observation -> capture; user-facing
  change to ship -> intent). The `/abcd:capture promote` flow hands the issue text
  to this create path (itd-46).
- **`GL002` — a glossary-driven forbidden-synonym gate for the record lint.**
  The lint now reads each glossary term's `forbidden_synonyms` and flags an
  *enforced* synonym used as a standalone word in live prose, so terminology
  drift is caught by a detector instead of by eye (itd-43). Enforcement is a
  deliberate subset (`epic` first): most forbidden synonyms are common English
  words whose false-positive rate would sink the gate, and each enforced word
  must be one the glossary actually forbids. Matching uses explicit Unicode word
  boundaries — not the ASCII-only regexp `\b` — and skips code spans, YAML
  frontmatter, dated/historical records, and the glossary term files themselves.
- **Bare `abcd ahoy` now names the next step for the folder it classified.** An
  unmanaged git repo report points at `/abcd:ahoy install` as the way to adopt
  it, and a plain (non-git) folder report states there is nothing to act on —
  the read-only classification never mutates either (itd-40).
- **Synthesis over the record — `abcd disembark principles`, `press-release`,
  and `oracle`.** Three post-pack verbs interpret a packed lifeboat, each in
  one of two self-recorded modes. Without a payload they run **deterministic
  mode**: principles distilled evidence-only from the packed ADRs'
  Decision/Consequences bullets, the press release composed from the brief's
  own page (or the spine, or an honest placeholder), and the oracle scoring
  mechanically — a failed manifest verification is a `MAJOR_RETHINK` verdict,
  not an error; more blanks than grounded sections is `NEEDS_WORK`; a healthy,
  verified lifeboat ships `SHIP` — the first code home of abcd's registered
  review-verdict vocabulary. With `--*-json <file|->` they ingest a
  host-delegated agent's output behind the same trust guards as an intent
  verdict, under cite-or-be-dropped (a principle or oracle finding citing no
  live record id, graveyard finding, or packed path is dropped and reported; a
  press release citing nothing resolvable is refused whole). The binary stamps
  the oracle's attestation fields itself, so a model cannot fabricate a
  manifest hash. All synthesis artifacts live outside `manifest_sha256`, are
  fully replaced per run, and carry no wall-clock — the audit is keyed by the
  lifeboat's own manifest hash. The four agents (`principle-distiller`,
  `graveyard-interpreter`, `press-release-composer`, `lifeboat-oracle`) ship
  under `agents/` with itd-5's prompt discipline: versioned prompts in the 0.x
  calibration band, `reads_untrusted_input` declared, and an injection-canary
  fixture each (itd-88, adr-35).

- **`abcd embark` — a lifeboat comes ashore.** `embark probe <lifeboat> [target]`
  is the read-only reconciliation: it refuses a lifeboat whose provenance schema
  is newer than the binary (with an upgrade message), re-hashes every archived
  file against the pinned `manifest_sha256` so a tampered or truncated lifeboat
  is caught before anything is read in anger, and reports — in one bulk report,
  not a per-file barrage — every conflict with the target repository. `embark
  from <lifeboat> [target]` is the write path: it refuses entirely on any
  conflict (identical bytes are an idempotent skip, and a re-embark is a no-op),
  writes only the record families (ADRs, issues, intents, specs) through
  `os.Root` containment plus independent path validation — two layers, so a bug
  in one is not an escape — and never copies lifeboat prose into `CLAUDE.md`:
  it re-injects the *current* abcd marker block instead. The rendered result
  leads with the coverage report's blanks and their questions — the handoff to
  the human who must answer them. The packer now also carries the spec store
  (`rescue/specs/`), and every lifeboat's provenance records
  `record_manifest_sha256`, the seal over exactly the record-derived families
  that must survive a round-trip byte-for-byte: pack → embark → re-pack
  reproduces it, and embarking into a byte-copy of the source reproduces the
  full original manifest hash (itd-88, adr-35; closure re-scope in the
  2026-07-16 decision log).

- **The graveyard — what the project abandoned, in three strictly-ordered
  layers.** Every packed lifeboat now carries `graveyard/archaeology.json`
  (deterministic git archaeology: reverted commits, branches never merged into
  the default branch ranked by divergence age, paths deleted after substantial
  history, dependencies adopted then dropped, wholesale-rewrite commits — pure
  evidence, no interpretation, from any git repo) and `graveyard/abandoned.json`
  (what the record itself declared dead: superseded intents and ADRs, wontfix
  issues with their reasons, each ADR's Alternatives-Considered options, and
  rejected options named in the decision log). A new
  `abcd disembark graveyard <lifeboat-dir> --lessons-json <file|->` verb ingests
  a host-delegated interpretation over those two layers into
  `graveyard/lessons.json` under a **cite-or-be-dropped** validator: every
  lesson must cite live layer-1/2 evidence ids or it is dropped (reported, never
  fatal), low-confidence lessons are quarantined under
  `graveyard/low-confidence/` instead of the main file, and the untrusted
  payload is read behind the same trust guards as an intent verdict (size cap,
  no symlinks, unknown fields refused, schema version gated). Each ingest fully
  replaces the prior interpretation, so a promoted or later-dropped lesson
  leaves nothing stale behind. The validator — not the model's good intentions
  — is the difference between a graveyard and a séance (itd-88, adr-35).

- **`abcd disembark probe <repo>` — a read-only coverage probe over any
  repository.** It walks a repo without touching it and reports, per brief
  section, whether a lifeboat could ground it: `grounded` / `partial` / `blank`,
  with the tier it was grounded from, a confidence, and the evidence cited. A
  blank is a first-class result — it carries what abcd searched and the question
  a human must answer, so the report is a to-do list, not a shrug. Adapters
  degrade across three tiers — git (any repo), conventions (README, docs,
  CHANGELOG, manifests, ADRs wherever they live), and abcd-native (`.abcd/`) — so
  a richer repo grounds more, and the `graveyard` section grounds from git
  history alone (reverts, deleted files, dependency churn). Every read is
  contained to the repo (`os.Root`), bounded, and non-blocking, and the source
  tree is byte-identical afterwards — the probe never writes to a source. A
  companion `abcd disembark coverage <report.json>...` reduces several probe
  reports to one section×repo table with an always-blank verdict per section:
  the delta between a record-rich repo and a git-only one is what keeping a
  record is worth, legible as a number. Both are read-only operator verbs (no
  `/abcd:disembark` command surface yet); the packer that writes a lifeboat is a
  later milestone (itd-88, adr-35). The dependency-manifest detector spans Go,
  Node, Rust, Python (pip/poetry/pdm/uv/pipenv), Ruby, and PHP, so a real project
  is not reported as having no dependencies merely because the probe did not know
  its packaging tool (found probing a Python/uv repo in the M2 cross-repo run).

- **`abcd disembark plan <repo>` — a dry run of the packer.** It shows the
  complete file set a lifeboat pack would write — brief citation maps for the
  grounded sections, `coverage.json`/`coverage.md`, verbatim copies of the ADRs
  and the issue ledger, the rescue spine (the intent corpus where one exists, a
  git-derived summary where it does not), and a `_provenance.json` carrying a
  pinned `manifest_sha256` over every other file — and writes **nothing**. Plan
  and the eventual packer are one code path, so the dry run cannot describe a pack
  a real pack would not perform; a re-plan of an unchanged source is byte-for-byte
  identical (the manifest carries no timestamp). `--json` emits the manifest
  (paths, sizes, and the hash — never file content). Still a read-only operator
  verb (no `/abcd:disembark` command surface yet); the destination write path is
  a later milestone (itd-88, adr-35).

- **`abcd disembark pack <repo> <dest>` — writes a lifeboat out-of-tree, and the
  `/abcd:disembark` command surface.** It writes the planned file set to `<dest>`
  and never to the source (a test hashes the source tree before and after).
  Everything that stops a pack destroying real work is enforced: a **destination
  safety gate** refuses unless `<dest>` is absent, an empty directory, or an
  existing lifeboat abcd produced (it carries a parseable `_provenance.json`) —
  and refuses a symlinked destination, one inside a `.git/` directory, or one that
  overlaps the source tree. The planned bytes are **secret-scanned before any
  write** and a hard-fail refuses the whole pack — a secret is fixed at source,
  never redacted into the artefact. Files are written into a staging directory
  through `os.Root` (no crafted path or symlink escapes it) and renamed into
  place, so a crash leaves staging, never a half-lifeboat; `_provenance.json` is
  written last. Any abcd marker block in a copied record is stripped so embarking
  the lifeboat cannot plant a stale rules-loader. Each pack appends one line to an
  append-only voyage ledger at `~/.abcd/voyage/<source-root-sha>/disembark/history.jsonl`,
  keyed on the source's root-commit SHA and carrying the manifest hash. `--json`
  emits the result (destination, file/byte counts, hash, voyage status).

- **Session transcripts are captured automatically when a session ends.** A
  `SessionEnd` hook now runs `abcd hook session-end`, which redacts the session
  transcript through the existing two-stage, fail-closed scanner and files it in
  the local per-repo store — no flag to pass, no command to type. `abcd history
  list` shows the records. It is wired to `SessionEnd` (which fires once when a
  session terminates) rather than `Stop` (which fires once per assistant turn) —
  the transcript grows through a session, so a `Stop`-wired capture would store a
  fresh, larger superset every turn. The store has existed since the native transcript
  corpus landed (adr-29) and was **called by nothing**: `history.Capture` was
  built, correct, and unused, so no session was ever stored. That gap was the one
  cost on the board that could not be recovered later — a session that ends
  without being captured cannot be reconstructed by any amount of future work.
  The hook is operator-internal, never blocks the host, and always exits `0`: a
  malformed payload, a missing or non-regular `transcript_path`, a hostile
  session id, or a directory that is not a git repo each capture nothing, say why
  on stderr, and exit cleanly — a `Stop` hook that errors or hangs would wedge
  the session, which is strictly worse than a missed transcript. Re-capture is
  idempotent (a `Stop` hook may fire more than once per session), the transcript
  open is non-blocking so a FIFO cannot hang the hook, and nothing is ever
  written to stdout. It needed a new verb because `history capture` cannot be
  wired to a `Stop` hook: from stdin it *requires* `--session <id>`, and a `Stop`
  hook delivers its session id inside a JSON payload (itd-89, adr-29).

- **A session that starts in a repo where abcd is not installed now says so.** A
  `SessionStart` hook runs `abcd hook session-start`, which — when the current
  repo is a git repository whose transcript store has not been bootstrapped —
  prints a one-line notice telling the user their sessions will not be captured
  and how to fix it (`abcd ahoy install`). Without it the automatic-capture hook
  above fails silently: the plugin is enabled, the user assumes their transcript
  corpus is accruing, and it is not. The notice rides `SessionStart`'s visible
  channel (stderr on a non-zero exit) and never blocks the session; every case
  that is *not* a bootstrappable-store problem — a non-git cwd, a malformed or
  empty payload, an already-installed store — stays completely silent and exits
  `0` (iss-95, itd-89).

- **`abcd audit` — a read-only repo-conformance check.** One command reports
  whether a repository follows the working conventions: the three-tier `.abcd/`
  layout, an `AGENTS.md` router, decisions durable in a committed
  `.abcd/work/DECISIONS.md`, docs currency (reusing the docs-lint engine where
  `docs/` exists), and privacy hygiene (no absolute local paths in committed
  files, waivable per line with `abcd-audit:allow`). It runs against any repo
  given only a working directory, prints a grouped human report with a fix per
  gap or machine JSON (`--json`, stable rule ids, `{ "findings": [] }` when
  clean), and exits with a tri-state code — `0` clean, `1` warnings only, `2`
  any error — so it gates CI as well as onboarding. It answers a different
  question from `abcd ahoy doctor`: `doctor` is tool-setup health, `audit` is
  repo conformance. `/abcd:prepare-this-repo` now runs `abcd audit` for its
  Phase 2 gap report instead of hand-auditing (itd-85, iss-86).

### Fixed

- **`--json` and stderr command errors no longer leak the developer's home or
  working-directory paths.** `cli.Run` routes every command error through the
  machine envelope, so identity-bearing paths reached it three ways: an
  `os.PathError`/`os.LinkError` (e.g. `memory ask --page-json` on a missing file),
  a path `fmt`-formatted into a core error (e.g. `capture` on a symlinked ledger
  dir), and a custom error type (e.g. history's home-rooted store path). The
  `Run()` boundary now redacts the working-directory and home roots (to `.` and
  `~`) and reduces any remaining `PathError`/`LinkError` path to its base name.
  Generalises the per-branch fix made in `iss-29` (iss-76). A verb echoing a
  user-supplied absolute path outside both roots is out of scope, tracked in
  `iss-81`.
- The `intent_lifecycle` record-lint rule now **blocks duplicate intent ids**.
  Id allocators are branch-local — parallel agents on separate branches each
  scan for `max + 1` and mint the same id — so two intents both claimed
  `itd-82` and both merged with every gate green. The rule flags *every* file in
  a colliding set, not just one: the linter cannot know which claimant is
  authoritative, and flagging a single file would imply the others are fine. The
  collision itself is resolved (the later claimant renumbered to `itd-83`); the
  underlying minting scheme is tracked as `iss-80`.
- `memory ingest --keep-original` writes the stored source copy through the
  canonical `fsutil.WriteFileAtomic` (temp + fsync + **chmod + parent-directory
  fsync**) instead of an inline temp+rename that omitted both — the fifth
  divergent atomic write the `iss-32` consolidation left untouched. The
  one-canonical-primitive detector now also flags inline `os.O_EXCL`+`os.Rename`
  sequences, not just named primitives (iss-79).

### Added

- **Four reviewer agents ship with the plugin**: `abcd:ruthless-reviewer`
  (correctness, resource handling, error paths, dead code),
  `abcd:security-reviewer` (adversarial review of a trust boundary),
  `abcd:docs-currency-reviewer` (every user-facing claim verified against the
  code), and `abcd:sota-researcher` (evidence-tiered state-of-the-art research).
  Every repo with the abcd plugin enabled gets the same review bar, versioned in
  the repo rather than in a per-machine harness config. Each renders a binary
  verdict, and every finding it emits must carry a concrete failure scenario —
  the LLM-judge calibration discipline (itd-81).
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

- `abcd intent plan` seeds a new native spec with a clear author-guidance
  placeholder in its `## Summary`, rather than a bare `TODO` (iss-68).
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
- The atomic-write and real-directory primitives are consolidated onto
  `internal/fsutil` (iss-32): the ahoy, capture, and memory store writers no
  longer keep their own divergent temp-file+rename copies. Two observable
  effects of routing through the canonical primitive: the ahoy and capture
  writers now fsync the parent directory after the rename (a crash-durability
  strengthening they previously lacked), and memory pages are written at a
  fixed `0644` (an explicit chmod, where the old writer left the mode subject to
  the process umask). A `TestNoNonCanonicalAtomicWritePrimitives` guard keeps a
  fifth copy from reappearing.

### Removed

- The `created` and `updated` frontmatter fields on issues. Git is the canonical
  source of an issue's timeline; the ledger no longer duplicates it.

### Fixed

- **Launch dogfood gate — identity false positive and resolver race** (iss-31).
  The secret/PII scanner no longer hard-fails on a system path such as
  `/dev/null` when the machine username collides with a system directory name
  (e.g. a user called `dev`): a local-username match is suppressed only when it
  is the top segment of an absolute system path, so genuine username leaks
  (nested under a home root, or bare) are still caught. The launch bundle's
  compiled-glob cache is now guarded by a mutex, removing a data race when the
  transport-agnostic core resolves bundles concurrently.
- **Memory-ingest boundary — partial-failure reporting and CRLF parity**
  (iss-30, continued). When `abcd memory ingest --keep-original` fails to store
  the original *after* the pages and registry are durably written, it no longer
  reports total failure: the successful ingest is reported (pages listed) with a
  warning and a non-zero exit, and the failure message names only the
  repo-relative store location — no absolute path, in text or `--json`. CRLF
  documents now split identically to their LF form (`splitFileFrontmatter`
  normalises line endings like its sibling parsers), so a `\r\n` closing `---`
  delimiter is no longer rejected and hashes/summaries no longer silently
  degrade.
- **Fail-closed capture surface** (iss-29). A mistyped `capture` subcommand
  (e.g. `abcd capture resovle iss-1 …`) is no longer swallowed as free text and
  filed as a new issue; it is refused with a did-you-mean and writes nothing.
  Errors requested with `--json` are now emitted as a `{"error": …}` envelope
  rather than raw Go text, and `abcd docs lint` with a missing or unreadable
  config reports a clean, repo-relative diagnostic instead of a raw file error
  that leaked the absolute config path.
- `abcd` status now reports `IsGitRepo` correctly in a linked git worktree or a
  submodule, where `.git` is a regular gitfile rather than a directory (iss-72).
- `abcd intent plan` now refuses an `## Acceptance Criteria` section with no
  top-level `-`/`*` bullet, matching the ingest gate — an intent can no longer be
  planned into a state where every fidelity verdict dead-letters for having zero
  positional criteria. The intent template's Audit Notes placeholder is cleared
  when the first review block lands, so a populated audit carries no stale "Empty"
  claim (iss-67).
- The frontmatter scanner (`internal/core/frontmatter`, used by `abcd intent`/
  `spec` and record-lint) now tolerates a trailing space or tab on the `---`
  delimiters; previously a `--- ` closing delimiter went unrecognised and every
  body line after it was misread as a frontmatter field. `record-lint` no longer
  keeps a divergent copy of the scanner — it routes through the canonical one and
  inherits this fix (iss-69).

### Security

- **Memory-ingest fetch/read hardening** (iss-30). `abcd memory ingest` now treats
  a non-2xx HTTP response as an error instead of storing the 404/500 error page as
  source content; the SSRF guard additionally rejects NAT64 (`64:ff9b::/96`) and
  6to4 (`2002::/16`) IPv6 addresses that embed a metadata/loopback/private IPv4; a
  local source file is size-capped like the URL path; and a `~user` path is left
  literal rather than being mangled into `home`+`user`.
- **Spec-store hardening** (iss-68). The spec-store reader now opens a file once
  with `O_NOFOLLOW`+`O_NONBLOCK` and validates the file descriptor before reading,
  closing a symlink-swap window (and never blocking on a FIFO leaf). `NextID` fails
  closed on an intent `spec_id` that carries no parseable reservation number (e.g.
  `spc-` with no digits) instead of silently dropping it from the id-reservation
  scan (which could hand out a colliding id); a `spc-N` or `spc-N-<slug>` form still
  reserves N, consistent with record-lint. (The leaf-only ancestor-symlink guard
  and the atomic-rename clobber check are documented as accepted under the
  trusted-worktree model.)
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
