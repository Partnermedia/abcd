# Decision notes

Short, narrowly-scoped decision artifacts that are NOT full ADRs.

An **ADR** ([`../adrs/`](../adrs)) records a settled architecture decision —
the *why* plus the alternatives rejected — for a choice that is reversible in
principle. A **note** here records something narrower and more operational:

- an **accepted residual** (a known fail-open / gap we have deliberately
  decided to live with) and its **precise blast-radius bound**, or
- a focused trade-off that does not rise to an architecture decision but must
  still be written down so it is not silently re-litigated.

A note is the durable, committed artifact for a residual whose primary working
record lives in the gitignored `.abcd/.work.local/` ledger. If a note's subject later hardens
into a reversible architecture choice, promote it to an ADR and link back.
