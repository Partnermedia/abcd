---
name: docs
description: Lint this repo's documentation for currency — change-narration ("previously", "formerly", …), broken relative links, and stray root markdown — by invoking the abcd binary. Read-only; performs zero writes.
argument-hint: "[lint]"
---

# `/abcd:docs` documentation-currency lint

Run the abcd binary's docs-currency lint for the current repo and present the
result. This command performs **zero writes**.

Run:

```bash
abcd docs lint --json
```

Then summarise the JSON for the user:

- `blockers` — how many blocker findings exist; any blocker fails the gate.
- `findings` — for each, its `File`, `Line`, `RuleID`, `Severity`, and
  `Message`; group them so the user sees what to fix.

The lint enforces present-tense docs: unambiguous change-narration (`previously`,
`formerly`, `renamed from`, `has been replaced`, `we switched`, `to be
implemented`) blocks, while phrases that also describe present state
(`deprecated`, `no longer`, `migrated from`) warn advisorily rather than block.
It also checks that relative links resolve and that no stray markdown sits at the
repo root (it belongs under `docs/`). Point the user at the offending file and
line for each finding, and note whether it is a blocker or a warning.

If `blockers` is zero the docs are currency-clean.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd docs lint --json` from the repo root, or tell the user to
build it with `make build`.

**User input:** $ARGUMENTS
