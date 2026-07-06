# `/abcd:reflect` — Phase Retrospective

`/abcd:reflect <phase-id>` composes a structured retrospective for a **completed
phase** of the voyage (itd-24). It is **phase-only grain**: the per-intent form
was dropped (per-intent reflection is the `intent-fidelity-reviewer`'s Role 1).
The command markdown performs ZERO writes — every write goes through the
deterministic reflect-writer capability, which renders
`.abcd/retrospectives/<phase-id>/README.md`.

This surface doc records the design contract; the runtime behaviour (contract
verification, README write, consumed-receipt-path + phase/audit/member-spec
links) is owned by task `fn-83-operator-surfaces-manifest-lockstep.3` and the
command file `commands/abcd/reflect.md`.

## Argument

The command takes exactly one positional argument: a **phase id** (e.g.
`phase-1-substrate`, `phase-5-roundtrip`). It is NOT `itd-N` and NOT a
milestone/`fn-N` id. `/abcd:reflect <itd-N>` is refused — reflection is
phase-grained only.

Bare `/abcd:reflect` (no argument) renders help/state and writes nothing.

## What it does

1. Selects the **latest** fn-66 phase-audit receipt whose `phase_id` matches the
   argument (newest `timestamp` wins) at
   `.abcd/logbook/audit/phase-<ts>/report.json`.
2. Runs the `reflection-composer` agent as a seeded single-pass interview: five
   seeded questions drawn from the receipt's per-bullet acceptance verdicts.
   Thin answers trigger one clarifying question; a deliberately-empty section
   renders an explicit "none recorded" line.
3. Shells the deterministic writer with the collected answers as JSON. The
   writer renders the five-section README, records the consumed audit-receipt
   path, and links to the phase doc + audit report + member specs ONLY.

## The five-section template (enforced)

The retrospective README always carries these five sections, in this order:

| # | Section | Content |
|---|---------|---------|
| 1 | Went well | Successes and strengths, with specific examples |
| 2 | Could improve | Issues and gaps |
| 3 | Lessons learned | Transferable insights framed for future-you |
| 4 | Decisions made | Architectural / design choices crystallised in the phase |
| 5 | Metrics | Qualitative + simple counts (no DORA/velocity telemetry) |

An empty section triggers a clarifying question in the interview; a
deliberately-empty section renders an explicit "none recorded" line rather than
being omitted.

## Refusals

The writer refuses (each with a message naming the phase-audit prerequisite):

| Refusal | Condition |
|---------|-----------|
| Non-phase argument | `itd-N` or free text — phase-only grain |
| No reflection answers | A bare `{}` on stdin (hollow all-"none recorded") — refused unless `--allow-empty` |
| No fn-66 audit receipt | No phase-audit receipt exists for the named phase |
| Empty-audited latest receipt | `member_specs` empty OR `done_total.total == 0` — nothing shipped to reflect on |
| Re-run without `--overwrite` | A retrospective already exists for the phase |

The writer also enforces write-site containment (defense-in-depth): the resolved
target must be inside `.abcd/retrospectives/`, and receipt-supplied
`member_specs[].spec_id` values are validated against the `fn-NN-slug` shape
before they are rendered into link text.

## Invocation model

The HOST session runs the reflection-composer interview per
`agents/reflection-composer.md`, collects the structured answers as a JSON object
(one key per section: `went_well`, `could_improve`, `lessons_learned`,
`decisions_made`, `metrics`), and pipes that JSON to the writer. The writer is
fully testable WITHOUT the agent — JSON answers in, README out. The writer is the
SINGLE dispatch target and the ONLY writer.

## Output path and single-source-of-truth

Output is fixed at `.abcd/retrospectives/<phase-id>/README.md` (a peer of
`.abcd/intents/` and `.abcd/logbook/`), committed as part of the phase's
permanent record. The README LINKS to the phase doc, the audit report (its
receipt path recorded in the README), and each member spec — it never copies
their bodies. v1 links are limited to those three: the fn-66 receipt carries no
intent ids, so intent links are a recorded future extension.

Canonical glossary terms (`voyage`, `persona`) are used in body prose; the README
records `glossary_terms_used: core/voyage, core/persona`.

## Lifeboat forward requirement (grill Q6)

The lifeboat must pack EVERY phase retrospective a voyage produced, so the full
reflection arc travels between voyages. disembark/embark are unbuilt (fn-17
stubs), so this is a **documented forward requirement on the future disembark
spec** — NOT a behaviour this surface implements. It is recorded here and in the
itd-24 intent acceptance so a later reader treats it as a requirement, not a
shipped capability.

## v1 scope

- Single seeded interview pass (per-bullet verdicts → five questions); multi-turn
  depth is a recorded future extension.
- No auto-triggering after phase close; reflection is deliberately on-demand.
- The draft's "offer to run the phase-fidelity-reviewer inline" on a missing
  audit is deferred — v1 refuses instead (recorded in the itd-24 intent).

## Related documentation

- Command file: `commands/abcd/reflect.md`
- Agent: `agents/reflection-composer.md` (the 16th catalog agent — see
  [`../05-internals/01-agents.md`](../05-internals/01-agents.md))
- Intent: `itd-24` (`../intents/…/itd-24-reflect-command.md`)
- The fn-66 phase-audit contract reflect consumes:
  `scripts/abcd/schemas/phase_review_report.schema.json`
- Naming / VR001 registration: [`../02-constraints/04-naming.md`](../02-constraints/04-naming.md)
