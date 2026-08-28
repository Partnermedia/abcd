# abcd Specs

The native spec store — one engineering design record per planned intent, and
the document that intent's fidelity review audits against.

---

## What's a Spec?

A **spec** is the *how* layer of the three-layer model — the brief is the whole
picture, an intent is the user-facing why, a spec is how it gets built (see
[adr-1](../decisions/adrs/0001-three-layer-mental-model.md)). It states the
scope, the approach, and the delivery of exactly one intent's promised
capability.

Specs are minted by the machinery, never by hand. `/abcd:intent plan <itd-N>`
calls `spec.Create` ([`internal/core/spec/store.go`](../../../internal/core/spec/store.go)),
which writes `open/spc-N-<slug>.md` carrying the back-link to the intent and a
`## Summary` placeholder for the author to replace. That placeholder is
load-bearing: `abcd intent ready <itd-N>` reports the intent not ready while the
body still carries it, because an unwritten design record is not implementable.

---

## Spec IDs

Filenames follow `spc-N-<slug>.md` — `N` unpadded (mirroring `itd-N`), the slug
kebab-case (`^[a-z0-9]+(?:-[a-z0-9]+)*$`) and inherited from the intent, since
`Create` is handed the intent's own slug. The store's loader recognises exactly
`^(spc-[0-9]+)-.*\.md$` inside a bucket directory, while the lint's filter is
looser (`^spc-\d+.*\.md$`, no required hyphen), so a malformed `spc-N.md` is
skipped by the store yet still linted. A file matching neither shape — this
README included — is not a spec to either.

**The mint.** `Create` validates the intent id and the slug *before* any path is
built (the slug becomes a filename), then performs the whole allocation under an
exclusive advisory lock: `flock(2)` on the `specs/` directory's own file
descriptor, opened `O_NOFOLLOW`, waited on for up to five seconds. Holding the
lock across the scan *and* the write is what stops two concurrent `intent plan`
runs from observing the same maximum and minting a duplicate — the filenames
differ by slug, so neither the atomic write nor the clobber guard would catch it.
Locking the directory descriptor itself leaves no lock artefact in the committed
record tree.

Under that lock, `NextID` allocates:

```
max( N over spec-store files
   ∪ N over every intent's `spec_id` frontmatter across the five intent buckets
   ∪ N over spec-store filenames on every other git ref ) + 1
```

Each term earns its place. The intent-reservation term keeps a fresh mint off a
*reserved* id: itd-3 carries `spec_id: spc-1` with no spec file behind it, so a
store-only scan would hand `spc-1` out again — folding reservations in is why the
first minted id is `spc-2`. The cross-ref term (`recordid.MaxAcrossRefs`) keeps
two branches from re-minting the same id once either commits, seeing a spec file
that is absent from the current working tree. When that ref scan degrades to
working-tree-only — git absent, not a repository, a failed ref query — `Create`
returns a mint warning that its caller must surface, so the weaker guarantee is
never silent. A count near the integer ceiling is refused outright rather than
wrapped into a negative id.

**The lock is scoped to one checkout.** An advisory lock on one worktree's
`specs/` directory cannot see a sibling worktree, and an uncommitted mint is on
no ref, so two checkouts minting in the same window still allocate the same
`spc-N` by construction. Say which family you are about to mint into, or mint
from one checkout.

**IDs are capture-stable.** A spec keeps its `spc-N` for life; closing it moves
the file and renumbers nothing. The `spec_id_unique` record-lint rule is the
backstop that flags a number claimed by two files across both buckets.

---

## Lifecycle — the bucket directory is the status

| Directory | Meaning |
|---|---|
| [`open/`](open/) | In flight. Minted by `/abcd:intent plan`, not yet closed. |
| [`closed/`](closed/) | Delivered. `abcd spec close <spc-N>` renames the file here and, in the same synchronous call, reconciles the linked intent `planned/` → `shipped/` and emits its OWED fidelity-review receipt. |

There is deliberately **no `status:` frontmatter field**. Directory location is
the single source of truth for lifecycle state
([adr-3](../decisions/adrs/0003-directory-as-truth-for-lifecycle.md)); the spec
store applies that principle exactly as the intent store does. The record lint
enforces it — `validateSpec` in `internal/core/lint/lint.go` flags any spec
carrying a `status:` key under the `spec_lifecycle` rule, with the message
*"status: key forbidden; spec status is the bucket directory (open/closed), not
a field"*. A second source of truth is a source that drifts; the file move is the
transition, and `git log` is its history.

