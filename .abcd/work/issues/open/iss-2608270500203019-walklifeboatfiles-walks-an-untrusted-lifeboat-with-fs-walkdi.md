---
schema_version: 1
id: "iss-2608270500203019"
slug: "walklifeboatfiles-walks-an-untrusted-lifeboat-with-fs-walkdi"
severity: "major"
category: "security"
source: "agent-finding"
found_during: "github-ledger-dedup-2026-08-27"
found_at: "internal/core/lifeboat/embark.go"
---

walkLifeboatFiles walks an untrusted lifeboat with fs.WalkDir (unbounded ReadDir(-1) per directory), so one wide directory materialises and sorts its entire entry list before the maxEmbarkFiles cap can fire. GitHub mirror: #343