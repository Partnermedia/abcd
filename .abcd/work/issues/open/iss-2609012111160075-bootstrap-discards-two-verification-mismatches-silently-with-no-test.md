---
schema_version: 1
id: "iss-2609012111160075"
slug: "bootstrap-discards-two-verification-mismatches-silently-with-no-test"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "ship-audit-itd-130-itd-132-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "hooks/bootstrap.sh"
---

itd-132 ac-5 promises that every promotion of a cached artefact re-verifies against its recorded binary_sha256 and 'refuses loudly' on a mismatch. Two branches discard silently instead: the migration-seed mismatch is ignored by spec design (spc-35 line 145) with no notice, and the bootstrap's PATH-refresh branch skips a copy whose hash no longer matches the recorded provenance with no notice and no test covering that branch. Both are refusals in effect, neither is loud, and a user whose PATH copy quietly stopped being refreshed has no signal until abcd update or ahoy tells them it is foreign. Surfaced by the itd-132 fidelity audit (receipt rcp-acde3e9ce729). Loud-staging says a stage that no-ops must say so.
