---
schema_version: 1
id: "iss-103"
slug: "gitutil-checkignored-fail-open-contract"
severity: "nitpick"
category: "tech-debt"
source: "agent-finding"
found_during: "multi-agent-bughunt"
---

gitutil.CheckIgnored returns an empty map for every non-zero git exit, conflating 'nothing ignored' (exit 1) with 'git cannot answer' (exit 128 not-a-repo/corrupt, or git absent). This is intentional fail-open and correct for the audit-rule callers (which guard with InRepo). The one reachable defect was the launch bundle's finalize calling it unguarded, so a git failure silently admitted every gitignored file to the release bundle — that is already fixed (was B18: launch now uses a fail-closed checkIgnoredStrict). Recorded from the multi-agent bug hunt (was B23) for the record: consider lifting a strict, error-returning CheckIgnoredStrict into gitutil so the launch package need not keep its own copy, while preserving the fail-open CheckIgnored for audit. Nitpick / tech-debt; no live defect remains.