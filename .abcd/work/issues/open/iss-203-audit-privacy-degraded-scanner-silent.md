---
schema_version: 1
id: "iss-203"
slug: "audit-privacy-degraded-scanner-silent"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "bug-hunt loop round 9 (state issue #197), contract-fidelity hunt angle + independent adversarial verification"
found_at: "internal/core/audit/rule_privacy.go:87 (err == nil guard), internal/adapter/scanner/scanner.go:92-124 (New, degradation paths)"
---

audit's privacy-hygiene rule guards on 'if sc, err := scanner.New(ctx.RepoRoot); err == nil' expecting a fallback to the built-in pattern set on error, but scanner.New never returns a non-nil error — every degradation path (unreadable/unparseable/uncompilable pii.json override) returns unavailable=true with a nil error instead — so the guard is always true, the documented fallback branch is dead code, and the rule never calls sc.Unavailable(); a broken per-repo severity override (malformed JSON, a directory at the config path, or an uncompilable override regex) silently drops the repo's raised severities and downgrades the audit exit code from 2 to 1 with zero finding or diagnostic explaining why, unlike history.Capture and the CLI identity path which both check Unavailable() explicitly on the same scanner. Independently verified via two-fixture test (valid override raises severity to error/exit-2; broken override silently drops to warn/exit-1 with no diagnostic) for all three degradation branches (malformed JSON, directory at path, uncompilable regex)