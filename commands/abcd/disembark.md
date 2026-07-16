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

## Graveyard interpretation (layer 3)

A packed lifeboat carries a **graveyard** of what the project tried and left
behind: `graveyard/archaeology.json` (deterministic git evidence — reverts,
unmerged branches, deleted paths, removed dependencies, wholesale rewrites) and
`graveyard/abandoned.json` (what the record itself declared dead — superseded
intents and ADRs, wontfix issues, rejected options). These are evidence only; no
interpretation.

To add interpreted lessons, run the `graveyard-interpreter` agent over those two
files, have it emit a lesson JSON document (each lesson **citing** the finding ids
it rests on), write that document to a file, then:

```bash
abcd disembark graveyard <lifeboat-dir> --lessons-json <path>   # or - for stdin
```

The verb is a **cite-or-be-dropped** gate. A lesson survives only if at least one
of its `evidence` refs resolves to a live finding id from layers 1/2; a lesson
that cites nothing (or only dead refs) is **dropped** — reported in the result,
never fatal. Surviving lessons are written to `graveyard/lessons.json`, sorted by
id. A lesson marked `confidence: "low"` is routed to
`graveyard/low-confidence/<id>.json` instead, kept apart from the confident set.
Re-ingesting **replaces** the prior interpretation: each run rewrites layer 3
from the current payload's survivors, so a lesson promoted low→high or dropped
from a later payload leaves nothing stale behind.

Report the result to the user: `written` (into `lessons.json`), `low_confidence`
(routed aside), and `dropped` (with the reason for each). The verb **exits 0 even
when every lesson was dropped** — an honest "nothing cited" is a valid outcome.
It exits non-zero only on a structural fault: the directory is not an abcd
lifeboat, its graveyard files are unreadable, or the lesson payload is
unreadable, oversize, or malformed. The lesson files are a later, mutable
interpretation and are **not** part of the lifeboat's `manifest_sha256`.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd disembark ...` from the repo root, or build it with `make build`.

**User input:** $ARGUMENTS
