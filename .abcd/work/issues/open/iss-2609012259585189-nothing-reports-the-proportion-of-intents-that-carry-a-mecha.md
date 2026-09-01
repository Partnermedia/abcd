---
schema_version: 1
id: "iss-2609012259585189"
slug: "nothing-reports-the-proportion-of-intents-that-carry-a-mecha"
severity: "minor"
category: "observation"
source: "agent-finding"
found_during: "iteration-2 conformance audit against the design framework v4 and the readings companion v4, 2026-09-01"
origin: researcher-authored
production_mode: dictated-and-formatted
found_at: "internal/core/reading"
---

Nothing reports the proportion of intents that carry a mechanism claim alongside the entailment reading's findings

The readings companion (v4, section 6.6) bounds the entailment reading: the
causal claim is prompted and nullable under the recording gradient, so where
`## Mechanism` is null the reading has nothing on the causal side to work from,
and "the yield of this position depends on how many intents carry a mechanism.
The proportion should be reported alongside the findings."

Nothing computes that proportion. The assembler's size report counts bytes and
estimated tokens per material kind; the manifest enumerates items by path and
field; the ingest verb validates what came back. None of them says how many of
the projected intents carried a `## Mechanism` section, how many carried the
`None stated.` nullity, and how many carried neither. Today the answer for the
cold-reading workstream's own fourteen shipped intents is zero, zero and
fourteen, which is exactly the bound the companion wants stated beside the
findings rather than discovered by a reader afterwards.

The natural home is the assembly result at the entailment position, where the
intent projection already knows which fields each file contributed: a field a
file does not carry contributes no item, so the count is already implicit in the
manifest and only needs stating. The number belongs in the size report's
rendering and in the manifest, so it is on the record for the run rather than
recomputed at write-up.
