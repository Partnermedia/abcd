---
schema_version: 1
id: "iss-68"
slug: "spec-store-hardening"
severity: "minor"
category: "bug"
source: "agent-finding"
found_during: "clean-slate-sweep"
found_at: "internal/core/spec/store.go"
---

spec store hardening (itd-80): spec.Create writes a Summary/TODO body — decide the minimal native-spec body and make record-lint tolerant (spec.go:135, seed4); ensureDir/ensureRealDir only Lstat the leaf so a symlinked ancestor (specs/, intents/) is followed — literal guard or honest comment (seed5, spec+intent); readRepoFile Lstat then ReadFile re-resolves the path (symlink TOCTOU) and the size cap is not atomic — open-once O_NOFOLLOW+fstat+LimitReader (store.go:225, P7); Close Lstat-exists-then-Rename is racy and os.Rename silently overwrites — no-replace primitive (store.go:212, P8); maxIntentSpecNum silently drops a non-null unparseable spec_id, fail-open reservation scan (store.go:146, P5). Corpus: seed4, seed5, P7, P8, P5.