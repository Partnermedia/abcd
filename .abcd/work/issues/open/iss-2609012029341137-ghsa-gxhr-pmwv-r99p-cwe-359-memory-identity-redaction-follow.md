---
schema_version: 1
id: "iss-2609012029341137"
slug: "ghsa-gxhr-pmwv-r99p-cwe-359-memory-identity-redaction-follow"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/redact.go"
---

GHSA-gxhr-pmwv-r99p (CWE-359): memory identity redaction follows the repo git identity, not the caller. memory.newStoreRedactor builds scanner.New(repoRoot) whose ProbeIdentity reads the single effective user.name and user.email, so with a work persona in the repo the caller global identity is stored verbatim in the page body and the kept original. Evidence: memory/redact.go newStoreRedactor, scanner.ProbeIdentity. One fix in the probe (union of every configured scope) covers the memory, history and capture stores; neither identity may survive in any store write.
