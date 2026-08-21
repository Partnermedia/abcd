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
"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy --json
```

Then summarise the JSON for the user:

- `folder_kind` — `managed-repo`, `unmanaged-repo`, or `unmanaged-folder`.
- `plugin_root_status` and `root_sha` — where abcd is anchored.
- `signals.install_mode` — the PATH-entry install mode: `dev (tip build)` when
  the track-latest dogfood shim is installed, `pinned` for the abcd-owned copy
  of the verified release binary, empty when there is nothing on `PATH` yet.
  Either non-empty value may carry a trailing ` (shadowed on PATH)` when another
  `abcd` comes first on `PATH`. Report it so a dev install — or an entry abcd
  wrote that is not the one that runs — is never invisible.
- `vintage` and `staleness` — the running binary's build revision (in a source
  checkout) or pinned version, and whether it is up to date, stale, or of an
  undeterminable vintage relative to the on-disk reference. Report them so a
  binary running behind its own source is never silent. The comparison is
  disk-only — no network.
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
"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install --json
```

**This writes.** It applies the actionable gaps the detection pass found — the
marker block, the `.abcd/` scaffolding, the owned `PATH` entry. Report the
returned `status`, what changed, and any `notes` — a note is a refusal, stating
something abcd deliberately did not do and why. The engine prompts before an
ambiguous adoption, so surface any prompt to the user rather than answering it
for them.

The `PATH` entry goes to `~/.local/bin` (created when absent), or to an
abcd-owned entry already on `PATH`, which is adopted exactly where it stands.
`--bin-dir <dir>` names a different directory — the only way to reach a
system-wide one — and fails loudly when it is not writable. abcd never escalates
privileges, so never suggest re-running any of this under `sudo`; report the
failure and let the user pick a directory they own. If the report carries a
`path.bin_dir_not_on_path` gap, relay its one-line `export` fix verbatim and
leave the user's shell profile alone.

A `symlink.shadowed` gap (or a note saying the same) means another `abcd` comes
first on `PATH`, so the entry abcd just wrote is NOT what runs — typically a
binary an older install copied into a system directory. Relay it prominently:
the install is not finished from the user's point of view. abcd will not remove
that binary, and neither should you offer to; state the two remedies it gives
(delete the stale one, or install ahead of it with `--bin-dir`) and let the user
choose.

Prompts read stdin whether or not stdin is a terminal, so an answer can be
relayed without one:

```bash
yes | "${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install
```

**One line answers one question.** The install asks one approval per gap
category present — often several — and every line after the last one you supply
reads end-of-input and DECLINES. `yes` is the reliable form because it never
runs out; a single `printf 'y\n'` answers the first question only and silently
declines the rest. The questions come in a fixed order (dependency,
safe-autocreate, config-change, user-state, plugin-owned), so a scripted stream
of specific answers lines up with them. Each answer is echoed back, so the
transcript shows what was asked and what it was answered — read it back rather
than assuming. Under `set -o pipefail` the pipeline reports 141: `yes` takes
SIGPIPE when abcd stops reading, by design — judge the run by abcd's own output
and exit status, not the pipeline's.

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
`yes |` form above as the way to apply it.

For dogfooding abcd itself, `abcd ahoy install --dev` installs a track-latest
shim instead of the pinned owned copy: the `PATH` entry rebuilds abcd from
the source tip on every call and fails loudly on a broken build. Re-running
`abcd ahoy install` without `--dev` switches back to the pinned owned copy.

## `uninstall` — reversible marker-only removal

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy uninstall --json
```

**This writes.** It removes the BEGIN/END marker block and abcd's own `PATH`
entry — the owned copy (or a legacy pinned symlink), found wherever it sits on
`PATH`, along with its provenance record — and leaves `.abcd/` intact, so the
repo's record survives. The persistent download cache is left to the harness's
own uninstall to delete. Report `marker.removed` and the entry note; the
receipt's `symlink.target` is already rendered in tilde form, so relay it as
given rather than expanding it. It never touches `hooks.json`. An entry that was
installed with `--bin-dir` into a directory outside `PATH` cannot be found by a
`PATH` scan — pass the same `--bin-dir <dir>` to `uninstall` to remove it.

## `doctor` — the full read-only report

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy doctor --json
```

Runs the same detection pass plus a read-only audit sweep and reports **every**
gap, including user-scope state the bare render leaves out. Writes nothing.
Report the folder kind, the detection-gap count, and the audit-gap count, then
the per-gap detail from the JSON. This is the sub-verb to reach for when the bare
render says a repo is healthy and the user's experience says otherwise.

## `dry-run` — the canonical detection envelope

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy dry-run
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

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
