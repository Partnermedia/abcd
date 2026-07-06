# `/abcd:memory` — Multi-Upstream Curated Knowledge Substrate

User-facing command for the per-project compounding-curated knowledge substrate at `.abcd/memory/`. Design target per itd-36 (idea-1 final shape after 5-round adversarial RP review); the write core (ingest/ask/bare) shipped via fn-38 (`scripts/abcd/memory.py`) and the lint family via fn-39 (`scripts/abcd/memory_coverage.py` + the `MQ`/`MS`/`ML` codes).

For the **substrate spec** (page-class enum, source-class taxonomy, curator behaviour, lifecycle class, integration with itd-26 loot), see [`05-internals/07-memory.md`](../05-internals/07-memory.md). This file is the surface contract: what the user types and what happens.

## Sub-verbs

Bare `/abcd:memory` shows status + help + render of current memory state — never mutates state. Per the [bare-command-as-render discipline](../02-constraints/04-naming.md). Current sub-verbs (each does something bare cannot):

- **Bare `/abcd:memory`** — render: page count by class (e.g., "23 session_memory + 8 external_pdf + 4 oracle_review + 2 spec_modification_grammar"), last-ingest timestamp, recent contradictions surface, suggested next actions. No mutation. Quotation-budget headroom per source renders READ-ONLY from the fn-39 `.coverage_index.json`: when the index is present AND fingerprint-fresh (a read-only crawl recomputes the current fingerprint and matches the stored one) it shows per-source warn/block headroom; fingerprint drift shows a "stale — run /abcd:memory lint" hint; an absent index an info line; a malformed index or crawl failure a non-fatal "headroom unavailable" line. The bare render never rebuilds or mutates the index.
- **`/abcd:memory ingest <path-or-url>`** — read external source (PDF / transcript / article / URL), distil into typed entity/topic pages with citation frontmatter, append to ingest log. **Default: do NOT store original.** Flag-shaped modifier: `--keep-original` (opt-in storage at `.abcd/memory/sources/<sha256>.<ext>`; the lifeboat licence gate — `/abcd:disembark`, NOT launch, per [adr-18](../../decisions/adrs/0018-launch-payload-excludes-memory-gate-scoped-to-lifeboat.md) — refuses publish without an explicit allowlist entry; launch excludes `.abcd/**` wholesale per [`04-launch.md § 2`](04-launch.md#2-payload-manifest-default-deny)).
- **`/abcd:memory ask <question>`** — query memory by domain + class; synthesise an answer with citations (every citation references `source.class` + `citation` + `source_hash`); optionally file the result back as a new memory page (interactive prompt).
- **`/abcd:memory lint` (fn-39 — shipped)** — full-store curator health-check: per-page quotation budgets (`MQ001`), cumulative source coverage (`MQ002`), coverage-unavailable diagnostic (`MQ003`, info), source-class single-class advisory (`MS001`), cross-class without weighting note (`MS002`), missing licence on `external_*` (`ML001`). ALWAYS crawls the full workspace store, rebuilds the regenerable `.coverage_index.json`, emits findings to `.abcd/logbook/memory/lint-<ts>/report.{json,md}`. Exit: blockers → nonzero; warn-only → 0 (curator advisory — see [`06-lint.md §2`](../05-internals/06-lint.md#2-severity-model)). Mutates no memory-store state (coverage index + logbook report are its only writes). Per ADR-13's write/lint split, the fn-38 write core ships ingest/ask/bare; fn-39 ships this lint family ONLY — contradictions are rendered by fn-38's reconciliation (surfaced by the bare render), orphan/stale-claim audits are deferred.

## 1. Default flow — distil, cite, discard

```
/abcd:memory ingest <path>
    │
    ▼
PROBE
  - Compute sha256 of source content
  - Look up in .abcd/memory/.sources_index.json (the provenance substrate per
    itd-36 — shipped via fn-38, `scripts/abcd/provenance.py`; distinct from the
    ahoy history store that keys session transcripts on the root-commit SHA)
  - If found: bump ingest_count, update last_ingest, return cached citation
  - If new: continue
    │
    ▼
LICENCE DETECT (per 05-internals/09-provenance-substrate.md § 1)
  - Parse source for SPDX-ID (LICENSE file, package manifest, file headers, HTTP header)
  - On ambiguous / missing: prompt user for explicit licence (or `unknown`)
  - Reject if licence is restrictive AND project is public (--accept-licence-risk override)
    │
    ▼
DISTIL (principle-distiller curator)
  - Read source content
  - Produce N entity/topic pages: <type>_<domain>_<slug>.md
  - Each page carries source: { class, citation, licence, source_hash, ingested_at, weighting_note? }
  - Cross-reference to existing memory pages (topic-hash dedup)
  - Apply per-page quotation budget as curation discipline (the MQ001 lint
    that enforces it computes at LINT time — fn-39's `/abcd:memory lint` —
    never at ingest)
    │
    ▼
WRITE
  - .abcd/memory/<type>_<domain>_<slug>.md (new pages or updates)
  - .abcd/memory/index.md (regenerated catalog)
  - .abcd/memory/log.md (append: ## [YYYY-MM-DD HH:MM] external_pdf | <slug> — <summary>)
  - .abcd/memory/contradictions.md (if curator surfaces conflict with existing pages)
  - .abcd/memory/.sources_index.json (registry update)
    │
    ▼
DISCARD ORIGINAL (default behaviour)
  - Source path + hash recorded for re-ingest only
  - Original NOT stored at .abcd/memory/sources/
  - Log entry includes "original discarded; use --keep-original to retain"
```

`--keep-original` opts the user into storing the original at `.abcd/memory/sources/<sha256>.<ext>`. The fn-38 restrictive-licence gate refuses to publish anything under `.abcd/memory/sources/` unless `.abcd/launch-allowlist.json` explicitly names the file. Per adr-18 this gate is the **lifeboat's** (`/abcd:disembark`), NOT launch's — launch excludes `.abcd/**` wholesale and never publishes `.abcd/memory/sources/`; the gate is future/inert at launch.

## 2. Acceptance Criteria (Given-When-Then, per itd-1)

See [the full acceptance criteria](../../roadmap/intents/drafts/itd-36-memory-unification.md#acceptance-criteria) in itd-36's intent spec. Surface-level summary:

- **Bare**: bare `/abcd:memory` renders current state; never mutates.
- **Ingest default-no-original**: original NOT stored unless `--keep-original`; citation + source_hash recorded; quotation budget applied per page (enforced at lint time by fn-39's `MQ001`, never at ingest).
- **Ingest with `--keep-original`**: original stored at `.abcd/memory/sources/<sha256>.<ext>`; the lifeboat licence gate (`/abcd:disembark`, not launch — adr-18) refuses publish without allowlist.
- **Ask**: synthesises answer with per-citation provenance (class + citation + source_hash); optionally files result back.
- **Lint (fn-39 — shipped; not fn-38 behaviour)**: emits `MQ001` / `MQ002` / `MQ003` / `MS001` / `MS002` / `ML001` codes; cumulative coverage uses span-level dedup. fn-38 writes `licence: unknown` explicitly; fn-39's `ML001` is what lints it.
- **Schema extension on existing**: existing flat-named pages preserved; `index.md` generated over them; `source.class: session_memory` backfilled as default.
- **Cross-consumer registry**: the provenance substrate's `.abcd/memory/.sources_index.json` (per itd-36 — shipped via fn-38, `scripts/abcd/provenance.py`; not to be confused with the ahoy history store) is shared with itd-26 loot (a later phase, not yet built); same hash → same registry entry.

## 3. Logbook layout

```
.abcd/logbook/memory/
├── ingest-<utc-ts>/
│   ├── ingest-report.{json,md}     # source path, sha256, distilled page count, citation, licence
│   └── distil-trace.json           # principle-distiller per-page output trace (debug)
├── ask-<utc-ts>/
│   └── ask-report.{json,md}        # question, retrieved page slugs, synthesised answer with citations
└── lint-<utc-ts>/
    └── report.{json,md}            # lint findings: MQ001/MQ002/MS001/MS002/ML001 with locations
```

Per universal pattern 6 (logbook-as-reports-only); coordination locks live at `.abcd/coordination/` not under logbook.

## 4. Composition with adjacent surfaces

- **`/abcd:disembark`** exports curated project memory/provenance into the lifeboat (existing behaviour; itd-36 doesn't change disembark's source-mapping). What the lifeboat carries today is the curated provenance surface named in [`02-disembark.md §5`](02-disembark.md) — `research/pitfalls.{json,md}`, `assets/_manifest.json`'s provenance/classification, and root `_provenance.json` — **not** a verbatim `.abcd/memory/` payload; declaring an exact `.abcd/memory/`-verbatim payload is deferred to the disembark spec that wires the lifeboat packer (adr-18). The recovery-humility framing on disembark/embark applies: the lifeboat is the floor of recoverable theory, not theory itself.
- **`/abcd:embark`** unpacks `.abcd/memory/` into the receiving repo. Source-class enum carries forward; receiver runs `/abcd:memory lint` (fn-39) post-unpack to verify quotation budgets and licences haven't drifted.
- **`/abcd:launch`** does **not** consume the provenance substrate's licence gate (adr-18): the public launch payload excludes `.abcd/**` — including `.abcd/memory/**` — wholesale as policy, so launch never publishes the files the gate checks. The restrictive-licence gate's real consumer is the **lifeboat** (`/abcd:disembark`, above), the surface that publishes curated project memory/provenance; it refuses publish on restrictive-licence files and warns on `licence: unknown`. At launch the gate is future/inert — `/abcd:launch dry-run` renders its verdicts only as a diagnostic preview (see [`04-launch.md § 2`](04-launch.md#2-payload-manifest-default-deny) and [adr-18](../../decisions/adrs/0018-launch-payload-excludes-memory-gate-scoped-to-lifeboat.md)).
- **`/abcd:dredge`** (a later phase, itd-25) writes synthesis output to `.abcd/memory/<type>_<domain>_<slug>.md` with `source.class: dredge_synthesis`. Distinct verb (storage vs operation per dredge-pushback in idea-1 R4); shared destination namespace.
- **`/flow-next`** specs inherit from itd-37 modification grammar — at spec completion, `principle-distiller` extracts the spec's `## Modification Grammar` section into `spec_modification_grammar_<spec_id>.md` (append-only) and updates curator-merged `modification_grammar_<domain>.md` (compounding-curated). User does not invoke `/abcd:memory` for this — the extraction is automatic on spec completion.

## 5. Cost shape

| Item | Cost |
|---|---|
| Schema extension | 3 new sibling files (`index.md`, `log.md`, `contradictions.md`) + typed `source:` frontmatter; no migration of existing flat-named files |
| Curator role on `principle-distiller` | Role extension (existing agent; itd-31 precedent — agent count stays 15) |
| Provenance/licence substrate | Separable spec at [`05-internals/09-provenance-substrate.md`](../05-internals/09-provenance-substrate.md); shared with the later-phase itd-26 loot |
| Lifeboat licence-gate extension (adr-18) | `.abcd/memory/sources/` allowlist + restrictive-licence detection over the lifeboat's gated payload (`/abcd:disembark`); NOT a launch payload gate — launch excludes `.abcd/**` wholesale, so the gate is future/inert at launch |
| Lint codes | New family: `MQ001` (per-page quotation), `MQ002` (cumulative coverage), `MS001` (source-class single-class), `MS002` (mixed-class without weighting note), `ML001` (licence missing) |

**Tight coupling** (logged as a principal risk): itd-36 is structurally non-decomposable — sub-verbs need the schema; schema needs the curator role; curator role needs the lint codes. Partial ship not meaningful.

## References

- [`05-internals/07-memory.md`](../05-internals/07-memory.md) — substrate spec (page-class enum, source classes, curator behaviour)
- [`05-internals/09-provenance-substrate.md`](../05-internals/09-provenance-substrate.md) — provenance/licence subsystem (shared with the later-phase loot verb)
- [`../../roadmap/intents/drafts/itd-36-memory-unification.md`](../../roadmap/intents/drafts/itd-36-memory-unification.md) — full intent spec with acceptance criteria + adversarial worked-example ship gate
- [`research/related-work.md § Karpathy LLM Wiki`](../../research/related-work.md#karpathy-llm-wiki--pattern-source-for-abcdmemory) — pattern source
