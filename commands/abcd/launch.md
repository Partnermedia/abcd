---
name: launch
description: Preview the public launch — the file bundle, the secret/PII scan, and the release gates — by invoking the abcd binary in dry-run mode. Strictly read-only; performs zero writes and never publishes.
argument-hint: "[dry-run]"
---

# `/abcd:launch` release preview

Run the abcd binary's launch dry-run for the current repo and present the
result. This command performs **zero writes** and never publishes anything.

Run:

```bash
abcd launch --dry-run --json
```

Then summarise the JSON for the user:

- `version` — the version the release would carry.
- `files` — how many files the bundle would include.
- `scan.hard_fails` — secret/PII findings that would block the release.
- `would_publish` — whether every gate passes.
- `would_refuse_on` — if non-empty, the gates that would refuse, so the user
  knows what to fix before a real launch.

This is preview-only: publishing is not driven from this command.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd launch --dry-run --json` from the repo root, or tell the user
to build it with `make build`.

**User input:** $ARGUMENTS
