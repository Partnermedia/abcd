# Changelog

All notable changes to abcd are recorded here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and abcd
uses [Semantic Versioning](https://semver.org/spec/v2.0.0.html) with a
leading `v`.

Before v1.0.0, minor releases may make breaking changes; each one is
called out in a **Breaking** section.

## [Unreleased]

### Breaking

- **The lifeboat verdict says what it is.** `abcd disembark oracle` is now `abcd disembark review`, its flag `--oracle-json` is `--review-json`, and the verdict artefact moves from `audit/oracle-<manifest12>.{json,md}` to `review/review-<manifest12>.{json,md}` (the `audit_path` key in the JSON envelope is now `review_path`). The `lifeboat-oracle` agent becomes `lifeboat-reviewer`. The verb emits family-1 change-judgement verdicts (`SHIP`/`NEEDS_WORK`/`MAJOR_RETHINK`) — a *review* in adr-40's vocabulary — and the planning investigation found it never invokes the oracle seam at all, so naming it for that seam claimed something untrue and collided with the reserved `/abcd:oracle ask`. **The seam itself is untouched**: adr-25 stands, `oracle` remains the model-access seam's name, `ahoy install --oracle-backend` is unchanged, and the `oracle_review` task-class token stays. A re-run over a lifeboat reviewed before this release replaces cleanly — it writes the new artefact, removes that manifest's stale pair, and prunes `audit/` if empty, never touching another manifest's files. Verdict logic, the enum gate, cite-or-be-dropped, and attestation stamping are behaviour-frozen. (itd-125, spc-30; adr-40 §5 amended in place)

- **The conformance check calls itself lint.** `abcd audit` is now `abcd lint` (and `/abcd:audit` is `/abcd:lint`): the verb applies deterministic rules about a repo's form, which the record's vocabulary names a *lint*, and the `audit` name returns to its reserved seat — itd-16's hash-chain fidelity surface (adr-40). The tri-state exit contract (0 clean / 1 warnings only / 2 any error), the rule set, and the JSON envelope are unchanged; the conformance core moves to `internal/core/repolint`. The privacy waiver's current spelling is `abcd-lint:allow`, and every committed `abcd-audit:allow` line stays honoured forever. **Managed repos:** re-download the binary, and re-run `prepare-this-repo`/`launch scaffold` where a repo's own instructions or CI referenced `abcd audit`. (itd-124, spc-29; the lint-engine merge question is iss-251)

- **The intent audit says audit.** `abcd intent review` is now `abcd intent audit`, and `intent review ingest` is `intent audit ingest`: the verb emits family-2 promise-vs-reality verdicts (`MET`/`NOT_MET`/…), which the record's own vocabulary rules an *audit*, not a review (adr-40). The `intent-fidelity-reviewer` agent renames to `intent-auditor` with it. Clean break, no alias: the old spelling is refused (with the successor named) and never swallowed as a free-text intent create. Stored artefacts are format-frozen — the `abcd-review:` audit-note markers, the `abcd/intent-fidelity-verdict/v1` payload type, and every previously ingested verdict remain valid; only the verb, the agent, and the code identifiers move. (itd-123, spc-28)

### Added

- **The surface registry can no longer wave its hands.** Every surface file under the brief's `04-surfaces/` carries a machine-checked `## Sub-verbs` table recording two facts per verb — its adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a non-assessment verb) and whether it exists (`shipped` / `staged`) — and the `surface_coverage` record-lint rule gains a sub-verb pass that checks every row against the committed command-tree snapshot in both directions: a `shipped` row must be registered, a `staged` row must not be, a registered sub-command must have a row, and a sub-command-bearing verb cannot hide without a table. Exemptions (host-delegated surfaces, operator-internal verbs, the bare command) are explicit config, never silent skips; a duplicate table heading and a malformed row are findings too. The bucket enum is registered as reserved vocabulary, closed and PR-to-extend, and every sub-verb now has one machine-checked home that a stale prose claim can be corrected against. (itd-122, spc-27; closes iss-246)

- **Type the id, get your next move.** `abcd <id>` dispatches on any record id — `iss-N`, `itd-N`, `spc-N`, `adr-N` — and reports, strictly read-only, what the record is, its links, and the concrete next move for its lifecycle state: a draft intent points at the planning interview, a planned one at its spec body or at implementation (the `intent ready` checks decide), an open issue at `capture promote`/`resolve`/`wontfix`, a promoted one at the intent it graduated into, and decisions are read. Bare `abcd` answers *what can I do*; `abcd <id>` answers *what is this*. Any other positional stays on the unknown-command path unchanged, and every recommended verb is pinned to the live command tree by a test, so a future rename breaks the build instead of shipping stale advice. (itd-121, spc-26)

- **A resolved issue points at what fixed it.** `abcd capture resolve` gains optional provenance flags — `--intent itd-N`, `--spec spc-N`, `--commit <sha>` — that write the structured `resolved_by` pointer the schema has modelled all along, in the same atomic transition as the resolution note and impact. Ids must exist in their record store; the sha is shape-checked only; an unknown or malformed value refuses the whole resolve and writes nothing. Without the flags the record stays byte-identical to a plain resolve. `wontfix` is untouched — a non-action points at nothing. (itd-120, spc-25; the `resolved_by` half of iss-245)

- **An issue graduates into an intent without retyping.** `abcd capture promote <iss-N>` is the native verb for step 2 of the record walk: one invocation mints an intent draft — slug reused from the issue, body carrying a by-id pointer to the issue rather than a copy, `promoted_from: iss-N` in its frontmatter — and stamps the issue's `promoted_to` with the minted `itd-N`. Promotion works from any status folder and never moves the issue, a second promote is refused with the existing `itd-N`, and a stamp failure after the mint names the orphan draft and its repair: `capture promote <iss-N> --intent <itd-N>`, the stamp-only mode that links an existing draft instead of minting. (itd-119, spc-24; the `promoted_to` half of iss-245)

### Fixed

- **A skewed plugin install no longer blocks the session it is meant to guard.** On a host hook an exit status is an instruction, not a diagnostic: the host reads 2 as *block this action*. Since every command parent began refusing an unknown sub-verb at cobra's usage status, a renamed hook sub-verb made `abcd guard <old-name>` exit 2 — so the guard claimed a hazard verdict it never reached and blocked **every shell command** in the session, and `abcd hook <old-name>` blocked **every prompt**. The manifest and the binary can genuinely skew, because `hooks/hooks.json` ships with the plugin clone while the binary is fetched from the latest release, and the PreToolUse wrapper could not rescue it: it treats 2 as a recognised code, so its *FAILED TO RUN … UNGUARDED* net never fired. The unknown-sub-verb path under the two parents a hook reaches now refuses at exit 1 — loud, non-blocking, naming the skew and its remedy. A mistyped sub-verb still never reads as success; only the code moves. **Scoped, not universal:** a stray positional on a leaf, an unknown flag, and an unknown top-level token still exit 2, none of them reachable from today's manifest (iss-269 carries that gap). One consequence worth knowing: `abcd guard <typo>` now exits 1, which is also `guard check`'s code for a block verdict, so a script keying on the exit code alone can no longer tell a typo from a hazard — the stderr text distinguishes them. The tree-wide usage-error mapping no longer overwrites an exit code a validator chose deliberately, which is what had been silently undoing this. A new test resolves every invocation in `hooks/hooks.json` against the live command tree, enforcing the doctrine that actually survives the skew: the manifest's spellings are **frozen**, and a rename is absorbed by an alias in the binary — because the clone runs ahead of the binary, an alias only ever helps the new side, so editing the manifest to a new spelling is the move that strands older binaries. (iss-267)
- **The attribution gate now catches the footer it was written to catch.** Its banned-footer rule anchored on optional leading whitespace, but the footer a hosted agent platform actually appends is wrapped in markdown emphasis — a leading underscore — so the precise shape that motivated the gate went straight through the pull-request-body check. It was found live on two pull requests whose failing leg happened to be something else, which is how it stayed hidden. The rule now admits an optional emphasis run — italic, bold or bold-italic — with the robot-emoji form allowed on either side of it, since a footer may put the emoji inside the italics or before them. The marker must be attached to the word, exactly as markdown reads it (`*text*` is emphasis, `* text` is a list item), so a bullet describing the banned footer stays documentation rather than a violation — the same writable-about property the line anchor already protects. Eight new corpus cases pin the refused forms and two pin the list-item forms that must keep passing; narrowing the run, dropping either emoji position, or allowing a space after the marker each fails the corpus, so every clause earns its place. (iss-262)
- **A mistyped sub-verb no longer reads as success.** A cobra parent with no `RunE` is not runnable, so cobra printed help and exited **0** without ever running the command's argument validator: `abcd docs nonsense`, `abcd guard nonsense`, `abcd embark nonsense`, `abcd ideate nonsense`, `abcd docs cite nonsense` and `abcd hook nonsense` all reported success to a calling script. Every parent in the tree is now runnable, so its declared argument validator actually runs and refuses an unknown sub-verb at exit 2, while bare invocation still prints help at exit 0. Refusing needs both halves — a runnable parent that declares no validator falls through to cobra's arbitrary-args default and exits 0 anyway — so the guarding test asserts runnability *and* a declared validator across the live command tree, and derives the parents it exercises from that tree rather than a hand-kept list, so a parent added later cannot reintroduce the hole by missing either half. `capture` and `intent` are unchanged — their positional is free text by design, and they keep their own suspected-typo guard. (iss-266; the sweep spc-30 owed after fixing the `disembark` parent alone)


## [0.5.1] - 2026-08-16

### Added

- **abcd is citable, and its sources are on the record.** A root `CITATION.cff` (CFF 1.2.0) powers the forge's cite-this-repository box; the References & sources section of `ACKNOWLEDGEMENTS.md` records the academic literature the design record draws on as a curated, alphabetically ordered list of primary sources; and the canonical CSL-JSON metadata lives at `.abcd/development/research/references.csl.json`, with a documented on-demand `.bib` export for LaTeX toolchains — generated, never committed. (iss-257)

### Fixed

- **A missing plugin binary no longer degrades the whole session in silence.** Session start was the only hook that ever provisioned the binary, so when that one hook did not fire the session stayed broken for good, with every later hook failing as a raw `No such file or directory`. Every binary-invoking hook — the prompt router, the shell guard, the compaction reset, the session-end capture — now guards on the plugin root, attempts a rate-limited silent bootstrap salvage when the binary is absent (one attempt per ten-minute window), resolves the binary plugin-root first and then from `PATH`, and on continued absence prints one plain line naming exactly what is degraded and the install remedy. Session start keeps its loud primary-provisioner role unchanged. (iss-254)
- **`ahoy install` no longer hides a repo's committed record tiers from git.** The public visibility block used to ignore `/.abcd/` wholesale, so on any repo that commits the durable tiers — this one included — new intents and issues vanished from `git status` and `git add` refused them. When `.abcd/` holds tracked files, the block's `.abcd/` entry — and only that entry — now narrows to the local-ephemeral tier, and the install receipt says out loud that the committed record tiers remain published. Narrowing needs positive evidence: a directory whose git cannot be asked, or no repo at all, keeps the declared set, and the `memory/` snapshot fence is unchanged. (iss-255)

## [0.5.0] - 2026-08-16

### Breaking

- **The documented install location is `~/.local/bin`, and nothing abcd runs asks
  for administrator rights** (iss-171). The README one-liner drops `sudo` and
  copies the verified binary into `~/.local/bin`; `abcd ahoy install` writes its
  `PATH` entry there too, creating the directory when absent. **What existing
  users see depends on how their binary got there.** An abcd-owned symlink — the
  thing `ahoy install` writes — is found anywhere on `PATH` and adopted in place,
  so nothing changes. A binary the old one-liner *copied* into `/usr/local/bin`
  is a plain file abcd does not own: it is never adopted, and a new entry in
  `~/.local/bin` lands behind it on `PATH`, so the copy is still what runs. That
  case now reports a `symlink.shadowed` gap naming the copy, both from `ahoy`
  and on the install run itself; the remedy is to delete the stale copy
  (`rm /usr/local/bin/abcd`, which needs the rights that put it there) or to
  install ahead of it with `--bin-dir`. abcd will not remove it — it never
  touches a binary it does not own. A system-wide directory is reachable only
  through an explicit `--bin-dir`, which fails loudly when it is not writable
  rather than re-running itself with privilege: abcd escalates nothing, so there
  is no fallback to hide the refusal behind. Two more gaps arrive with it:
  `~/.local/bin` not on `PATH` is its own named gap carrying the one-line
  `export PATH="$HOME/.local/bin:$PATH"` fix (abcd prints it and never edits a
  shell profile), and an abcd-owned `PATH` entry whose binary has gone is
  reported as dangling rather than silently trusted. Install refuses to create a
  link whose target does not exist, because a dangling `abcd` early on `PATH`
  shadows every working one behind it, and it reports every such refusal on the
  result rather than leaving a gap to speak for it. The old detector recognised
  exactly one blessed target, so a working `~/.local/bin/abcd` reported
  `symlink.missing` while the detector was itself running as that very binary,
  and "fixing" it would have written the shadowing link.

### Added

- **A stale abcd never answers silently** (itd-111). Every surface now knows its
  own vintage and says so. `abcd version` and `abcd ahoy` print the install mode,
  the running binary's vintage (its build revision in a source checkout, its
  pinned version otherwise), and whether it is up to date, stale, or of an
  undeterminable vintage relative to the on-disk reference. At session start, a
  binary behind (or a dirty rebuild atop) its own source tip is named with the
  one-command rebuild fix. `abcd ahoy install` refuses before any write when the
  binary is stale against its tip or its vintage cannot be determined — the trap
  where month-stale install logic silently ran against a machine — with
  `--allow-stale-binary` as the documented override. And `abcd version --check`
  is the one command that reaches the network: it fetches the latest release
  once, compares, and reports with its source named. Every other path reads only
  what is on disk.
