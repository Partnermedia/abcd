# Install experience — release plan, two cuts (2026-08-11)

**Status:** the backlog for the next implementation cycle, consumed by the
generic protocol at
[`2026-07-12-abcd-run-protocol.md`](2026-07-12-abcd-run-protocol.md). Scoped
to one theme: what a new user goes through between deciding to try abcd and
having a working `/abcd:*` surface.

**Release framing (maintainer interview, 2026-08-11): one theme, two cuts.**
Cut A makes the install path *work* on today's packaging. Cut B makes it
*small* by repointing the marketplace at the curated release artefact
([itd-108](../intents/drafts/itd-108-the-plugin-installs-from-the-curated-release-artifact-not-th.md)).
The split is not caution for its own sake — it is
[iss-205](../../work/issues/open/iss-205-every-one-of-the-17-files-under-commands-resolves-the-binary.md)'s
own recorded prerequisite: the `${CLAUDE_PLUGIN_ROOT}` binding must land **and
be verified on a real install** before the payload slims, so there is never a
release in which the `go run` fallback is gone and the primary path is not yet
proven. Cut B is gated on the verification in §4, not merely on Cut A merging.

**The version is derived, never declared.** `abcd launch ship` derives it from
the shipped impacts. Cut A carries a breaking change (the documented binary
install location moves), so derivation from v0.4.2 predicts **v0.5.0**; Cut B
is breaking for users installed under the relative-path source, predicting
**v0.6.0**. Both are predictions, not overrides.

## Decisions settled at interview (2026-08-11)

Recorded here because several contradict what the source records currently
assume, and the contradictions are the point.

1. **Two cuts, repair then repoint.** Per iss-205's prerequisite note.
2. **itd-108's update signal: the catalog is published as a release asset.**
   This is the "third shape" itd-108's Open Questions hoped for, and it
   dissolves the conflict the intent could not resolve. The release job renders
   `marketplace.json` with the version and payload URL stamped and uploads it
   alongside the binaries; users add the marketplace by the stable
   `releases/latest/download/marketplace.json` URL. Consequences: **no commit
   ever lands on a branch**, so `release.yml`'s armed no-branch-commit tripwire
   stays armed and un-carved; [adr-19](../decisions/adrs/0019-plugin-json-version-carve-out.md)
   stays intact rather than needing amendment; and no maintainer edits anything
   between the tag and the user's update. The rejected alternatives and why:
   *digest-signalled* contradicts adr-19's recorded **ACCEPT** outcome and the
   render step that already implements it (`internal/core/launch/dryrun.go:66`
   keeps the DEV polarity in-tree and proves the PUBLIC polarity over the
   rendered payload; `lockstep.go:43` requires the version present and
   strict-SemVer there), so it would mean amending an accepted ADR to work
   around a mechanism that already works; *SemVer with a committed catalog*
   buys automation by disarming a guard this project deliberately built.
3. **`docs/` leaves the payload.** 1.3 MB of the 1.6 MB bundle. This widens
   itd-108's scope, which currently declares `launch-payload.json` out of
   scope — the intent must be edited to match, not silently exceeded.
4. **The binary-resolution ladder's last rung becomes a source-checkout-only
   note.** Rung 1 `"${CLAUDE_PLUGIN_ROOT}/abcd"`, rung 2 `PATH`, rung 3
   `go run ./cmd/abcd` *explicitly qualified as applying only in a source
   checkout*. Honest for contributors without printing an instruction a plugin
   user cannot follow.
5. **`ahoy install` defaults to `~/.local/bin`, never escalates privileges.**
   The README one-liner drops `sudo`. This is the breaking change in Cut A.
6. **Prompts read stdin when stdin is not a TTY.** Smallest surface; makes the
   thing an agent already tried (piping `y`) work.
7. **The version-skew notice is retired in Cut B, not repaired.** Under
   decision 2 a plugin update creates a fresh plugin root and `bootstrap.sh`
   fetches the matching binary from the same release, so surface-ahead-of-binary
   cannot arise in steady state. Repairing the hex gate would pay for a
   mechanism the same release removes.
8. **iss-163/iss-164 (plain-language prompt help, persona-readable summary) are
   out of both cuts**, deferred to
   [itd-63](../intents/planned/itd-63-setup-wizard-explains-installs.md).
   Real, and larger than everything else here combined; keeping them out keeps
   both cuts reviewable.

## Run contract

- `backlog:` the workstreams below; each item's stable id is its issue/intent
  id. Issue bodies in `.abcd/work/issues/open/` are the item spec; this file
  adds ordering, collision notes, and the interview decisions above.
