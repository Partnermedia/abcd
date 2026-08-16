# Process coherence and walkability

**Date:** 2026-08-16
**Status:** planning complete, intents not yet filed
**Decision record:** [adr-40](../decisions/adrs/0040-review-audit-lint-are-three-verbs.md)
**Issues:** iss-244, iss-245, iss-246

## Why this exists

abcd cannot be handed to a tester, because the process it exists to run cannot be
followed through once. Two things block it, and they are independent.

**The process does not complete.** Two of the twelve steps in the record
lifecycle have no working verb.

**The vocabulary contradicts itself.** Three shipped surfaces are named for the
opposite of what they do, so a maintainer cannot reason about which verb to reach
for.

Fixing the second without the first produces a beautifully-named system nobody
can walk through. Walkability comes first.

## The walk, traced against the binary

| # | Step | Verb | State |
|---|---|---|---|
| 1 | observe something | `abcd capture "…"` | ships |
| 2 | decide it is a capability | `abcd capture promote iss-N` | **missing** |
| 3 | or file directly | `abcd intent "…"` | ships |
| 4 | planning interview | host-run, `commands/intent.md` | ships |
| 5 | sign-off | `abcd intent plan itd-N` | ships |
| 6 | write the spec body | manual edit | n/a |
| 7 | may I build? | `abcd intent ready itd-N` | ships |
| 8 | implement | the agent | n/a |
| 9 | close the work | `abcd spec close spc-N` | ships |
| 10 | fidelity verdict | host agent | ships |
| 11 | record it | `abcd intent review ingest` | ships |
| 12 | close the issue | `abcd capture resolve` | ships, but **loses the trail** |

Ten of twelve work. Step 2 has no verb. Step 12 cannot write `resolved_by`, so a
resolved issue asserts it was fixed in prose while the slot for pointing at the
intent, spec, or commit that fixed it sits parsed-but-unwritable (iss-245).

Neither gap is closed by the vocabulary work. Conversely, closing both would make
the process walkable without renaming anything.

## The two jobs

**Walkability** — the process completes end to end, and a tester can observe
where any record stands.

**Coherence** — every verb says what it does, and the record cannot drift back.

## The six intents

Each needs a planning interview with the maintainer present, because
`abcd intent plan` is a human-session-only act and an autonomous run hitting
`intent ready` exit 1 will correctly skip the work. **No intent here may be
planned unattended.**

### Walkability

1. **`abcd capture promote <iss-N>`** — an issue graduates into an intent
   without retyping it. The schema already models this: `PromotedTo`
   (`internal/core/capture/capture.go`) is validated against `^itd-[0-9]+$` and
   documented in the issues README, but no verb writes it. The
   `04-surfaces/README.md` row already marks `promote` a design target. This is
   the verb that makes "capture now, decide later" safe, and it closes the
   forced-choice the "Which ledger?" note in `commands/intent.md` currently
   imposes at the moment of lowest information.

2. **`abcd capture resolve` writes `resolved_by`** — `ResolvedBy{Intent, Spec,
   Commit}` is parsed on read but `Resolve` writes only `resolution` and
   `impact`, and exposes no flag to supply it. Without this the ledger cannot
   point at what fixed an issue, while the intent side carries a SHA-256
   receipt: one record store, two evidence standards.

3. **`abcd <id>`** — dispatch on a record id and report what it is and what to do
   next. `iss-`, `itd-`, and `spc-` prefixes are already globally unique and
   regex-validated, so the id is the routing. The mental model the maintainer
   set: bare `/abcd` answers *what can I do*, `abcd <id>` answers *what is this,
   and what is my next move*. This is what makes the twelve-step walk observable
   rather than a directory hunt, and it dissolves the need to know that shipping
   an intent is a `spec` verb. SD001-safe: a positional on the namespace root is
   not a `show` sub-verb.

### Coherence

4. **Sub-verb tables plus extended `surface_coverage`** — each surface file under
   `04-surfaces/` carries a sub-verb table recording two facts per verb: which
   bucket it is (lint / review / audit / gate, per adr-40) and whether it exists
   (`shipped` / `staged`). `surface_coverage` extends to check each row against
   registered cobra sub-commands, both directions. This is the detector adr-40
   needs and the fix for iss-246 in one table. **It must land armed, with rows
   reflecting current names, before any rename** — so the renames are proved
   complete by a gate rather than asserted complete by an agent.

