---
name: version
description: Print the installed abcd version by invoking the abcd binary. Read-only.
---

# `/abcd:version`

Report the installed abcd version. This command performs **zero writes**.

Run:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" version --json
```

Then tell the user the `name` and `version` from the JSON.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
