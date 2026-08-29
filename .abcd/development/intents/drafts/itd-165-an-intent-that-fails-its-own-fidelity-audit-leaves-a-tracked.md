---
id: itd-165
slug: an-intent-that-fails-its-own-fidelity-audit-leaves-a-tracked
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
---

# An intent that fails its own fidelity audit leaves a tracked issue, not a paragraph. Today the audit ingest counts a NOT_MET verdict into a rollup, renders it into the shipped intent's Audit Notes, and stops: no capture, no exit code, no gate reads it, so an intent can ship with every acceptance criterion failed and the only consequence is prose nobody queries. This intent gives the verdict teeth without making the audit a blocker: on ingest, each NOT_MET auto-captures one ledger issue stamped with the receipt id and deduped on it, so a re-ingest cannot mint a second issue for the same failure, and the issue id is written back into the Audit Notes so the edge is two-sided like every other record link. An INCONCLUSIVE verdict deliberately does NOT capture, because it means the auditor was under-fed, which is an input bug rather than a product defect, and auto-capturing it at volume would fill the ledger with noise and train a reader to ignore captures. A baseline ratchet then converts the verdict from advisory to binding the way the dangling-reference gate already does: today's failures are baselined, a NEW failure fails the gate, and a fixed one invites a baseline shrink.

## Press Release

> _Seeded from a quoted-text intent capture. Expand into the full press-release narrative before planning._

## Why This Matters

An intent that fails its own fidelity audit leaves a tracked issue, not a paragraph. Today the audit ingest counts a NOT_MET verdict into a rollup, renders it into the shipped intent's Audit Notes, and stops: no capture, no exit code, no gate reads it, so an intent can ship with every acceptance criterion failed and the only consequence is prose nobody queries. This intent gives the verdict teeth without making the audit a blocker: on ingest, each NOT_MET auto-captures one ledger issue stamped with the receipt id and deduped on it, so a re-ingest cannot mint a second issue for the same failure, and the issue id is written back into the Audit Notes so the edge is two-sided like every other record link. An INCONCLUSIVE verdict deliberately does NOT capture, because it means the auditor was under-fed, which is an input bug rather than a product defect, and auto-capturing it at volume would fill the ledger with noise and train a reader to ignore captures. A baseline ratchet then converts the verdict from advisory to binding the way the dangling-reference gate already does: today's failures are baselined, a NEW failure fails the gate, and a fixed one invites a baseline shrink.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