- **`/abcd:intent` decomposes a proposal before filing it** (itd-84, MVP
  rung). The surface page now carries a hand-run protocol that routes each
  part of a proposal to its record home (capability → intent, trust rule →
  ADR plus brief invariant, stance → principle, plumbing → brief), surfaces
  typed links to existing records, and renders an advisory FILE-AS-IS /
  SPLIT / HOLD verdict a human confirms; the planning interview runs the same
  step. Not yet automated: the deterministic pre-pass and the capture-time
  validator are future rungs, and every hand-run is graded into a calibration
  corpus that gates them.
- **The attribution gate reads the git identity, not only the message.** A commit
  authored and committed as `Claude <noreply@anthropic.com>` carried a fully
  compliant message — `Assisted-by:` trailer, no banned footer — and sailed
  through the gate, because the gate read only messages and bodies. The
  contributor graph is built from commit authorship plus `Co-authored-by:`
  trailers, so that one identity put an AI at #2 in the graph twice over: once
  for the commit itself, and again on every squash merge, where the forge
  auto-appends a `Co-authored-by:` for any branch author who is not the PR
  author. The commits half now refuses an AI author or committer identity —
  whole-name match on the assistant names AI tools stamp by default, plus the
  vendors' address space, vendor-agnostic in intent like the co-authorship ban —
  while the bot exemption is untouched and a human whose name merely contains an
  assistant's name still passes. The corpus at
  `scripts/check-attribution-cases.sh` grows a commits-mode section that proves
  all of it against a scratch repository.

- **The attribution gate accepts the trailer forms actually in use, and refuses AI
  co-authorship whoever wrote it** (iss-214, iss-215). Two defects in the gate's first cut. Its trailer
  pattern had no bracket in its character class, so `Assisted-by:
  Claude:claude-opus-5[1m]` — an exact model identifier, carried by 38 commits of
  this repository's own history — was rejected while the unbracketed form passed,
  so an agent disclosing its precise model followed the prose and failed the check.
  And the co-authorship ban named one vendor, so `Co-authored-by: ChatGPT` passed
  while `Co-authored-by: Claude` did not, leaving the gate effective only for as
  long as the assisting model was Claude. The trailer now admits an optional
  bracketed context suffix; the co-authorship ban matches the trailer KEY rather
  than a vendor name; and the "generated with" ban matches the footer's shape, so
  `Generated by [tool]` is caught alongside `Generated with [tool]`. A test corpus
  at `scripts/check-attribution-cases.sh` pins all of it — including the historical
  identifiers and the shell-metacharacter body — and runs in the gate's own
  workflow, so the rules are proven rather than asserted. A bare vendor with no
  version stays refused, and refusing every co-authorship trailer refuses a human
  one too: abcd defers DCO until the repo is public or takes an outside
  contribution, so there is no such case today.

- **The AI-attribution convention is enforced, not merely written down.** A new
  `attribution` workflow fails a pull request whose commit messages or body break
  the rule `AGENTS.md` and `CONTRIBUTING.md` state: the kernel trailer
  `Assisted-by: Claude:<model-version>`, never `Co-Authored-By:` for an AI, and
  never a tool's own "Generated with <tool>" footer. `scripts/check-attribution.sh`
  holds the logic and runs locally as `make check-attribution` for the commit half.
  The convention had been prose since the beginning and drifted anyway — a
  reconciliation sweep across 78 pull requests was needed once before, after PR
  bodies picked up a tool's default footer. The body half is the half that slips,
  because it comes
  from a tool default rather than from a contributor's habit, so the trigger
  includes `edited`: a body corrected or broken after opening is re-checked. Bot
  authors are exempt — announced in the log rather than silent, and applied inside
  the steps so the job always runs and the check always reports, which is what makes
  the workflow safe to mark required: a required check that never reports leaves a
  pull request waiting on it forever. The gate binds only once it is added to the
  branch's required status checks.

- **A dependency bump that lands on a release workflow can be carried into the
  template it was rendered from** (iss-209). `.github/workflows/release.yml` and
  `auto-release.yml` are rendered from
  `internal/core/launch/scaffold/templates/*.tmpl`, and `TestSelfScaffoldParity`
  holds the two byte-identical so that abcd's own release exercises the machinery
  a managed repo receives. Dependabot sees only the rendered side — its
  github-actions ecosystem discovers `.github/workflows/*` and composite action
  manifests, and no ecosystem scans a `.tmpl` under `internal/` — so every action
  bump it opens broke that parity with no mechanical way back. `make
  scaffold-sync` now propagates the pinned refs from the committed workflows into
  the templates, and `make scaffold-sync-check` reports the same drift without
  writing. Only the ref moves: indentation, list markers and version comments
  survive byte-for-byte, an action the workflow does not pin is left alone, and an
  action pinned to two different refs in one workflow is refused rather than
  guessed. The propagation is one-way by design — re-rendering the template over
  the workflow would revert the bump. `TestSyncRepoPinsIsCleanToday` fails
  preflight when a workflow pin has moved without the template following, and
  names the command that fixes it.

### Fixed

- **The guard now reads the execute-a-string wrapper family, and a warn is no
  longer silent on the hook** (iss-200, iss-231). A command carried as a data
  argument — `sh -c '<payload>'`, `bash -lc "<payload>"`, `eval '<payload>'`, and
  GNU `env -S<value>`/`--split-string` — used to sail past every blocker: the
  payload was one opaque token the matchers never looked inside, so `env -S 'gh
  repo delete owner/repo'` and `sh -c 'git push --force'` were accepted. The guard
  now expands each payload once and matches it as if it had been typed inline, so a
  hazard inside blocks (or, for a warn-tier hazard, warns) exactly as the bare
  command would. The `sh`/`bash`/`dash` `-c` command string is read as the shell's
  first non-option operand, so an option wedged after `-c` (`sh -c -x '<payload>'`,
  `bash -c -- '<payload>'`) can no longer hide it, and `eval` drops a leading `--`
  before joining; when options make the operand impossible to locate the guard warns
  rather than allowing. `env -S` is read on the raw tokens in every spelling — separate,
  glued, `--split-string=`, its abbreviations, and bundled short clusters — at
  every `env` in a wrapper chain, and its value is split only when it decodes to a
  provably plain command; anything else (an expansion, a quote, a leading option, an
  escape env itself would reject) is refused rather than guessed. A payload the
  guard cannot read splits by posture: an uninspectable `sh -c`/`bash -c`
  (a `$(...)` substitution, a pipe into an interpreter) is a loud warning, while an
  uninspectable `env -S` or a nest deeper than two layers is blocked outright. And a
  warn on the pre-tool-use hook now exits non-zero-but-non-blocking so its message
  is actually seen — a hook that exits zero has its stderr discarded, so every warn
  (for example `git reset --hard`) had been running as if allowed with nobody told.
- **The one instruction that resolves the no-binary-on-`PATH` state can now be run
  in that state** (iss-207). The bootstrap's success notice and the README both
  said to run `abcd ahoy install` once — a command whose whole premise is that
  `abcd` is not a name the shell can resolve, so it failed with "command not
  found" for precisely the reader it was written for. On the first manual install
  the consequence was not cosmetic: the agent reading the notice could not run the
  printed command, invented a `go run` incantation reaching into the harness's
  plugin cache, and told the user to run that instead — a source-build path
  needing a Go toolchain, and not the documented install at all. The notice now
  prints the absolute plugin-root path the script already holds, shell-quoted so
  a plugin root containing a space, an apostrophe, a `$` or a backtick still
  pastes as one word, and with the invocation last on the line so it stays
  copy-pasteable to the end. The README carries the same form with the one part a
  committed file cannot know left as a placeholder, says what that placeholder is
  in host-agnostic terms, and points a reader who cannot instantiate it at the
  install one-liner, which needs no plugin root. CI holds both surfaces — every
  `ahoy install` either one prints must be reached through a path, not a bare
  name, and the printed command is handed to a real shell against a hostile path
  to prove it runs as pasted — while the end-to-end reading of it on a real
  plugin cache remains the manual install gate.

