---
schema_version: 1
id: "iss-147"
slug: "guard-load-reads-abcd-guard-json-from-the-working-tree-so-a"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

guard.Load reads .abcd/guard.json from the WORKING TREE, so a disabled:true (or a retiered blocker) takes effect on the very next command, before anyone reviews it — spc-16 and the shipped docs both say the only escape is a committed, reviewable override, and nothing enforces the committed half. Reachable in one move: an agent writes .abcd/guard.json and the guard itself allows that write. Mitigated at the front door in itd-103 wiring (a disabled registry now warns UNGUARDED on every command, and abcd ahoy reports OFF) but not enforced. Proper fix is core-side: refuse a disabled:true that is not in HEAD, or drop the committed claim from spc-16. Found by the security reviewer on the itd-103 wiring branch.