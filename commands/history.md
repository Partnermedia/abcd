---
name: history
description: Manage the native session-transcript store for this repo by invoking the abcd binary. list and show are read-only; capture is the redacting write path. The store is keyed on the repo's root-commit SHA and every stored transcript is redacted on write.
argument-hint: "list | show <session-id-or-filename> | capture <transcript-file>"
---

# `/abcd:history` — session-transcript store

The native session-transcript store at
`~/.abcd/history/<root-sha>/transcripts/`, keyed on this repo's root-commit
SHA. `list` and `show` **perform zero writes**; `capture` is the only path that
writes, and it redacts on write — no live secret or absolute home path can
survive capture.

## List

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" history list --json
```

Summarise each record newest-first: `captured_at`, `session_id`, `source_kind`,
and the `redacted_secrets` / `redacted_home_paths` counts. An empty list means
no transcripts are stored for this repo yet.

## Show

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" history show <session-id-or-filename> --json
```

Fetch one record's metadata and its full redacted `body`, matched by session id
(newest when a session has several records) or by the record filename. Present
the metadata and, if the user wants it, the body.

## Capture

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" history capture <transcript-file> --json
"${CLAUDE_PLUGIN_ROOT}/abcd" history capture --session <id> - < transcript.txt
```

Read a raw transcript from a file argument (or stdin with `-`), redact it
through the scanner in a two-stage fail-closed pass, and store the record. The
session id defaults to the transcript filename; reading from stdin requires
`--session`. `--kind` selects the source kind (`native` — the default — or
`specstory-import`). The write is idempotent on the source's content hash: an
identical transcript already stored is a no-op. If any hard-fail secret or the
caller's own home path survives redaction, capture refuses to write.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
