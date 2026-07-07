---
name: ahoy
description: Detect abcd's install/update state for the current repo — folder kind, plugin-root status, and outstanding gaps — by invoking the abcd binary. Strictly read-only; performs zero writes.
argument-hint: "[status]"
---

# `/abcd:ahoy` install/update detector

Run the abcd binary's read-only detection pass for the current repo and present
the result. This command performs **zero writes**.

Run:

```bash
abcd ahoy --json
```

Then summarise the JSON for the user:

- `folder_kind` — `managed-repo`, `unmanaged-repo`, or `unmanaged-folder`.
- `plugin_root_status` and `root_sha` — where abcd is anchored.
- `gaps` — how many are outstanding, and for each actionable one its `title`,
  `category`, and `fix_hint`; call out which are `required`.

If there are actionable gaps, tell the user to run `abcd ahoy install` to apply
them. If `folder_kind` is `unmanaged-folder`, note there is nothing to act on
(not a git repo, no abcd markers).

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd ahoy --json` from the repo root, or tell the user to build it
with `make build`.

**User input:** $ARGUMENTS
