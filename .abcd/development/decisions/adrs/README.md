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
- User-facing capability — those are intents (`../roadmap/intents/`).
- Bug fixes, refactors, content edits — git log + commit messages cover deltas.

---

## ADR IDs

ADR IDs follow the pattern `adr-N` (unpadded, mirrors `itd-N` / `rfc-N` / `fn-N`). Filenames: `adr-N-<slug>.md`.

IDs are capture-stable. Once assigned, an ADR's ID never changes — superseding ADRs use new IDs and link backwards.

---

## Lifecycle (Status Field)

| Status | Meaning |
|---|---|
| `proposed` | Draft; the decision is not yet locked. Rare — most ADRs are written after the fact. |
| `accepted` | The decision is in force. Default for retrospective ADRs. |
| `superseded` | Replaced by a newer ADR. Frontmatter `superseded_by: <adr-N>` points to the successor. |
| `deprecated` | The decision no longer applies but no successor replaces it (the surface itself was removed). |

Transitions are deliberate. An ADR is never deleted — superseded or deprecated, but always retained as historical record.

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
| `adrs/adr-N-<slug>.md` | `related_intents: [itd-N, ...]` (intents whose framework this ADR justifies) |
| `adrs/adr-N-<slug>.md` | `related_rfcs: [rfc-N, ...]` (RFCs that informed this decision) |
| `adrs/adr-N-<slug>.md` | `supersedes: adr-N` / `superseded_by: adr-N` (chain) |
| `intents/{drafts,planned,shipped,disciplines}/itd-N-<slug>.md` | `related_adrs: [adr-N, ...]` (when an intent references an ADR) |
| `rfcs/rfc-N-<slug>.md` | `related_adrs: [adr-N, ...]` (when an RFC references an ADR or its resolution becomes one) |

`intent_lint.py` extends to verify these reciprocally.

---

## Index

> **Index maintenance (fn-56/itd-44):** the `reserve-and-write-adr` writer
> allocates `adr-N` and materialises the ADR file race-safely, but it does
> **not** touch this index table — appending the new row is a manual edit (or
> a future tooling pass). Auto-updating the index under the allocation lock
> (atomic rewrite + rollback) is **out of scope** for the thin adoption; the
> follow-up is recorded in `.work/issues.md` (2026-06-14). Until then, add the
> row by hand when an ADR is captured.

| ID | Title | Status | Date |
|---|---|---|---|
| [adr-1](./adr-1-three-layer-mental-model.md) | Three-layer mental model (brief / intent / spec) | accepted | 2026-05-04 |
| [adr-2](./adr-2-three-intent-kinds.md) | Three intent kinds (standalone / bundle-member / discipline) | accepted | 2026-05-07 |
| [adr-3](./adr-3-directory-as-truth-for-lifecycle.md) | Directory location is the source of truth for lifecycle state | accepted | 2026-05-07 |
| [adr-4](./adr-4-lifeboat-as-regenerable-output.md) | Lifeboat is regenerable output; voyage is the operations namespace | accepted | 2026-05-04 |
| [adr-5](./adr-5-brief-is-current-state.md) | Brief is the current state; no version label, no archive directory | accepted | 2026-05-08 |
| [adr-6](./adr-6-rp-review-storage-and-architecture.md) | RP review storage (hybrid commit/ignore) and abcd-side wrapper architecture | accepted | 2026-05-10 |
| [adr-7](./adr-7-grill-skill-and-glossary.md) | `/abcd:intent grill` — one sub-verb with two inseparable phases; cite-or-fail lint; bounded-context glossary structure | accepted | 2026-05-11 |
| [adr-8](./adr-8-dual-backend-review-asymmetric-trust.md) | Dual-backend review (RP + Codex CLI) with asymmetric trust — scoped reviewer's verdict gates; mandatory stopping rule | accepted | 2026-05-16 |
| [adr-9](./adr-9-phase-as-product-layer.md) | Phase as a product-reflection layer between brief and intent; replaces plugin-version language | accepted | 2026-05-16 |
| [adr-10](./adr-10-phase-negotiator-grounded-tradeoffs.md) | The phase negotiator — a Socratic agent that proposes phases and grounds every trade-off in the DAG / phase acceptance | accepted | 2026-05-16 |
| [adr-11](./adr-11-spec-terminology-rename.md) | One canonical word for a specced block of work — rename the *how* layer to "spec" | accepted | 2026-05-18 |
| [adr-12](./adr-12-issue-ledger-live-vs-structured.md) | `.work/issues.md` stays the live operational ledger; structured `iss-*` store deferred until fn-22 is re-planned | accepted | 2026-06-06 |
| [adr-13](./adr-13-fn38-memory-single-writer-and-write-lint-split.md) | fn-38 memory is single-writer (atomic-rename, no txn/recovery); split into fn-38 (write) + fn-39 (lint); re-plan via flow-next with triple-backend review | accepted | 2026-06-09 |
| [adr-14](./adr-14-fn40-guard-fail-closed-full-required-manifest.md) | Guard degraded fallback fails closed to the full required manifest (integrity not coverage); floor beneath, never empty | accepted | 2026-06-10 |
| [adr-15](./adr-15-abstraction-boundary-warn-not-block.md) | Abstraction boundary warns, never blocks — argv-sentinel live discriminator (fn-37.3), artifact-only static detection, PreToolUse hook deferred | accepted | 2026-06-11 |
| [adr-16](./adr-16-fn43-autodrain-boundary-and-gate-defaults.md) | Autodrain fires at the Ralph post-iteration edge only (no Claude Code hook); the gate reports, never blocks; drain cost-bounded by processed entries | accepted | 2026-06-11 |
| [adr-17](./adr-17-rp-chat-send-override-supersedes-acj1-env-skip.md) | `rp chat-send` becomes a declared abcd override (scoped supersession of fn-33 AC-J1's env-skip) — fixed budget pre-flight on every path, driven-path reversal, durable `--selected-paths`, vestigial SKIP export as rollback path | accepted | 2026-06-11 |
| [adr-18](./adr-18-launch-payload-excludes-memory-gate-scoped-to-lifeboat.md) | The public launch payload excludes `.abcd/memory/**` as policy; the restrictive-licence gate is scoped to the lifeboat, future/inert at launch (no override may re-include memory) | accepted | 2026-06-13 |
| [adr-19](./adr-19-plugin-json-version-carve-out.md) | The plugin version lives only in the published snapshot; dev files stay unversioned, and the version location is chosen by a schema-validated decision artifact, not hard-coded | accepted | 2026-07-01 |
| [adr-20](./adr-20-manifest-version-lockstep.md) | The two published manifests stay version-consistent via a read-only anti-drift checker over a pinned per-tree path list; dev stays unversioned; `--allow-dirty` must never bypass manifest consistency (wiring policy); the marketplace changelog entry gets a committed schema | accepted | 2026-07-03 |
