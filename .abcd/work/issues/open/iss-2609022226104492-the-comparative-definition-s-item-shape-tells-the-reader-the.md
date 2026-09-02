---
schema_version: 1
id: "iss-2609022226104492"
slug: "the-comparative-definition-s-item-shape-tells-the-reader-the"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "Iteration 2 opening run, comparative ingest"
origin: researcher-authored
production_mode: hand-written
found_at: "agents/cold-reading-comparative.md"
---

The comparative definition's item shape tells the reader the criterion field is the declared criterion quoted from the passed material, and the criteria discipline declares each criterion as its name, a dash and a gloss on one line, so a reader quoting the passed material returns the whole line; the ingest's undeclared-criterion check accepts the declared name alone, as the companion's section 7.4 has it (the criterion as declared; the gloss is what it means in this project). The two halves of the instrument disagree and the first real comparative run of Iteration 2 was refused whole on it. The definition's wording is the half to move, in a PATCH of its prompt_version; until it moves, the host's reader prompt states the field's form.
