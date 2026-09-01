---
schema_version: 1
id: "iss-2608311518250688"
slug: "three-smaller-defects-in-the-ingest-verb-each-cheap-and-each"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "itd-185 fidelity audit rcp-fe3450ca55ff"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/reading/ingest.go"
resolution: "The definition-drift refusal routes through refuse(), the one writer of a refusal record, so a run refused after its identity is proven for a definition that did not resolve leaves refusal.json with no regime and the definition named; the refusal list renders under one rule on both surfaces (ItemRefusal.Render) and a zero ordinal is omitted, so no durable record carries item 0; and U+2800 BRAILLE PATTERN BLANK is folded away by the package's one invisible-rune predicate, which now also judges the instrument's three parts."
impact: fix
resolved_by:
  commit: "82fc295e2886fd398bef03c398b33f432207484b"
---

Three smaller defects in the ingest verb, each cheap and each an instance of a class this workstream has already recorded. A definition-drift refusal returns bare after instrument identity is proven, so the run is refused and no refusal record is written, contradicting both the criterion that says a list-level refusal writes one and the plugin surface's own corrected sentence. The refusal-list elision entry reaches the DURABLE record carrying ordinal zero, so a refused run writes item 0 into its refusal record although there is no item 0; the terminal renderer has the branch that suppresses it and the record writer does not. And the braille pattern blank is accepted as a provenance value: it renders as nothing but is neither a format rune nor a default-ignorable nor a variation selector nor whitespace, so it clears the blankness test. That last one does not breach the criterion's letter, which speaks of an empty or absent field, but it does breach the residue's claim that every invisible rune is folded away.

## Grounds

- pursued: each of the three defects was a small contained fix on the same verb, each reproduced by a failing test asserting the durable record's content before the change and passing after; the sibling sweep found the instrument's blank check on bare TrimSpace and closed it on the same predicate
