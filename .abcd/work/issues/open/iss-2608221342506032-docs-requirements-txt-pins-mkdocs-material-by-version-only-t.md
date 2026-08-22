---
schema_version: 1
id: "iss-2608221342506032"
slug: "docs-requirements-txt-pins-mkdocs-material-by-version-only-t"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "agent-finding"
found_at: "docs/requirements.txt"
---

docs/requirements.txt pins mkdocs-material by version only; the deploy render job runs pip without hash-pinning, so a --require-hashes freeze is the remaining supply-chain hardening