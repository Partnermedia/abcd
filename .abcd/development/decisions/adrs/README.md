# abcd ADRs

Architecture Decision Records — retrospective records of settled decisions, their context, alternatives rejected, and consequences.

---

## What's an ADR?

An **ADR** captures a *settled decision* — written after the decision is made, recording why it was made and what was rejected. ADRs exist to keep decisions intelligible to future readers (and future selves) who weren't in the room when the decision happened.

ADRs are used when **all three** are true:

1. **Hard to reverse** — the cost of changing the decision later is meaningful.
2. **Surprising without context** — a future reader will wonder "why this way?"
3. **Result of a real trade-off** — there were genuine alternatives and one was picked for specific reasons.

If any of the three is missing, skip the ADR. File-scoped rationale (why this section reads this way) lives inline in the brief; project-scoped framework decisions earn an ADR.

ADRs are *not* used for:

- Forward-looking discussion — those are RFCs (`../roadmap/rfcs/`).
- User-facing capability — those are intents (`../../intents/`).
- Bug fixes, refactors, content edits — git log + commit messages cover deltas.

---

## ADR IDs

ADR IDs follow the pattern `adr-N` (unpadded, mirrors `itd-N` / `rfc-N`) as the prose handle. Filenames are the zero-padded four-digit sequential form `NNNN-<slug>.md` (`adr-7` lives in `0007-<slug>.md`), a stable cross-reference handle per [ADR-30](0030-record-information-architecture.md).

IDs are capture-stable. Once assigned, an ADR's ID never changes — superseding ADRs use new IDs and link backwards.

---

## Lifecycle (Status Field)

| Status | Meaning |
|---|---|
| `proposed` | Draft; the decision is not yet locked. Rare — most ADRs are written after the fact. |
| `accepted` | The decision is in force. Default for retrospective ADRs. |
| `superseded` | Replaced by a newer ADR. Once the successor is in place, the superseded ADR is pruned from the record — the successor's `supersedes` note records what it replaced. |
| `deprecated` | The decision no longer applies but no successor replaces it (the surface itself was removed). |

Transitions are deliberate. Accepted ADRs are retained in the record; a superseded ADR is pruned once its successor lands and carries the transition rationale — git history preserves the original text.

---

## Format

Every ADR has frontmatter (machine-readable) plus a Markdown body following this structure:

```markdown
---
id: adr-N
slug: <kebab-case-slug>
status: accepted                 # proposed | accepted | superseded | deprecated
date: YYYY-MM-DD
supersedes: null                 # adr-N if this replaces an earlier ADR
superseded_by: null              # adr-N if a later ADR replaced this one
related_intents: []              # [itd-N, ...] cross-references
related_rfcs: []                 # [rfc-N, ...] cross-references
related_adrs: []                 # [adr-N, ...] sibling decisions
---

# ADR-N: <Title — short noun phrase, the decision in one line>

## Context

What forced the decision? What was the world look like before? What constraints
were already locked?

## Decision

What did we decide? Stated as a positive declaration: "We will X."

## Alternatives Considered

2–4 options laid out fairly, including the chosen one. For each: what it would
have looked like, why it was rejected (or chosen).

## Consequences

What follows from the decision — both gains and costs. Honest about trade-offs.
What's now easier; what's now harder; what new obligations the decision creates
(lint rules, vocabulary terms, audit gates).
```

---

## Bidirectional Linking

| File | Frontmatter field |
|---|---|
| `adrs/NNNN-<slug>.md` | `related_intents: [itd-N, ...]` (intents whose framework this ADR justifies) |
| `adrs/NNNN-<slug>.md` | `related_rfcs: [rfc-N, ...]` (RFCs that informed this decision) |
| `adrs/NNNN-<slug>.md` | `supersedes: adr-N` / `superseded_by: adr-N` (chain) |
| `intents/{drafts,planned,shipped,disciplines}/itd-N-<slug>.md` | `related_adrs: [adr-N, ...]` (when an intent references an ADR) |
| `rfcs/rfc-N-<slug>.md` | `related_adrs: [adr-N, ...]` (when an RFC references an ADR or its resolution becomes one) |

The intent lint (a Go implementation) extends to verify these reciprocally.

---

## Index

> **Index maintenance:** allocating the next `adr-N` and materialising the ADR
> file is race-safe, but appending the row to this index table is a manual edit;
> add the row by hand when an ADR is captured.

