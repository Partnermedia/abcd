---
schema_version: 1
id: "iss-40"
slug: "glossary-unification"
severity: "major"
category: "inconsistency"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/development/brief/glossary"
---

one canonical glossary: 02-constraints/04-naming.md maintains a term registry that competes with the brief glossary/ directory; glossary/ is itself stale (contexts and terms missing from its own index, a validation command that does not exist) and is invisible to the brief — neither README.md nor 00-meta.md mention it. Detector: a single-registry rule enforced by a glossary lint — one canonical home, terms in the record resolve to it, the index derived rather than hand-kept, and a banlist entry for the retired registry location once merged. Acceptance corpus: the competing registry, the missing index entries, and the phantom validation command.