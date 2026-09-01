---
schema_version: 1
id: "iss-2609012029335717"
slug: "ghsa-5qr6-f78x-g2cx-cwe-312-the-memory-store-redacts-only-th"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/redact.go"
---

GHSA-5qr6-f78x-g2cx (CWE-312): the memory store redacts only the PEM BEGIN header and stores the key body. memory.storeRedactor.redactText scans with the header-only pem_private_key pattern, scanner.Redact masks that one span, and the stage-two BlockingResidual rescan re-runs the same rule over a fingerprinted header and is clean by construction; a text source with a key block lands in the page body and in the --keep-original copy (redactOriginalBytes) with body and END line verbatim. Evidence: memory/redact.go redactText and redactOriginalBytes. The fix is the shared block consumer in scanner.Redact so memory, history and capture cannot drift; the page body and the kept original must carry no body line.
