---
schema_version: 1
id: "iss-2608221342506046"
slug: "abcd-site-build-reuses-a-non-empty-output-directory-without"
severity: "minor"
category: "bug"
source: "user-observation"
found_during: "agent-finding"
found_at: "internal/core/site"
---

abcd site build reuses a non-empty output directory without purging, so files from an earlier build survive into the new tree; a purge needs a build-marker guard so it can never remove a directory the build did not write