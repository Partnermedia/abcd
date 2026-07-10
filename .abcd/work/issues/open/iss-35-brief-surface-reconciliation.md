---
schema_version: 1
id: "iss-35"
slug: "brief-surface-reconciliation"
severity: "critical"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/brief/05-internals/08-skills.md"
---

brief-vs-shipped-surface reconciliation: 05-internals/08-skills.md claims abcd ships zero user-facing skills and six top-level commands while 04-surfaces/README.md itself tables nine and /abcd:consult and /abcd:ingest are shipped; the shipped skills violate the brief criterion that any artefact mutation is a command, not a skill; the skills/ layout described (abcd-ahoy, commit-attribution, secrets-and-pii) is fictional vs the real consult/ and ingest/; the implemented, user-reachable abcd docs lint and abcd history verbs have no home in 04-surfaces at all; the operator-internal paragraph contradicts the commands/ directory that exists. Detector (per spec-moves-with-the-surface): a record-lint cross-check that every entry under commands/ and skills/ resolves to a brief surface row, and every brief surface row resolves to a shipped or explicitly staged surface. Acceptance corpus: each falsified claim above — the check fails on all of them today. Fix amends the criterion or the surface in one change, never silently.
---

**Detector run 1 (2026-07-10, autonomous run, workflow MVP per
script-first):** the bidirectional cross-check ran as 22 checker agents
(11 brief docs verified against the shipped binary, 11 real surfaces
seeking their brief home). **150 unique discrepancies**: false-claim 77,
fictional-layout 29, undocumented-surface 24, criterion-violation 15,
stale-count 5. Every falsified claim enumerated above reproduced, plus:
no `abcd init`/`config`/`run` verbs exist though the brief cites them
unmarked; launch's brief row claims artifact-cutting while the binary is
read-only preview; the bare-command-as-help "universal convention" is met
by no shipped verb; `docs` and `history` verbs have no brief home. Worst
docs: 05-intent (17), 07-memory (13), 01-ahoy/04-launch/06-capture/
08-abcd/08-skills (10 each). Full corpus: the run's local log
(iss35-discrepancies.json) — re-derivable by re-running the workflow.
Next: reconcile per doc behind this detector (amend criterion or surface,
never silently), then graduate the check to a record-lint rule
(spec-moves-with-the-surface).
