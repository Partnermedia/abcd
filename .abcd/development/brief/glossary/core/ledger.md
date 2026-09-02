<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: ledger
bounded_context: core
definition: An append-or-move store a command writes and a human reads back — the issue ledger under .abcd/work/issues/ when the word stands bare. Four further ledgers exist (provenance, voyage, sources, and the local framing side), and each is named in full because none of them is the issue ledger.
aliases: ["issue ledger"]
forbidden_synonyms: ["database", "log file", "queue", "tracker"]
status: stable
introduced_in: itd-4
starts_when: null
ends_when: null
not_to_be_confused_with: core/record
versions: null
---

# ledger

A **ledger** is a store a command writes on the user's behalf and a human reads back later.
Bare, the word means the **issue ledger**: the structured per-repo capture store at
[`.abcd/work/issues/`](../../../../work/issues/README.md), working-tier data by
[adr-32](../../../decisions/adrs/0032-issue-ledger-is-working-tier-data.md), with
`open/`, `resolved/` and `wontfix/` as its buckets.

## Senses

| Sense | The one spelling | Where it lives |
|---|---|---|
| The captured-issue store | **the issue ledger**, or bare "the ledger" | [`.abcd/work/issues/`](../../../../work/issues/README.md), via [`commands/capture.md`](../../../../../commands/capture.md) |
| Origin, grounds and disposition stamped at the point of commitment | **the provenance ledger** | [`roadmap/phases/phase-7-ledger.md`](../../../roadmap/phases/phase-7-ledger.md) |
| The append-only record of every disembark and embark run | **the voyage ledger** | [voyage](voyage.md), `~/.abcd/voyage/<source-root-sha>/` |
| Source-to-decision provenance for the local sources corpus | **the sources ledger** | [`04-surfaces/13-consult.md`](../../04-surfaces/13-consult.md) |
| The uncommitted side where framing traces stay | **the local ledger side** | [adr-50](../../../decisions/adrs/0050-framing-traces-never-enter-the-record.md), [adr-55](../../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md) |

**"Ledger content" in the reading chapter means the second sense.** The cold-reading surface
says that nothing in the repository's tiering prevents a reading reaching ledger content — the
material the provenance ledger holds, which a reading must not see. Read as the issue ledger,
that sentence says something else entirely, so the qualifier carries the whole claim.

**Where the record disagrees.** The fifth sense is a side, not a store: adr-50 and adr-55 use
"the local ledger side" for where framing traces live, which is a place no `ledger` command
writes to. The glossary keeps the phrase because both ADRs are ratified, and marks it as the
one sense that names no store.

## When to use

Use bare "the ledger" only where the issue ledger is unambiguous — inside `capture` surface
prose, or beside a `iss-N` id. Name the ledger in full everywhere else.

## When NOT to use

Do not use "ledger" for the append-only decision log `.abcd/work/DECISIONS.md`, which the record
calls a decision log. Do not use it for the native transcript store or the memory substrate:
both are stores, neither is append-or-move over identified records.

## Examples

- "A half-formed observation goes to `/abcd:capture`; the issue ledger is where it lands."
- "The provenance ledger adds conjecture granularity as a grounds argument on the selection surfaces."
- "Writing to the destination passes the safety gate, and the run is appended to the voyage ledger."

## Related terms

- [record](record.md) — what a ledger holds
- [voyage](voyage.md) — the operations namespace the voyage ledger lives in
- [construal](construal.md) — the framing statement whose history stays on the local ledger side
- [reading-position](reading-position.md) — what "ledger content" is withheld from
