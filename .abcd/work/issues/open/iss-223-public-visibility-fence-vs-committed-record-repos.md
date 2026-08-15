---
schema_version: 1
id: "iss-223"
slug: "public-visibility-fence-vs-committed-record-repos"
severity: "minor"
category: "inconsistency"
source: "manual-test"
found_during: "manual-capture"
found_at: "internal/core/ahoy/gitignore.go"
related_issues: ["iss-169"]
---

the managed .gitignore visibility table has no mode for a public repo that commits its record: on abcd-cli (public, single-repo-curated-release, .abcd/** deliberately in-tree) ahoy install applied the public policy and fenced /.abcd/ and /memory/, contradicting the repo's own boundary that the record is present in every checkout. The visibility table needs a committed-record declaration (config or marker) that suppresses the fence