- `gate:` `make preflight` (sole deterministic authority), plus `gofmt -l .`
  before pushing — CI's format gate sits outside preflight.
- `budget:` one item per burst; stop cleanly at the first bound reached.
- `commit_trailer:` `Assisted-by: Claude:<model-id>` (kernel format; never
  `Co-Authored-By` for AI).
- `reviewers:` correctness `abcd:ruthless-reviewer` on every item; security
  `abcd:security-reviewer` on A1, A4, B1 and B2 (binary resolution, PATH/
  symlink writes, release-artefact publication, checksum coverage).
- `strike_limit:` 3.
- One PR per item; never commit to main directly. Fixing a ledger item includes
  moving it open → resolved (`abcd capture resolve`) in the same PR.
- Every new behaviour gets a test **watched fail before the change and pass
  after**. For A1 and A2 that test is partly manual (§4) — say so in the PR
  rather than implying CI proved it.

---

## Cut A — the install path works

Ordered. A1 is the keystone; A2 and A3 are independent and can land in any
order; A4 is separable but belongs in the same cut because it is the other
front door.

### A1 — bind the command surface to the plugin root (iss-205, critical)

**Verified 2026-08-11:** 19 files under `commands/`, **0** occurrences of
`${CLAUDE_PLUGIN_ROOT}`, **17** carrying a `go run ./cmd/abcd` fallback.

Today `hooks/bootstrap.sh` installs a checksum-verified binary *into the plugin
root* while itd-105 deferred PATH to `ahoy`, so the two halves never meet: a
fresh install works only because the marketplace clone happens to carry `cmd/`
— 54 seconds and a Go toolchain. Under Cut B there is no `cmd/`, no `go.mod`
and no source, so this becomes a hard failure with no recovery path on the
user's machine.

- Rewrite the resolution ladder in all 19 files per decision 4.
- The plugins reference confirms `${CLAUDE_PLUGIN_ROOT}` substitutes in skill
  and agent content wherever the placeholder appears — so this is a text
  change, not a new mechanism.
- Add a check that fails if any file under `commands/` names a binary without
  the plugin-root rung first. Without a detector this regresses on the next
  command added.

### A2 — one SessionStart entry, success reported as success (iss-204 + iss-208, major)

**Verified 2026-08-11:** `hooks/hooks.json` places `bootstrap.sh` and both
binary-gated hooks in a single `SessionStart` group and relies on list order.
The harness runs matching hooks **in parallel** — spc-21's warrant that
"ordering within one event's hook list is preserved" is false and load-bearing.
Both gated hooks lose the race against a ~10.7 MB download, print "the plugin
binary is not installed", and genuinely do not run — on every fresh install
*and every plugin update*, since each update lands in a fresh cache directory
with no binary. itd-105 AC#1 fails on both.

iss-208 compounds it: `bootstrap.sh`'s `notice()` exits 2 for a sound reason
(only a non-zero exit puts stderr in front of the human), but the harness
renders any non-zero `SessionStart` exit as a hook *error*. There is no notice
channel distinct from error, so the checksum-verified happy path is labelled a
fault. A first-time user's opening screen is three consecutive hook errors, one
of which is success.

- Collapse the three entries into **one** `SessionStart` command that runs the
  bootstrap and then the two binary calls in a single shell, so sequencing is
  owned here rather than assumed of the harness.
- Lead the visible first line with the success — the transcript shows only the
  first line of stderr.
- The honest-failure posture for the genuinely-no-binary window is unaffected.
- Correct spc-21's warrant in the same change; a false warrant left standing
  will be trusted again.

### A3 — print an instruction that can be run in the state it describes (iss-207, minor)

`bootstrap.sh`'s success notice and README both say run `abcd ahoy install`
once — a command that cannot run in the no-binary-on-PATH state it addresses.
Observed consequence on the 2026-08-10 install: the agent invented a `go run`
incantation reaching into the harness plugin cache and told the user to run
that. Print the absolute plugin-root path the script already holds as
`$binary`; give README the same concrete form.

### A4 — no-sudo, field-standard binary install (iss-171, major) — **Breaking**

Today `binTarget` defaults to `/usr/local/bin/abcd` and symlink detection
recognises only that exact target, so a working `~/.local/bin/abcd` — the
single-user location uv/pipx/rustup-class tools use — reports
`symlink.missing` while the detector is *itself running as `abcd` from PATH*.
Letting `ahoy install` "fix" that would write a symlink to
`<plugin-root>/abcd` without validating the target exists: a dangling symlink
shadowing a working install.

