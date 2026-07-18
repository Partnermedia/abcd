# Bug-hunting principles — a where-to-look playbook

Distilled from the accumulated secret/PII/launch/lifeboat hardening hunts (the
2026-07 `Unreleased` Security + Fixed run) and one further gap-driven adversarial
round on 2026-07-17. Every principle below is grounded in at least one concrete
fix (`file:symbol`) and anchored to a CWE / SOTA reference. Nothing here is
generic checklist filler: if a line is not falsifiable against the code, it was
deleted.

The loader auto-injects this doc on prompts matching the `BUGHUNT` domain (see the
bottom of this file). Read the spine, then the checklist; the provenance and
frontier sections are for a deeper pass.

---

## The spine

**Almost every defect was a wrong *implicit assumption* at a *trust boundary*,
and it failed *silently* — a gate that let something through, never a crash.** The
code assumed the world was narrower than it is:

> ASCII-only · SHA-1-only (40 hex) · case-sensitive · one control-char range ·
> first-occurrence-wins · a clean environment · a parseable config · a
> case-sensitive filesystem · an injective normaliser.

Each narrow assumption is a *false negative in a fail-closed control*: the leak
ships, the secret is missed, the kill switch is ignored, the wrong repo is
queried. A crash would have been safer — it would have been noticed.

**The second spine, surfaced by the 2026-07-17 round: a fix is not done until
every sibling site is swept.** Four of that round's confirmed findings were the
*same defect the diff had just fixed, at a site the fix missed*: the case-fold fix
skipped `home_path_self`; the Unicode-boundary fix left a fixed-length secret
regex dead; the SHA-256 voyage-key fix skipped the history store's identical key;
the `IsolatedEnv` env-scrub fix skipped two live git call sites. **A pattern fixed
at 2 of N locations leaves N−2 latent — grep for the pattern, not the instance.**

---

## Hunting checklist (value-ordered)

Each row: the **smell** you scan for · the **failing question** that decides it ·
the **CWE/SOTA anchor**. Ordered by demonstrated yield.

| Smell | Failing question | Anchor |
|-------|------------------|--------|
| A pattern was just fixed in one spot | Did the fix land at *every* sibling site? `grep` the pattern across the tree. | — (the completeness meta-rule) |
| `exec.Command("git", …)` with `os.Environ()` / default env | Can an inherited `GIT_DIR`/`GIT_WORK_TREE`/`GIT_CONFIG_*` redirect it at another repo or inject config? | CWE-426, CWE-88 |
| RE2 `\b` / `\B` over human or identifier text | Can the token begin/end with a non-ASCII letter? RE2 `\b` is ASCII-only. | CWE-176, CWE-20 |
| A fixed-length regex whose class includes a non-word char, anchored with trailing `\b` | Is there a valid value ending in that char, with no shorter match to satisfy the boundary? (a *dead* detector) | CWE-20 |
| A case-sensitive regex/compare on an email, hostname, hex, username, path-on-case-folding-FS | Is this compared case-insensitively in the real world? | CWE-178 |
| A detector that requires a component the leak itself omits (trailing sep, surname, second word) | Does the *minimal* leak satisfy every required sub-part? | CWE-20 |
| A hardcoded width — `{40}`, fixed field, truncation | Does a newer valid form (SHA-256 = 64) or a large value violate it? | CWE-1284, CWE-190 |
| `n, _ := strconv.Atoi(...)` feeding an id/index/max | Does an over-range value clamp to `MaxInt64` and then wrap on `+1`? | CWE-190, CWE-252 |
| An id/key/slug built by *deleting* or lossily replacing chars, then used as a map/dedup key | Do two distinct inputs collapse to one key (non-injective)? | — (collision → shadowed finding) |
| An equality/dedup compare | Set or ordered? Does the spec say one and the code do the other? | CWE-697 |
| `os.ReadFile`/`os.Open` (no `O_NOFOLLOW`, no cap) on a file an attacker could have committed as a symlink | Is this path on a hostile/cross-repo path (embark/probe/pack), vs the trusted local worktree? | CWE-59, CWE-400 |
| `x, _ := readConfig()` / ignored parse error on a gate, ledger, or kill switch | If it can't be parsed, does it proceed as *empty/permissive* instead of failing closed? | CWE-636, CWE-252 |
| A scanner/validator that returns "clean" | Did it actually scan, or mistake "couldn't read / git error / 0 results" for compliant? | CWE-390 |
| Untrusted repo content (commit subjects, refs, paths, page prose) rendered to a terminal/report | Does it pass a control-char sanitizer? Which ranges — C0, **C1 (U+0080–009F)**, DEL, bidi (U+202E)? | CWE-150, CWE-1007 |
| Untrusted string as a git/exec *positional* arg | Can it start with `-`? Is there a `--` separator or a leading-dash guard? | CWE-88 |
| `stat`/`Lstat`-then-open, chmod-by-name-after-close, find-then-write outside the lock | Can the target be swapped, or two runs collide, between check and use? | CWE-367, CWE-362 |
| A redaction/suppression span that hides one finding under another | Does masking a low-severity finding suppress a *high*-severity one beneath it? | — (severity downgrade) |
| A test assertion — `Contains(out,"cap")`, `NoError`, `len>=0` | Would it still pass if the code under test were reverted? Does it exercise the *attack* input? | — (vacuous test / mutation testing) |
| A "fix" that isolates/hardens a data source | Does the hardening also break the thing the source feeds? (e.g. neutralising global git config blinds the identity probe) | — (fail-safe must preserve function) |

