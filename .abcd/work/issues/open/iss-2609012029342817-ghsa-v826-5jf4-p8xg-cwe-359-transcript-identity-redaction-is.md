---
schema_version: 1
id: "iss-2609012029342817"
slug: "ghsa-v826-5jf4-p8xg-cwe-359-transcript-identity-redaction-is"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/history/history.go"
---

GHSA-v826-5jf4-p8xg (CWE-359): transcript identity redaction is keyed on the repo effective git identity, storing the caller other git identity in clear text. history.Capture uses scanner.New(repoRoot), whose ProbeIdentity resolves one user.name and user.email under -C repoRoot with global config in effect (gitutil.ScrubbedEnv); git last-wins resolution means a repo-local or includeIf persona replaces the unconditional global identity, so a transcript carrying both stores the personal one verbatim with the frontmatter reporting no identity redaction. Evidence: scanner.ProbeIdentity, history.Capture. The fix must union the effective identity with every other value git resolves for the key (git config --get-all), keeping the persona redacted and redacting the global identity too.
