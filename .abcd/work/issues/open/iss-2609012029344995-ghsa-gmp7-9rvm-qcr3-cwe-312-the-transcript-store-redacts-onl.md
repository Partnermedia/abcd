---
schema_version: 1
id: "iss-2609012029344995"
slug: "ghsa-gmp7-9rvm-qcr3-cwe-312-the-transcript-store-redacts-onl"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/history/history.go"
---

GHSA-gmp7-9rvm-qcr3 (CWE-312): the transcript store redacts only the PEM BEGIN header and stores the key body verbatim. history.Capture stage one scans with the header-only pem_private_key pattern and scanner.Redact seals the header span alone; stage two rescans with the same detector, the fingerprinted header no longer matches, and the write proceeds with every body line and the END line in the record while the frontmatter counts one secret. Header and body on one line leak the same way because byte-span sealing stops at the header. Evidence: history.Capture, scanner.DefaultPatterns pem_private_key, scanner.Redact. The fix must consume the body through the END line in the shared primitive with a bound, mask the span whole, and keep the prose after the block; the record mode (0644) is a separate observation.
