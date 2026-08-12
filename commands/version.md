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
provisions the binary into the plugin root. If that path is absent, fall back to
`abcd` on `PATH`, and run `"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install` to put one
there. In a source checkout of this repo, and only there, `go run ./cmd/abcd` is
a third rung; the published plugin payload carries no `cmd/`, so it cannot fire
for a plugin user.

**User input:** $ARGUMENTS
