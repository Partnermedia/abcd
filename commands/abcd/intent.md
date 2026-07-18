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
abcd intent --json
```

Summarise the JSON for the user: counts per bucket, open/closed spec counts,
and the intent↔spec links. Nothing is created or moved by this invocation.

**Which ledger?** A half-formed observation, question, or nitpick goes to
`/abcd:capture "…"`; a user-facing change you want to ship goes to
`/abcd:intent "…"`.

## Create a draft

```bash
abcd intent "<text>" --json
```

Files `drafts/itd-N-<slug>.md` seeded from the text. Report the new `id` and
`path`, and tell the user the seeded Acceptance Criteria section is a
placeholder that must be replaced with real Given-When-Then bullets — via the
planning interview below — before the draft can be planned.

## THE RULE: no implementation without a planned, specced intent

Before implementing ANY `itd-N` — or whenever the user asks you to "build",
"implement", or "work" an intent — run the gate first:

```bash
abcd intent ready <itd-N> --json
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
2. **Press release:** confirm or refine the user moment with the human.
3. **Open questions:** resolve each with the human, or record an explicit
   deferral in the draft. An open question that gates scope blocks planning.
4. **Acceptance criteria:** walk EVERY Given-When-Then bullet; the human
   accepts, edits, or strikes each, and adds what is missing. Seeded criteria
   are proposals, never approvals.
5. Edit the draft file to the confirmed content.
6. Only after the human explicitly confirms the criteria are theirs, run:

   ```bash
   abcd intent plan <itd-N> --json
   ```

   This invocation IS the maintainer's sign-off act — never run it unattended
   or infer consent. It mints the spec stub, links both sides, and moves the
   intent `drafts/ → planned/`.
7. **Spec build:** replace the minted spec body's `_Draft:` placeholder with
   the real design record — scope, approach, and how it satisfies each
   acceptance criterion.
8. Re-run `abcd intent ready <itd-N>` and report READY to the user.

## Autonomous runs

In an unattended run, exit 1 from `ready` is a SKIP: journal the rendered
findings and move to the next item. The planning interview, acceptance-criteria
authoring, and `abcd intent plan` are human-session-only acts.

## Link

```bash
abcd intent link <itd-N> <spc-N> --json
```

Retroactively writes a planned intent's `spec_id` when a spec already claims
it (the one-sided-link remedy `ready` reports). Report the linked pair.

## Review / ingest

```bash
abcd intent review <itd-N> --json                       # re-emit a shipped intent's review request
abcd intent review ingest --verdict-json <file> --json  # apply a host-produced verdict
```

Ingest is fail-closed: report the returned status (`ingested`, `dead_letter`,
or `noop`) and, for `dead_letter`, the reason.

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd intent …`
from the repo root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
