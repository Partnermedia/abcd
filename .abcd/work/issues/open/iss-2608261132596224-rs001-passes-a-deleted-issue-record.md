---
schema_version: 1
id: "iss-2608261132596224"
slug: "rs001-passes-a-deleted-issue-record"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "bughunt-round-8"
found_at: "scripts/check-issue-resolution.sh:77"
---

check-issue-resolution RS001 counts a bare deletion from open as a resolution: the unconditional D branch unions with ids_entering_closed, so a trailer plus git rm passes the gate with no terminal landing