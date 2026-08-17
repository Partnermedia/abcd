---
name: lint
description: Check whether the current repo conforms to the working conventions — the three-tier .abcd/ layout, an AGENTS.md router, durable decisions, current docs, privacy hygiene — by invoking the abcd binary. Strictly read-only; performs zero writes.
argument-hint: ""
---

# `/abcd:lint` repo-conformance check

Run the abcd binary's read-only conformance lint for the current repo and
present the result. This command performs **zero writes** — it reports gaps, it
never fixes them (remediation stays with `/abcd:prepare-this-repo`).

Run:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" lint --json
```

Then summarise the JSON for the user. Its shape is `{ "findings": [ … ],
"skipped": [ … ] }`:

- `findings` — each has a stable `ruleId`, a `severity` (`error` or `warn`), a
  `file` and `line` (line `0` means the finding is not tied to one line), a
  `message`, and a `fix`. Group them by severity: report `error` findings first
  (these fail conformance), then `warn` findings (advisory). For each, give the
  `file:line`, the `message`, and the `fix`.
- `skipped` — rule ids that did not apply to this repo (e.g. `docs-currency`
  when there is no `docs/`). Mention them as "not applicable", not as failures.

State the outcome plainly: if there are no findings the repo conforms; otherwise
lead with how many errors and warnings there are. The process exit code is the
Conftest tri-state — `0` clean, `1` warnings only, `2` any error — so
`abcd lint` can also gate a repo's CI.

A finding on a deliberately illustrative line can be waived by adding
`abcd-lint:allow` on that line (the earlier `abcd-audit:allow` spelling is honoured too).

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
