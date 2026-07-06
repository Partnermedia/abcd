<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->
---
term: disembark
bounded_context: core
definition: The act of packing a lifeboat — the `/abcd:disembark` surface distils a project's settled artefacts, decisions, and configuration into a portable lifeboat directory that a fresh context can later unpack via `/abcd:embark`.
aliases: ["lifeboat packing", "disembarkation"]
forbidden_synonyms: ["export", "backup", "dump", "snapshot"]
status: stable
introduced_in: phase-1
starts_when: null
ends_when: null
not_to_be_confused_with: null
versions: null
---

# disembark

**Disembark** is the project-lifecycle surface that packs a [lifeboat](lifeboat.md): a
portable, highest-fidelity proxy of a project's theory that can be carried across a session,
machine, or team boundary. Bare `/abcd:disembark` shows status and help only; `to <path>`
performs the packing pass, while `probe` and `dry-run` inspect without writing.

## When to use

Use "disembark" when describing the act of packing a lifeboat — the outbound half of the
lifeboat lifecycle. It is the counterpart to [embark](../interview/embark.md)'s inbound
opening; together they bracket the portability boundary.

## When NOT to use

Do not use "export", "backup", or "snapshot" — these lose the navigational, theory-preserving
connotation. Disembark packs the *floor* the project can carry forward, not a byte-for-byte
copy.

## Examples

- "The disembark pass distilled the shipped intents and decision timeline into the lifeboat."
- "A `disembark probe` lists the sources that would be packed without writing anything."

## Related terms

- [lifeboat](lifeboat.md) — the artefact disembark packs
- [voyage](voyage.md) — the project journey a lifeboat preserves across a boundary
