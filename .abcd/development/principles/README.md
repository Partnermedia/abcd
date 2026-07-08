# principles/

Distilled cross-cutting design principles — the rules that hold across the whole
system (e.g. "transport-agnostic core", "wired or it isn't done", "host-delegated by
default"). A first-class abcd artefact: the lifeboat packs *decisions, principles,
pitfalls, and the spine*, so principles live here — distinct from `../decisions/`
(ADRs: the ratified *why we chose*) and `../intents/` (the user-facing *why it
matters*).

One principle per file. Populated during the Phase 0.5 content reconciliation.

**Promotion path.** A principle here holds a value or convention *without* an
enforcement hook. The moment a principle gains a mechanical gate (a lint code,
a hook, a CI check), it is promoted to a **discipline-kind intent** — the
lifecycle'd, spec-inherited form (see [`../intents/disciplines/`](../intents/disciplines)).
Enforced principle ⇒ discipline; this directory is the not-yet-enforced layer.
