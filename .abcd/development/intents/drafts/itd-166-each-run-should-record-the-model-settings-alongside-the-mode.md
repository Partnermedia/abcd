---
id: itd-166
slug: each-run-should-record-the-model-settings-alongside-the-mode
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
severity: minor
promoted_from: iss-2608290811463906
---

# Each run should record the model SETTINGS alongside the model id, because the id alone does not determine behaviour and so does not make a run interpretable or reproducible. Whatever the current model exposes belongs in the record: reasoning or thinking depth, temperature and other sampling parameters, a seed where one exists, the context-window variant, and any speed or quality mode the harness offers. Two runs stamped with the same model id can differ substantially on all of these, so a record that names only the id invites a false comparison between them: an audit verdict, a benchmark, or a regression blamed on a code change may in fact be a settings difference. This matters most where a verdict is treated as evidence. The intent-audit verdict already carries a verifier block of id and version, which is the natural place to widen; the per-run transcript store is the natural home for the settings of an ordinary session; and the Assisted-by trailer is NOT the place, since its accepted shape is a vendor and model pair and widening it would break the gate that reads it. The hard constraint is the same one iss-2608290810032799 records: a model cannot reliably introspect the settings it is running under, so the harness has to supply them. That makes this a capability of the dispatching layer rather than something an agent can be instructed to self-report, and any design that asks the agent to state its own settings inherits exactly the truthfulness problem that the attribution gate already has.

## Press Release

> _Seeded by promotion from iss-2608290811463906. Expand into the full press-release narrative before planning._

## Why This Matters

Graduated from `iss-2608290811463906`: Each run should record the model SETTINGS alongside the model id, because the id alone does not determine behaviour and so does not make a run interpretable or reproducible. Whatever the current model exposes belongs in the record: reasoning or thinking depth, temperature and other sampling parameters, a seed where one exists, the context-window variant, and any speed or quality mode the harness offers. Two runs stamped with the same model id can differ substantially on all of these, so a record that names only the id invites a false comparison between them: an audit verdict, a benchmark, or a regression blamed on a code change may in fact be a settings difference. This matters most where a verdict is treated as evidence. The intent-audit verdict already carries a verifier block of id and version, which is the natural place to widen; the per-run transcript store is the natural home for the settings of an ordinary session; and the Assisted-by trailer is NOT the place, since its accepted shape is a vendor and model pair and widening it would break the gate that reads it. The hard constraint is the same one iss-2608290810032799 records: a model cannot reliably introspect the settings it is running under, so the harness has to supply them. That makes this a capability of the dispatching layer rather than something an agent can be instructed to self-report, and any design that asks the agent to state its own settings inherits exactly the truthfulness problem that the attribution gate already has.. Read that issue record for the source observation.

## Acceptance Criteria

> _Required (the itd-1 discipline): add at least one Given-When-Then bullet describing the verifiable bar for "shipped" before this draft can be planned._

## Open Questions

_None recorded yet._

## Audit Notes

_Empty. Populated by intent-auditor when intent moves to shipped/._
