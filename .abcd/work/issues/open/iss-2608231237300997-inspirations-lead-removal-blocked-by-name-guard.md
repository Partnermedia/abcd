---
schema_version: 1
id: "iss-2608231237300997"
slug: "inspirations-lead-removal-blocked-by-name-guard"
severity: "minor"
category: "documentation"
source: "user-observation"
found_during: "user-observation"
found_at: "ACKNOWLEDGEMENTS.md"
---

The Inspirations lead sentence ('Ideas and methodologies that shaped the design — not code abcd depends on.') is removed from ACKNOWLEDGEMENTS.md at the maintainer's request, but the edit CANNOT BE COMMITTED on this machine: the private name-guard's entry-14 blocks every commit touching that file, which NEXT.md records as needing the maintainer to add '# abcd-banlist: keyed' to the private names file and anchor that entry. The edit sits uncommitted in the working tree. The guard was not weakened and no workaround was attempted.