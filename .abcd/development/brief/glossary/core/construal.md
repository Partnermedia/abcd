<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: construal
bounded_context: core
definition: What a situation is treated as — the frame a piece of work reasons inside, stated in one or two sentences at the top of the brief's framing chapter. The construal as it presently stands is committed record; the history of how it came to stand is not.
aliases: ["construal statement", "framing statement"]
forbidden_synonyms: ["opinion", "assumption", "hypothesis", "framing trace"]
status: stable
introduced_in: adr-55
starts_when: null
ends_when: null
not_to_be_confused_with: core/brief
versions: null
---

# construal

A **construal** says what the situation is being treated *as*. It is the macro-why: the frame
the work reasons inside, written where a reader finds it rather than reconstructed from scope
notes. The brief's framing chapter opens with one, under a `## Construal` heading, in a
sentence or two.

The word carries **one sense**, and it earns a glossary entry because it is coined in the
framing chapter and used freely outside it — in the ADRs, in the reading definitions, and in
this glossary's own README, where curating a term is called a framing act because it changes
how a later reader construes the record (iss-2609012245352480).

## The split that makes it usable

[adr-55](../../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md)
divides the word from its history, and both rules are unconditional:

- **The construal as it presently stands is committed record** — the framing section's
  statement, the glossary's committed terms, committed scope and vocabulary. A widening
  reading reads against that statement and nothing behind it.
- **Its history is not.** Declined construals, superseded terms and the reasoning that settled a
  terminological dispute stay on the local ledger side, which no automated reviewer reads
  ([adr-50](../../../decisions/adrs/0050-framing-traces-never-enter-the-record.md)).

## When to use

Use "construal" for the stated frame — what a piece of work treats the situation as. Use it
when the question is *why this shape of answer* rather than *what was built*.

## When NOT to use

Do not use it for a mechanism claim ("we expect X because Y", recorded on an intent) or for
grounds ("why this is being pursued now", recorded at the readiness gate). Those are
falsifiable claims about a specific piece of work; a construal is the frame they sit inside. Do
not use it for a framing trace — the deliberation that produced the construal, which the record
deliberately does not hold.

## Examples

- The framing chapter's own construal: the ledger work is treated as a gap in what the repository records, so it adds the missing record rather than restricting what agents may do.
- "The glossary is a deliberate frame surface: what the project chose to name is the part of its construal a machine can read."

## Related terms

- [reading-position](reading-position.md) — the widening position reads against the construal as it stands
- [ledger](ledger.md) — the local ledger side, where the construal's history stays
- [brief](brief.md) — the document whose framing chapter states the construal
