---
schema_version: 1
id: "iss-2608210923436594"
slug: "local-agent-memory-masks-portable-convention-gaps"
severity: "minor"
category: "architectural-insight"
source: "user-observation"
found_during: "memory-portability audit"
---

User-local agent memory is an unarmed, non-portable convention store: several working conventions (attribution footer handling, cloud-routine git identity, citation integrity) live only in one developer's per-project memory, so abcd works here partly because that memory compensates — and none of it travels to other users' managed repos. abcd already owns the portable homes (the bundled rules domains, .abcd/development/principles/, the per-project .abcd/memory/ substrate, validators); what is missing is the deliberate graduation loop: a convention that appears twice in local memory is a candidate for a rules-domain line, a principle, or a validator, and nothing currently prompts that promotion. Same shape as the itd-114 lesson — safety must live in the tool, not in every minter's prompt or any one user's memory