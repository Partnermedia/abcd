---
schema_version: 1
id: "iss-277"
slug: "ruleset-drift-check-is-unbuilt-the-applied-branch-rulesets-a"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
---

Ruleset drift check is unbuilt: the applied branch rulesets are mirrored under .abcd/work/rulesets/ but nothing diffs the live rulesets against the committed JSON, so drift is invisible until a human re-reads the console; itd-92 owns the verb