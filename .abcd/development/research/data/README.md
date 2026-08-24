# Research data

Measured corpora kept as data: detector and checker output that grounds a
research note, an issue, or an intent, and that no other tier holds.

A semantic-gate receipt under [`work/reviews/`](../../../work/reviews) records
that a gate ran and what it decided. It is deliberately narrow — subject,
class, disposition — and it is not a copy of the evidence. When the evidence
itself is what a record rests on, and it would otherwise live only in
ephemeral local storage, it lands here instead.

What belongs here:

- Output of a pinned detector or checker, with its content commit, manifest
  hash and tier recorded in the payload.
- Corpora that a note or an intent cites as its empirical base.

What does not:

- Verdicts and dispositions — those are the receipt's, or the decision
  record's.
- Anything reproducible on demand from committed inputs. Re-running is cheaper
  than storing.

One directory per corpus, `<YYYY-MM-DD>-<scope>/`, each carrying a `README.md`
that states its provenance and what the payload does and does not claim.

| Path | What it holds |
|---|---|
| [`2026-08-23-brief-surface-crosscheck/`](2026-08-23-brief-surface-crosscheck) | Three full-tier `iss35-brief-surface-crosscheck` runs over the brief's surface chapters (itd-147) |