- Detection scans PATH, resolves symlinks (`EvalSymlinks`, the same seam as
  iss-170), classifies dev-shim / owned / foreign.
- Install defaults to `~/.local/bin` (created if absent), adopts an existing
  owned install in place, does system-wide dirs only behind an explicit
  `--bin-dir` that fails loudly when unwritable, never escalates privileges,
  and refuses to create a symlink whose target does not exist.
- "`~/.local/bin` not on PATH" becomes its own loud gap with a printed one-line
  fix. Script-first: print the `export` line, do not auto-patch shell profiles.
- README one-liner drops `sudo`. **Breaking** section entry required.

### A5 — answer prompts without a TTY (iss-167, major; carries iss-166)

Piped `y` and pseudo-TTY attempts all arrive as declines, so a host agent
cannot drive the interactive path at all and must hand the step back to the
human — the exact seam this plan exists to smooth. Read answers from stdin when
stdin is not a terminal; leave the interactive path unchanged when it is.

Fold in **iss-166** while here: `--yes` reports "already up to date" yet
silently skips the optional identity-pin category. Either `--yes` covers
optional categories or its help and completion output say they are excluded and
how to apply them. Decide and state it; do not leave it ambient.

### A6 — repo-relative paths in the install receipt (iss-177, minor)

Every apply step notes its write with an absolute path and the CLI prints them
verbatim, so a receipt pasted into an issue carries the developer's home
directory and username. Sibling verbs already route error text through
`scrubPaths`; the receipt does not. One change across every apply step — report
repo-relative, or scrub before rendering. Included here because A4 touches the
same apply path and adds more such notes.

---

## §4 — the verification gate between cuts

**Cut B does not start until this passes.** Not a formality: A1 and A2 fix
failures that CI structurally cannot observe, because the harness, the parallel
hook execution and the plugin cache are not present in CI.

On a machine (or a clean container/VM) with **no Go toolchain**:

1. Install the released Cut A binary by the new README one-liner. `abcd
   version` succeeds without `sudo`.
2. Add the marketplace and install the plugin. Record the opening transcript
   verbatim.
3. Assert: **zero** hook-error lines. The bootstrap's success reads as success.
   Both binary-backed `SessionStart` hooks actually ran.
4. Assert: a bare `/abcd` resolves via `${CLAUDE_PLUGIN_ROOT}` — sub-second, no
   Go invoked. Confirm by temporarily removing `go` from PATH.
5. Assert: `ahoy install` completes when driven non-interactively by the agent,
   and its receipt contains no absolute home path.

Record the transcript under `.abcd/.work.local/scratch/` and cite it in the Cut
B PRs. A failure here is a Cut A defect, not a Cut B blocker.

---

## Cut B — the install is small (itd-108)

**Prerequisite: §4 green.**

### B1 — publish the curated payload and the catalog as release assets

`internal/core/launch` today contains **no zip or archive code** (verified
2026-08-11) — `launch --dry-run` computes the bundle (54 files, 0 scan
hardfails) and discards it. `release.yml:306` uploads `bin/abcd-*` and
`bin/checksums.txt` only. So packaging is new work, not wiring.

- Package the payload declared by `.abcd/config/launch-payload.json` into a
  zip, attached to the release with its digest recorded next to the binaries'.
- Render `marketplace.json` with the version and payload URL stamped, and
  upload it as a release asset (decision 2). The committed
  `.claude-plugin/marketplace.json` stays the unversioned source template and
  keeps satisfying the DEV polarity.
- **Drop `docs/` from `launch-payload.json`** (decision 3), taking the bundle
  to roughly 300 KB. Two knock-ons: the payload README's
  `<img src="docs/assets/img/...">` tags break — rewrite them to absolute
  GitHub URLs in the rendered payload, or docs-lint's broken-relative-link rule
  fires on the artefact; and itd-108's "What's Out of Scope" must be edited,
  since it currently takes the manifest as given.
- Note `launch-payload.json` also includes `.gitignore`, which itd-108's quoted
  payload list omits. Reconcile the intent to the file.
- Assert in CI: the published zip's contents match what `abcd launch --dry-run`
  reports, so preview and artefact cannot disagree.
- Assert: the no-branch-commit tripwire still passes — the release job pushed
  nothing to any branch. Under decision 2 this should hold by construction;
  prove it anyway.

### B2 — repoint the install path

- README install section points at the `releases/latest/download/marketplace.json`
  add-command, and states the **minimum host version** an `archive` source
  requires (v2.1.224), so a failure below it names the version rather than
  presenting as a broken plugin.
