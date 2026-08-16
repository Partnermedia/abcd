---
id: itd-114
slug: abcd-mints-collision-proof-record-ids-across-parallel-agents
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
---

# Two Agents Can Mint At The Same Instant And Never Collide

## Press Release

> **abcd record ids stop colliding when the work goes parallel.** Every id abcd
> mints — issues, intents, specs, decisions — is collision-proof by
> construction, with no coordination between agents: a native default that
> needs no network stamps each id with a time-ordered prefix and a content
> hash, so two agents minting the same instant on different branches produce
> different ids and neither is ever renumbered. Where a team would rather let a
> forge own the numbers, an optional backend records straight to GitHub issues,
> where the number is allocated server-side and is atomic by definition — the
> same `capture` verb, a different store. The id stays human-legible and
> time-sortable either way; collision-safety is not paid for in readability.
>
> "We had three agents draining plans in parallel and every few captures two of
> them would mint the same `iss-N` and one branch would fail the merge check,"
> said Bob, a maintainer. "Now they just run. Alice's routine and Carol's
> routine never step on each other's ids, and I still read the record at a
> glance."

## Why This Matters

Record-id allocation is branch-local: each agent scans the refs for the
family's highest id and takes the next one ([iss-80], resolved). `MaxAcrossRefs`
closed the common commit-then-branch case, but its own resolution names the
residual and defers the real fix: *"two branches that BOTH mint before either
commits still collide … a collision-free allocation scheme (forge-minted /
random-suffix / timestamp / reserve-registry — SOTA under research)."* This
intent is that deferred scheme.

The pain is no longer hypothetical. Once work runs across parallel agents — two
subscriptions draining two plans, plus an autonomous bug-hunt — minting is
constant and concurrent, and the manual mitigation is agents messaging each
other *"what is your highest in-flight id?"* before every mint. That does not
scale to unattended routines. The record-lint uniqueness detectors make a
collision fail-safe (it is caught at merge, never silent) but not free: it
forces a renumber, which breaks the *ids-are-capture-stable* invariant the same
record depends on. Collision-proof minting removes the coordination tax and the
renumber both.

## SOTA

Anchors: **UUIDv7 / ULID** (time-ordered, collision-resistant ids generated with
no central coordinator — the current answer for distributed id generation);
**git's short-hash / full-sha** dual id (a human-friendly handle over a
collision-proof stable core); and **forge server-side issue numbering**
(GitHub/GitLab allocate `#N` atomically, so concurrent creation cannot collide).

**Declared path: 2 — native floor, adapter seam.** The native default is a
time-ordered id plus a content-hash discriminator (UUIDv7/ULID-shaped, adapted
to the record families) — no network, no coordinator, in-tree, offline, and
part of the durable record, exactly as abcd's [basics-built-in /
adapter-over-native-default] stance requires. The forge backend (GitHub issues,
server-allocated numbers) is the **optional adapter** behind the same `capture`
seam — opt-in, never a runtime dependency, so this path proceeds without the
new-dependency stop. The seam is one id-minting interface with two
implementations (native / forge); the independent fit-challenge runs at plan
time.

## Acceptance Criteria

> _BDD (the [itd-1] discipline)._

- **Given** two agents on separate branches each `capture` an issue at the same
  instant, **when** both mints complete before either pushes, **then** the two
  ids differ and no merge-time uniqueness check ever forces a renumber.
- **Given** no network is available, **when** an id is minted, **then** it is
  collision-proof by the native scheme alone (no forge call, no failure).
- **Given** the optional forge backend is configured, **when** `abcd capture`
  runs, **then** the record is created on the forge with its server-allocated
  number and abcd references that number stably.
- **Given** a record carrying a legacy sequential id (`iss-200`), **when** the
  new scheme is in effect, **then** the legacy id stays valid and stable — no
  forced renumber of history.
- **Given** a human reads or sorts records, **when** they need recency or
  identity at a glance, **then** the id is time-sortable and legible — the
  readability property of the old sequential ids is preserved, not traded away
  for collision-safety.

## Decomposition (itd-84 hand-run, 2026-08-16)

Verdict **FILE-AS-IS with flags** — one capability, no separate records to file
now:

| Part | Type | Home |
|------|------|------|
| Collision-proof id minting, native default + optional forge backend | capability | this intent |
| The native time+hash scheme and the forge-adapter seam | SOTA declaration (path 2) | this intent's SOTA section (above) |
| "Ids are collision-proof by construction; native-default, forge-optional" | stance | embodies [basics-built-in / adapter-over-native-default / prefer-sota] — no new principle |
| Changing the id **format** (sequential `iss-N` → time+hash) pervades the record | architecture decision, **only if adopted** | a **future ADR at adoption**, not now |

Typed links: `refines` [iss-80] (the deferred "SOTA under research" scheme it
names); driven by parallel autonomous routines (the itd-107 lineage).

**Advisory reversal flag (human confirms at planning):** a pure time+hash id
would reverse the *human-sequential-readable, orderable* property of today's
`iss-N`. The last acceptance criterion is written to forbid that reversal — the
scheme must stay legible — but whether the answer is a time-sortable id, a
git-style dual id (collision-proof stable core + human short display id), or a
forge number is a genuine design fork for the grill, not settled here.

## Open Questions

- **Readability vs collision-safety — the central fork.** Candidate
  reconciliations: (a) a time-ordered id (ULID-shaped) that sorts by recency;
  (b) a git-style dual id — a collision-proof stable id plus a short
  human/sequential display id assigned single-writer at merge; (c) the forge
  number as the id when the adapter is on. Resolve at the grill.
- **Migration.** Do legacy `iss-N`/`itd-N`/`spc-N` stay as-is (dual-scheme, new
  ids in the new format) or is there a one-time remap? The stable-id invariant
  argues for leave-as-is.
- **Which families.** All of `iss-N`/`itd-N`/`spc-N`/adr, or start with the
  highest-churn one (captures)?
- **Forge-adapter scope.** Does `abcd capture --backend github` also *read*
  (list/resolve) via the forge, and how does folder-as-status map onto GitHub
  open/closed + labels? Ties to the record-lint enforcement that assumes in-tree
  files.
- **Interaction with the record-lint uniqueness detectors** — do they retire, or
  become a cheap assertion that the scheme held?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

[iss-80]: ../../../work/issues/resolved/iss-80-record-id-allocators-itd-n-spc-n-iss-n-are-branch-local-para.md
[itd-1]: ../disciplines/itd-1-acceptance-gates.md
[basics-built-in / adapter-over-native-default / prefer-sota]: ../../principles/prefer-sota.md
