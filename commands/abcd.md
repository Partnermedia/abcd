---
name: abcd
description: Top-level where-am-i status board and record-id dispatch. Bare `/abcd` renders a read-only snapshot of the current directory; `/abcd <record-id>` (iss-N, itd-N, spc-N, adr-N) reports what that record is and the next move. Strictly read-only.
argument-hint: "[<record-id>]"
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
exist.

## Record-id dispatch

Bare answers *what can I do*; `abcd <id>` answers *what is this, and what is
my next move*. A positional matching `^(iss|itd|spc|adr)-[0-9]+$` locates the
record in its store — any status folder or bucket — and renders it read-only:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" <record-id> --json
```

Summarise the `id`, `family`, `status`, `title`, `path`, the `links` edges
(`spec_id`, `intent`, `promoted_to`, `resolved_by.*`, `superseded_by` as
present), and each entry in `next_moves` — the concrete lifecycle move
(e.g. a draft intent points at the planning interview and `intent plan`; an
open issue points at `capture promote` / `resolve` / `wontfix`; decisions are
read). A shape-matching id found in no store exits non-zero naming the stores
searched. Any other positional is refused as an unknown command (exit 2) —
there is no `status` alias.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
