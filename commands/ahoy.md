---
name: ahoy
description: Detect and repair abcd's install/update state for the current repo — folder kind, plugin-root status, and outstanding gaps — by invoking the abcd binary. Bare invocation performs zero writes.
argument-hint: "[status | install | uninstall | doctor | dry-run]"
---

# `/abcd:ahoy` install/update detector

Run abcd's install/update engine for the current repo and present the result.
Bare invocation and the `doctor` and `dry-run` sub-verbs perform **zero writes**;
`install` and `uninstall` are the two that change the repo, and each says so
before it runs.

Read `$ARGUMENTS` for the sub-verb. No argument, or `status`, is the bare
read-only detection pass below.

## Bare — read-only detection

Run:

```bash
abcd ahoy --json
```

Then summarise the JSON for the user:

- `folder_kind` — `managed-repo`, `unmanaged-repo`, or `unmanaged-folder`.
- `plugin_root_status` and `root_sha` — where abcd is anchored.
- `signals.install_mode` — the PATH-entry install mode: `dev (tip build)` when
  the track-latest dogfood shim is installed, `pinned` for the built-binary
  symlink, empty when there is nothing on `PATH` yet. Report it so a dev install
  is never invisible.
- `banlist` — the two-layer name guard, when the folder is a repo: `hook` and
  `merge_hook` (`installed` / `absent` / `foreign` / `unreadable`), whether this
  clone is armed (`hooks_path_armed`), `public_family`, and the private layer's
  state on this machine (`private_store`, and its shape — `private_keyed`,
  `private_entries`, `private_unparsed`). Never report a `foreign` hook as abcd's
  guard, and never report a committed hook as a running one. Relay `reach`
  verbatim: it is the one sentence stating what the private layer does NOT cover,
  and a paraphrase drops the half that matters.
- `gaps` — how many are outstanding, and for each actionable one its `title`,
  `category`, and `fix_hint`; call out which are `required`.

If there are actionable gaps, tell the user to run `/abcd:ahoy install` to apply
them. If `folder_kind` is `unmanaged-folder`, note there is nothing to act on
(not a git repo, no abcd markers).

## `install` — apply the outstanding gaps

```bash
abcd ahoy install --json
```

**This writes.** It applies the actionable gaps the detection pass found — the
marker block, the `.abcd/` scaffolding, the owned `PATH` symlink. Report the
returned `status` and what changed; the engine prompts before an ambiguous
adoption, so surface any prompt to the user rather than answering it for them.

Prompts read stdin whether or not stdin is a terminal, so an answer can be
relayed without one:

```bash
yes | abcd ahoy install
```

**One line answers one question.** The install asks one approval per gap
category present — often several — and every line after the last one you supply
reads end-of-input and DECLINES. `yes` is the reliable form because it never
runs out; a single `printf 'y\n'` answers the first question only and silently
declines the rest. The questions come in a fixed order (dependency,
safe-autocreate, config-change, user-state, plugin-owned), so a scripted stream
of specific answers lines up with them. Each answer is echoed back, so the
transcript shows what was asked and what it was answered — read it back rather
than assuming.

That is a channel for passing on an answer the user has GIVEN — ask first, then
pipe; it is never a licence to answer on their behalf. Note that `yes |`
approves EVERY question, so only reach for it once the user has agreed to all of
them.

**Stdin must end, or the prompt waits.** With stdin at end-of-input every
question declines, so a run that was told nothing writes nothing — but a stdin
that is held open and silent (a pipe from a still-running command) makes the
prompt WAIT for the answer that never comes, rather than declining. For a run
that must not block and must not prompt, close stdin or pre-answer everything:
`abcd ahoy install --yes --refuse-adopt < /dev/null`.

`--yes` approves every resolvable category but never adopts the optional
git-identity pin, because the pin records whatever git identity is currently
configured. When the result carries `optional_skipped`, report it and offer the
`yes | abcd ahoy install` form above as the way to apply it.

For dogfooding abcd itself, `abcd ahoy install --dev` installs a track-latest
shim instead of the pinned-binary symlink: the `PATH` entry rebuilds abcd from
the source tip on every call and fails loudly on a broken build. Re-running
`abcd ahoy install` without `--dev` switches back to the pinned symlink.

## `uninstall` — reversible marker-only removal

```bash
abcd ahoy uninstall --json
```

**This writes.** It removes the BEGIN/END marker block and abcd's own `PATH`
symlink and leaves `.abcd/` intact, so the repo's record survives. Report
`marker.removed` and the symlink note. It never touches `hooks.json`.

## `doctor` — the full read-only report

```bash
abcd ahoy doctor --json
```

Runs the same detection pass plus a read-only audit sweep and reports **every**
gap, including user-scope state the bare render leaves out. Writes nothing.
Report the folder kind, the detection-gap count, and the audit-gap count, then
the per-gap detail from the JSON. This is the sub-verb to reach for when the bare
render says a repo is healthy and the user's experience says otherwise.

## `dry-run` — the canonical detection envelope

```bash
abcd ahoy dry-run
```

Renders the canonical `DetectionResult` JSON envelope and writes nothing — the
same pass `install` would apply, shown rather than applied. `dry-run` always
emits JSON, so it takes no `--json` flag. Use it when the user wants to see
exactly what an install would do before letting it run.

## Scoping note: `identity-check` is CLI-only

`abcd ahoy identity-check` exits non-zero when the git commit identity diverges
from the committed pin. It exists to be wired into a pre-commit hook or CI, where
its exit code is the whole point, so it stays a bare-CLI entrypoint rather than a
plugin sub-verb; report it only if a user asks how the identity gate fails
closed.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd ahoy --json` from the repo root, or run
`go run ./cmd/abcd ahoy install` to put a binary on `PATH`.

**User input:** $ARGUMENTS
