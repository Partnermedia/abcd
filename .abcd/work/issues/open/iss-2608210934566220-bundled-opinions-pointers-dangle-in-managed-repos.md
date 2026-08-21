---
schema_version: 1
id: "iss-2608210934566220"
slug: "bundled-opinions-pointers-dangle-in-managed-repos"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "memory-graduation principle work"
---

The six bundled OPINIONS rules each end with a pointer to a file under .abcd/development/principles/ that exists only in the abcd repo itself — a managed repo inherits the injected lines verbatim, so every prompt whose recall matches the domain hands its agent six dangling references (the principles corpus is not part of adoption). Either the bundled lines need self-contained phrasings with the pointer marked as abcd-repo-only, or adoption (prepare-this-repo / ahoy) should ship a distilled principles set the pointers can resolve against. Found while adding the memory-graduation rule, whose line was written self-contained for exactly this reason