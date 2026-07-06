---
id: adr-4
slug: lifeboat-as-regenerable-output
status: accepted
date: 2026-05-04
supersedes: null
superseded_by: null
related_intents: []
related_rfcs: []
related_adrs: [adr-1, adr-5]
---

# ADR-4: Lifeboat is regenerable output; voyage is the operations namespace

## Context

`/abcd:disembark` packs a portable artefact — the *lifeboat* — that captures a project's theory at the highest fidelity we can leave behind. Early framing treated the lifeboat as a long-lived archive: each disembark would land alongside previous ones (`lifeboat-v1/`, `lifeboat-v2/`, …) so historical snapshots would accumulate in-place.

This led to four problems:

1. **Disk pressure** — lifeboat snapshots include verbatim ADR / spec / docs copies and full audit reports; accumulating them per-run quickly bloated the repo.
2. **Stale-snapshot drift** — older lifeboats no longer matched current source; readers had to reason about which snapshot was authoritative.
3. **Naming collision** — operations metadata (which adapter ran, which oracle backend was used, which manifest hashes) needed somewhere to live, and `lifeboat/` was both the artefact directory and (effectively) the operations log.
4. **Embark ambiguity** — when a target repo had `lifeboat-v1/` AND `lifeboat-v2/`, embark had no clean rule for which to trust.

## Decision

**The lifeboat is regenerable output.** `.abcd/lifeboat/` holds *only the latest* disembark snapshot and is overwritten in place on each run (with `.bak` safety net). It is gitignored unless `visibility=private`.

**Operations metadata lives in a separate namespace: `.abcd/development/voyage/`.**

- `voyage/disembark/history.jsonl` — append-only log of every disembark run: `manifest_sha256`, file list, oracle backend used, verdict.
- `voyage/embark/provenance.json` — last embark's source path + manifest hash.
- `voyage/embark/from/<timestamp>/` — opt-in (`embark --archive`) verbatim copy of the input lifeboat (for the case where the source repo will disappear).

Hash chain: each disembark's `_provenance.json` (inside the lifeboat) matches the `manifest_sha256` recorded in `voyage/disembark/history.jsonl`. History audit-traceable; snapshot itself is regenerable.

**Naming distinction is load-bearing:**
- `lifeboat/` = the artefact (noun-side; what gets carried).
- `voyage/` = operations (verb-side; what we did to produce it).

## Alternatives Considered

1. **Keep accumulating snapshots (`lifeboat-v1/`, `lifeboat-v2/`).** Rejected: disk pressure, drift, and embark ambiguity all compound over time. Every problem above lives in this option.
2. **One snapshot, no history at all.** Rejected: audit trail is load-bearing for the recovery-humility frame (per Naur). Knowing *when* a disembark happened, what it captured, and which oracle verified it is the minimum forensic record. Without it, recovery from a partial or contested lifeboat has no provenance.
3. **History inside `lifeboat/` itself** (single namespace). Rejected: collides on overwrite. If the lifeboat is overwritten in place, the history would be too; if the history is preserved through overwrite, the lifeboat is no longer "just the latest snapshot." Two distinct lifecycles (regenerable vs append-only) want two namespaces.
4. **Manifest-only history (no opt-in archive).** Rejected: there's a real use case where the source repo disappears (project shut down, machine lost, contractor hands over). For that case, `embark --archive` saving a verbatim copy at `voyage/embark/from/<timestamp>/` is the only recovery path. Made opt-in so the default stays lean.

## Consequences

**Gains:**
- `.abcd/lifeboat/` is always the current state; no version-resolution logic.
- Full provenance (hash chain) without storing stale snapshots.
- Operations namespace (`voyage/`) is reusable for future verbs (e.g., a v2 `redisembark` or `relaunch` would log to `voyage/<verb>/`).
- Embark contract is unambiguous — read the lifeboat at `<path>`, full stop.

**Costs / obligations:**
- Hash-chain integrity must be maintained — `intent_lint.py` (or a sibling audit) verifies `_provenance.json.manifest_sha256` matches the latest line in `voyage/disembark/history.jsonl` for the same run.
- Reserved vocabulary: `voyage/`, `lifeboat/`, `manifest_sha256`, `_provenance.json`, `history.jsonl` — registered in `02-constraints/04-naming.md`.
- The `embark --archive` flag is opt-in; users who don't know about it will lose the source-disappearance recovery path. Documented prominently in `04-surfaces/03-embark.md`.

**Downstream decisions enabled:**
- ADR-5 (brief is current state) — the same lifecycle principle (regenerable + append-only-history-elsewhere) applies one level up: the brief is the project's current state, history lives in `git log`, no `archive/<NN>/` directories.
