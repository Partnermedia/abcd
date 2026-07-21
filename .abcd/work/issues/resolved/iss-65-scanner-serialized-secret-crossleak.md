---
schema_version: 1
id: "iss-65"
slug: "scanner-serialized-secret-crossleak"
severity: "critical"
impact: fix
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/adapter/scanner/finding.go"
resolution: "Byte-span snippet masking closes the same-line secret cross-leak (both disjoint and overlapping matches); isText mid-rune trim; unreadable->Unscanned. PR #30 merged (main 5a49696). P10 skip-list coverage floor deferred: the zero-coverage sentinel is already the floor. ruthless SHIP + security PASS (security caught an overlap BLOCK on the first fix)."
---

scanner serialized-finding secret cross-leak (BLOCK) + scan hygiene: Finding.MarshalJSON rebuilds the snippet from the full source line but masks only THIS finding own token, so a second secret on the same line (minified JSON, collapsed .env, a=X; b=Y) is serialized verbatim (finding.go:124). Also isText caps the UTF-8 sniff at 8192 bytes and utf8.Valid fails when the cut splits a multibyte rune, misclassifying a valid >8KB text file as binary and skipping the scan (scanner.go:338); ScanBundle drops an unreadable file with a bare continue instead of surfacing it in Unscanned (scanner.go:269); per-repo pii.json skip-lists have a severity floor but no coverage floor (scanner.go:128). Detector: two-secrets-one-line redaction test; mid-rune isText case; unreadable-file-surfaced case; skip-list coverage floor. Corpus: sweep C14/C17, C15, C18, P10.