---
schema_version: 1
id: "iss-2608300931349990"
slug: "provenance-plugin-pages-over-claim"
severity: "minor"
category: "documentation"
source: "impl-review"
found_during: "itd-178 adversarial ruthless review, 2026-08-30"
found_at: "commands/capture.md, commands/intent.md, internal/README.md"
---

itd-178 prose findings: the plugin pages say every record a command writes carries the two keys, but capture disposition writes a dsp record with neither (narrow to intent, spec and issue records, as spc-56 scoped); the same pages assert the keys are excluded from every reading by the input assembler's field projection, describing code not in this tree (the sentence lives in spc-56's out-of-scope list — delete it from the user-facing surface); the internal package map omits the new provenance leaf while listing its sibling leaves.
