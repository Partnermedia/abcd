---
schema_version: 1
id: "iss-2608230751434661"
slug: "site-redundant-gen-dateline-under-titles"
severity: "minor"
category: "ux"
source: "user-observation"
found_during: "user-observation"
found_at: "internal/core/site/explorer.go"
---

Explorer pages print a grey dateline (date · version · commit) under every page title, duplicating the header pill and footer stamp. Remove the gen line from the explorer shell (report 5).