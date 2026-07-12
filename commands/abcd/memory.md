---
name: memory
description: Query and curate the per-project memory substrate at .abcd/memory/ by invoking the abcd binary. Bare invocation is a read-only status render; ingest/ask/lint curate, synthesise, and health-check the store.
argument-hint: "[<empty>] | ingest <path-or-url> [--keep-original] | ask <question> | lint"
---

# `/abcd:memory` — curated knowledge substrate

The per-project compounding-curated knowledge substrate at `.abcd/memory/`.
Bare invocation **performs zero writes**.

## Status (bare)

```bash
abcd memory --json
```

Summarise the JSON: `pages` and `by_class` (page count per source class),
`last_ingest`, any `contradictions`, and per-source `headroom` lines. The bare
render never rebuilds or mutates the coverage index.

## Ingest a source

Distil an external source (PDF / transcript / article / URL) into typed,
cited memory pages. **You** are the distiller: read the source, produce the
`DistilledPage` JSON array, and pass it to the binary via `--pages-json`
(a file, or `-` for stdin). The binary computes the provenance, licence, and
content hash, validates every page, and writes atomically.

```bash
abcd memory ingest <path-or-url> --pages-json distilled.json --json
```

Add `--keep-original` to retain the source at
`.abcd/memory/sources/<sha256>.<ext>` (the lifeboat licence gate — not launch —
governs its export). Report `status`, `licence`, and the written `pages`. An
already-known source re-ingests from the registry with no `--pages-json`.

## Ask memory

Deterministic retrieval over the store, then a cited answer:

```bash
abcd memory ask "<question>" --json
```

The default answer is the deterministic citation-renderer over the top-ranked
pages; every citation references `source.class`, `citation`, and `source_hash`.
Optionally file the answer back as a new page with `--file-back --page-json
<file|->` (one `DistilledPage` object you produce from the retrieved matches).
Report the `answer` and, if present, the `file_back` result.

## Lint

Full-store curator health-check — per-page quotation budgets, cumulative source
coverage, source-class and licence advisories:

```bash
abcd memory lint --json
```

It rebuilds the regenerable `.coverage_index.json` and writes a report under
`.abcd/.work.local/logs/memory/lint-<ts>/`. Summarise `summary.blockers` /
`summary.warnings` / `summary.infos` and each finding's `code` and `message`.
Blockers exit nonzero; warn-only exits 0.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd memory …`
from the repo root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
