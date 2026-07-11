# `/abcd:history` — Session-Transcript Store

`/abcd:history` manages the native session-transcript store — a per-repo,
redact-on-write archive of raw session transcripts, keyed on the repo's
**root-commit SHA**. The store lives outside the repo at
`~/.abcd/history/<root-sha>/transcripts/`, with a per-repo `meta.json`
(`root_commit`, `name`, `github`, and a corpus block) alongside it. `list` and
`show` **perform zero writes**; `capture` is the only write path, and it redacts
on write — no live secret or absolute home path survives capture.

## Sub-verbs

- **`/abcd:history list`** — list stored transcripts for this repo, newest
  first. Each record reports `captured_at`, `session_id`, `source_kind`, and the
  `redacted_secrets` / `redacted_home_paths` counts. An empty list means nothing
  is stored for this repo yet.
- **`/abcd:history show <session-id-or-filename>`** — show one record's metadata
  and its full **redacted** body, matched by session id (newest when a session
  has several records) or by record filename.
- **`/abcd:history capture [<transcript-file>|-]`** — redact and store a raw
  transcript, read from a file or from stdin (`-`). Capture is **fail-closed on
  redaction** and **idempotent on content hash** (re-capturing identical content
  does not duplicate). Flags: `--kind` (`native` | `specstory-import`, default
  `native`) and `--session` (the record's session id; defaults to the transcript
  filename, and is **required** when reading from stdin).

Bare `abcd history` prints command usage — it does **not** render a status board.
The global `--json` flag emits machine-readable output for every sub-verb.

## Redaction boundary

Capture is a trust boundary: the raw transcript is untrusted input, and the
store is a durable artefact that may later feed the memory substrate. Every
transcript is redacted on write (secrets and absolute home paths), the redaction
counts are recorded on the record, and a redaction failure refuses the write
rather than storing unredacted content.

## Composition

The store is the substrate the transcript-harvest path (and, later, the memory
distiller) reads from — history captures raw sessions; `memory` distils curated
knowledge from them. The store is keyed per repo, so transcripts never leak
across projects.

## References

- Plugin command: [`commands/abcd/history.md`](../../../../commands/abcd/history.md)
- Store + redaction engine: `internal/core/history`
- Install-time provisioning of the per-repo store: [`01-ahoy.md`](01-ahoy.md)
