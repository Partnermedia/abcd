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
abcd --json
```

Then summarise the JSON for the user: the directory, whether it is a git repo,
whether the abcd development record is present, and which `.abcd/` work tiers
exist. `status` is a positional alias for the same bare render.

If the `abcd` binary is not on `PATH`, tell the user to build it with
`make build` (or run `go run ./cmd/abcd --json` from the repo).

**User input:** $ARGUMENTS
