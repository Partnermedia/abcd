---
name: version
description: Print the installed abcd version, install mode, and vintage by invoking the abcd binary. Read-only unless --check is passed.
---

# `/abcd:version`

Report the installed abcd version, install mode, and vintage — the running
binary's build revision (in a source checkout) or pinned version, and whether it
is up to date, stale, or of an undeterminable vintage relative to the on-disk
reference. This command performs **zero writes** and touches no network.

Run:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" version --json
```

Then tell the user the `name`, `version`, `install_mode`, `vintage`, and
`staleness` from the JSON.

**Checking for a newer release.** Only when the user explicitly asks whether a
newer version exists, add `--check`:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" version --check --json
```

`--check` reaches the network — it fetches the latest release once, compares,
and reports under `check` (with its `source` named). The only other verb that
does is `update`, which completes what `--check` reports; every other path
reads only what is on disk.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
