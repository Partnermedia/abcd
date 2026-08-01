---
name: ahoy
description: Detect abcd's install/update state for the current repo — folder kind, plugin-root status, and outstanding gaps — by invoking the abcd binary. Strictly read-only; performs zero writes.
argument-hint: "[status]"
---

# `/abcd:ahoy` install/update detector

Run the abcd binary's read-only detection pass for the current repo and present
the result. This command performs **zero writes**.

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

If there are actionable gaps, tell the user to run `abcd ahoy install` to apply
them. If `folder_kind` is `unmanaged-folder`, note there is nothing to act on
(not a git repo, no abcd markers).

For dogfooding abcd itself, `abcd ahoy install --dev` installs a track-latest
shim instead of the pinned-binary symlink: the `PATH` entry rebuilds abcd from
the source tip on every call and fails loudly on a broken build. Re-running
`abcd ahoy install` without `--dev` switches back to the pinned symlink.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd ahoy --json` from the repo root, or tell the user to build it
with `make build`.

**User input:** $ARGUMENTS
