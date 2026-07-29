---
id: itd-105
slug: a-fresh-plugin-install-just-works-the-hook-binary-survives
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: [itd-67]
severity: critical
---

# A fresh plugin install just works: the hook binary survives install and every update

## Press Release

> **Installing the abcd plugin from its marketplace now yields a working hook
> surface on the very first session — and keeps working across every plugin
> update.** Today the hooks call `$CLAUDE_PLUGIN_ROOT/abcd`, but the plugin
> root is a git clone of the repo, and the repo (correctly) commits no binary:
> every fresh install and every update produces a plugin root with no `abcd`
> in it, so every hook fails with "No such file or directory" until someone
> hand-copies a binary in. With this intent, the binary gap closes itself —
> either the surface bootstraps the checksum-verified release binary into the
> plugin root when it is missing, or the hooks resolve `abcd` from PATH so the
> documented one-line install is sufficient — and an update never reintroduces
> the gap, because each update lands in a fresh commit-stamped directory that
> starts out binary-less.

> "I installed the plugin exactly as the README says, opened a session in a
> clean repo, and got two hook errors before I typed anything," said Alice,
> trying abcd on a second machine. "The README told me to install the binary
> to my PATH and I had — but the hooks don't look there. And when I turned on
> auto-update so a teammate's fixes would reach me, every update broke it
> again. Now the first session after install just works, and so does the
> first session after every update."

## Why This Matters

The distribution story ([[itd-67-installable-versioned-plugin]]) made the repo
its own marketplace, and the README documents both halves: install the
checksum-verified release binary to PATH, and install the plugin from the
marketplace. But the two halves do not meet. `hooks/hooks.json` invokes
`"$CLAUDE_PLUGIN_ROOT/abcd"` — a path inside the harness's plugin cache — while
the PATH install puts the binary in `/usr/local/bin`. Nothing in the repo,
the payload, or the `ahoy` machinery provisions the plugin-root binary: `ahoy
install` only creates the PATH symlink *pointing at* `$pluginRoot/abcd`,
assuming it already exists. A fresh marketplace install therefore fails every
hook (`SessionStart`, `UserPromptSubmit`, the Bash guard) in every session,
and the guard failure means shell commands run unguarded — the surface
degrades exactly where it claims to protect.

The update path makes this structural rather than one-off. The harness caches
each installed plugin version in a directory stamped with the source commit
(`cache/<marketplace>/abcd/<sha>/`). Every update — manual or auto — is a
fresh clone into a fresh directory, which starts with no binary; a hand-copied
fix evaporates on the next update. A maintainer who enables marketplace
auto-update precisely so that merged fixes flow to their machines gets the
opposite: every merge re-breaks the hook surface. There is also a version-skew
corollary: the plugin surface (hooks, commands, agents) tracks the repository
tip, while the newest binary is the last tagged release — fixes that merge
without a release cut can leave the surface ahead of any binary a user can
download ([[itd-73-derived-versioning]], [[itd-69-plugin-metadata-lockstep-update]]).

## What's In Scope

- A defined, tested answer to "how does `$CLAUDE_PLUGIN_ROOT/abcd` come to
  exist" for a marketplace install. The two candidate shapes, one to be
  chosen by the spec: (a) **bootstrap-on-missing** — a hook entry (or guarded
  preamble in each hook command) that, when the plugin-root binary is absent,
  downloads the release binary for the host OS/arch, verifies its SHA-256
  against the release `checksums.txt`, refuses on mismatch, and installs it
  into the plugin root; or (b) **PATH resolution** — the hook commands resolve
  `abcd` from PATH (with a clear, actionable error naming the README install
  when absent), making the documented binary install sufficient.
- Update survival: whichever shape is chosen must hold across a plugin update
  into a fresh commit-stamped cache directory — the first session after an
  update has a working hook surface with no manual step.
- Failure honesty: while the binary is genuinely unavailable (offline, no
  release asset for the platform, checksum mismatch), hooks fail with one
  clear message naming the fix — never a raw shell "No such file or directory",
  and the existing guard-health machinery (`abcd ahoy`) reports the gap.
- Version-skew visibility: when the installed surface (plugin commit) is newer
  than the newest published binary, the surface says so rather than silently
  running old-binary/new-surface.
- The README plugin section documents the chosen contract — what a user must
  install by hand (possibly nothing) and what provisions itself.

## What's Out of Scope

- The release/versioning pipeline itself — how versions are derived, bumped,
  tagged, and published stays with [[itd-73-derived-versioning]] and
  [[itd-69-plugin-metadata-lockstep-update]]; this intent only consumes
  published release artefacts.
- Marketplace auto-update policy — whether a user enables the harness's
  auto-update is their choice; this intent makes either answer safe, it does
  not toggle harness settings.
- Building from source as the recovery path — the fallback contract is about
  released binaries; dev-mode (`ahoy install --dev`) already covers the
  source-tip workflow for maintainers.
- Any change to the transport-agnostic core's behaviour — this is a
  distribution/wiring concern at the plugin surface.

## Acceptance Criteria

> _Given-When-Then per the itd-1 discipline._

- **Given** a machine that has never installed abcd, **when** a user adds the
  marketplace, installs the plugin, and starts a session in any repo, **then**
  every hook (`SessionStart`, `UserPromptSubmit`, `PreToolUse` guard,
  `PreCompact`, `SessionEnd`) executes successfully — no "No such file or
  directory", no unguarded-shell warning.
- **Given** a working install, **when** the plugin updates to a new commit
  (fresh cache directory), **then** the first session after the update has a
  working hook surface with no manual step.
- **Given** the bootstrap shape, **when** the downloaded binary's SHA-256 does
  not match the release `checksums.txt` (or the manifest lacks the platform),
  **then** nothing is installed and the hook failure message names the
  mismatch — a corrupted or unpublished artefact never runs.
- **Given** the binary is genuinely unavailable (offline first session),
  **when** hooks fire, **then** the failure is a single clear message naming
  the recovery step, and `abcd ahoy` reports the binary gap in guard health.
- **Given** an installed surface newer than the newest published binary,
  **when** a session starts, **then** the skew is surfaced to the user rather
  than silently running a stale binary against newer surface expectations.

## Open Questions

- Which shape wins — bootstrap-on-missing (self-healing, but hooks gain a
  network dependency and a trust boundary) or PATH resolution (simple, but
  couples hook correctness to an out-of-band install and loses the pinned
  binary-per-plugin-version property the `ahoy` symlink design implies)?
  A hybrid (PATH fallback while bootstrap is pending) may be the first slice.
- Should the bootstrap pin the binary matching the installed plugin commit's
  release (coherence) or always take `releases/latest` (freshness)? The skew
  question above depends on this.
- Does the guard hook's fail-open/fail-closed posture change while the binary
  is missing? Today the wrapper reports "shell commands run UNGUARDED" —
  is that acceptable for the bootstrap window, or should the first session
  block Bash until the guard is real?
- Where does the bootstrap live so it needs no binary to run — a committed
  POSIX-sh script under `hooks/`, given the binary is exactly what is missing?
