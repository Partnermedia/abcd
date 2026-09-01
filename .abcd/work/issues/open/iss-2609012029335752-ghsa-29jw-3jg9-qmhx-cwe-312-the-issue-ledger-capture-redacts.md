---
schema_version: 1
id: "iss-2609012029335752"
slug: "ghsa-29jw-3jg9-qmhx-cwe-312-the-issue-ledger-capture-redacts"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/capture/redact.go"
---

GHSA-29jw-3jg9-qmhx (CWE-312): the issue-ledger capture redacts only the PEM private-key BEGIN header and commits the key body. The bundled pem_private_key pattern (scanner.DefaultPatterns) matches the header line alone and scanner.Redact seals matched spans per line, so a pasted key block reaches the committed record with its base64 body and END line verbatim while the JSON reports redacted 1; the slug is derived after redaction, so the body prefix becomes the filename, a one-line key in a capture resolve note leaks the same way, and capture promote inherits the body-prefixed slug into the minted intent. Evidence: capture.redactLedgerText, workflow.go deriveSlug, promote.go. The fix must extend redaction over the body through the END line in the shared primitive, bounded so a truncated block cannot swallow the record, mask the span whole, and leave body, END, filename, note and promoted intent free of key bytes.
