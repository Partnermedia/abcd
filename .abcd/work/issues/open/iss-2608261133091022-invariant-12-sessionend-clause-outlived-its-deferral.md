---
schema_version: 1
id: "iss-2608261133091022"
slug: "invariant-12-sessionend-clause-outlived-its-deferral"
severity: "minor"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: ".abcd/development/brief/02-constraints/03-invariants.md:37"
---

invariant 12 and spc-35 still withhold the SessionEnd no-network clause pending iss-2608210934566223, which resolved with a pinning test, so the register under-asserts a property the tree holds