The transition is a rename, not an edit. `spec.Close` fails closed on a spec
that is missing or already closed, and refuses to overwrite a file of the same
name already sitting in `closed/`.

---

## The Intent Back-Link

| File | Frontmatter field |
|---|---|
| `specs/{open,closed}/spc-N-<slug>.md` | `intent: itd-N` — required, scalar, `^itd-[0-9]+$` |
| `intents/{drafts,planned,shipped}/itd-N-<slug>.md` | `spec_id: spc-N` (scalar, or `null` while in `drafts/`) |

`intent:` is the load-bearing field, and it is fail-closed at both ends.
`spec.Validate` rejects any record whose id or intent link is malformed, and
`spec.Load` aborts the *whole* store on a single bad file rather than serving a
partial one.

The record lint's `spec_lifecycle` rule (`checkSpecLifecycle`) then makes three
checks per spec, over the one traversal `ScanSpecLinks` performs of both trees:

1. the id and the intent link are well-formed;
2. the named intent **exists** in some intent bucket;
3. the load-bearing cross-check — that intent points **back**, its `spec_id`
   carrying the same number as this spec's id. Drift either way (the intent
   names a different spec, or `null`) is reported as *bidirectional drift*.

Comparison is on the spec *number*, so `spc-9`, `spc-9-widget` and `spc-009` are
one handle everywhere. A spec exempted from the content rules is still held to
well-formedness: being historical excuses a record from how it is written, never
from being well-formed, and `spec.Load` has no exemption concept at all.

Both directions exist once `/abcd:intent plan` runs. Every spec in the store
names exactly one intent; the bundle direction, where several intents share one
spec, is declared on the intent side and its shared spec is not yet minted (see
the [intents charter](../intents/README.md)).

---

## Format

```markdown
---
id: spc-N
slug: <kebab-case-slug>
intent: itd-N
---

# <slug>

## Summary

<What this spec delivers for its intent — scope, approach, and how it satisfies
the intent's Acceptance Criteria.>
```

Three frontmatter keys, and only three, across the whole store. `## Summary` is
the one heading every spec carries; beyond it the corpus converges on `## Scope`,
`## Approach`, an acceptance-criteria mapping, and `## Out of scope` without
mandating them — the spec body is a design record, not a form.

---

## Two `spc-N` Namespaces — Always Qualify Above the Ceiling

The live store holds **spc-2 … spc-42** — 41 records, one per number. `spc-1` is
reserved by itd-3 and has never been a file in this history.

A **retired predecessor store** minted its own `spc-1 … spc-83`, and prose across
the record still cites those records — spc-56, spc-66, spc-75 and the bundle
handle `spc-83-operator-surfaces` among them. The two namespaces overlap, so a
bare `spc-N` is ambiguous wherever `N` falls inside the live range. Three rules
follow:

- **Every reference to a predecessor record carries a qualifier.** The
  convention in the record is the parenthetical *"(predecessor store)"* —
  "spc-12 (predecessor store)", or "the predecessor store's spc-66".
- **An id above the live ceiling belongs to the predecessor store by
  construction.** `spc-43` … `spc-83` name no live record, so prose citing one
  means that store even where the qualifier is absent.
- **Two ids collide outright.** `spc-33` and `spc-37` exist in both namespaces on
  different subjects: live spc-33 is itd-114's collision-proof record-id mint and
  live spc-37 is itd-135's abcdev.app landing page, while the predecessor's
  spc-33 and spc-37 are the harness-security records itd-51 cites. An unqualified
  reference to either resolves, in the store, to a spec about something else — so
  it must never be written unqualified.

The backlog of draft intents citing predecessor ids as though they were live is
tracked as iss-239 in the issue ledger.

---

## No Listing Here

This charter describes the store by *rule* — the grammar, the mint, the
lifecycle, the link, and the namespace. It transcribes neither the open set nor
the closed set: the directories are the listing, and a copy kept here goes stale
on the next mint or close (adr-5, derive don't store).
