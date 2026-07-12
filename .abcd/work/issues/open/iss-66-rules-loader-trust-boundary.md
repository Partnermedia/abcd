---
schema_version: 1
id: "iss-66"
slug: "rules-loader-trust-boundary"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/rules/rules.go"
---

rules loader trust-boundary (itd-3): an untrusted per-repo .abcd/rules.json can silence default guardrail domains — Merge() ORs the global kill switch (rules.go:156) and mergeDomain lets an override set any default domain dormant, so a cloned repo suppresses all rule injection (P15); session-state lives in a predictable shared /tmp path (os.TempDir/abcd-rules-state, sha256(session).json, session defaults to constant default) letting a local co-tenant suppress injection fail-open (inject.go:99, P14); Load() Lstats rules.json then os.ReadFile re-resolves by name, a symlink-swap TOCTOU (rules.go:132, C19). Detector: override cannot disable a default guardrail/kill-switch; state under UserCacheDir+ownership check; open-once O_NOFOLLOW+fstat. Corpus: P15, P14, C19.