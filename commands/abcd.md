---
name: abcd
description: Top-level where-am-i status board. Bare `/abcd` renders a read-only snapshot of the current directory — git repo, whether the abcd development record is present, and which .abcd/ work tiers exist. Strictly read-only.
argument-hint: "[status]"
---

# `/abcd` where-am-i

Run the abcd binary's read-only status board for the current repo and present the
result. This command performs **zero writes**.

Run:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" --json
```

Then summarise the JSON for the user: the directory, whether it is a git repo,
whether the abcd development record is present, and which `.abcd/` work tiers
exist. `status` is a positional alias for the same bare render.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root. If that path is absent, fall back to
`abcd` on `PATH`, and run `"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install` to put one
there. In a source checkout of this repo, and only there, `go run ./cmd/abcd` is
a third rung; the published plugin payload carries no `cmd/`, so it cannot fire
for a plugin user.

**User input:** $ARGUMENTS
