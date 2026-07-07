---
name: capture
description: Capture issues to the structured per-repo ledger and query them, by invoking the abcd binary. Bare invocation is a read-only status render; list/resolve/wontfix act on the ledger.
argument-hint: "[text] | list --open|--resolved|--wontfix|--all | resolve <iss-N> <note> | wontfix <iss-N> <reason>"
---

# `/abcd:capture` — issue ledger

The lightweight write side of the structured issue ledger under
`.abcd/development/activity/issues/`. Every issue gets a stable `iss-N` id,
schema-checked frontmatter, and folder-as-status (`open/`, `resolved/`,
`wontfix/`). Bare invocation **performs zero writes**.

## Status (bare)

To render recent captures and counts:

```bash
abcd capture --json
```

Summarise the JSON for the user: `open_count` / `resolved_count` /
`wontfix_count`, and for each entry in `recent_open` its `id`, `severity`, and
`slug`. No `iss-*.md` file is created, moved, or mutated by this invocation.

## Capture an issue

Append a structured issue from free-form text:

```bash
abcd capture "<text>" --json
```

Provide provenance and taxonomy through flags when known (each falls back to a
default): `--severity` (`nitpick|minor|major|critical`, default `minor`),
`--category` (default `observation`), `--source` (default `user-observation`),
`--found-during` (session/command context, default `manual-capture`),
`--found-at` (optional repo-relative path), `--slug` (overrides the slug derived
from the text). Report the new `id`, `status`, and `path` from the JSON.

## Query the ledger

`list` is the one earned filter-flag exception — a filter is **required**:

```bash
abcd capture list --open --json      # or --resolved / --wontfix / --all
```

The unfiltered form `abcd capture list` exits 2 with a "choose a filter"
message; there is no implicit default. Summarise each issue's `id`, `status`,
`severity`, and `slug`, sorted numerically by `iss-N`.

## Resolve / wontfix

```bash
abcd capture resolve <iss-N> "<resolution-note>" --json
abcd capture wontfix <iss-N> "<reason>" --json
```

Each moves the issue out of `open/` and records the note; report the `id` and
the `from_status -> to_status` transition from the JSON.

Promoting an issue to an intent (`/abcd:capture promote <iss-N>`) is
skill-orchestrated, not a binary sub-verb — it runs the intent-new interview
seeded with the issue body.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd capture …`
from the repo root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