5. **`abcd intent review` → `abcd intent audit`** — it emits family 2
   (`MET`/`NOT_MET`/…); the brief already calls it the intent audit. The
   `intent-fidelity-reviewer` agent and the `intent_review` task-class token move
   with it. Breaking; ~37 files reference the current name.

6. **`abcd audit` → a lint-shaped name** — it is deterministic rule-checking,
   and `02-constraints/04-naming.md` reserves `/abcd:audit` for itd-16's
   hash-chain fidelity checks. Breaking; ~57 files reference the current name.
   **The replacement name is an open question** (below).

## Settled in the planning grill

Recorded in full in [adr-40](../decisions/adrs/0040-review-audit-lint-are-three-verbs.md).

- **Four buckets** — lint, review, audit, gate — separated by what each compares.
- **One surface, one act.** Multi-act surfaces are split at design time, which is
  cheap because every multi-act surface abcd has is unbuilt.
- **Closed list, PR-to-extend**, not a criterion — an autonomous run given a
  criterion interprets it; a table lookup either matches or fails.
- **`intent ready` survives as a `gate`** — it names a decision, not a comparison.
- **Determinism is orthogonal** and never names a bucket.
- **Automatic vs manual never enters a verb name** — the trigger is the
  facilitator's, per itd-97.
- **Clean break, no aliases.** Pre-1.0.0; `--impact breaking` drives version
  derivation (iss-171 precedent). Users re-download. An alias would additionally
  collide with the documentation discipline's ban on change-narration.

## Open questions for the interview

1. **What replaces `abcd audit`?** `abcd lint` matches the existing
   `record-lint` / `docs-lint` / `lint-reviews` family and is honest about the
   act, but `abcd docs lint` already exists as a sub-verb, so `abcd lint` as a
   top-level repo-conformance verb sits slightly oddly beside it. Alternatives
   considered and not chosen: `conform`, `check`.
2. **Does `disembark oracle` rename?** It emits family 1 but the brief calls its
   output "the oracle audit" in five places. `oracle` names the model-access seam
   (adr-25), not an assessment, so the recommendation is to fix the prose to
   "oracle review" and leave the verb. Not yet ruled on.
3. **Does the `Status` enum gain `partial` at surface grain**, or is sub-verb
   granularity sufficient to retire the ambiguity? Sub-verb rows may make a
   surface-grain third value redundant.

## Ruled at the planning interview (2026-08-16)

All three open questions were settled with the maintainer present, and every
intent below is planned, specced, and `ready` exit 0:

1. **`abcd audit` → `abcd lint`** — matches the `record-lint` / `docs-lint` /
   `lint-reviews` family; `abcd docs lint` reads as the same word at a
   narrower scope.
2. **`disembark oracle` → `disembark review`** — the investigation found the
   binary verb never invokes the oracle seam (it is a compute-or-ingest
   verdict endpoint), so the plan's keep-the-verb recommendation was reversed
   and adr-40 §5 amended in place. This adds a **seventh intent**. The
   artefact moves to `review/review-<manifest12>.*` with clean replacement
   across the rename; the agent becomes `lifeboat-reviewer`.
3. **No `partial`** — the `Status` enum stays two-valued at both grains; the
   sub-verb rows carry the granularity.

Id map: itd-119/spc-24 (`capture promote`), itd-120/spc-25 (`resolved_by`),
itd-121/spc-26 (`abcd <id>`, adr-N included read-only), itd-122/spc-27
(sub-verb tables, with binding bucket pre-rulings: identity = audit, launch
changelog guardrail = gate, `guard check` = gate), itd-123/spc-28
(`intent audit`, agent → `intent-auditor`), itd-124/spc-29 (`abcd lint`,
package → `internal/core/repolint`; engine merge captured as iss-251),
itd-125/spc-30 (`disembark review`).

## Sequencing

1. Intents 1–3 planned and specced (walkability)
2. Autonomous run builds 1–3
3. Maintainer walks the twelve steps end to end
4. Intents 4–6 planned and specced (coherence), validated against a process
   actually completed
5. Autonomous run builds 4, armed, before 5 and 6

## STOP conditions for any autonomous run over this plan

- `abcd intent ready <itd-N>` exits 1 — SKIP, journal, never improvise criteria
  and never run `abcd intent plan`.
- A rename would land while the extended `surface_coverage` is not yet armed.
- `make preflight` or `gofmt -l .` fails and the fix is not obviously mechanical.
- Any change would touch `.abcd/development/brief/` semantics rather than status
  markers — the brief is the maintainer's.
- The four-bucket classification of a surface is genuinely ambiguous. Record the
  ambiguity and stop; do not guess a bucket into a closed list.
