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
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
