# principles/

Distilled cross-cutting design principles — the rules that hold across the whole
system (e.g. "transport-agnostic core", "wired or it isn't done", "host-delegated by
default"). A first-class abcd artefact: the lifeboat packs *decisions, principles,
pitfalls, and the spine*, so principles live here — distinct from `../decisions/`
(ADRs: the ratified *why we chose*) and `../intents/` (the user-facing *why it
matters*).

One principle per file. Populated during the Phase 0.5 content reconciliation.

**Promotion path.** The full ladder has three rungs: a **principle** — the
normative statement (a value, a definition of proven/done/good, or a rule of
action) that survives any particular mechanism; beneath it an **enabling
convention, script, or file format** — the MVP, the smallest unenforced
enabler; and above it a **discipline-kind intent or core absorption** — the
tool, which makes the practice enforced or cheap at the price of becoming a
maintained artefact with its own lifecycle (false-positive budget, gaming
surface, saturation, kill criterion). An unenforced convention remains a
principle-layer artefact in this record; "MVP" names the enabling artefact
beneath a principle, not a third governance category. The moment a principle
gains a mechanical gate (a lint code, a hook, a CI check), it is promoted to
a discipline-kind intent — the lifecycle'd, spec-inherited form (see
[`../intents/disciplines/`](../intents/disciplines)); enforced principle ⇒
discipline, and this directory is the not-yet-enforced layer. The ladder also
runs downward: a tool demotes to advisory on stale calibration and archives
to regression-only on saturation.

**Intake.** A candidate's layer is its *entry rung* — repo-relative and
dated: the rung where the actionable delta sits given the current record,
with every rung marked exists/partial/absent. Evidence artefacts (acceptance
corpora, calibration fixtures, baselines) are attachments to a rung, never
rungs. Record-shaped work may declare a degenerate ladder (principle = MVP,
or topping out at MVP). Two rules: articulate the full ladder for every
candidate, and never fabricate an absent rung. Provenance:
[`../research/notes/2026-07-09-practice-mvp-tool-extraction.md`](../research/notes/2026-07-09-practice-mvp-tool-extraction.md).
