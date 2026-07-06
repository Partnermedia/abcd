---
id: itd-46
slug: abcd-intent-quoted-text-create-symmetric
spec_id: spc-30-symmetric-abcdintent-and-abcdcapture
kind: standalone
suggested_kind: standalone
reclassification_history: []
related_adrs: []
---

# `/abcd:intent "<text>"` And `/abcd:capture "<text>"` Become Symmetric Create Paths

## Press Release

> **abcd's two ledger commands gain a symmetric, sub-verb-free create path.** Typing `/abcd:intent "I want users to feel the loyalty card respects their time"` files the new intent directly — no `new` sub-verb required. Typing `/abcd:capture "the export button looks dead-on dark mode"` works the same way for a quick observation. The bare command (zero arguments) keeps doing what the universal abcd convention says it must: shows status and help, never mutates state. The three invocation shapes are crisp: bare → status, quoted text → create, sub-verb [+ID] → act on existing. The product thinker reaches for the same shape no matter which ledger they want.
>
> "I stopped having to remember `new`," said Carol, product lead. "Quotes mean create. That's the whole rule for both commands. The cognitive overhead of `/abcd:intent new "…"` vs `/abcd:capture "…"` was tiny, but it was the difference between a flow that felt designed and a flow that felt grown — and now it feels designed."

## Why This Matters

The `.work/issues.md` 2026-05-22 design discussion landed on a clear conclusion: **drop the `new` sub-verb from both `/abcd:intent` and `/abcd:capture`.** Quoted-text-as-create-signal is symmetric across the two commands and aligned with abcd's universal "bare = status + help, never mutates" convention. The full rationale is recorded in that ledger entry; this intent captures the work to implement it.

Three reasons the change matters beyond aesthetics:

1. **It honours the universal convention.** `docs/reference/commands.md` line 3
   states: "Every command, when called with no arguments, shows help and a
   status summary… Bare invocation never mutates state; every state-changing
   operation requires an explicit sub-verb." The current `/abcd:capture` is
   already shaped correctly (`/abcd:capture "<text>"` is the documented
   create path). `/abcd:intent` is the asymmetric outlier with its `new`
   sub-verb. This intent removes the asymmetry.
2. **It clarifies the bridge between the two commands.** `/abcd:capture promote
   <iss-N>` already hands an issue to `/abcd:intent new` as a seed. With the
   sub-verb removed, the promote path calls `/abcd:intent "<text>"` directly,
   and the conceptual model becomes "promote means: capture's content becomes
   intent's seed" rather than "promote means: invoke a specific sub-verb on
   intent". Cleaner mental model.
3. **It is a small, well-scoped surface change with no behaviour change at
   the *output* level.** The artefact a `/abcd:intent "<text>"` invocation
   produces is byte-identical to what `/abcd:intent new "<text>"` produced.
   Only the routing logic and the docs change.

This intent is project-agnostic in the same sense the rest of the command surface is — it shapes how the `/abcd:intent` and `/abcd:capture` commands behave in every project that installs abcd.

## What's In Scope

- **Update `abcdDev/commands/abcd/intent.md`** — remove the `new` row from the
  sub-verb table; add a `"<text>"` row with the same description; update the
  routing logic so the first argument being a quoted string (not a known
  sub-verb) dispatches to the create path.
- **Update `abcd/docs/reference/commands.md`** `/abcd:intent` table — remove
  the `new "<text>"` row, replace with a `"<text>"` row carrying identical
  description.
- **Update `/abcd:capture promote` plumbing** — the promote path currently
  hands content to `/abcd:intent new`; change it to call `/abcd:intent`
  directly with the issue text as the quoted-text argument.
- **Add a one-line decision rule to both commands' bare-form help output** so
  users know which ledger to reach for:
  - *Half-formed observation, question, or nitpick?* → `/abcd:capture "…"`
  - *User-facing change you want to ship?* → `/abcd:intent "…"`
- **Confirm `/abcd:capture` already follows the shape** — the
  `docs/reference/commands.md` `/abcd:capture` row for `"<text>"` already
  exists; verify the implementation matches the documented shape (per
  `.work/issues.md` 2026-05-22 the implementation is already correct here
  — confirm and move on).

## What's Out Of Scope

- **Renaming `iss-N`/`itd-N` to `C01`/`IN01` or any new ID scheme.** The
  2026-05-22 ledger entry explicitly rejects this as a third partial
  terminology migration (sibling to itd-43's epic→spec sweep).
- **Changing the sub-verb-required rule for mutating sub-verbs.** `refine`,
  `grill`, `plan`, `ship`, `review`, `reclassify`, `link`, `resolve`,
  `wontfix`, `promote` continue to require their target's ID. Corpus-scope
  scan verbs (`list`, `consistency`, `shape`) continue not to require an ID.
  Already-correct shape; this intent does not modify it.
- **Merging `/abcd:intent` and `/abcd:capture` into one command.** The
  ledger entry explicitly rejects a unified `/abcd:capture <kind>` design;
  this intent implements the conclusion of that discussion, not the
  rejected alternative.

## Acceptance Criteria

- *Given* a contributor running `/abcd:intent "I want users to feel X"`, *when* the command executes, *then* a new intent file is created under `intents/drafts/` with content seeded from the quoted text — byte-identical to what `/abcd:intent new "I want users to feel X"` would have produced.
- *Given* a contributor running `/abcd:intent new "I want users to feel X"`, *when* the command executes, *then* it either (a) routes to the create path as before (backwards-compatible alias), or (b) errors with a clear message naming the new shape. Decide at plan; lean (a) for a transition period, (b) eventually.
- *Given* a contributor running `/abcd:intent` (bare), *when* the command executes, *then* status + help is shown and no file is created — the universal bare-invocation convention is preserved.
- *Given* a contributor running `/abcd:capture promote iss-N`, *when* the command executes, *then* the issue text is handed to the intent create path and a new intent file appears under `intents/drafts/` — same outcome as before, different plumbing.
- *Given* a contributor reading the bare-form help output of either command, *when* they scan the help text, *then* the one-line decision rule ("nitpick → capture; user-facing change → intent") is visible.

## Open Questions

- **Backwards-compatibility window for `/abcd:intent new`.** Keep it as an
  alias for one phase to ease the transition, then remove it? Or remove
  immediately on landing? Lean: alias-for-now with a deprecation warning,
  remove in a later sweep.
- **Documentation rewrite extent.** Updating `commands.md` is trivial; the
  question is whether the supporting docs (the `facilitator.md` role doc,
  the `intent.md` skill files, any tutorial content) need parallel
  updates. Likely yes — record at plan.
- **Promote path testing.** `/abcd:capture promote <iss-N>` is a relatively
  rarely-exercised path; regression tests for it should land with this
  intent's spec, not as a side note.

## Related

- **`.work/issues.md` 2026-05-22 line 413** — the design discussion that
  produced this intent.
- **`docs/reference/commands.md`** — the doc that gets updated.
- **`commands/abcd/intent.md`** — the command-routing surface.
- **itd-4** (issue-capture intent) — the broader intent for `/abcd:capture`'s
  existence; this intent refines `/abcd:intent` to match `/abcd:capture`'s
  already-correct shape.
- **itd-43** (epic→spec terminology) — precedent for *not* introducing a
  partial terminology migration; this intent preserves `iss-N`/`itd-N`.
