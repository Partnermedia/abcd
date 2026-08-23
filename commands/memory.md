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
"${CLAUDE_PLUGIN_ROOT}/abcd" memory --json
```

Summarise the JSON: `pages` and `by_class` (page count per source class),
`last_ingest`, any `contradictions`, and per-source `headroom` lines. The bare
render never rebuilds or mutates the coverage index.

## Ingest a source

Distil an external source (transcript / article / URL) into typed, cited
memory pages. PDF is a later-phase seam: the binary rejects a PDF source with a
clear error, because no text-extraction dependency is wired. **You** are the distiller: read the source, produce the
`DistilledPage` JSON array, and pass it to the binary via `--pages-json`
(a file, or `-` for stdin). The binary computes the provenance, licence, and
content hash, validates every page, and writes atomically.

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" memory ingest <path-or-url> --pages-json distilled.json --json
```

Add `--keep-original` to retain the source at
`.abcd/memory/sources/<sha256>.<ext>` (the lifeboat licence gate — not launch —
governs its export). Report `status`, `licence`, and the written `pages`. An
already-known source re-ingests from the registry with no `--pages-json`.

## Ask memory

Deterministic retrieval over the store, then a cited answer:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" memory ask "<question>" --json
```

The default answer is the deterministic citation-renderer over the top-ranked
pages; every citation references `source_class`, `citation`, and `source_hash`.
Optionally file the answer back as a new page with `--file-back --page-json
<file|->` (one `DistilledPage` object you produce from the retrieved matches).
Report the `answer` and, if present, the `file_back` result.

## Lint

Full-store curator health-check — per-page quotation budgets, cumulative source
coverage, source-class and licence advisories:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" memory lint --json
```

It rebuilds the regenerable `.coverage_index.json` and writes a report under
`.abcd/.work.local/logs/memory/lint-<ts>/`. Summarise `summary.blockers` /
`summary.warnings` / `summary.infos` and each finding's `code` and `message`.
Blockers exit nonzero; warn-only exits 0.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
