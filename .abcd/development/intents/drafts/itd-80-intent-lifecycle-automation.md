---
id: itd-80
slug: intent-lifecycle-automation
kind: null
suggested_kind: standalone
bundle: null
spec_id: null
reclassification_history: []
builds_on: [itd-34, itd-1]
related_adrs: [adr-25, adr-26, adr-27]
severity: major
---

# An Intent Ships Itself: `planned → shipped` Follows Its Spec Closing, With A Fidelity Audit Attached

## Press Release

> **abcd ships intent-lifecycle automation: an intent moves from `planned/` to `shipped/` as a side-effect of its linked spec closing, and arrives carrying a fidelity audit.** The developer runs `abcd intent plan itd-N` to link an intent to a freshly-minted native spec (`spc-M`) and commit it to `planned/`. When the work is done and they run `abcd spec close spc-M`, a deterministic reconcile detects the change through the intent↔spec link and moves the intent to `shipped/` — no `move` verb, no hand-editing, because the directory *is* the lifecycle state. The move emits a host-delegated fidelity-review request over the intent's Acceptance Criteria and the delivered code; the host's verdict is ingested back into the intent's `## Audit Notes` as per-criterion verdicts (MET / MET_WITH_CONCERNS / NOT_MET / INCONCLUSIVE) plus a three-bucket prose audit (honoured / diverged / missing), every claim carrying a cited `file:line` evidence pointer.
>
> "I stopped babysitting the roadmap," said Maya, an autonomous-development practitioner. "I plan an intent, I close its spec, and the intent ships itself — and the audit that lands with it tells me, criterion by criterion with evidence, whether what I built actually matches what I promised. The lifecycle stopped being a thing I maintained by hand and became a thing that maintains itself."

## Why This Matters

Until now, an intent's journey through `drafts → planned → shipped` was a set of manual file moves and hand-written audits — the very kind of undisciplined, drift-prone bookkeeping abcd exists to eliminate. itd-3 was driven through the lifecycle entirely by hand (its `## Audit Notes` were authored manually) precisely because no machinery existed. That manual precedent is the reference the automation must reproduce.

This intent builds the automation. Directory location stays the single source of truth (adr-26 directory-as-truth): there is no `status:` field to drift and no `abcd intent move` to invite manual state-setting. The lifecycle advances only as a *consequence* of a real event — a spec closing — detected through an ID-anchored link that survives file moves. Because abcd cannot run an LLM itself (adr-25 host-delegated default), the fidelity review is asynchronous: abcd emits a request and parks a receipt; the host runs the reviewer; the verdict is ingested through a validated transport. The result is a lifecycle that advances itself and a shipped record that carries its own evidence.

## What's In Scope

- **A minimal native spec store** (`internal/core/spec`): directory-as-truth spec records, `spc-N` id minting that does not collide with reserved ids, an ID-anchored bidirectional link to an intent (`intent: itd-N`), and status-by-directory (open vs closed).
- **`abcd intent` verb family**: bare `abcd intent` renders lifecycle status (never mutates); `abcd intent plan <itd-N>` mints a spec, writes the bidirectional link, sets `kind` (default `standalone`), and moves the intent `drafts → planned`; `abcd intent link <itd-N> <spc-N>` retroactively links a pre-existing spec; `abcd intent review ingest` ingests a verdict.
- **`abcd spec` verb family**: bare render; `abcd spec close <spc-N>` marks a spec done and runs the reconcile.
- **A deterministic reconcile**: on spec close, detect the linked intent via the immutable link, move it `planned → shipped` with an atomic same-filesystem rename, and repair any drifted derived `spec_id`. Fail closed on a missing or ambiguous link — never a silent or partial move. Keeps the `intent_lifecycle` record-lint contract green (the `shipped/` `spec_id ^spc-` requirement satisfied, not relaxed).
- **A host-delegated `intent-fidelity-reviewer` markdown agent** (new `agents/` directory): a single-document (Role 1) fidelity reviewer that reads an intent's Acceptance Criteria + the delivered diff and returns a VSA-shaped verdict JSON with per-criterion verdicts and a honoured/diverged/missing audit, each entry citing evidence.
- **An async verdict-ingest transport**: emit a review-request + park an `OWED` receipt at the ship-move; ingest schema-gates the verdict, semantically checks it (every `criterion_id` exists in the intent, every verdict in-enum), and appends idempotently (keyed on the receipt id) to `## Audit Notes`. A malformed/absent verdict after retry → receipt `DEAD_LETTER`, affected criteria recorded `INCONCLUSIVE`, raw payload retained.
- **Dogfooding**: this intent (itd-80) is the pipeline's first real payload — it is driven `drafts → planned → shipped` through the machinery it specifies, and its `## Audit Notes` are produced by the automation, not by hand.