---

## Compare to SOTA (alignment · differentiators · gaps)

**Alignment.** These map cleanly onto the taint/trust-boundary model (OWASP),
the CWE injection family (CWE-88 argument injection, CWE-150 terminal escape),
CWE-59 symlink-follow and CWE-367 TOCTOU (the `O_NOFOLLOW` + fstat-on-fd +
`fchmod` idiom is textbook), CWE-178/CWE-176 (case/Unicode), and fail-closed
design as the explicit axis. Consolidating onto one hardened primitive
(`IsolatedEnv`/`ReadGuarded`/`WriteFileAtomic`) is the canonical "no divergent
copies" move.

**Differentiators — sharper than a generic CWE sweep.** (1) *Injectivity as a
security property* — recognising that a normaliser keying a dedup map must be
injective (`idClean`). (2) *Second-order severity interactions* — a redaction
*suppression* silently downgrading a hard-fail (`localSuppressionSpans`). (3)
*Mutation-testing the tests themselves.* (4) *Honest threat-model scoping* — the
"trusted-worktree model" documents what is *deliberately* out of scope rather than
pretending all in-tree input is hostile.

**Gaps / frontiers (the next hunt's targets).** See below — the 2026-07-17 round
opened three; two (terminal-escape sanitization, attack-input tests) are now
closed and the trust-boundary read sweep remains the top open frontier.

---

## Known gaps / next frontiers

Remaining after the 2026-07-17 fixes (including an ultracode sweep-completeness
pass that finished the env sweep — `identity.gitConfig` + `capture.discoverRepoRoot`
— and guarded the five embark/pack reads over an arbitrary target repo):

0. **`ahoy.registerRepo` — unlocked read-modify-write of the cross-repo
   `~/.abcd/history/index.json` (CWE-362 lost update).** Confirmed in two rounds.
   Two concurrent `abcd install` runs each load → mutate → write the shared index,
   so one repo's entry is lost. `history/store.go` already flocks its per-repo
   store; the ahoy index has no lock. **Design wrinkle blocking a quick fix:**
   `registerRepo`'s RMW contains an interactive `prompter.Confirm` (re-founding
   lineage), so a flock must NOT be held across it — a correct fix locks the
   load/write but resolves the prompt outside the lock (or re-reads under the lock
   after confirming). Track as an issue rather than rushing a concurrency change.

