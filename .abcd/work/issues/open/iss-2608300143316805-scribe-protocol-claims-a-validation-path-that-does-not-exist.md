---
schema_version: 1
id: "iss-2608300143316805"
slug: "scribe-protocol-claims-a-validation-path-that-does-not-exist"
severity: "major"
category: "inconsistency"
source: "impl-review"
found_during: "itd-188 adversarial review, 2026-08-30"
found_at: ".abcd/development/brief/05-internals/01-agents.md, agents/scribe.md, agents/CHANGELOG.md"
---

The scribe protocol, the definition's Delivery section, and the changelog entry claim record_schema judges a scribe-emitted reading or disposition record the moment it is committed and that the validation path exists today; on this base no reading or disposition schema or store exists (spc-58 is open), so record_schema refuses the directory as an undeclared bucket and a malformed record is indistinguishable from a well-formed one. The dependency on spc-58 must be stated in all three places.
