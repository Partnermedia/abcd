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
abcd history list --json
```

Summarise each record newest-first: `captured_at`, `session_id`, `source_kind`,
and the `redacted_secrets` / `redacted_home_paths` counts. An empty list means
no transcripts are stored for this repo yet.

## Show

```bash
abcd history show <session-id-or-filename> --json
```

Fetch one record's metadata and its full redacted `body`, matched by session id
(newest when a session has several records) or by the record filename. Present
the metadata and, if the user wants it, the body.

## Capture

```bash
abcd history capture <transcript-file> --json
abcd history capture --session <id> - < transcript.txt
```

Read a raw transcript from a file argument (or stdin with `-`), redact it
through the scanner in a two-stage fail-closed pass, and store the record. The
session id defaults to the transcript filename; reading from stdin requires
`--session`. `--kind` selects the source kind (`native` — the default — or
`specstory-import`). The write is idempotent on the source's content hash: an
identical transcript already stored is a no-op. If any hard-fail secret or the
caller's own home path survives redaction, capture refuses to write.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd history …`
from the repo root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