1. **Trust-boundary read sweep (remaining `os.ReadFile` sites) — still the top
   frontier for any NEW code.**
   Many read `.abcd/config/*.json` and record files in the *local* worktree —
   trusted under the documented trusted-worktree model, so *not* uniformly in
   scope. The ones that genuinely cross the boundary are on the embark/probe/pack
   path over an *arbitrary target* repo (e.g. `lifeboat/embark.go:embarkMarker`
   reading a target `CLAUDE.md`; `cli` `disembark coverage <report.json>` user-typed
   operands) and `scanner.New` reading `pii.json` under a hook. **Next step: triage
   each by "is the source repo attacker-controlled here?" and guard only those with
   `fsutil.ReadGuarded`.**
2. **JSON decode-depth stack exhaustion (CWE-674) — CLOSED by the toolchain
   (verified 2026-07-17, Go 1.25.6).** `encoding/json` now enforces a max nesting
   depth: a `[[[[…` payload returns a clean `"exceeded max depth"` error long
   before the stack is at risk (confirmed empirically at 1 MB / ~500k levels).
   No depth-limited decoder is needed on this toolchain — adding one would be
   redundant dead defence duplicating the stdlib. Re-open only if the pinned Go
   version regresses below 1.25.
3. **The hand-rolled frontmatter parser (`internal/core/frontmatter/Fields`) has
   never been audited.** No YAML dependency exists, so this parser processes
   attacker markdown from embarked/packed lifeboats and ingested memory. Check
   duplicate-key precedence, unbounded key/value growth, and control-char/newline
   smuggling in field values that later flow to output or gate decisions. (The
   `capture` dup-key fix was in capture's own parser, not this one.)
4. **Minor consistency gap — FIXED (2026-07-17).** `memory.LoadRegistry` now folds
   the `O_NOFOLLOW` symlink refusal (`syscall.ELOOP`) into the same
   `RegistryFormatError` branch as `ErrNotRegular`/`ErrTooBig`, so a planted
   symlink is classified like every other guarded refusal and the raw ELOOP
   syscall detail no longer escapes. Covered by
   `TestLoadRegistryRefusesSymlink` (now asserts the typed error + no leak).

**Burst 2 (2026-07-17, session hardening loop) — the last lstat→read TOCTOU on the
embark path, now closed:**
- `lifeboat/graveyard_lessons.go:readGraveyardFile` read packed layer-1/2 files
  (`archaeology.json`/`abandoned.json`) from an arbitrary target dir with the
  `os.Lstat`(symlink+size)-then-`os.ReadFile` anti-pattern — the same swap window
  the embark/pack conversions closed, but this one was missed. A regular in-cap
  file could pass the checks then be swapped for a symlink-to-/dev/zero or a grown
  file that `os.ReadFile` follows/reads unbounded. Converted to
  `fsutil.ReadGuarded` (O_NOFOLLOW + regular-file-on-fd + size cap, one call);
  ELOOP folded into the not-regular message so no raw syscall detail leaks.
  Static inputs are handled identically (so no watched-fail); the fix closes the
  race by construction, pinned by `TestIngestLessonsSymlinkedLayerFileRefused`.
  **Sweep note:** a fresh grep confirms this was the *last* `lstat`/`Lstat`-then-
  `ReadFile` pair over an attacker-reachable path in `internal/core/lifeboat`.

**Burst 3 (2026-07-17, session hardening loop) — unguarded CLI operand reads:**
- `cli.readSource` (the `--pages-json`/`--page-json` memory-ingest transport) and
  `disembark coverage <report.json>` read untrusted operands with a raw
  `os.ReadFile` — no symlink refusal, no size cap. `readSource` was doubly
  exposed: its stdin branch used `io.ReadAll` with NO `LimitReader` either.
  A symlinked operand was FOLLOWED (verified: old `os.ReadFile` returns the
  link target's content — a read-what-you-point-at gap on host-produced /
  cross-repo JSON), and `/dev/zero`/a huge file read unbounded (OOM/hang).
  Both routed through `fsutil.ReadGuarded` (O_NOFOLLOW + regular-file + 8 MiB
  cap) and the stdin branch bounded by a `LimitReader`, matching
  `readLessonsPayload`/`readSynthesisPayload`/`readHookInput`. Watched-fail:
  `TestReadSourceRefusesSymlinkAndOversize` (symlink followed before, refused
  after). **Sweep note:** git subprocess surface audited clean this burst —
  every `exec.Command("git",…)` uses fixed subcommands, env-scrub (`gitEnv`/
  `ScrubbedEnv`), and passes paths via `--stdin -z` (check-ignore), so no
  argument-injection or `-`-leading-flag vector. `readLessonsPayload`/
  `readSynthesisPayload` still carry a benign lstat→read TOCTOU (guarded, but
  the swap window remains) — a candidate follow-up, lower priority as the path
  is user-typed.

**Burst 6 (2026-07-18, session hardening loop) — config-override surface audited
CLEAN; last read-guard TOCTOU closed.** The suspected config-downgrade bypass does
NOT exist: a repo `.abcd/config` cannot lower a bundled hard_fail pattern below its
floor (`applyFloor` clamps severity on BOTH patterns and identity kinds; rank
ordering correct, unknown severities rank least-severe and fail to floor), a
bundled pattern's REGEX is non-replaceable (closing the never-match-regex
downgrade), a malformed override regex fails closed, and the merge is
all-or-nothing. The config skip-list is additive but bounded by the trusted-
worktree model (the local `.abcd/config` is the repo owner's). Only remaining
read-guard item closed: `readLessonsPayload`/`readSynthesisPayload` converted from
guarded-but-lstat→ReadFile (a benign swap window on a user-typed path) to the
shared `readGuardedOperand` (fsutil.ReadGuarded), so EVERY CLI operand read now
uses the one-call guarded primitive. **Convergence:** after six bursts the
security-critical surfaces — redaction patterns (bursts 4-5), config merge (6),
identity matchers, the Redact rewrite, filesystem reads (1-3), and the git
subprocess surface (3) — are audited and hardened; what remains is nitpick-tier.

**Fixed since (was a frontier, now closed):**
- **Terminal-escape sanitization** — the escape/C1 sanitiser is now the shared
  `internal/termsafe.Sanitize`, extended with bidi/RTL-override + zero-width
  (Trojan-Source, CWE-1007), and applied at `audit`, `docs lint`, `capture list`,
  the `intent`/`spec` boards, and `memory` (bare + `ask`) as well as lifeboat.
- **Attack-input regression tests** for 7 of the ~8 landed fixes (`refIsSafe`,
  `checkIgnoredStrict` env-scrub — via `GIT_WORK_TREE`, empirically the var that
  redirects `check-ignore`, not `GIT_DIR` — `rulesRoot`, `notPresent` ENOTDIR,
  `resolveSymlinkDest`, `LoadRegistry` symlink guard, `GitExistingTags` prerelease
  filter). The memory-ingest guarded read is covered by proxy (same
  `fsutil.ReadGuarded` primitive, tested directly and at `LoadRegistry`).

**Checked, currently clean:** `slack_token`'s trailing `\b` is safe (its `{10,}` is
variable-length, so RE2 can end the match at a boundary — unlike `google_api_key`'s
fixed `{35}`); the `IsolatedEnv` config-injection scrub covers `GIT_CONFIG_KEY_*`/
`VALUE_*` by prefix; a raw invalid-UTF-8 byte (e.g. a lone `0x9b`) decodes to
U+FFFD, so the live C1 hazard is a *2-byte-encoded* U+009B — test fixtures must use
the escaped `\u009b`, not `\x9b` (a real vacuous-test trap hit and fixed this round).

