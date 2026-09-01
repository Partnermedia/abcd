---
schema_version: 1
id: "iss-2609012029342529"
slug: "ghsa-rvhr-3455-c5jw-cwe-359-the-issue-ledger-identity-redact"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/redact.go"
---

GHSA-rvhr-3455-c5jw (CWE-359): the issue-ledger identity redaction keys on the repo effective git identity and commits the caller other identity in clear text. scanner.ProbeIdentity runs git config --get for user.name and user.email under -C repoRoot, the single last-wins value, so a repo-local persona or a global includeIf persona replaces the global identity in the matcher set; a capture naming both identities redacts the persona and commits the personal address and name, and the post-redaction slug carries the mailbox into the filename while the CLI notice says identifiers are never committed. Evidence: scanner.ProbeIdentity, capture.redactLedgerText, workflow.go deriveSlug. The fix must union every scope git resolves (git config --get-all, includes evaluated) so a persona adds an identity and never displaces one; body and filename must carry neither identity.
