# One canonical primitive

**The rule.** Infrastructure primitives — durable write, frontmatter parse,
path expansion, directory validation — have exactly one home
(`internal/fsutil` and its peers). A needed variant is an option on the
canonical implementation, never a local copy. A second copy is a flagged
consolidation target; a third is forbidden.

**Why.** Copies do not merely duplicate — they silently *weaken*. Of the four
atomic-write implementations the 2026-07-08 review found in the tree, two omit
the parent-directory fsync that makes the rename crash-safe, while their doc
comments still claim full durability. Each copy also forks the bug-fix
surface: a durability fix landed in one home reaches every caller; landed in
one of four, it reaches a quarter of them. A sibling project pairs this rule
with a canonical-primitives table and a "what NOT to build" section in every
plan, so the canonical home is discoverable at the moment of temptation.

**Bounds.**

- The rule covers infrastructure primitives, not domain logic: two verbs may
  legitimately interpret the same data differently, but they may not each own
  a rename-into-place.
- Mode preservation, exclusive-create naming, and similar needs are options on
  the canonical primitive — evidence for extending it, not for forking it.

**Live instance.** `internal/fsutil`'s package doc already flags the ahoy
copies as the consolidation target; the capture and memory copies extend that
target.

**Promotion.** A lint or vet check that flags private redefinitions of the
canonical names (`writeFileAtomic`, `isRealDir`) would make this a discipline.