---

## Per-principle provenance

Fixes from the prior hunts (the `Unreleased` diff):

- **RE2 `\b` ASCII-only** — `scanner/identity.go` `wordBounded`/`leadingBoundaryOK`;
  `audit/rule_privacy.go` `absPathRe`.
- **Case-sensitive on case-insensitive value** — `scanner/identity.go` `m.email`/
  `m.github` (`(?i)`).
- **Env-inheritance hijack** — `launch/bundle.go:checkIgnoredStrict`,
  `launch/retention.go:GitExistingTags` → `gitutil.IsolatedEnv`.
- **Argument injection** — `lifeboat/graveyard_archaeology.go:refIsSafe`.
- **Incomplete sanitization (C1)** — `lifeboat/coverage.go:sanitize`.
- **Non-injective normalization** — `lifeboat/graveyard.go:idClean` (delete →
  percent-encode).
- **Set-vs-ordered** — `memory/schema.go:equalStringSets`.
- **Case-folding FS gate** — `lifeboat/pack.go:within`/`caseFoldingFS`.
- **SHA-1-only width** — `lifeboat/voyage.go:rootSHARe`.
- **Unguarded trust-boundary read** — `memory/ingest.go`, `memory/provenance.go` →
  `fsutil.ReadGuarded`.
- **Fail-open on unparseable config** — `ahoy/apply.go:stepConfigValues`.
- **Config resolved from wrong root** — `cli/cli.go:rulesRoot`.
- **Parser/serializer disagreement** — `capture/parse.go` (dup key),
  `capture/validate.go` (`resolved_by` type).