- **The session-start hooks run after the bootstrap that provisions their binary,
  and a successful install reads as success** (iss-204, iss-208). The hook
  manifest listed the bootstrap and the two binary-backed commands as three
  sibling `SessionStart` entries and relied on list order; the harness runs every
  hook matching an event in parallel, so both gated entries raced a ~10.7 MB
  download, lost, printed "the plugin binary is not installed", and genuinely did
  not run — on every fresh install and every plugin update, since an update lands
  in a fresh cache directory with no binary. The three entries are now ONE
  command that runs the bootstrap and then both binary calls in a single shell,
  so the sequencing is owned by the manifest rather than assumed of the harness.
  Chaining them makes two further properties load-bearing, and both are held
  explicitly: the hook payload is read once and piped to each call separately,
  because every hook verb consumes the whole of stdin and a shared stdin would
  leave `session-start` reading EOF and silently disabling its notices; and
  `session-start` runs ahead of `prompt-router-reset`, whose unconditional
  success diagnostic would otherwise be the one line the transcript renders.
  The bootstrap's own message is emitted first, which is what the transcript
  renders: on a fresh install the visible line is the checksum-verified success
  rather than one of two missing-binary complaints, and the two complaints
  collapse into one. The honest-failure posture is unchanged — a refusal keeps
  its message and its exit code, a binary that is genuinely absent is still said
  out loud, and the binary calls' stdout still reaches the model untouched. The
  spec that shipped the bootstrap carried the false warrant ("ordering within one
  event's hook list is preserved by the harness") as a load-bearing claim; it is
  corrected in place, with the brief's two descriptions of the manifest. Parallel
  hook execution and the plugin cache are not present in CI, so the end-to-end
  proof is the manual install gate; what CI holds is the manifest's shape and the
  chained command's behaviour against fixtures.
- **`ahoy install` prompts read a piped answer, in a fixed order, and `--yes` says
  what it does not cover** (iss-167, iss-166). The prompter attached to stdin only
  when stdin was a terminal, so `yes | abcd ahoy install` — the first thing an
  agent reaches for — arrived as a decline on every question, and the interactive
  path could not be driven at all: the agent reported failure and handed the step
  back to the human. Prompts now read stdin whether or not it is a terminal, and
  off a terminal each question's answer is echoed to stderr, so a piped run leaves
  a transcript of what was asked and answered. Piped answers are positional, so
  the approval questions are now asked in a **fixed order** — dependency,
  safe-autocreate, config-change, user-state, plugin-owned, the order the apply
  pass acts in — where the walk previously ranged over a map and handed out a
  fresh permutation on every run: the same command approved a different category
  each time, exiting 0 and reading as a clean install. One line answers one
  question, which is why `yes` is the documented form. The safe default is
  unchanged:
  answers that run out read as EOF, and EOF declines every confirm and takes the
  default for every prompt, so an unattended run still adopts nothing it was not
  told to adopt. The interactive path at a terminal is untouched. Folded in:
  `--yes` deliberately does not adopt the optional git-identity pin — the pin
  records whatever identity is currently configured, and a blanket approval would
  canonicalise a sandbox or agent identity, the very value the identity gate
  exists to reject — but it reported "already up to date" without mentioning the
  skip. The exclusion is now stated in the flag's own help, carried in the install
  envelope as `optional_skipped`, and printed with the way to apply it —
  `yes | abcd ahoy install` — which the piped answer makes available to a
  non-interactive caller for the first time. A run that must neither block nor
  prompt closes stdin and pre-answers
  (`abcd ahoy install --yes --refuse-adopt < /dev/null`); the plugin surface says
  so, because reading a non-terminal stdin means a stdin held open and silent
  makes a prompt wait rather than decline.
- **An `ahoy install` receipt is safe to paste** (iss-177). Every apply step
  reported its write as an absolute path and the CLI printed them verbatim, so a
  receipt pasted into an issue or a transcript carried the developer's home
  directory and username — while the sibling verbs already routed their error
  text through a shared path scrub the receipt did not use. The receipt now
  reports a repo write repo-relative (`.abcd/config.json`) and a user-scope write
  home-relative (`~/.abcd/history/index.json`), leaving a location that names no
  developer (`/usr/local/bin/abcd`) exactly as written — the same limit the error
  scrub already states, through the same primitive, which moved to
  `internal/fsutil` so the two cannot drift. The scrub sits at the one seam every
  step reports through rather than in each step's string, and a test holds that
  seam to being the only writer of the receipt, so a step added later cannot
  reintroduce an absolute path by forgetting.
- **The command surface reaches the binary a plugin install actually provisions**
  (iss-205). Every command file resolved the binary as a bare `abcd` on `PATH`
  with a `go run ./cmd/abcd` fallback, and none named `${CLAUDE_PLUGIN_ROOT}` —
  while the bootstrap hook installs its checksum-verified binary *into the plugin
  root* and leaves nothing on `PATH`. The two halves never met: a fresh install
  worked only because the marketplace clone happened to carry `cmd/`, costing 54
  seconds and a Go toolchain, and on a machine without Go the whole `/abcd:*`
  surface was non-functional despite a healthy binary sitting in the plugin root.
  The resolution ladder in all 17 binary-invoking command files now runs
  `"${CLAUDE_PLUGIN_ROOT}/abcd"` first, `abcd` on `PATH` second, and `go run
  ./cmd/abcd` third and explicitly only in a source checkout — the published
  payload carries no `cmd/`, so an unqualified third rung prints an instruction a
  plugin user cannot follow. Every fenced command line, which is what an agent
  runs verbatim, carries the plugin-root form.
  `TestCommandSurfaceResolvesBinaryFromPluginRoot` keeps it that way: it fails if
  any file under `commands/` names the binary without the plugin-root rung first,
  hands over a fenced invocation that resolves any other way, leaves a `go run`
  rung unqualified, or drops the ladder paragraph. It runs under `go test ./...`,
  so `make preflight` and CI already execute it rather than needing a target
  anyone can forget to wire.

- **The build plumbing's own comments describe the gate suite that runs**
  (iss-182). The `Makefile` preflight comment claimed the target ran "the same
  steps CI's check job runs" and named only the reviews-charter gate, though the
  recipe takes three lint prerequisites (`lint-reviews`, `record-lint`,
  `docs-lint`) and CI additionally runs a `gofmt -l .` format gate that preflight
  does not; the `.githooks/pre-push` comment called what the hook enforces a
  "check-job trio", undercounting those lint gates, and listed only the
  secret-scan and workflow-audit lanes as Actions-only; and the `ci.yml` header
  omitted the check job's gofmt, record-lint and docs-lint steps along with the
  reviews-charter and smoke jobs. All three now state what actually runs.
  Comment-only — no recipe, hook logic, workflow step, or gate behaviour moves.

- **The record describes the `.abcd/**` exclusion by the channel it is true of**
  (iss-183). The blanket present-tense claim — `.abcd/**` "never ships", or is
  "excluded from the release artifact by packaging" — survived across
  `CONTEXT.md`, `AGENTS.md`, both `.abcd/` READMEs, four brief sections, the
  release glossary term, the Phase 1 expectation and a scanner code comment,
  after the README alone was corrected. The exclusion is implemented but has
  never run on a release: the launch bundler denies the `.abcd` namespace
  structurally, while a marketplace install takes the repository root and GitHub
  attaches an auto-generated source archive to every release, so only the
  released binaries omit the directory. Each descriptive instance now says so —
  present in every repository checkout, marketplace installs and release source
  archives included, never in the released binaries — and names the bundler as
  the implemented mechanism it is. Decision records, dated plans and intent
  bodies keep their original wording.

- **Merging a rule-set overlay onto a base that declares no domains no longer
  panics** (iss-187). `rules.Merge` promises that new domain keys are added, but
  it wrote them into a map it never allocated when the base carried no domains of
  its own — a valid rule set that the validator accepts — so the merge crashed
  instead of returning the overlay's domains. It now allocates that map before
  adding the overlay's keys, matching the guard registry's loader. No behaviour
  changes for the rules a repo loads today, where the base is always the bundled
  default set.

- **`abcd capture resolve`/`wontfix` no longer strand an issue across two status
  directories on a failed move** (iss-186). The transition wrote the destination
  file, then removed the source; if the removal failed for a reason other than
  "already gone" (a read-only remount or a restrictive attribute on the source
  directory), the error surfaced after the destination already existed, leaving
  the same issue id present in both `open/` and its target directory. From then
  on every later transition on that id was refused as a duplicate, with no
  repair verb — the file had to be deleted by hand. The move now rolls the
  destination back when the source can't be removed, so a failed transition
  leaves the ledger exactly as it was before the attempt and a retry is all
  that's needed.

## [0.4.2] - 2026-08-06

### Added

- **`context_citation_currency` — the orientation doc cannot ground a live
  caveat in a record that is finished** (iss-42). `CONTEXT.md` is the first file
  a session reads and its sharp-edges list is the paragraph a reader trusts
  before they have read anything else, yet nothing required that list to be
  revisited when the record it cites moved — so the staleness was structural, not
  incidental, and recurred every time a cited record shipped, resolved, closed,
  or was superseded. The rule resolves every `iss-N`, `itd-N`, `spc-N`, and
  `adr-N` handle the sharp-edges section names against the store that holds it,
  and blocks any citation whose target sits in a terminal lifecycle state —
  `resolved`/`wontfix` for an issue, `shipped`/`superseded` for an intent,
  `closed` for a spec, and a declared supersession or a retired status for an
  ADR, whose flat store carries its lifecycle in frontmatter rather than in a
  directory. Scope is one section of one document on purpose: a shipped intent's
  evidence trail, a supersession chain, and a dated plan all name closed records
  legitimately, and only the living orientation doc claims to describe what is
  true right now. A handle that resolves to no record is left alone — dangling
  cross-references are `record_schema`'s question, already answered there. It
  fails closed on the three ways an armed gate could check nothing: no target to
  read, no sharp-edges section in it, and configured stores that resolve no
  records at all, which reads exactly like a clean section. Any abcd-managed repo
  enables it by declaring its own `target` and `record_stores` in
  `.abcd/record-lint.json`, and may name a different section with `section`.
- **The development record map's `research/` row is derived rather than
  hand-kept.** An `index_drift` region holds the row to the directory's actual
  subdirectories, so a routing claim naming a child that does not exist — or
  omitting one that does — fails the record gate instead of quietly misdirecting
  a reader.
- **`delivery_state` — the changelog cannot credit an intent the record calls
  unbuilt** (iss-41). An intent in `drafts/` is a captured idea nobody has
  committed to build, so the intent tree and a delivery entry citing it say
  opposite things about the same work — and the tree is the side nobody re-reads.
  The rule reads each version entry's delivery sections — `Added` and `Changed`,
  plus any heading a repo names in `delivery_sections`, which is unioned with
  those rather than substituted for them so a config can widen the gate but never
  narrow it — and blocks any `itd-N` citation whose intent still sits in
  `drafts/`. Non-delivery sections are out of scope on purpose: an id
  under `Fixed` is normally provenance for a defect — which draft two branches
  minted at once — not a claim that the intent is built. Its two remedies are the
  two truths a finding can be reporting: the intent shipped whole and was never
  promoted out of `drafts/`, or the entry credits an intent for less than it
  promises, in which case it describes the capability without the citation, since
  an intent is delivered whole or not at all. It fails closed on the three ways an
  armed gate could check nothing — no changelog to read, no intents store to
  resolve against, and a store holding none of the lifecycle buckets, which is
  what a root pointed one level too deep looks like and reads exactly like a clean
  corpus. Any abcd-managed repo enables it by declaring its own `changelog` and
  `intents_root` in `.abcd/record-lint.json`.
- **`index_drift` gains `dir_entry`, so a listing can enumerate records by id.**
  The rule compared a document's entries against whole filenames, which only ever
  agreed when the document transcribed slugs — a listing nobody writes by hand.
  `dir_entry` is the mirror of `entry` on the directory side: a regexp that
  reduces each file's stem to the part the document enumerates — a record's id
  rather than its id-plus-slug filename — and drops any stem it does not
  describe. That is what lets the brief's later-phase list be gated by the same
  one rule instead of a second one shaped like it. It is rejected in `absent`
  mode, which resolves listed paths rather than reading a directory.
- **`record_schema` — the design record is checked across its stores, not just
  inside each one** (iss-39). Every existing record rule asks a question about one
  store: is this intent's bucket schema right, does this spec agree with its
  intent. None of them asked the questions that only make sense between stores,
  which is where a record actually drifts. The new rule reads the ADR, intent,
  spec, and issue stores in one pass and holds four invariants: a cross-reference
  field (`supersedes`, `related_adrs`, `related_intents`, `builds_on`,
  `blocked_by`) names a record the corpus has — or one a successor declares it
  superseded, since a pruned record is accounted for rather than lost; a
  supersession is declared from BOTH sides, so a record can no longer contradict
  itself about which decision is in force; a filename and the id inside it agree,
  so one record cannot answer to two handles; and every directory that should
  hold records is enumerated — an undeclared bucket in a store root, and a
  subdirectory inside a bucket or a flat store — so a directory nobody declared is
  a finding rather than a place records hide from every check. A supersession may
  cross stores — an ADR that redecides the question an
  intent rested on retires that intent — so `superseded_by` accepts `adr-N` as
  well as `itd-N`. That widening makes `intent_lifecycle`'s own target check
  looser on its own: it reads only the intent tree, so a repo arming it WITHOUT
  `record_schema` accepts an `adr-N` successor without resolving it. The two are
  meant to be armed together, and `record_schema` is what resolves a cross-store
  target. Cross-references are checked in frontmatter rather than in prose on
  purpose: a frontmatter handle is a machine-readable claim that the record
  exists, while prose legitimately names ids that do not resolve. A retirement
  declaration is bounded by what the store has issued, so `supersedes:` cannot
  mint a phantom id that then resolves everywhere else. Any abcd-managed repo
  enables it by declaring its own `record_stores` in `.abcd/record-lint.json`.
- **`index_drift` — a hand-written directory index is gated, not trusted**
  (iss-38). A README that enumerates its sibling files by hand is a second copy
  of something the filesystem already knows, and it drifts the moment a file is
  added, renamed, or shipped — silently, because nothing was checking. The rule
  reads a region a document fences with `<!-- index: <id> -->` and holds it to
  the directory it enumerates: in `exact` mode the listing and the directory must
  agree in both directions, so a file added without a line and a line kept for a
  file that has gone are each a finding; in `absent` mode every listed path must
  still be missing from the tree, which is the shape a "planned seams" list has,
  where the drift is a seam that shipped while the list still calls it planned.
  Scoping is by explicit marker plus a configured entry pattern rather than by
  parsing arbitrary markdown, so unrelated prose around a list cannot produce a
  finding and no bespoke parser is needed per README. It fails closed on the
  three ways a gate could be quietly disarmed — a region deleted while its config
  entry remains, a region that parses to no entries at all, and a malformed index
  spec (no document, no directory, no entry pattern, an uncompilable pattern, an
  unknown mode) — the last as a loud configuration error rather than a pass. Any
  abcd-managed repo enables it the same way: an `index_drift` block naming its
  own document/directory pairs in `.abcd/docs-lint.json`.
- **`abcd banlist` — the names a repo must not publish, in two layers** (itd-74,
  spc-20). Enforcement splits by sensitivity, because a deterministic CI gate is
  the right tool for a public banned name and the wrong place for a private one:
  the rule would have to contain the very string it forbids. The **public** layer
  is the `banned_tokens` family of `.abcd/docs-lint.json` — the same primitive
  that already gates this repo's harness names, not a second mechanism — with
  verb-written entries under a `names/` id prefix that marks what the verb owns:
  `list` renders the whole family, and a removal is refused for a hand-curated
  entry. Config edits are byte surgery on the located array rather than a
  re-marshal, so an add is one inserted line, a remove is one deleted line, and
  add-then-remove returns the file to its exact bytes. The **private** layer is a
  gitignored per-machine store read by the committed pre-commit guard, and its
  visibility follows: entries render by key only, never their pattern, and the
  redaction is structural — the entry type carries no pattern field, so no
  rendering can leak one. A private pattern is entered by piping it on stdin
  (`printf %s 'PATTERN' | abcd banlist add --private KEY -`), which is the
  recommended form because an argument is world-readable in `/proc/<pid>/cmdline`,
  is captured by process auditing, and lands in shell history. Each layer is
  validated against the engine that enforces IT — a private pattern by the guard's
  own grep, a public one through the linter's compile path — so an entry cannot be
  stored as healthy while it matches nothing; and `add --private` refuses outright
  if git does not ignore the store's path, since the layer rests on that file being
  untracked. `add` and `remove` name their layer explicitly (neither flag and both
  flags exit 2); bare invocation and `list` are read-only, both state their reach
  plainly, including that CI cannot enforce the private layer, and `list --private`
  separates a line the guard cannot use from one it accepts but reads differently,
  because the first stops every commit and the second stops nothing.
- **`abcd ahoy` scaffolds the whole two-layer name banlist** (itd-74, spc-20). A
  repo becomes name-safe by being abcd-managed rather than by a maintainer
  hand-wiring the files. Install writes FIVE artefacts, each reported by name on
  the status board: `.githooks/pre-commit` (the guard), `.githooks/pre-merge-commit`
  (git runs no pre-commit for a merge commit, so a banned name would otherwise walk
  into history the moment a merge commit carrying it is made), an appended
  `.gitattributes` line keeping both hooks at LF (a `core.autocrlf` checkout
  rewrites a script git executes, and its shebang stops resolving),
  `.abcd/docs-lint.json` carrying an empty public banned-names family, and the
  documented private stub in the gitignored local tier. A clone arms the hooks once
  with `git config core.hooksPath .githooks`; abcd never sets it, and no surface
  reports a committed hook as a running one. The merge half is written ONLY beside
  abcd's own guard, identified by a whole `# abcd-name-guard: v1` line: beside a
  maintainer's own hook it would both claim coverage it has not got and silently
  start running that hook on merge commits, which git never did. The public family
  is seeded EMPTY on purpose — abcd cannot know which names a repo may not publish,
  and a ban nobody declared would fail a build over a word the maintainer never
  chose. Every write is create-if-absent and contained: a hook, a CI-gating config,
  and above all a populated private store are the maintainer's, and paths resolve
  through a containment root so a symlink committed at `.githooks` or at the local
  tier cannot land an artefact outside the repo while the surfaces report the
  in-repo path. The stub is written only where `git check-ignore` itself reports the
  store's path as ignored, not where a comparison of `.gitignore` text suggests it:
  a repo can carry a byte-perfect block and still track the store, and a stub git
  would track is the hazard rather than the remedy. Where git cannot be asked at all
  — missing from PATH, a corrupt `.git` — a repo-shaped directory fails closed
  rather than borrowing a plain folder's answer. The public config is held to the
  same rule: abcd does not write one into a path git ignores, because it would have
  to report it unenforceable in the same breath. The stub's worked examples are all
  commented out — a fresh scaffold parses to zero entries, and the guard says so
  loudly at commit time instead of looking like protection — and every illustrative
  value in it is a reserved documentation value (RFC 5737, RFC 3849, RFC 2606, RFC
  7042) or a persona-derived fixture host, judged by the repo's own
  network-identifier detector.
- **The name guard refuses copies of the private store, with a published escape**
  (itd-74, spc-20). The store-path refusal matched only the local tier, so a COPY of
  the private banlist anywhere else — `notes.txt`, a `.bak` beside it, or a `git mv`
  out of the tier — committed every pattern in clear while the guard announced a
  clean check: the entries cannot catch their own text, because they are escaped
  regular expressions and a pattern does not match itself. Three tests now: a staged
  path inside the local tier (including a rename's source path, and not escapable),
  a staged blob whose first line is the format declaration, and a staged path whose
  basename is the store's filename. They are shape tests on a mistake rather than a
  net against someone determined — a copy with the declaration stripped or displaced
  still commits — and they carry a per-file escape, because a repo that legitimately
  commits a store-shaped file (a fixture corpus, a doc quoting the declaration) needs
  one that is not `--no-verify`: a second line reading `# abcd-banlist-example`
  exempts a blob from the copy refusals and from nothing else, its content still
  scanned against every entry. The guard also pins `PATH`, `IFS` and xtrace and
  unsets any inherited shell function shadowing a command it runs, as its first
  statements, and BLOCKS loudly when a tool it needs is missing rather than failing
  as a mute exit 127; a machine with a nonstandard prefix extends the pin with
  `git config --local abcd.guardPath <dir>`, never an environment variable, since a
  repo-scoped environment is the hole the pin exists to close. The format declaration
  must be line 1 or nowhere: a blank line, a comment, a duplicate below it, or any
  prefix bytes before it is a damaged declaration rather than a silent downgrade to
  the legacy format, and both readers say so identically.
- **Every surface that describes the private layer states its reach** (itd-74,
  spc-20). `abcd ahoy` now reports the name guard's state — what occupies each hook
  path, whether the public family can actually be enforced, and what shape the
  private layer is in on this machine — and the status line, the JSON envelope, and
  the `abcd banlist` verb all carry the same sentence: CI cannot enforce the private
  layer; it protects only machines that have opted in, and only the commits git runs
  a hook for, so a fast-forward `git pull`, a rebase, a `git am`, a `git revert` or a
  cherry-pick bypasses it, as does `--no-verify`. The list is explicitly
  non-exhaustive and names nothing that is in fact covered. The reach travels inside
  the reported state rather than being added by a renderer, because a machine
  consumer reading "hook committed" beside a present store would otherwise draw
  exactly the wrong conclusion. Where a claim cannot be supported it is withdrawn
  rather than softened: a docs-lint config git ignores — the state
  `visibility: public` puts every repo in, since the installed fence ignores the
  whole `.abcd/` namespace — is reported as NOT ENFORCEABLE instead of as the
  committed, CI-enforced layer, with the placement question left for a maintainer to
  settle (iss-176). The status pass reads the private store's shape (its format and
  two counts) and never its content: the patterns are the secret, and a status board
  is the surface that must not hold them.
- **The private name guard refuses by key and says when it is inactive** (itd-74,
  spc-20). The committed `.githooks/pre-commit` guard checks the CONTENT of every
  staged file, read out of the index, and on a match refuses the commit naming the
  entry key alone: the matched text and the pattern value never reach stdout,
  stderr, or a log, because a refusal that echoed the string would defeat the layer
  at the moment it worked. The pattern reaches grep on stdin rather than in argv,
  for the same reason. Hostnames, IP and CIDR values, MAC addresses, and device
  names are ordinary entries, matched exactly as a name is, and so are binary
  blobs — a name in one is in history just the same. The store declares its own
  format on its first line: `# abcd-banlist: keyed` means every line is
  `KEY<space-or-tab>PATTERN`, and no declaration means every line is one whole-line
  pattern under a synthetic key, so an older store keeps matching exactly what it
  always matched and no part of any line is ever read — or printed — as a key. A
  line that does not parse, a pattern the engine refuses, and any git step that
  fails are each a refusal naming a step or a line number: an unusable entry is
  never skipped, and a check that could not run must never look like a check that
  passed. A store that is absent, or present with no entries, prints a loud warning
  that the layer is inactive on this machine and lets the commit through: it
  protects machines that opted in, and silence must never impersonate protection.
  It reads each staged blob stage-explicitly (so a file literally named
  `0:README.md` cannot hide behind git rev-magic), scans the staged PATH strings as
  well as content (a banned name in a filename enters history just the same), skips
  a staged gitlink rather than fail-closing on a submodule it cannot read, refuses
  to commit the private store itself, and announces the format and entry count it
  read before the scan so a stripped format declaration cannot silently downgrade it.
- **A citations family in `abcd docs lint`, with zero network in the gate**
  (itd-101, spc-17). Cited references rot silently — pages retitle, URLs
  redirect, whole platforms announce their own shutdown — but a gate that dials
  out to notice is a gate that flakes. So the checking splits in two. The lint
  side, landing here, reads only committed markdown, committed config, and a
  committed baseline: `citation_footnotes` holds a page's footnote markers and
  definitions in bijection (an unreferenced definition counts, being the way a
  reference page quietly stops meaning what it says); `citation_crosswalk_rows`
  requires every crosswalk table row to carry a footnote; `citation_url_syntax`
  checks cited URLs and DOIs are well-formed; and `citation_source_policy`
  refuses aggregator domains named in config — a list that ships empty, because
  naming one is a project's editorial policy and never something the gate
  invents on its behalf. A page's citations are its footnote definitions, not
  its prose, so ordinary body links stay `links_resolve`'s business.
  `citation_baseline` enforces the committed record offline — no cited URL
  without an entry, none recorded broken, none whose recorded final address has
  drifted from what the page cites, and a 180-day staleness warning. Each rule
  is opt-in per repo, and the whole family is inert until configured.
- **`abcd docs cite refresh` — the one verb that fetches, so the gate never has
  to** (itd-101, spc-17). It collects every cited URL through the same collector
  the lint uses, gives each exactly one bounded attempt, and rewrites the
  committed baseline. It never retries — a run's cost cannot become a function of
  how many links are failing, nor a burst against a struggling host — and it never
  reads a response body, because liveness is a property of the status line and a
  citation to a huge file should cost a response header, not a download. Sources
  that refuse automated fetchers (401, 403, 406, 429) are **not** recorded as
  broken: a refusal says the fetcher may not look, which is a different fact from
  the source being gone, so those URLs get no invented entry and are printed as a
  manual checklist naming exactly where each is cited. A current human-verified
  receipt is preserved verbatim and not even re-requested; once it ages past the
  staleness threshold it is re-checked like any other, because human and machine
  verifications share one clock. Receipts for addresses the documentation no
  longer cites are dropped.
- **`abcd docs cite confirm` — recording that a human cleared a queued link.**
  Name the URLs, or pass `--receipt` with a receipt file; both assemble the same
  schema, so the generated checklist page that lands later is a different producer
  of one input rather than a second pathway. Only URLs the documentation actually
  cites can be confirmed, and one bad line refuses the whole receipt rather than
  half-applying it. The receipt records **that** a human verified a citation and
  **when**, never how — the schema declares no such field and decoding rejects
  unknown keys outright.
- **The citation baseline surfaces where maintainers already look.** `abcd ahoy`
  reports its coverage and age on one line in any repo that has armed the rule;
  `abcd launch --dry-run` gains a `citation-baseline` gate that names entries
  approaching the staleness blocker while still letting a release cut, and refuses
  on ones that are overdue, broken, or unreceipted. `abcd docs lint
  --release-gate` promotes an overdue citation from a warning to a blocker — the
  flag is the trust root, so a repository cannot defang its own release by editing
  a committed config, and an ordinary commit is never blocked by the calendar.
  The flag is built and tested but **not yet passed by `release.yml`**, which
  still runs the plain `abcd docs lint`; wiring it into the release workflow is a
  CI change needing its own sign-off, and until it lands the 365-day threshold
  warns at release time rather than blocking.
- **The citation gate is armed for this repository.** 50 of the 51 URLs cited
  under `docs/` carry a receipt, and four citations that had silently drifted
  behind redirects now name the address they actually resolve to — rot nobody
  would have caught by reading. The fifty-first answers HTTP 403 to every
  automated fetcher and is waiting in the manual queue, so `abcd docs lint`
  reports it as unreceipted until a maintainer opens it and runs `abcd docs cite
  confirm`. That report is the mechanism working: no receipt is ever written on
  a human's behalf.
- **A schema-versioned citation baseline at `.abcd/citations-baseline.json`.**
  Per cited URL it records the final resolved address, when it was last checked,
  the outcome, and whether verification was automatic or manual with its date.
  It records nothing about *how* a human verified, and cannot be made to: the
  schema declares no such field, and loading rejects unknown keys outright, so a
  hand-added `method` or transcript is a refusal rather than a quietly-kept
  note. Manual entries age on the same clock as automatic ones, so a human
  confirmation buys no exemption from going stale.
- **The hazard registry refuses destructive GitHub remote operations and a force
  push spelled as a refspec** (iss-159, iss-148). Three shapes that verdicted
  `allow` now block. `gh repo delete` destroys the remote copy of a repository
  and everything kept with it — issues, pull requests, releases, review history —
  and nothing an agent can run brings any of it back, so the refusal names the
  human-only successor: tell the person who owns the repository, and archive it
  (`gh repo archive owner/repo`) where retiring it rather than destroying it was
  what was meant. `gh api -X DELETE repos/{owner}/{repo}` is the same deletion
  written as a raw API call, with no confirmation prompt anywhere in the way; it
  is depth-limited on purpose, so a `DELETE` deeper inside a repository — a
  branch ref, a release — is ordinary work and stays allowed. The path operand is
  normalised before that depth check — a `scheme://host` prefix, a query, and a
  fragment are dropped — so the same call written as
  `https://api.github.com/repos/owner/repo` is the same refusal; what remains
  stated rather than closed is a host that serves the API under a mount prefix
  (a GitHub Enterprise Server `/api/v3/…`), because matching a root segment
  wherever it appeared would falsely refuse
  `DELETE /teams/{id}/repos/{owner}/{repo}`, which removes a repository from a
  team and destroys nothing. And
  `git push origin +main:main` overwrites the remote branch exactly as `--force`
  does, without the flag anyone reviews for. Entry patterns gained the fields
  those shapes need, each optional and overridable per repo like the rest: a
  second-level subcommand (`gh repo list` and `gh repo delete` share their first
  level and only one of them is a hazard), a flag's SETTING rather than its
  presence (`-X DELETE` refused, `-X GET` allowed, all three spellings read and
  the case of the value ignored), an operand's path root and exact depth, and an
  operand prefix. Every new entry ships through the same admission gate as the
  rest: known-bad and known-good fixtures in the registry file itself, a 100%
  true-negative floor, and known-good at least 40% of its own corpus.
- **The plugin provisions its own binary, so a fresh install and every update
  yield a working hook surface** (itd-105, spc-21). The hooks call
  `$CLAUDE_PLUGIN_ROOT/abcd`, the plugin root is a clone of a repository that
  commits no binary, and the harness re-clones each update into a fresh
  commit-stamped cache directory — so every install and every update produced a
  plugin root with nothing to call, and each hook failed with a raw "No such
  file or directory" until someone hand-copied a binary in. `hooks/bootstrap.sh`
  closes that: committed POSIX sh, needing no abcd binary to run (the binary is
  exactly what is missing), wired as the FIRST `SessionStart` hook so it lands
  before the binary-backed ones in the same event. It downloads the latest
  release binary for the host platform plus `checksums.txt`, verifies the
  SHA-256 against the manifest, and installs it with an atomic rename from a
  temp directory on the same filesystem — a mismatch or an absent manifest line
  deletes the download and refuses, so a corrupted or unpublished artefact never
  reaches the binary path. The trust bar is the README one-liner's: same-origin
  checksums, with `go build ./cmd/abcd` documented as the full-trust route. The
  script fetches from two hardcoded origins and offers no way — no environment
  variable, no flag, no branch — to point it anywhere else, and it pins HTTPS
  including redirects on every request unconditionally: the binary and the
  `checksums.txt` that verifies it come from one origin, so anything able to
  name that origin would supply both the payload and its own manifest, and what
  is installed then runs unattended as the Bash shell guard on every tool call.
  Pinning the transport is not enough on its own, because `curl` reads a
  configuration surface the command line knows nothing about: every fetch
  therefore passes `-q` as its first argument, so `$CURL_HOME/.curlrc` and
  `$HOME/.curlrc` are never loaded — one `connect-to` or `resolve` line there
  re-points the connection while the URL still reads `https://github.com/…`, and
  the checksum then verifies the substitute against its own manifest. The names
  `curl` reads without being told to are removed from the environment before the
  first request for the same reason: the proxy variables (`HTTPS_PROXY`,
  `ALL_PROXY` and their lowercase forms, plus `HTTP_PROXY`/`http_proxy`) and the
  certificate-authority overrides (`CURL_CA_BUNDLE`, `SSL_CERT_FILE`,
  `SSL_CERT_DIR`) that would make such a route succeed on TLS. **The accepted
  cost: a machine that can only reach the network through a proxy does not
  bootstrap automatically** — install the release binary by hand or build from
  source with `go build ./cmd/abcd`, both of which the refusal message names. A
  plugin root that already holds an executable binary exits on one file test
  with no network, so steady-state sessions pay nothing; concurrent sessions
  serialise on an atomic `mkdir` lock whose loser exits quietly and whose stale
  remains (older than ten minutes, from a killed run) are broken and retaken. A
  lock that cannot be taken and is not there afterwards is not a lost race —
  the plugin root is unwritable, or something that is not a directory occupies
  the lock path — and that says so out loud instead of sharing the loser's
  silence, because it repeats every session and nothing else would report it. A
  platform outside the released matrix — darwin and linux on amd64 and arm64 —
  is reported, not retried: it states the matrix and changes nothing.
  Every failing path prints one plain-language message naming what is missing,
  what it costs (hooks cannot run, the shell-hazard guard is inactive and
  commands run UNGUARDED), and the three ways out; `abcd ahoy`'s guard health
  names the same script and its recovery step, so one fault has one story
  wherever it is read. Every message the bootstrap has for a person — the
  unsupported-platform statement and the one-time `abcd ahoy install`
  suggestion included — goes to stderr with a non-zero (non-blocking) exit,
  because a SessionStart hook's stdout becomes model context while only a
  non-zero exit puts its stderr in front of the human who can act on it. The
  two binary-backed `SessionStart` commands that follow the bootstrap check for
  the binary first, so a session that begins before one exists reports the gap
  in the same plain language instead of a raw "No such file or directory".
  Every message strips control characters from the values it echoes — a
  plugin-root path, `uname`'s answer, the release tag read off a redirect — so
  an escape sequence in one of them cannot recolour or visually rewrite a
  message whose whole job is to be believed.
- **A binary at a different commit from the surface it serves says so at session
  start** (itd-105, spc-21). The plugin surface tracks the repository tip while
  the newest binary is the last tagged release, so a fix can merge without a
  release cut and leave a session running old-binary logic against new-surface
  expectations. The bootstrap records what it installed in a `.binary-meta` file
  beside the binary — release tag, release commit, the SHA-256 it verified the
  binary against, fetch time, and the plugin commit it provisioned for, written
  by the same atomic rename the binary gets — and the session-start hook renders
  one line when the plugin commit and the release commit differ. The line names
  both directions the difference could run in and asserts neither, because
  comparing two commits establishes that they differ and never which is ahead.
  What the bootstrap could not resolve it records as `unknown`, and a commit
  that is not forty lowercase hex characters — unresolved, or truncated by a
  crash mid-write — produces no line at all: a skew notice that guesses is worse
  than no notice. The plugin commit comes from the cache directory's name, which
  is an assumption about the harness this repository cannot verify, so the raw
  basename is recorded beside the gated value: if that naming ever changes, the
  notice goes quiet for everyone and `.binary-meta` is the one place that says
  why. The release tag is sanitised before it is rendered, being the only value
  in the line that arrives unchecked from off the machine.

### Changed

- **`abcd ideate record` files its verdict record where dated research notes
  live.** The record it writes is a dated research note by spc-18's own
  reasoning, and `research/README.md` files those under `notes/`; it was landing
  in `research/` root, so the verb was a standing producer of a convention
  violation the record then had to clean up by hand. It now writes
  `.abcd/development/research/notes/<date>-ideate-<slug>.md`. The path is still a
  constant, not a parameter — a configurable target would let a verdict land
  somewhere no future session looks.
- **The intent tree's schema rules stop honouring the `superseded/` lint
  exemption** (iss-39). Being historical excuses a record from how it is
  WRITTEN — the banned tokens, the persona roster — never from being
  well-formed. The exemption used to reach the intent-lifecycle schema too, so
  the `superseded/` bucket excused its own records from the one rule that
  checks a supersession is recorded: two intents sat there carrying a prose
  note and no `superseded_by` field at all, invisible for as long as the
  exemption held. Both rules that read the intent tree — `intent_lifecycle` and
  `intent_impact_valid`, which share one scan of it — now run everywhere,
  `superseded/` included; if you arm `intent_impact_valid` with `exempt_paths`,
  an exempt record carrying an illegal `impact:` value is a finding on upgrade
  where it was silent before. The spec-store rules (`spec_lifecycle`,
  `spec_id_unique`) still honour the exemption as before — widening those is
  separate work. The new `record_schema` is cross-store and never consulted the
  exemption at all.

### Fixed

- **The README layout line says where the development record travels**
  (iss-43). `.abcd/` was described as "never shipped", which holds for the
  released binaries alone: a marketplace install takes the repository root, and
  every release carries an auto-generated source archive, so both hold the
  directory. The line divides on the real boundary — any repository checkout
  against the released binaries — and the section on marketplace installs names
  those source archives alongside the binaries and checksums a release
  publishes.
- **The installed plugin surface is the surface the documentation describes**
  (iss-44, iss-160, iss-161, iss-162). The verb files lived in
  `commands/abcd/`, a subdirectory named after the plugin itself, and a harness
  maps each `commands/` subdirectory to a namespace segment — so every verb
  registered as `/abcd:abcd:<verb>`, the plugin name twice, and each documented
  `/abcd:<verb>` was an unknown command. The verb files move flat into
  `commands/`, which also makes the surface's own next-step guidance runnable:
  it already spelled `/abcd:<verb>` throughout, so it was correct prose about a
  name nothing registered. `commands/README.md` shipped as a spurious
  `/abcd:README` because the loader reads every markdown file under `commands/`
  as a command, requiring no frontmatter and exempting no name; the directory's
  documentation moves out of the auto-discovery root into the brief's surface
  registry, which already enumerates the same surface. `/abcd:ahoy` reaches the
  sub-verbs the binary registers — `install`, `uninstall`, `doctor` and
  `dry-run` each have a section, and `identity-check` an explicit note saying
  why it stays CLI-only — where the command file previously drove only the bare
  read-only detect. Every "the `abcd` binary is not on `PATH`" remedy named
  `make build`, which cross-compiles `bin/abcd-<goos>-<arch>` and never a
  PATH-resolvable `abcd`; each now names the `go run ./cmd/abcd …` fallback the
  same paragraphs already carried. And `abcd memory ask` run from a plain
  terminal no longer heads its output with a plugin slash command: the renders
  live in the transport-agnostic core, so they name the binary invocation and
  stay true whichever front door produced them. A surface-parity test holds all
  of it — the command surface is one flat level of markdown commands and
  nothing else, every binary sub-verb is reachable from its command file or
  carries a scoping note, and no remedy offers a build that cannot put `abcd`
  on `PATH`.
- **The developer docs describe the gate suite that actually runs** (iss-37).
  `AGENTS.md`, `README.md` and `CONTRIBUTING.md` name the three lint gates
  `make preflight` runs (`lint-reviews`, `record-lint`, `docs-lint`) alongside
  build, vet, test and the race-enabled internal tests; `AGENTS.md` and
  `CONTRIBUTING.md` attribute the format gate to CI's own `gofmt` step rather
  than to preflight; and the `ci.yml` step comment states that the record-lint
  step blocks.
- **The record can say what has been delivered, and the brief's later-phase list
  is derived again** (iss-41). Two entries credited an intent that never left the
  uncommitted bench: the v0.1.0 `abcd docs lint` entry cited `itd-60`, whose
  deterministic layer shipped while its semantic layer is still five open
  questions, and the v0.2.0 `abcd audit` entry cited `itd-85`, which shipped
  whole and was never promoted. Both entries now describe the delivered
  capability without the citation, and `intents/README.md` states the rule the
  new `delivery_state` gate holds: leaving `drafts/` is what represents delivery,
  and nothing else does. `itd-85`'s missing promotion is `iss-180` — it is
  blocked on the `shipped/` schema, which wants a spec id the audit verb has
  never had. The brief's canonical later-phase list, which claimed to be derived
  from `drafts/` and not hand-counted, had drifted nineteen entries: three intents
  it still listed had moved to `planned/`, `shipped/`, and `superseded/`, and
  sixteen captures since the last hand edit were missing. It is regenerated and
  now sits inside an `index_drift` marked region, so the claim is enforced rather
  than repeated. The four later-phase items that never had an intent id move out
  of the derived list, which could not have contained them.
- **One canonical glossary, and its index is derived** (iss-40). The record held
  two artefacts each calling itself the glossary: `.abcd/development/brief/glossary/`,
  a directory of per-term files with the frontmatter the `GL002` forbidden-synonym
  rule reads, and `02-constraints/04-naming.md`, whose vocabulary-registration
  section declared that terms are registered in *it*. Nothing said which one won
  for a given term, so "the glossary" resolved to two places. The naming file now
  states the split it actually implements — the glossary directory holds
  cross-cutting prose vocabulary, that file's own tables hold the naming
  convention, controlled enums, and spec-pinned reserved names — and stops
  claiming to be a glossary; its ~60 reserved-vocabulary rows are untouched,
  because a closed enum is not a glossary term. The `VR001` descriptions that
  carried the old framing across the brief and itd-37 are corrected with it.
  The glossary's own index was stale in the way a hand-kept index always is: a
  whole bounded context (`distribution/`, three terms) existed on disk and
  appeared in neither the context list, the directory tree, nor the term index.
  Both blocks are now RENDERED from the term files by `internal/core/glossary`
  and held to them by a drift test, so a term file that lands without an index
  row fails the build — the same generate-and-gate shape the brief↔lifeboat
  mapping table already uses. The README also promised a JSON schema
  (`internal/core/schema/terminology.schema.json`), a CLI verb
  (`abcd lint terminology`), and a config allowlist key
  (`terminology_exclude_files`), none of which exist; it now describes what does
  — the frontmatter tables as the shape's own specification, `GL002` through
  the record-lint gate, and the index drift gate — and says plainly that neither is a
  full schema validator. `glossary/core/brief.md` no longer calls the brief
  "immutable once approved", which contradicted adr-5's living-record decision,
  and the brief's own navigation table and directory tree now list `glossary/`,
  which the generated brief↔lifeboat mapping had been alone in knowing about. A
  public banlist entry (`names/record/glossary-self-identification`) blocks the
  retired self-identifying phrase from returning to the published surface.
- **The plugin root is resolvable when abcd runs through the PATH symlink
  `ahoy install` itself pins** (iss-170). Resolution falls back to walking the
  executable's ancestors for a `hooks/` directory, but the executable path was
  taken as invoked rather than canonicalised — and the installed layout puts a
  symlink on PATH (`/usr/local/bin/abcd` -> `<plugin-root>/abcd`), so the walk
  climbed `/usr/local/bin`, `/usr/local`, `/usr` and never reached the plugin
  root (the walk's own guard excludes `/`, so it was never a candidate either).
  The path is now resolved through its symlinks before the walk — falling back
  to an absolutised form of the original path, then the original path itself,
  only if resolution errors, which `os.Executable` never does on a shipped
  platform, so no case that worked before can regress.
  The gap was invisible inside the harness, where `CLAUDE_PLUGIN_ROOT` is set and
  answers first; it appeared in a plain shell, on exactly the layout the
  installer writes. Linux is unaffected — `os.Executable` already resolves
  symlinks there; the unresolved path this fix accounts for is macOS's.
- **Two frontmatter-parser defects that silently dropped a memory page's
  provenance** (iss-30). Both were found by the detectors written for that
  issue's last acceptance instances, and both broke the same invariant the parser
  documents at the top of `internal/core/memory/yaml.go`: what the writer emits,
  the reader gives back. (1) A block scalar opened with `key: |` and left without
  an INDENTED body took the next unindented line as its content and went on
  consuming the rest of the document — every later key, `source` and its hashes
  among them, vanished into that value with no parse error. The block's body must
  be indented deeper than the line that opened it; an unindented line now ends an
  EMPTY block and is read as the top-level key it is. (2) The dumper decided a
  string needed quoting on `\n` and `\t` but not `\r`, so a bare carriage return
  went into the YAML region raw — and since the reader normalises `\r` to `\n`
  before splitting lines, a value carrying `\r---` re-read as an early
  frontmatter terminator, pushing the keys below it into the page body. A page
  written that way reported a successful ingest while reading back with no source
  hashes at all, which is the state the reconcile/repair path treats as an
  orphan. `\r` now triggers the same quoting as `\n` and `\t` (the escape side
  already emitted `\r` correctly). Defect (2) was reachable from distiller
  output alone, so a page's provenance could be stripped without any
  hand-editing; defect (1) needs a hand-authored or externally-written page
  (a bare `key: |` cannot survive the dumper's own quoting of `|`).
- **`ahoy install` writes a `.gitignore` block that matches the tier layout it
  documents** (iss-169). The managed block ignored a root-level `.work/` under
  both visibilities — a path the three-tier layout does not have — while never
  naming `.abcd/.work.local/`, the one tier that must be gitignored. So a fresh
  install ignored nothing that existed and left the local-ephemeral tier tracked:
  precisely the state `abcd audit`'s `three-tier-layout` rule then reports as an
  error, the installer and the auditor disagreeing about the same convention. The
  block now follows the brief's visibility table exactly — a private repo ignores
  `.abcd/.work.local/` alone, because the rest of the `.abcd/` namespace is meant
  to be committed; a public repo ignores `.abcd/` outright, one switch with no
  per-subdirectory exceptions, plus the legacy root-level `memory/` snapshot. An
  existing repo carrying the old block needs no migration step: the block reads
  as drift, `abcd ahoy` reports it, and one apply replaces it, leaving the
  repository's own ignore rules untouched. A repository still on the historical
  root-level `.work/` layout stops ignoring `.work/` when the block refreshes;
  the layout migration in `/abcd:prepare-this-repo` is the path off that state.
- **The `PII` rules domain fires on network work, and says never to commit a
  hostname or an address** (iss-156). Its recall keywords were the vocabulary of
  credentials — secret, token, credential, pii, redact, hostname, email — so a
  session investigating a mesh VPN's reachability, a firewall rule, or an
  address in a network config matched nothing and the hook injected nothing.
  The keywords now cover that context — `ip`, `ips`, `ipv4`, `ipv6`, `vpn`,
  `tailscale`, `tailnet`, `wireguard`, `firewall`, `network`, `reachability`,
  `reachable`, `unreachable`, `dns`, `ssh`, `subnet`, plus the `mac address`
  phrase alias — and
  one new rule line forbids committing hostnames, IP or MAC addresses, and other
  live network identifiers: redact or omit them, and reach for a reserved
  documentation value (RFC 5737, RFC 3849, RFC 2606, RFC 7042 — the same set the
  audit privacy-hygiene rule cites) only where an illustrative
  example is actually needed. That rule previously existed only in a repo's own
  instructions file, which meant the one discipline this domain most needed to
  state was the one it never said — an agent could recall `PII` and still read
  nothing about the machine identifiers in front of it. Recall matching is
  word-bounded already, so the bare `ip` keyword matches a standalone token and
  never the middle of a word such as "script" or "zip"; the plural, the
  version-qualified forms and the `-able` adjective need their own entries
  because the stemmer has a three-character floor and does not bridge
  `-ability` to `-able`. `mac` is an alias phrase rather than a bare keyword so
  that an Apple Mac does not recall the domain.
- **`abcd audit` flags local-tier artefacts sitting in a committed tier**
  (iss-155). The `three-tier-layout` rule verified that the committed tiers
  exist and that `.abcd/.work.local/` is gitignored, but never the reverse
  containment: a `NEXT.md`, `scratch/` or `logs/` placed directly in
  `.abcd/work/` or `.abcd/development/` passed clean — which is exactly how a
  handover file carrying machine-local detail rides a committed tier into
  public history. Those three names are the local-ephemeral tier's
  conventional contents, so their presence in a committed tier is now an
  error, one finding per misplacement, each naming the offending path and
  fixing it with a move to `.abcd/.work.local/`. Presence is checked on the
  filesystem, like the tiers themselves: an untracked `NEXT.md` in
  `.abcd/work/` is one `git add -A` from being committed, and the audit
  should say so before that happens, not after.
- **Network identifiers are detected, redacted and audited — one pattern set,
  three surfaces** (iss-154, iss-157, iss-125, iss-153). The scanner carried
  token-shaped secrets and identity matchers but nothing for addresses or
  hostnames, so a lifeboat pack or launch bundle carrying a tailnet address and
  device names shipped clean, a stored transcript kept a LAN hostname verbatim,
  and `abcd audit` passed the whole class in silence. The detector is an
  allowlist inversion rather than a leak-recognition heuristic: because every
  illustrative identifier in a committed file comes from a reserved range, the
  small closed set of values that are ALLOWED is what a pattern can recognise,
  and everything outside it is a finding. Allowed are the documentation ranges
  (RFC 5737, RFC 3849, RFC 2606/6761, RFC 7042), device names derived from the
  persona registry, and the values that name no individual host at all —
  loopback, unspecified, netmasks, masked CIDR prefixes, and the IANA
  special-use ranges for link-local, multicast, benchmarking, protocol
  assignments and the NAT64 well-known prefix. Flagged is what identifies
  private topology: private ranges, CGNAT/tailnet, IPv6 unique-local, 6to4,
  public unicast, LAN suffixes (`.local`, `.lan`, `.fritz.box`) and
  device-hostname shapes. The set is built once and folded into the scanner's
  defaults, so Stage-1 redaction and the launch/lifeboat scan inherit it and the
  audit rule consults exactly the same patterns — the surfaces cannot disagree
  about what a leak is. Addresses are hard_fail; the two hostname shapes are
  warn, and the audit surface carries that split through rather than flattening
  it. A line carrying `abcd-audit:allow` stays exempt, so a deliberately
  illustrative value needs no weakening of the patterns. An identifier is masked
  WHOLE — to a readable placeholder in redacted text, and fully starred in a
  serialised finding — never to the head-and-tail fingerprint a credential gets:
  on a MAC or a hostname that fingerprint preserves the vendor bytes and the
  head, which is enough to re-identify the machine the masking was meant to hide.
  The two hostname shapes tell a host from source code, because Stage-1
  redaction rewrites every finding and a false positive there corrupts a stored
  transcript: a mixed-case suffix where code puts a value (`zone := time.Local`),
  a selector position (an assignment target, a block head, a call) and a
  determiner before a device noun ("a synology-nas") are all read as code or
  prose rather than machines. Mixed case on its own exempts nothing — a shifted
  key in a command or a URL would otherwise be a bypass anyone could type — and a
  fixture host must be spelled the way the persona registry spells it, in lower
  case, because a capitalised given name in front of a device noun is how macOS
  names a real person's machine. A digest is told from an address by the length
  of the colon-separated run around it, counting only the groups that carry hex:
  an address is eight of them and may sit beside one more (the port a tool
  prints), while the shortest fingerprint is sixteen. A
  repo that wants the hostname shapes to block raises their severity in
  `.abcd/config/pii.json`, and `abcd audit` reads that merged set, so the
  override reaches the surface that reports it. The transcript store's
  verification rescan refuses the write on ANY surviving identifier, not only a
  hard_fail one, so a warn-severity hostname cannot reach disk in silence, and
  the refusal it prints names the surviving kinds rather than a severity it does
  not gate on.
- **`/Users/Shared` and `/Users/Guest` stop reading as usernames** (iss-153).
  privacy-hygiene flagged every segment under a `/Users` root, so the macOS
  system directories that live there were reported as though they named a
  person, taxing product code that legitimately writes to one. The exemption is
  narrow: it is scoped to the `/Users` root, it covers the system directory
  itself and not what sits beneath it (a name nested under one — `Shared/<user>`
  — is still a user, and still flags), a segment that merely begins with a
  system-directory name is still a leak, and an empty or dots-only segment in
  between (`Shared/../<user>`) restores no shield. The audit rule holds those
  terms for both spellings it matches, the POSIX one and the Windows one — and
  because Windows accepts either separator within one path, a name written after
  a forward slash inside a Windows path is read as the nested name it is; the
  scanner's own identity matcher shares the allowlist for the POSIX paths it
  recognises, which are the only ones it has ever matched.
- **A wrapper carrying its own flags no longer defangs the hazard entry behind
  it** (iss-148). Of the
  matcher's gaps this was the one a facilitator would reasonably assume was
  covered: only the wrapper NAME was stepped over, so `sudo <hazard>` was seen
  and `sudo -u bob <hazard>` was not — `-u` was read as the command name, and one
  extra token turned an entry the registry does describe into an allow. The same
  held for `env -i` and `time -p`. A wrapper's own arguments are now stepped over
  with it, off two explicit tables: the flags each wrapper documents as taking a
  value, and the mandatory operand that no flag stepping can reach —
  `timeout [OPTIONS] DURATION COMMAND...`, where `timeout 30 rm -rf /` read as a
  command called `30`. `xargs`, `timeout` and `exec` join the wrapper set they
  were missing from. What stays unseen is stated rather than implied: a wrapper
  outside the known set, a value-taking flag the per-wrapper table does not
  name (a bundled `sudo -Hu bob`), where the miss is a non-match and never a
  false block, and a backtick command substitution, which stays a disclosed v1
  gap: the fourth part of iss-148 was attempted on this branch and reverted, and
  the issue stays open on it alone.
- **The guard no longer treats an arithmetic shift as a here-document, so a
  command after it is no longer silently unchecked** (iss-184). An unquoted
  arithmetic left shift with an identifier operand (`$((1<<shift))`) parsed
  as a heredoc delimiter; the tokenizer then scanned for a closing line and,
  finding one — even one an attacker supplies on purpose — consumed every
  line up to it as unchecked body text, so a later `git push --force` or
  `rm -rf` never reached command position and the guard returned an allow
  verdict with no error and no signal at all. The delimiter word is now
  recognised as never real when a bare `(` or `)` immediately follows it
  with no separator (the shape `$((expr<<ident))` always produces), so the
  arithmetic expression is read correctly and the rest of the command is
  checked normally regardless of what later lines contain. A genuinely
  unterminated heredoc (a real delimiter whose closing line never appears)
  still surfaces as an unparsable command rather than silently swallowing
  the rest of the input, the same error class an unterminated quote already
  uses.
- **The secret scanner now detects a fixed-length secret token that
  immediately abuts another, same-family or not, with no separator, instead
  of silently missing it** (iss-185). Every bundled secret pattern anchors
  its start on a leading `\b`; when two such tokens are concatenated with no
  separating byte, the byte just before the second token is itself a word
  character — the first token's own last byte — so that boundary can never
  hold and the second token was never matched at all. Because detection
  missed it, `Redact` left its whole body raw in the output, and the
  fail-closed residual re-scan that stage-two history capture relies on
  reported the redacted text clean anyway, letting a live secret reach disk.
  Fixed at the scan pass: immediately after each match, every pattern's
  `\A`-anchored, `\b`-stripped variant is tried (within a small bounded
  window, so one attempt can never cost more than a constant amount of work
  regardless of what follows it) at the exact byte offset where the match
  ended — anchored by the adjacency itself rather than by `\b`, so it cannot
  introduce a false match anywhere else in the line. A pattern whose
  quantifier is open-ended rather than fixed-length can still greedily
  consume into a following token before this recovery ever runs; that is a
  separate, broader gap, tracked as iss-188.
- **The secret scanner now also detects an abutting token whose leading bytes
  the preceding pattern's greedy quantifier had already swallowed** (iss-188).
  A pattern with an open-ended length bound consumes as many class-matching
  bytes as it can find, including the next token's own prefix when those bytes
  fall in the same character class — so the boundary between the two tokens
  sits before the reported end of the match, where the adjacency probe above
  never looks. Two concatenated GitHub PATs were reported as one over-long
  finding and the second token's tail survived redaction raw, with the
  fail-closed residual re-scan reporting the output clean, so a live
  credential still reached disk. Before a match's end is accepted as final,
  the scan now walks back over a bounded window looking for a cut where the
  shortened match is still valid for its own pattern and another token can
  begin, and probes there as well. Whether a pattern needs that search is
  decided from the pattern itself — one whose match cannot survive losing a
  byte has a rigid length and is skipped — so a pattern added later is covered
  without being annotated. The backward window, like the forward one, is a
  small fixed size and candidates come from one combined probe over the whole
  pattern set, so the work done per match never scales with what follows the
  match — a legitimately huge match (a base64 blob, a minified line) keeps the
  whole scan linear in the length of the line rather than quadratic. The
  recovery is bounded, not exhaustive: a recovered token longer than the probe
  window is still truncated to it (or, for a pattern whose required separators
  fall outside the window, missed), so an abutting token behind one of those is
  not yet covered — tracked as iss-190.

## [0.4.1] - 2026-07-28

### Added

- **`abcd identity` — one canonical identity block, and every surface held to
  it** (itd-102, spc-19). A project's positioning fragments silently: the README
  strapline, the package or plugin manifest description, and the conventions
  file's opening are edited at different moments until three surfaces tell three
  stories (the recorded iss-143 drift, which is this check's acceptance corpus).
  A repo now records one canonical markdown block — `- **Title:**`,
  `- **Tagline:**`, and an optional, wrappable `- **Pitch:**` under a recorded
  heading — and `.abcd/positioning.json` records only where that block lives, how
  loudly the family reports, and which surfaces render from it. Markdown stays
  the single source of truth. The registry is data, not branches: a surface names
  candidate files (the first present wins, so one entry covers several manifest
  formats), a locator (a capture-group regexp or a top-level JSON field), the
  block fields it must carry, and the template a proposal renders from; the three
  defaults are the README strapline, the manifest description, and the
  conventions opening, and a declared list replaces them so nothing is registered
  silently. A new `identity-positioning` rule runs the check on **every**
  `abcd audit`, naming the file, the exact drifted line, and the canonical line it
  should carry — warn-tier by default (it highlights, it never gates) and
  upgradeable per-repo to blocker. `abcd identity` renders the block and each
  surface's verdict; `abcd identity render` prints the proposed correction as a
  unified diff and **writes nothing** (autonomous rewriting is permanently out of
  scope — adopting a proposal is always the maintainer's move); `abcd identity
  init` records the block and the pointer at onboarding, adopting an existing
  block rather than re-interviewing over it. `/abcd:prepare-this-repo` gains the
  interview, detect-first. Every path the registry names is read and written
  inside an OS-enforced containment root, so an audited repo that commits a
  symlinked directory cannot make the check read a file the repository does not
  own and quote it into the report.

- **`/abcd:ideate` — the idea-admission gauntlet** (itd-104, spc-18). A big,
  unproven idea can be put through three legs before it becomes a record entry:
  primary-source research (each load-bearing claim checked against its **primary**
  source, never a secondary citation), a grill against the existing record, and an
  adversarial review that is fresh-context, off-policy, and receives the idea as an
  artefact of unknown authorship. The legs are host work; `abcd ideate record
  <idea-slug> --verdict-json <file|->` is the deterministic frame that validates
  them and writes the durable verdict — the dated record under
  `.abcd/development/research/` plus one pointer line in `.abcd/work/DECISIONS.md`.
  The verdict is recorded **whether the idea survives or dies**, and its rejected
  alternatives may be empty only behind an explicit marker, because silence and
  "nothing was weighed" read the same to a session tempted to re-propose the idea.
  Every record-grill hit is **cited by id and proved to resolve** in the
  repository: an id naming no record refuses the whole verdict and names the id.
  The legs travel as an ordered array so "three legs, in order" is checked rather
  than assumed, and refusals are whole-document — nothing is written unless
  everything validates. Ideate is **optional and never a gate**: the `intent` and
  `capture` routing help names it for big unproven ideas, and nothing requires it
  or warns when it is skipped.

- **`abcd guard` — the shell-hazard guard, wired into a live session** (itd-103,
  spc-16). `abcd guard check --command "<line>"` evaluates one candidate command
  against the hazard registry and reports allow, warn, or block: a blocker exits
  1 and answers with the plain-language why and the safe successor, a warn exits
  0 with the warning rendered, and a guard that cannot be evaluated at all exits
  2 rather than letting silence read as clearance. `abcd guard hook` is the host
  adapter: it reads a pre-tool-use hook payload, refuses a matching command with
  the successor as the block message, and lets everything else through. The
  installed hook entry wraps the binary call so a missing or broken abcd **fails
  open, loudly** — the command runs and the session carries an unmissable
  UNGUARDED warning, never a silent no-op and never a stuck session. `abcd ahoy`
  gains a `guard:` line reporting the three things that can independently be
  false — hook installed, binary reachable, registry loadable — plus a
  deliberately disabled registry, so a broken guard is visible from outside the
  session too. Every unguarded state is loud, including the deliberate one: a
  registry switched off in `.abcd/guard.json` makes each command it lets through
  carry the warning, so "off" can never pass for "clear". Per-repo overrides live
  in that one file and nowhere else — no flag, environment variable, or prompt
  disarms the guard for a session, so the change lands in a diff. An allow means
  no entry matched, never that a command is safe: a hazard reached another way —
  a string handed to an interpreter (`eval`, `sh -c`), a launcher the guard does
  not step over, a backtick substitution, or a form no entry describes — is not
  seen. Coverage is what the registry names, and the command reference says so.
  A registry that is switched off answers `abcd guard check` with a fault rather
  than a clearance, so a script using the verb as a gate cannot be waved through
  by an edit to `.abcd/guard.json`.
- **`abcd launch scaffold` — the changelog-driven release-gate scaffolder**
  (itd-93, spc-14). Writes the fixed release machinery into a managed repo that
  lacks it: `.github/workflows/release.yml` (verify → build → publish, the verify
  gate armed against the reviewed **content** commit so the first public release
  cannot hit the receipt-vs-tag self-reference), `.github/workflows/auto-release.yml`
  (newest dated CHANGELOG heading → tag that commit → call `release.yml`), and the
  adr-37 runbook — wired to the repo's own default branch and Go version,
  `GITHUB_TOKEN`-only and injection-safe. The workflows ship from a **single
  embedded template** that abcd-cli's own release workflows are regenerated from
  (self-scaffold parity, proven by a byte-exact test), so every abcd release
  exercises the exact machinery a managed repo receives. The scaffolded
  `release.yml` carries a `workflow_dispatch` **rehearsal** that arms the full gate
  against a simulated changelog roll and reviewed-content commit and publishes
  nothing — a green rehearsal is the runbook precondition for the first real
  release. A bare repo with no semantic detector degrades cleanly to the
  deterministic gates and a generic build. The verb is idempotent and fail-safe: a
  re-run on current machinery is a no-op, a hand-edited file is refused rather than
  clobbered (unless `--confirm`), and it refuses rather than half-writing.
- **A public terminology crosswalk at `docs/reference/terminology.md`** (itd-100).
  One reference page maps 26 established agentic-AI terms — protocols, the core
  loop, context, safety, governance, operations — to abcd's position on each:
  USES (naming the native verb or principle), ADAPTS (the sharper native name and
  why), REJECTS (with the recorded reason), or WATCHING (with the record id).
  Every established definition carries a footnote citation to a primary source
  (specification, standards body, DOI-bearing paper, or origin engineering doc);
  every abcd claim is grounded in the committed record. Vendor names appear only
  inside citation footnotes, keeping the page body host-agnostic.

- **`abcd ahoy install --dev` — a track-latest dogfood install mode** (iss-75).
  Normal `ahoy install` symlinks the pinned built binary, so tracking live
  development meant hand-rolling a `~/.local/bin/abcd` wrapper that ran
  `go build -C <repo> && exec` on every call. That manual workaround now dies:
  `--dev` installs a shim at the same `PATH` target that rebuilds abcd from the
  source tip on every invocation and execs the fresh binary. A broken build fails
  loudly and never execs a stale binary (loud-staging). `abcd ahoy` status reports
  the mode as `install: dev (tip build)`, detected from the installed shim itself
  (never recorded in the tracked repo config, so it can never go stale), so a dev
  install is never invisible. Installing over an existing
  install applies-as-update in either direction — `--dev` replaces the pinned
  symlink with the shim, a plain re-install restores the symlink — and a foreign
  occupant is still never clobbered.
- **Record-id minting now sees every branch, and a spec-id uniqueness lint closes
  the class** (iss-115, iss-120). Sequential ids (`iss-N`, `itd-N`, `spc-N`) were
  minted from the local working tree only, so two branches cut from the same base
  silently minted the same next id — invisible on each branch and surfacing only
  at merge. Minting now folds in the highest id committed on every local and
  remote-tracking branch (a single canonical refs-union scan), so once one branch
  commits an id, the other mints past it. When git cannot be read over a present
  repository the mint degrades to working-tree-only and says so loudly on stderr
  (never a silent fallback); a directory that is not a repository has no branches
  to collide with and mints quietly. The residual window — two branches that both
  mint before either commits — is caught by the record-lint uniqueness rules on
  the merged pull request, which now cover spec ids too: the new `spec_id_unique`
  rule flags every file claiming a duplicate `spc-N`, mirroring the existing
  `issue_id_unique` and intent-id guards.
- **`abcd capture resolve` and `abcd intent "<text>"` can now stamp a product
  `impact`** (iss-117). A resolved issue and a shipped intent are in the release
  set, so the `issue_impact_valid` and `intent_impact_valid` record-lint blockers
  require a valid `impact` on those records — but the verbs that mint them had no
  way to set one, so the tool's own path produced records its own gates rejected.
  `capture resolve` now takes a mandatory `--impact <additive|breaking|fix|internal>`
  (there is no default: an absent or misspelled value is refused, not guessed),
  and `abcd intent "<text>"` takes an optional `--impact <additive|breaking|fix>`
  that is stamped onto the seeded draft and travels unchanged through planning to
  `shipped/`. `internal` is rejected on an intent (a press-release-first intent is
  user-facing by definition). `capture wontfix` is unchanged — a non-action ships
  nothing, so `wontfix/` carries no impact.

### Fixed

- **Untrusted prose can no longer open raw HTML in a record abcd writes.** The
  shared cleaner every host-delegated ingest routes through — the release
  changelog ingest, the lifeboat synthesis writers, and the ideate verdict —
  neutralised newlines and HTML comment markers but left a bare tag intact. In
  CommonMark a `<` followed by a letter, `/`, `!`, or `?` opens an HTML block, and
  several of those block types run to the end of the document: a single
  `<script>` in one model-supplied field made every later section of the record
  render as inert text inside an unclosed element, while a forged section above it
  rendered normally — so an artefact whose whole value is that a later session
  trusts it could hide its own evidence. Every tag opener is now broken apart in
  the one canonical primitive, so the fix lands at all three boundaries at once.
  The neutralisation runs **after** terminal sanitisation, which closes a second
  hole: sanitising substitutes `?` for a masked rune, so `<` before an escape byte
  previously became `<?` — a processing instruction — after the old ordering.
- **The disembark probe's recursive file walk is bounded per directory, opens
  each child in O(1), and skips the common ecosystems' dependency trees**
  (iss-112, iss-114, iss-116). The walk now reads every directory with a bounded
  `ReadDir` (the same 50 000-entry guard `ListDir` uses), so a single directory
  of millions of entries can no longer balloon memory before the file cap
  applies. It holds a sub-root per directory (`os.Root.OpenRoot`) instead of
  re-resolving every path from the containment root one component at a time, so a
  deep tree costs O(entries) rather than O(entries × depth) — a 48 000-directory
  depth-30 tree walks in ~1.4 s where the old walk took ~7 s, and the cost is now
  independent of depth. The `os.Root` containment guarantee is unchanged: a
  symlink is still refused rather than followed out of the tree. The skip set
  widens beyond Node and Go to the common dependency, cache, and build-output
  trees — Python (`.venv`, `venv`, `.tox`, `__pycache__`), Rust and generic
  build output (`target`, `build`, `dist`), and CocoaPods (`Pods`) — so a
  vendored `TODO` is no longer cited as the project's own open question, and a
  large dot-prefixed dependency tree can no longer exhaust the walk cap before
  the project's own `src/` is reached.
- **The open-questions marker scan no longer reads documentation about markers as
  open questions** (iss-111). The pattern that grounds `evidence/open-questions`
  admitted a bare uppercase `TODO`/`FIXME` followed by whitespace, so on a
  repository that documents its own conventions every prose mention of a marker
  was cited as a work marker — 14 such false positives across the durable record
  (`.abcd/development/`) on this repository, all documentation, none a real
  marker, down to 3 after the fix (each an irreducible prose quotation of the
  literal `TODO:` form). `TODO` and `FIXME`
  now require a trailing `:` or `(` (the conventional `TODO:` / `TODO(alice):`
  spellings), which is how genuine markers are almost always written; `XXX`,
  `HACK`, and `BUG` still admit their conventional bare form, because they are
  rarely written as bare words in prose and carry no measured false-positive
  cost.
- **Concurrent runs can no longer drop a repo registration or delete a
  just-committed issue file** (iss-101, iss-102). Two `abcd ahoy install` runs
  from different worktrees shared one `~/.abcd/history/index.json`, and its
  registration was an unlocked load-modify-write: atomic rename kept the file
  intact but the last writer clobbered the other's update, silently erasing a
  repo entry or a re-founding lineage link. The history registry now serializes
  its load-modify-write behind an inter-process lock and re-loads inside it, so
  concurrent registrations compose instead of overwriting; the store bootstrap
  creates `index.json` with an exclusive create, so exactly one racing run seeds
  it. The re-founding lineage confirmation is still asked before the lock is
  taken — never across an interactive prompt — and the state it validated is
  re-checked under the lock, surfacing a conflict rather than writing a link the
  user approved against a stale index. Separately, the capture ledger's orphan
  sweep and its commit write now take the same ledger lock, closing a window in
  which a capture stalled more than sixty seconds could have its committed issue
  file swept away after the capture reported success. The inter-process lock is a
  single shared primitive; the capture allocator and the history registry both
  route through it.

## [0.4.0] - 2026-07-22

### Breaking

- **The record-lint banlist now requires a machine-readable successor and
  context.** Every `banned_tokens` entry must declare a non-empty `successor`
  (what to use instead of the retired token) and a non-empty `allow_context`
  (where the token is legitimately allowed); a config whose entry omits either
  is rejected at load rather than lints with a replacement that lived only in
  prose. Each finding auto-cites its successor, so the reader is told the
  replacement inline. The bundled record-lint and docs-lint banlists carry the
  new fields.

### Added

- **Generated CLI reference with a drift gate** (iss-47). `docs/reference/cli/`
  now holds `commands.md`, a per-command Markdown reference walked deterministically
  from the Cobra command tree (`cli.GenerateReference`) — usage line, summary, and
  flags for every user-facing verb, with the operator-internal hook subtree omitted.
  Refresh it with `go generate ./internal/surface/cli`; a `go test` drift gate
  regenerates the tree and fails the build whenever the committed page and the tree
  disagree, so the reference can never silently go stale. The walker is hand-rolled
  over stdlib and the existing Cobra dependency — no new module dependency.
- **`agent-observation` is now a valid `--source` value for `abcd capture`.**
  An autonomous run's self-observation had no honest surfacing channel and was
  reusing `agent-finding`. `agent-observation` parallels it ("an agent
  observed") without being tied to one run mode (iss-57).

- **`issue_id_unique` record-lint rule — a duplicate issue id is a blocker.**
  The lint pass now scans the issue ledger's three status directories
  (`open/`, `resolved/`, `wontfix/`) and flags any `iss-N` id claimed by two or
  more files, mirroring the existing intent-id uniqueness check (both share one
  `validateIDUnique` primitive). The capture allocator already rejects a
  duplicate on the reservation path, but a hand-added issue file that bypassed
  it slipped straight through until now; this is the backstop that catches it.
  Every file in a colliding set is flagged, since the linter cannot know which
  claimant is authoritative.
- **`abcd intent ready <itd-N>` — the implement-readiness gate.** A read-only
  verb reporting whether an intent may be implemented now: planned
  (directory-as-truth), enumerable Acceptance Criteria, a bidirectional spec
  link, and a spec body written past its minted stub. Every check carries a
  reason and, when failing, the exact remedy command. The exit code is the
  machine seam: 0 ready, 1 not ready (the rendered report is the output), 2
  structural fault — so an autonomous run can gate on it (step 0 of the run
  protocol) instead of re-deriving readiness from prose, and an unplanned
  intent is refused rather than improvised against.
- **`/abcd:intent` plugin command surface** (`commands/abcd/intent.md`),
  covering the full verb family — status, quoted-text create, `ready`, `plan`,
  `link`, `review`/`ingest` — plus the host-run planning interview an unready
  intent is routed to: the human confirms the press release, resolves open
  questions, and accepts, edits, or strikes every acceptance criterion before
  `abcd intent plan` is run as their sign-off act.

### Fixed

- **The published plugin now ships its agents and hooks.** The launch payload's
  include list named neither the `agents/` nor the `hooks/` surface directory,
  so the released bundle omitted both — the plugin installed without its agents
  and without its hook wiring. Both directories are added to the payload
  includes, and a bundle-completeness check now fails if any auto-discovered
  plugin-surface directory present on disk is neither included nor explicitly
  excluded with a reason, so a newly-added surface can never be silently dropped.
- **The history-store bootstrap error names the verb that exists.** When a
  transcript capture found the store's owned directories absent, the preflight
  error told the user to "run `abcd install`" — a verb that does not exist. The
  remediation now names `abcd ahoy install`, the verb that actually bootstraps
  the store (iss-58).
- **`abcd ahoy` no longer overclaims `managed-repo` for a stray `.abcd/`
  directory (iss-88).** Folder classification treated the mere presence of an
  `.abcd/` directory as a strong managed signal, so a repo with an unregistered,
  markerless `.abcd/` reported `managed-repo`. Only index registration or a
  marker block now promotes a folder to managed; a stray `.abcd/` reports
  `unmanaged-repo` (or `unmanaged-folder` outside a git repo).
- **The identity pin round-trips through the self-contained commit guard.** The
  pin was stored with Go's default JSON encoder, which escapes `&`, `<`, `>` (and
  always `"`, `\`, control characters); the pre-commit identity guard reads the
  raw bytes between the quotes with a naive parse, so an escaped value never
  matched `git config` and fail-closed a correctly configured identity (e.g. a
  `user.name` of `Marks & Spencer`). The pin is now marshalled without HTML
  escaping so `&`/`<`/`>` are stored literally, and the characters a parse can
  never read back (`"`, `\`, control) are refused at pin time with a clear remedy
  — keeping the commit guard zero-dependency rather than delegating it to a
  possibly-stale binary.
- **`abcd ahoy install` now honours an explicit config-value flag on an
  already-configured repo (apply-as-update).** Passing `--visibility`,
  `--docs-target`, `--oracle-backend`, or `--scan-deep` on a repo whose value
  was already set and valid was silently dropped: the persisted value
  short-circuited the install before the override was consulted. An
  explicitly-passed flag whose value differs from the persisted one now
  overwrites it, echoes the change (`changed: visibility: private -> public`),
  and — for visibility and docs-target — refreshes the `.gitignore` block and
  marker files so nothing is left inconsistent. A re-install with no such flag
  is still an exact no-op and never clobbers a valid value.
- **GL002 no longer fires a spurious blocker when a line has a closed inline-code
  span before a stray backtick.** `stripInlineCode` restored the entire original
  line whenever it reached end-of-line with an unpaired backtick, un-masking an
  enforced synonym that sat inside an earlier, correctly-closed span. It now
  blanks matched backtick pairs only and leaves a trailing unpaired backtick (and
  its tail) literal, so the earlier span stays masked (iss-106). Full CommonMark
  double-backtick span parsing remains out of scope.
- **`abcd intent "<text>"` no longer files a draft from a mistyped subcommand.**
  A near-miss for an intent subverb (`intent paln`, `intent lnk itd-5`) is
  refused with a did-you-mean and writes nothing, mirroring `abcd capture`'s
  guard; a genuine prose title still files. The shared typo heuristic is now
  record-id aware (`iss`/`itd`/`spc`), which also sharpens `abcd capture`.

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
- **The last three ahoy reads route through the guarded primitive, closing a
  residual read-time TOCTOU.** The hook-manifest check and the two `.gitignore`
  reads (one at the attacker-influenced working-directory boundary) each ran a
  separate `lstat`, regular-file, and size-cap check and then a distinct
  `os.ReadFile`, so a type or symlink swap in the window between the check and the
  read was not refused on the descriptor that was read. All three now read through
  the single guarded open (`O_NOFOLLOW` + regular-file on the open fd + size cap),
  and every structured signal — the manifest's reason strings and the
  `.gitignore` overwrite refusals — is preserved unchanged.
- **Machine (`--json`) output no longer leaks an absolute local path.** The
  error-only path scrub covered failures but not the success envelope, so
  `capture`, `capture resolve`/`wontfix`, and `capture list` echoed the absolute
  ledger path in their `path` field; a successful `memory ingest` recorded the
  absolute (and, for a `~/…` source, home-rooted) source path in its
  `citation.origin`; and a `memory ingest` failure on a source outside the
  working directory embedded an absolute source path the scrub could not reach.
  Every such locator is now rendered repo-relative — in machine output and in the
  persisted citation alike — so no verb emits a developer-identity path into
  machine output.

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
  Phase 2 gap report instead of hand-auditing (iss-86).

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
- `abcd docs lint` — a deterministic docs-currency gate over `docs/` and the repo
  root: change-narration in a doc body, a broken relative link, or a stray
  top-level markdown file each fails the gate.
- `record-lint` — a deterministic drift gate for the `.abcd/development` design
  record (banned tokens, git-metadata, link resolution, intent lifecycle), wired
  blocking into CI and the pre-push preflight.
- Derived-versioning design record (intent itd-73 + ADR-31): the release version
  is derived from intents' declared impact, never hand-authored. The derivation
  itself lands in a later phase.
