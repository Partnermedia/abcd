---
name: capture
description: Capture issues to the structured per-repo ledger and query them, by invoking the abcd binary. Bare invocation is a read-only status render; list/promote/resolve/wontfix act on the ledger.
argument-hint: "[text] | list --open|--resolved|--wontfix|--all | promote <iss-N> [--intent <itd-N>] | resolve <iss-N> <note> --impact <additive|breaking|fix|internal> [--intent <itd-N>] [--spec <spc-N>] [--commit <sha>] | wontfix <iss-N> <reason>"
---

# `/abcd:capture` — issue ledger

The lightweight write side of the structured issue ledger under
`.abcd/work/issues/`. Every issue gets a stable `iss-N` id, schema-checked
frontmatter, and folder-as-status (`open/`, `resolved/`, `wontfix/`). Bare
invocation **performs zero writes**.

## Status (bare)

To render recent captures and counts:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" capture --json
```

Summarise the JSON for the user: `open_count` / `resolved_count` /
`wontfix_count`, and for each entry in `recent_open` its `id`, `severity`, and
`slug`. No `iss-*.md` file is created, moved, or mutated by this invocation.

**Which ledger?** A half-formed observation, question, or nitpick goes to
`/abcd:capture "…"`; a user-facing change you want to ship goes to
`/abcd:intent "…"`. For a big, unproven idea there is an optional third route:
`/abcd:ideate` runs the admission gauntlet and records the verdict either way.
It is a pointer, never a precondition — capture friction stays at one line.

## Capture an issue

Append a structured issue from free-form text:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" capture "<text>" --json
```

Provide provenance and taxonomy through flags when known (each falls back to a
default): `--severity` (`nitpick|minor|major|critical`, default `minor`),
`--category` (default `observation`), `--source` (default `user-observation`),
`--found-during` (session/command context, default `manual-capture`),
`--found-at` (optional repo-relative path), `--slug` (overrides the slug derived
from the text), `--blocked-by` (comma-separated `iss-N` ids this issue depends
on). Report the new `id`, `status`, and `path` from the JSON. Report `redacted`
too whenever it is non-zero: it counts the spans rewritten before the text was
written, and the user needs to know their wording was changed.

Priority is **derived, never stored**: an issue is ranked lower while any of its
`--blocked-by` targets is still open, and `blocked_by` records the dependency in
one direction only (the inverse is computed).

## Query the ledger

`list` is the one earned filter-flag exception — a filter is **required**:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" capture list --open --json      # or --resolved / --wontfix / --all
```

The unfiltered form `abcd capture list` exits 2 with a "choose a filter"
message; there is no implicit default. Summarise each issue's `id`, `status`,
`severity`, and `slug`. The list is returned in **derived-priority order**:
unblocked issues first, then by severity (`critical` → `nitpick`); rows still
blocked by an open dependency are demoted and annotated `[blocked-by iss-N,…]`.

## Resolve / wontfix

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" capture resolve <iss-N> "<resolution-note>" --impact <additive|breaking|fix|internal> --json
"${CLAUDE_PLUGIN_ROOT}/abcd" capture wontfix <iss-N> "<reason>" --json
```

Each moves the issue out of `open/` and records the note; report the `id` and
the `from_status -> to_status` transition from the JSON.

`resolve` requires `--impact`: a resolved issue is in the release set, so it
carries the product judgement the version derivation reads (`additive`,
`breaking`, `fix`, or `internal` — plumbing invisible to users). There is no
default; an absent or misspelled impact is refused rather than guessed, so the
record always satisfies the `issue_impact_valid` gate. `wontfix` takes no impact
(a non-action ships nothing).

`resolve` also takes optional provenance — the structured `resolved_by` pointer
to what fixed the issue: `--intent <itd-N>`, `--spec <spc-N>`, `--commit <sha>`,
in any combination. Supply them whenever the fixing record is known — that is
what keeps the trail six months later. Ids must exist in their record store
(any bucket); the sha is shape-checked only (7–64 hex chars — its home may be
the remote). An unknown id or malformed value refuses the whole resolve and
writes nothing. With no provenance flags the record is byte-identical to a
plain resolve: provenance is optional, never guessed. The written members come
back in the JSON as `resolved_by`. `wontfix` takes no provenance — a non-action
points at nothing.

## Promote an issue into an intent

When a one-line issue turns out to be a capability, graduate it without
retyping:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" capture promote <iss-N> --json
```

One invocation mints a new intent draft under
`.abcd/development/intents/drafts/` — slug reused from the issue, body carrying
a by-id pointer ("Graduated from `iss-N`"), never a copy of the issue body —
and stamps the issue's `promoted_to` with the minted `itd-N`. The draft's
frontmatter records `promoted_from: iss-N`, so the edge is two-sided. The issue
keeps its status folder: promotion is orthogonal to fix-status and is not
resolution. An issue already carrying `promoted_to` is refused with the
existing `itd-N`.

`--intent <itd-N>` is the stamp-only mode: it links an *existing* draft instead
of minting — the repair path when a stamp failed after the mint (the error
names the orphan draft and this exact remedy), and the path for "I already
filed the intent by hand; link them". Report the `issue_id`, the minted (or
linked) `intent_id`, and both paths from the JSON.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
