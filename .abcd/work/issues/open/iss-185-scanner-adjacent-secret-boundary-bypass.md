---
schema_version: 1
id: "iss-185"
slug: "scanner-adjacent-secret-boundary-bypass"
severity: "critical"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 1"
found_at: "internal/adapter/scanner/scanner.go:274"
---

scanner leaves a second back-to-back secret token undetected and unredacted, and the fail-closed residual re-scan reports clean anyway. Fixed-length token patterns (e.g. github_pat_ at patterns.go:93) require a leading \b; when two same-family tokens are concatenated with no separator, the first regex match consumes bytes up through the second token's first byte, leaving word/word at the junction so the second token's \b can never match. ScanText (scanner.go:274) therefore returns only one finding. Redact's fingerprintSpan keeps 2 raw alnum bytes at the tail of the masked span (scanner.go:393), so the junction stays word/word after redaction too — the second token survives completely raw in the output. Because the boundary never recovers, history.Capture's stage-two residual check (history.go:159, blockingResidual) re-scans the redacted text, finds nothing, and writes the record to disk with a live secret still in it, directly violating history's documented fail-closed guarantee ('a stored record can never contain a live secret'). Distinct from iss-65 (Finding.MarshalJSON snippet cross-leak in serialized findings, already fixed) — this is a ScanText/Redact detection gap in the scan pass itself. Reproducing test: internal/adapter/scanner/, TestConcatenatedSecretsBothDetected.