---
schema_version: 1
id: "iss-2608311518109233"
slug: "all-three-reserved-name-tables-in-the-supply-regime-gate-are"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "itd-185 fidelity audit rcp-fe3450ca55ff"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/ingest_regime.go"
---

All three reserved-name tables in the supply-regime gate are mutation-vacuous: emptying any of them leaves the whole package green, because every test iterates the table under test rather than naming what it should contain. No test anywhere pins the literal field names that three acceptance criteria enumerate, so the criteria are satisfied by a gate that would still pass if the names it refuses were deleted. The signature registry has a floor asserting a minimum count; the reserved tables have none. The earlier vacuity sweep on this branch swept code BRANCHES and never data tables, which is why the shape was invisible to it: a guard whose behaviour is driven by a table is only as bound as the table is pinned. Also unpinned by the same reasoning: the per-position body field sets.