## What's Out of Scope

- **The inter-spec dependency graph** and richer spec-store features (adr-26's full vision) — the store here is the minimal floor: link, status-by-directory, reconcile.
- **Bundle-member and discipline lifecycles** (itd-34) — slice 1 handles `standalone` only.
- **Reviewer Roles 2 and 3** (cross-document consistency, shape-classification; itd-48) and **loop-to-acceptance** (itd-50) — the reviewer is Role 1, one-shot, report-only.
- **Multi-judge juries / calibration harness** — a single host verdict; a calibration corpus is added only once real verdicts exist.
- **Auto-generation or auto-planning of the spec body** — `plan` mints the spec record and link; authoring the spec's content is not automated here.
- **A vendor/event "hook"** — the reconcile is a deterministic Go step inside `abcd spec close`, never a harness event.

## Acceptance Criteria

> _BDD format, per the [itd-1 discipline](../disciplines/itd-1-acceptance-gates.md). These gates are checked by `intent-fidelity-reviewer`'s single-document role when this intent moves to `shipped/` — this intent dogfoods that path._

- **Given** a `standalone` intent in `drafts/` with a valid `## Acceptance Criteria` section and `spec_id: null`, **when** the developer runs `abcd intent plan itd-N`, **then** a native spec `spc-M` is minted (with `intent: itd-N`), the intent gains `spec_id: spc-M` and a binding `kind: standalone`, the file moves `drafts → planned`, and `make record-lint` stays green.
- **Given** a planned intent linked to an open spec `spc-M`, **when** the developer runs `abcd spec close spc-M`, **then** a deterministic reconcile detects the linked intent through the immutable link and moves it `planned → shipped` with an atomic rename, satisfying the `intent_lifecycle` `shipped/` contract, and no unrelated file is modified.
- **Given** the ship-move, **when** reconcile runs, **then** exactly one VSA-shaped fidelity-review request (over the intent's Acceptance Criteria + the delivered diff) is emitted and exactly one `OWED` receipt carrying an idempotency key is parked.
- **Given** a host-produced verdict JSON that matches a parked receipt, **when** the developer runs `abcd intent review ingest`, **then** the payload is schema-gated and semantically checked (every `criterion_id` exists in the intent; every verdict is one of MET / MET_WITH_CONCERNS / NOT_MET / INCONCLUSIVE), and on success the per-criterion verdicts plus a honoured/diverged/missing audit — each entry carrying a cited `file:line` evidence pointer — are appended to the intent's `## Audit Notes`, and the receipt flips `OWED → INGESTED`.
- **Given** a verdict already ingested for a receipt, **when** ingest runs again with the same `receipt_id`, **then** it is a no-op: `## Audit Notes` is not duplicated (idempotent append keyed on the receipt id).
- **Given** a malformed/missing intent↔spec link at reconcile time, or a verdict referencing an unknown receipt at ingest time, **when** the operation runs, **then** it fails closed — no partial move, no partial append — with a clear error; and a verdict still invalid after retry marks the receipt `DEAD_LETTER`, records the affected criteria `INCONCLUSIVE`, and retains the raw payload.
- **Given** a crafted `itd-`/`spc-` id containing path-traversal or unexpected characters, **when** any verb resolves it to a file, **then** the id is validated against `^itd-[0-9]+$` / `^spc-[0-9]+$` and rejected otherwise — no file outside the intent/spec directories is read, written, or moved.

## Prior Art

- **[itd-34](../planned/itd-34-three-intent-kinds.md)** (three intent kinds) — defines the `kind` field and the standalone/bundle/discipline lifecycle paths; slice 1 implements the `standalone` path only. This intent is the lifecycle *automation* itd-34's ACs assume exists.
- **[itd-46](../planned/itd-46-abcd-intent-quoted-text-create-symmetric.md)** — the `abcd intent` create ergonomics (markdown surface); complementary, not overlapping (that is the create path; this is plan→ship→audit).
- **[itd-48](../planned/itd-48-intent-fidelity-reviewer-roles-2-3.md)** — the reviewer's Roles 2/3; this intent delivers Role 1 that they extend.
- **itd-3** (shipped, manual precedent) — its hand-authored `## Audit Notes` are the golden reference the automated audit must reproduce in shape.
- **adr-26** (native spec store — directory-as-truth), **adr-25** (host-delegated LLM default), **adr-27** (autonomous-run receipt gating) — the load-bearing decisions this intent instantiates.
- **`.abcd/development/plans/2026-07-11-intent-lifecycle.md`** — the SOTA-researched design plan this intent builds to.

## SOTA

> _Per the [sota-per-intent principle](../../principles/sota-per-intent.md):
> existing alternatives + rough maturity, then the chosen path. Harvested from the
> SOTA-researched design plan
> [`2026-07-11-intent-lifecycle.md`](../../plans/2026-07-11-intent-lifecycle.md)._

- **Spec store / lifecycle state.** Alternatives: git-native trackers (git-bug —
  CRDT, *mature*; Fossil — *mature*) avoid directory-state for multi-writer merge,
  a pressure a single-maintainer config tool does not have; spec-kit branch-per-spec
  (*usable*) fragments the durable `shipped/` record we want; the companion harness
  spec/task backend (*usable*) is the intended richer engine. → **Path 2**: a
  minimal directory-as-truth native floor (`os.Rename`, stdlib) with the adr-26
  seam to adopt an external spec engine later. No dependency.
- **Fidelity review (LLM-as-judge).** Alternatives: eval/judge harnesses (*mature*)
  plus the 2025–26 judge-calibration literature. The literature's conclusions
  (per-criterion + 4-value ordinal + explicit INCONCLUSIVE + report-not-block,
  cited evidence, pinned judge/prompt hashes) are adopted as *design*; the harnesses
  themselves force a judge and add heavy deps. → **Path 2**: host-delegated native
  emit/ingest (adr-25), the better external judge being whatever the host runs.
  No dependency.
- **Async ingestion.** The transactional outbox/inbox pattern (*de-facto standard*)
  is adopted as a stdlib pattern (`encoding/json` + hand-rolled validation); the
  verdict envelope reuses the repo's existing VSA attestation shape. Pattern reuse,
  not a dependency.

**Verdict — Path 2 on every axis.** No new dependency ⇒ no path-1 hard stop; the
seams are load-bearing (adr-25/26) ⇒ no path-3 review. This is exactly why the
build proceeds autonomously without a dependency gate. The design plan's headline:
abcd's ADRs (25/26/27 + the VSA shape) already sit on or slightly ahead of generic
SOTA here — the work is mostly naming what we have after its mature counterpart and
closing two gaps (cited evidence per criterion; pinned judge/prompt/rubric hashes).

## Open Questions

- **`spc-N` minting under reserved ids.** The corpus pre-references `spc-` ids in planning docs and itd-3 reserved `spc-1`. Slice 1 mints `max(spec-store files ∪ intent `spec_id`) + 1` to avoid live collisions; reconciling the store's sequential minting with the brief's aspirational spec numbering is deferred to the richer spec-store slice.
- **Where the review-request + receipt are parked.** A local ephemeral queue vs a committed slot — leaning local-ephemeral under `.abcd/.work.local/` for the request, with the receipt as the correlation record; settle at plan.

## Audit Notes

<Empty until intent moves to shipped/. intent-fidelity-reviewer populates this with per-criterion verdicts plus a three-bucket prose audit comparing delivered reality to the press release above. This intent dogfoods that automation.>
