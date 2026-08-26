---
schema_version: 1
id: "iss-2608261133097575"
slug: "phase-audit-tooling-claimed-shipped-in-go"
severity: "minor"
category: "documentation"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: ".abcd/development/roadmap/README.md:93"
---

roadmap prose and adr-9 claim the phase-audit reviewer and PA001 lint exist and run as Go tooling; nothing in the Go tree implements either, and sibling records correctly call them a design target from the predecessor store spc-66