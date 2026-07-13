---
schema_version: 1
id: "iss-89"
slug: "capture-no-home-for-dogfood-defects"
severity: "minor"
category: "future-work-seed"
source: "agent-finding"
found_during: "2026-07-13 B1 dogfood: prepare-this-repo audit of Manuscripts"
found_at: "internal/core/capture"
---

abcd defects found while dogfooding a target repo have no ledger home: abcd capture writes only to the cwd repo .abcd/work/issues/, so capturing an abcd bug found while onboarding repo X would either pollute X ledger (wrong repo) or require re-running with abcd-cli as cwd. Surfaced during the Manuscripts B1 dogfood -- these seven captures had to be routed to abcd-cli by hand. Seed: a capture routing option that targets the abcd repo (or an upstream/--repo flag) for tool-defect captures. Acceptance: capturing an abcd defect from within a target repo lands it in abcd ledger without a manual cd.