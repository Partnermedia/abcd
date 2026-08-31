---
schema_version: 1
id: "iss-2608311331229368"
slug: "spc-65-and-itd-187-both-prescribe-watching-testassembledinpu"
severity: "minor"
category: "drift"
source: "agent-finding"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

spc-65 and itd-187 both prescribe watching TestAssembledInputIsByteIdenticalAcrossRuns red by removing the assembler's walk sort. Run on a copy of the tree, that mutation does not make it red: removing the sort yields a different but still deterministic order, so two assemblies of one commit stay byte-identical and only the order oracle fires. The mutations that do make the identity assertion red are a genuinely nondeterministic walk (candidates rebuilt from a map) and an absolute repository path embedded in the bundle. The record's prescribed hand-run for ac-1 names a mutation that cannot falsify ac-1.
