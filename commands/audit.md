---
name: audit
description: Check whether the current repo conforms to the working conventions — the three-tier .abcd/ layout, an AGENTS.md router, durable decisions, current docs, privacy hygiene — by invoking the abcd binary. Strictly read-only; performs zero writes.
argument-hint: ""
---

# `/abcd:audit` repo-conformance check

Run the abcd binary's read-only conformance audit for the current repo and
present the result. This command performs **zero writes** — it reports gaps, it
never fixes them (remediation stays with `/abcd:prepare-this-repo`).

Run:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" audit --json
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
`abcd audit` can also gate a repo's CI.

A finding on a deliberately illustrative line can be waived by adding
`abcd-audit:allow` on that line.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root. If that path is absent, fall back to
`abcd` on `PATH`, and run `"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install` to put one
there. In a source checkout of this repo, and only there, `go run ./cmd/abcd` is
a third rung; the published plugin payload carries no `cmd/`, so it cannot fire
for a plugin user.

**User input:** $ARGUMENTS
