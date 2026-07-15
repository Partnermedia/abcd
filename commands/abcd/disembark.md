---
name: disembark
description: Pack a lifeboat from a repository into a destination directory — read-only over the source, behind a destination safety gate, secret-scanned before any write. Point it at any repo (including a dead or archived one) and write the lifeboat elsewhere.
argument-hint: "<source-repo> <dest> | plan <source-repo>"
---

# `/abcd:disembark` — pack a lifeboat

Mine a repository's record into a portable lifeboat at `<dest>`. The source is
**never written** — a probe and plan read it read-only, and the pack writes only
to the destination. This is the out-of-tree model (adr-35): point it at any repo,
touch nothing, write elsewhere.

## Dry run first (recommended)

Show the exact file set a pack would write, without writing anything:

```bash
abcd disembark plan <source-repo> --json
```

Report `file_count`, `total_bytes`, `manifest_sha256`, and any `omissions` (records
too large or unreadable to carry). Then pack for real.

## Pack

```bash
abcd disembark pack <source-repo> <dest> --json
```

Summarise the JSON result for the user:

- `dest` — where the lifeboat was written.
- `files_written` / `bytes_written` — the size of the lifeboat.
- `manifest_sha256` — the pinned hash over every file (matches `<dest>/_provenance.json`).
- `voyage_appended` — whether the operator-level voyage ledger recorded the pack
  (`~/.abcd/voyage/<source-root-sha>/disembark/history.jsonl`); `voyage_note`
  explains a skip (e.g. a source with no root-commit SHA).
- `omissions` — any records deliberately left out, declared rather than dropped.

## What the pack refuses

The **destination safety gate** protects real work. A pack refuses unless `<dest>`
is absent, an empty directory, or an existing lifeboat abcd produced (it carries a
parseable `_provenance.json`). It also refuses a symlinked destination, one inside
a `.git/` directory, or one that overlaps the source tree. And it **refuses on a
hard-fail secret** in the planned bytes — a secret is fixed at source, never
redacted into the artefact. Relay the refusal message so the user knows what to fix.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd disembark ...` from the repo root, or build it with `make build`.

**User input:** $ARGUMENTS
