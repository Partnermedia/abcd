---
schema_version: 1
id: "iss-38"
slug: "hand-maintained-index-drift"
severity: "major"
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/README.md"
resolution: "Three of the four named indexes were real drift and are fixed; the fourth (repo README omits skills/) does not hold — there is no skills/ directory, abcd ships zero skills by design (brief 05-internals/08-skills.md, iss-61), so the Layout section is correct as it stands and was left untouched. intents/README.md's four ASCII trees give way to prose plus links to the directories that ARE the listing (adr-5). commands/README.md and internal/README.md keep their enumerations, corrected, and ship with the detector: the new index_drift lint rule holds a marked region to the directory it enumerates (exact mode) or to those paths still being absent from the tree (absent mode, the planned-seams shape). The rule reaches users through the same engine abcd docs lint runs, so any abcd-managed repo can gate its own indexes by declaring them in .abcd/docs-lint.json."
impact: additive
---

hand-maintained index drift: intents/README.md corpus listings have drifted from the filesystem; commands/README.md lists three of the seven plugin verb files; internal/README.md still describes core as two capabilities with adapter/scanner as a planned seam; the repo README Layout omits skills/ from the plugin surface. Detector (per adr-5 derive-dont-store): hand enumerations of sibling files are generated or deleted — a lint flags a README that enumerates directory contents by hand, or the enumeration is emitted by tooling with a drift test. Acceptance corpus: the four stale indexes above.