| ID | Title | Status | Date |
|---|---|---|---|
| [adr-1](0001-three-layer-mental-model.md) | Three-layer mental model (brief / intent / spec) | accepted | 2026-05-04 |
| [adr-2](0002-three-intent-kinds.md) | Three intent kinds (standalone / bundle-member / discipline) | accepted | 2026-05-07 |
| [adr-3](0003-directory-as-truth-for-lifecycle.md) | Directory location is the source of truth for lifecycle state | accepted | 2026-05-07 |
| [adr-5](0005-brief-is-current-state.md) | Brief is the current state; no version label, no archive directory | accepted | 2026-05-08 |
| [adr-7](0007-grill-skill-and-glossary.md) | `/abcd:intent grill` — one sub-verb with two inseparable phases; cite-or-fail lint; bounded-context glossary structure | accepted | 2026-05-11 |
| [adr-9](0009-phase-as-product-layer.md) | Phase as a product-reflection layer between brief and intent; replaces plugin-version language | accepted | 2026-05-16 |
| [adr-10](0010-phase-negotiator-grounded-tradeoffs.md) | The phase negotiator — a Socratic agent that proposes phases and grounds every trade-off in the DAG / phase acceptance | accepted | 2026-05-16 |
| [adr-11](0011-spec-terminology-rename.md) | One canonical word for a specced block of work — spec | accepted | 2026-05-18 |
| [adr-12](0012-issue-ledger-live-vs-structured.md) | `.work/issues.md` (historical) stays the live operational ledger; structured `iss-*` store deferred until the native spec layer schedules the migration | accepted | 2026-06-06 |
| [adr-13](0013-fn38-memory-single-writer-and-write-lint-split.md) | Durable memory writes — single-writer, atomic-rename crash model | accepted | 2026-06-09 |
| [adr-19](0019-plugin-json-version-carve-out.md) | The plugin version lives only in the released artifact; the working tree stays unversioned, and the version location is chosen by a schema-validated decision artifact, not hard-coded | accepted | 2026-07-01 |
| [adr-20](0020-manifest-version-lockstep.md) | The two release manifests stay version-consistent via a read-only anti-drift checker over a pinned per-view path list; the source view stays unversioned; `--allow-dirty` must never bypass manifest consistency (wiring policy); the marketplace changelog entry gets a committed schema | accepted | 2026-07-03 |
| [adr-21](0021-rebuild-in-go.md) | Rebuild abcd as a Go binary | accepted | 2026-07-06 |
| [adr-22](0022-bundled-deps-as-pluggable-adapters.md) | Bundled dependencies become pluggable adapters over a native default (supersedes adr-14, adr-15, adr-17) | accepted | 2026-07-06 |
| [adr-23](0023-transport-agnostic-core.md) | A transport-agnostic Go core behind thin front doors | accepted | 2026-07-06 |
| [adr-24](0024-companion-harness-peer-via-conventions-and-mcp.md) | the companion harness is a peer integrated via conventions and MCP, not a code dependency | accepted | 2026-07-06 |
| [adr-25](0025-host-delegated-llm-default.md) | The LLM is host-delegated by default; oracles are opt-in adapters (supersedes adr-8) | accepted | 2026-07-06 |
| [adr-26](0026-native-spec-layer-ccpm-backend.md) | A native minimal spec layer with the companion harness `ccpm` as the primary deeper backend | accepted | 2026-07-06 |
| [adr-27](0027-autonomous-run-pluggable-seam.md) | The autonomous run is a pluggable seam, not a Ralph port (supersedes adr-16) | accepted | 2026-07-06 |
| [adr-28](0028-single-repo-curated-release.md) | One repository, a curated release artifact — no dev→public mirror (supersedes adr-18) | accepted | 2026-07-06 |
| [adr-29](0029-native-transcript-corpus.md) | A native local redacted transcript corpus | accepted | 2026-07-06 |
| [adr-30](0030-record-information-architecture.md) | Design-record information architecture — flat artefact-type folders | accepted | 2026-07-06 |
| [adr-31](0031-derived-versioning-from-intents.md) | The release version is derived from the intents in it, never authored (extends adr-19, adr-20) | accepted | 2026-07-07 |
| [adr-32](0032-issue-ledger-is-working-tier-data.md) | The issue ledger is working-tier data, not authored record — move to `.abcd/work/issues/`, drop git-inferable timestamps, derive priority | accepted | 2026-07-08 |
| [adr-33](0033-launch-phase-ownership-tiered.md) | Launch phase ownership is tiered — Phase 1 owns the curated-release cut; deepenings are separately scheduled intents; the phase index is the sole ownership source | accepted | 2026-07-08 |
| [adr-34](0034-lifecycle-and-scheduling-orthogonal.md) | Intent lifecycle and phase scheduling are orthogonal axes — scheduled ⇒ committed (`planned/`), but planned intents may be unscheduled | accepted | 2026-07-08 |
| [adr-35](0035-lifeboat-as-coverage-experiment.md) | The lifeboat is a coverage experiment — read-only, out-of-tree, and proven before it is packed (supersedes adr-4) | accepted | 2026-07-14 |
| [adr-36](0036-coverage-blanks-are-a-fillable-lifecycle.md) | Coverage blanks are a fillable lifecycle — authored is not extracted, and the interview is its own step | accepted | 2026-07-15 |
| [adr-37](0037-changelog-driven-releases.md) | Releases are changelog-driven — rolling `[Unreleased]` is the release decision, and automation tags exactly that commit | accepted | 2026-07-17 |