- A user installed under the previous relative-path source must either land on
  the archive-sourced install or be told plainly what to re-add. An install
  that silently stops updating fails itd-108's own acceptance criterion.

### B3 — retire the version-skew notice (iss-206)

Per decision 7: remove `binarySkewNotice()` and the `.binary-meta` `plugin_sha`
field with it. The gate was never wrong to be diagnosed — `bootstrap.sh`
recorded the raw ungated `plugin_root_basename` beside the gated field
precisely so "why has this never fired" had an answer in the file. It fired as
designed; the condition it guards is what disappears. Record the retirement so
itd-105's shipped claim stays honest.

### B4 — the record consequences

These are not cleanup; itd-108 makes an invariant false **by design** and it
must move in the same change.

- `AGENTS.md` states `.abcd/**` is "present in every repository checkout —
  **marketplace installs** and release source archives included." Cut B makes
  that sentence false. Rewrite it, and drop the concession that the bundler
  "denies the namespace structurally, though that filter has yet to run on a
  cut release" — after B1 it has run.
- [adr-28](../decisions/adrs/0028-single-repo-curated-release.md) decided the
  *what* (a curated artefact) but not *how it reaches a user*. An amendment
  recording the catalog-as-release-asset mechanism belongs with this work.
  "The repo is the marketplace" stays literally true — the catalog is still a
  file in this repo; only the fetch method changes.
- adr-19 needs **no** amendment under decision 2. State that explicitly in the
  amendment to adr-28, so a later reader does not re-open it.
- Promote itd-108 drafts → planned with the interview decisions folded into its
  Open Questions as settled, and its scope widened for `docs/`.

---

## Out of scope

- **iss-163 / iss-164** — deferred to itd-63 (decision 8).
- **iss-209** — the dependabot/scaffold-parity deadlock. Release-pipeline
  hygiene, not install experience. Its manual half (`make scaffold-sync`,
  `TestSyncRepoPinsIsCleanToday`) is the right stopping point; the entry
  records the `workflow_run` automation as a reviewed **dead end** and any
  future attempt needs the four conditions listed there. Do not re-attempt it
  inside this plan. It does, however, touch `release.yml` — see the collision
  note below.
- **The release/versioning pipeline itself** — stays with itd-73 and adr-31.
  Both cuts consume a cut; neither changes how one is decided.
- **The binary's distribution and trust bar** — itd-105 settled it. Note
  honestly that the payload zip rests on TLS-to-GitHub alone with no `sha256`
  pin, and that this is the same root of trust the binary's `checksums.txt`
  already has, since manifest and payload share an origin — a vacuity
  `hooks/bootstrap.sh` names in its own comments. Accept the gap explicitly in
  the adr-28 amendment rather than inheriting it by accident.
- **Windows support** — its own future intent.

## Collision notes

- **A4 and A1 both touch the binary-resolution story** — A1 owns what the
  command files say, A4 owns what `ahoy` does. Land A1 first so A4's "adopt an
  existing owned install" has a stable definition of what the commands look for.
- **B1 and iss-209 both touch `release.yml`.** iss-209's `make scaffold-sync`
  keeps the workflow and its scaffold template byte-identical
  (`TestSelfScaffoldParity`). B1 adds steps to `release.yml`, so those steps
  must be added to the **template** and rendered, not typed into the workflow —
  otherwise B1 breaks preflight the same way a dependabot bump does. Whoever
  takes B1 should read iss-209 first.
- **A2 and B3 both touch `hooks/bootstrap.sh`.** A2 restructures its invocation
  and its first stderr line; B3 removes the `plugin_sha` field it writes. Land
  A2 first (it is in the earlier cut anyway) and treat B3 as a deletion on top.

## Risks

- **§4 is manual.** If it is skipped or done sloppily, Cut B ships on an
  unproven Cut A and the failure mode is a dead command surface on the user's
  machine with no local recovery. This is the single largest risk in the plan
  and the reason for the two-cut split.
- **Decision 2 assumes the harness will re-fetch a URL-delivered catalog.** If
  a marketplace added by direct URL never refreshes without user action, "cut a
  release is the whole publication step" degrades to "cut a release, and users
  see it when they refresh the marketplace." That is still far better than
  today, but it is a warrant to test in §4's successor for Cut B, not to assume.
- **Two breaking changes in consecutive minors** (install location, then
  install source) for a pre-v1 tool with an experimental badge. Acceptable, but
  both need explicit **Breaking** entries and the Cut B one needs the
  re-add instruction B2 requires.
