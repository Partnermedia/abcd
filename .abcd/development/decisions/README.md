# abcd Decision Records

Retrospective decision records for the abcd plugin.

abcd uses two decision-record surfaces, each with a distinct job:

| Class | Direction | Surface | Use when |
|---|---|---|---|
| **ADR** (Architecture Decision Record) | Retrospective | [`adrs/`](adrs) | A decision is settled. Records the *why* + alternatives rejected. |
| **RFC** (Request for Comments) | Prospective | [`../roadmap/rfcs/`](../roadmap/rfcs) | A decision is contested. Discussion is the deliverable. |

Plus a third roadmap surface — **intents** ([`../intents/`](../intents)) — capture user-facing capability (forward-looking, press-release-shaped). Intents drive *what to build*; RFCs sharpen *whether to build it*; ADRs record *why we built it the way we did*.

The three surfaces together form the project's decision provenance. See [`adrs/README.md`](adrs/README.md) for ADR format + lifecycle.
