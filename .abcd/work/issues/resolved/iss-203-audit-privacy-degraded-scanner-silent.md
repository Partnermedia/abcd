---
schema_version: 1
id: "iss-203"
slug: "audit-privacy-degraded-scanner-silent"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "bug-hunt loop round 9 (state issue #197), contract-fidelity hunt angle + independent adversarial verification"
found_at: "internal/core/audit/rule_privacy.go:87 (err == nil guard), internal/adapter/scanner/scanner.go:92-124 (New, degradation paths)"
resolution: "privacy-hygiene reports an unusable per-repo scanner config instead of absorbing it. scanner.New returns a nil error on every degradation path, so the err == nil guard was always true and its documented fallback was dead; the rule now consults sc.Unavailable() and emits an error Finding naming the config and the reason. An adversarial review found the contract still bypassable one character away — blanking a new pattern's regex dropped the detector with no Unavailable() and no finding, and abcd lint then reported conforms at exit 0 over a file it had been catching — so mergeConfig refuses an empty regex on a new pattern name, bounded to new names so a bundled-pattern severity raise is unaffected. Both directions are pinned."
impact: fix
---

audit's privacy-hygiene rule guards on 'if sc, err := scanner.New(ctx.RepoRoot); err == nil' expecting a fallback to the built-in pattern set on error, but scanner.New never returns a non-nil error — every degradation path (unreadable/unparseable/uncompilable pii.json override) returns unavailable=true with a nil error instead — so the guard is always true, the documented fallback branch is dead code, and the rule never calls sc.Unavailable(); a broken per-repo severity override (malformed JSON, a directory at the config path, or an uncompilable override regex) silently drops the repo's raised severities and downgrades the audit exit code from 2 to 1 with zero finding or diagnostic explaining why, unlike history.Capture and the CLI identity path which both check Unavailable() explicitly on the same scanner. Independently verified via two-fixture test (valid override raises severity to error/exit-2; broken override silently drops to warn/exit-1 with no diagnostic) for all three degradation branches (malformed JSON, directory at path, uncompilable regex)