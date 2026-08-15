---
name: intent
description: Press-release intent lifecycle — status, quoted-text create, the implement-readiness gate, and the human planning interview that turns a draft into a planned, specced intent.
argument-hint: "[text] | ready <itd-N> | plan <itd-N> | link <itd-N> <spc-N> | review [<itd-N>]"
---

# `/abcd:intent` — intent lifecycle

The write side of the intent record store under `.abcd/development/intents/`.
Every intent gets a stable `itd-N` id and directory-as-truth lifecycle state
(`drafts/`, `planned/`, `shipped/`, `disciplines/`, `superseded/`). Bare
invocation **performs zero writes**.

## Status (bare)

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" intent --json
```

Summarise the JSON for the user: counts per bucket, open/closed spec counts,
and the intent↔spec links. Nothing is created or moved by this invocation.

**Which ledger?** A half-formed observation, question, or nitpick goes to
`/abcd:capture "…"`; a user-facing change you want to ship goes to
`/abcd:intent "…"`. For a big, unproven idea there is an optional third route:
`/abcd:ideate` runs the admission gauntlet and records the verdict either way.
It is a pointer, never a precondition — filing a draft without it is a normal
thing to do.

## Decompose before filing (itd-84, hand-run)

One proposal is rarely one record. Before running the create command below,
run the itd-84 decomposition as an advisory analysis the human confirms —
capture routes the pieces, it never files a monolith:

1. **Route.** Split the proposal into parts and route each to its record
   home — user-facing capability → an intent; trust-boundary rule → an ADR
   (plus a brief invariant); standing stance → a principle; plumbing → the
   brief. Render the result as a table: part | type | home.
2. **Link.** Surface the existing records the proposal touches, with typed
   links — `supersedes` / `reverses` / `duplicates` / `refines` — never
   "related". A reversal ("this reverses invariant X") is *flagged for the
   human to confirm*, never auto-classified.
3. **Verdict.** Propose one of three outcomes — FILE-AS-IS / SPLIT / HOLD.
   The human adopts the routing; only the part they confirm as an intent
   proceeds to the create command, and the other parts go to their homes in
   the same session (or are captured so they are not lost).
4. **Grade.** Append the confirmed table — and whether the initial routing
   survived the human's confirmation — to the dated decomposition-calibration
   note under the development record's `research/notes/`. That corpus (about
   50 graded captures) is what gates the automated rung.

**Not yet automated.** The deterministic pre-pass and the capture-time
validator are future rungs of the itd-84 discipline; until they ship, this
documented protocol is the gate.

## Create a draft

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" intent "<text>" [--impact <additive|breaking|fix>] --json
```

Files `drafts/itd-N-<slug>.md` seeded from the text. Report the new `id` and
`path`, and tell the user the seeded Acceptance Criteria section is a
placeholder that must be replaced with real Given-When-Then bullets — via the
planning interview below — before the draft can be planned.

`--impact` is optional: a draft is "not judged yet", so an unset impact writes
no field. When you do set it, the value is validated (one of `additive`,
`breaking`, `fix` — never `internal`, since an intent is user-facing by
definition) and stamped onto the draft, where it travels unchanged through
planning to `shipped/`, which the `intent_impact_valid` gate requires.

## THE RULE: no implementation without a planned, specced intent

Before implementing ANY `itd-N` — or whenever the user asks you to "build",
"implement", or "work" an intent — run the gate first:

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" intent ready <itd-N> --json
```

- **Exit 0 (ready):** proceed. The linked spec's body is the design record to
  build against.
- **Exit 1 (not ready): DO NOT IMPLEMENT.** Do not improvise acceptance
  criteria, do not write code toward the intent, and do not run
  `abcd intent plan` on your own authority. Tell the user plainly:
  "`<itd-N>` is not specced, so it cannot be implemented yet", present each
  failing check's `detail` and `remedy` from the JSON, and **offer the
  planning interview** below.
- **Exit 2 (fault):** the id is malformed, the intent is unknown, or a record
  is unreadable — report the diagnostic; there is nothing to gate.

## Planning interview (host-run, with the human present)

The interview turns a draft into an intent the maintainer has signed off. Run
it only in a live session with the human; deferral of any question is a valid
answer, but silence is not consent.

1. Read the draft record; summarise it back: the press release, why it
   matters, the current Acceptance Criteria (say explicitly when they are
   facilitator- or agent-seeded and unconfirmed), and any open questions.
2. **Decomposition (itd-84):** run the hand-run table above over the draft —
   parts → homes, typed links, advisory reversal flags. A part that is not
   this intent moves to its home (or is captured) before planning proceeds;
   grade the run into the calibration note either way.
3. **Press release:** confirm or refine the user moment with the human.
4. **Open questions:** resolve each with the human, or record an explicit
   deferral in the draft. An open question that gates scope blocks planning.
5. **Acceptance criteria:** walk EVERY Given-When-Then bullet; the human
   accepts, edits, or strikes each, and adds what is missing. Seeded criteria
   are proposals, never approvals.
6. Edit the draft file to the confirmed content.
7. Only after the human explicitly confirms the criteria are theirs, run:

   ```bash
   "${CLAUDE_PLUGIN_ROOT}/abcd" intent plan <itd-N> --json
   ```

   This invocation IS the maintainer's sign-off act — never run it unattended
   or infer consent. It mints the spec stub, links both sides, and moves the
   intent `drafts/ → planned/`.
8. **Spec build:** replace the minted spec body's `_Draft:` placeholder with
   the real design record — scope, approach, and how it satisfies each
   acceptance criterion.
9. Re-run `abcd intent ready <itd-N>` and report READY to the user.

## Autonomous runs

In an unattended run, exit 1 from `ready` is a SKIP: journal the rendered
findings and move to the next item. The planning interview, acceptance-criteria
authoring, and `abcd intent plan` are human-session-only acts.

## Link

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" intent link <itd-N> <spc-N> --json
```

Retroactively writes a planned intent's `spec_id` when a spec already claims
it (the one-sided-link remedy `ready` reports). Report the linked pair.

## Review / ingest

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" intent review <itd-N> --json                       # re-emit a shipped intent's review request
"${CLAUDE_PLUGIN_ROOT}/abcd" intent review ingest --verdict-json <file> --json  # apply a host-produced verdict
```

Ingest is fail-closed: report the returned status (`ingested`, `dead_letter`,
or `noop`) and, for `dead_letter`, the reason.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
