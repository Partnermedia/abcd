---
schema_version: 1
id: "iss-134"
slug: "readdirbounded-determinism-doc-claim"
severity: "nitpick"
category: "tech-debt"
source: "impl-review"
found_during: "iss-112/114/116 review (2026-07-24 run queue, burst 9)"
found_at: "internal/core/lifeboat/probe.go"
---

readDirBounded's doc comment claims sorted-by-name determinism, which holds only at or under the bound; a directory exceeding the bound yields a readdir-order subset that is then sorted, so WHICH entries survive is nondeterministic (loud via truncated, consistent with ListDir, pathological trigger only)