- **Suppression downgrades severity** — `scanner/identity.go:localSuppressionSpans`.
- **TOCTOU on mode/rename** — `fsutil/fsutil.go:WriteFileAtomic` (`fchmod`).
- **Path resolved against wrong base** — `ahoy/fsutil.go:resolveSymlinkDest`.
- **Vacuous test** — `cli/hook_session_end_test.go`.

Fixes from the 2026-07-17 gap-driven round (this doc's own evidence):

- **Env-inheritance, completing the `IsolatedEnv` sweep** —
  `scanner/identity.go:ProbeIdentity` → new `gitutil.ScrubbedEnv` (scrub hijack
  vars but keep global config, or the identity probe goes blind);
  `ahoy/store.go:runGit` → `gitutil.IsolatedEnv`.
- **Dead-corner secret regex** — `scanner/patterns.go:google_api_key` (dropped the
  trailing `\b` that a `-`-terminated 39-char key could never satisfy).
- **Case-fold, completing the `(?i)` sweep** — `scanner/identity.go` `m.homeSelf`.
- **SHA-256 width, completing the voyage-key sweep** —
  `history/store.go:rootSHARe`.
- **Non-atomic write, completing the iss-32 consolidation** —
  `identity/identity.go:WritePin` → `fsutil.WriteFileAtomic`.
- **Ignored `Atoi` overflow** — `spec/spec.go:specNum` (over-int64 → 0, not a
  wrapping `MaxInt64`).

---

## The single highest-leverage recommendation for the next hunt

**Run a "sweep-completeness" pass keyed off each hardened primitive.** For every
`gitutil.IsolatedEnv`, `fsutil.ReadGuarded`, `fsutil.WriteFileAtomic`, `refIsSafe`,
`(?i)` identity matcher, and `sanitize` call the codebase already has, grep for the
*unhardened* sibling — the raw `os.Environ()`, `os.ReadFile`, `os.WriteFile`,
positional git arg, case-sensitive matcher, raw terminal print — and prove each
remaining instance is either hardened or explicitly out of scope under the
trusted-worktree model. Four of this round's confirmed bugs were exactly such
stragglers; the primitives exist, but the sweep that retires their unsafe
predecessors is incomplete. This converts open-ended hunting into a finite,
greppable, verifiable checklist.
