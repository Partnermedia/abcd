---
schema_version: 1
id: "iss-192"
slug: "brief-surface-drift-v042-release-gate"
severity: "major"
category: "drift"
source: "drift-detection"
found_during: "v0.4.2-release-gate"
---

Brief surface chapters lag the shipped surface: the v0.4.2 release-gate crosscheck (full tier, 22 checkers) found 26 discrepancies across 01-ahoy, 04-launch, 05-intent, 07-memory, 15-prepare-this-repo, 16-audit and 05-internals/08-skills — banlist missing from the fourteen-command roster, hooks/bootstrap.sh absent from every hook enumeration, the shipped citation-baseline dry-run gate undocumented, ahoy's envelope/render enumerations missing banlist/guard-detail/citations fields, intent §6 attributing unshipped acceptance-criteria/bundle/supersession checks to the record lint, prepare-this-repo's stale commands/abcd/ layout and undocumented identity phase, audit's privacy-hygiene row understating network-identifier scanning, and memory's bare-render/schema/spec-range drift. Full finding list with evidence: the iss35-brief-surface-crosscheck receipt for the v0.4.2 content commit under .abcd/work/reviews/. Same residue class as iss-152 (v0.4.1).