---
schema_version: 1
id: "iss-40"
slug: "glossary-unification"
severity: "major"
category: "inconsistency"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/brief/glossary"
resolution: "One canonical glossary: brief/glossary/ keeps the name and 04-naming.md stops claiming it, keeping its reserved-vocabulary table as the distinct enum registry it is. The glossary index and directory tree are rendered from the term files by internal/core/glossary and gated by a drift test, so the missing distribution/ context cannot recur. The phantom terminology schema, abcd lint terminology verb, and terminology_exclude_files key are replaced by what ships (GL002 via abcd docs lint, plus the drift gate). core/brief.md matches adr-5, brief/README.md lists glossary/, and a public banlist entry blocks the retired phrase from the published surface."
impact: fix
---

one canonical glossary: 02-constraints/04-naming.md maintains a term registry that competes with the brief glossary/ directory; glossary/ is itself stale (contexts and terms missing from its own index, a validation command that does not exist) and is invisible to the brief — neither README.md nor 00-meta.md mention it. Detector: a single-registry rule enforced by a glossary lint — one canonical home, terms in the record resolve to it, the index derived rather than hand-kept, and a banlist entry for the retired registry location once merged. Acceptance corpus: the competing registry, the missing index entries, and the phantom validation command.