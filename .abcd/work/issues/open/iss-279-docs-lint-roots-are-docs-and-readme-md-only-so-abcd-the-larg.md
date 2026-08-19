---
schema_version: 1
id: "iss-279"
slug: "docs-lint-roots-are-docs-and-readme-md-only-so-abcd-the-larg"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "manual-capture"
found_at: ".abcd/docs-lint.json"
---

docs-lint roots are docs/ and README.md only, so .abcd/** — the largest public surface — is scanned by no name gate; retire-the-name's banned_tokens cannot reach CONTRIBUTING.md, AGENTS.md, or scripts/ either