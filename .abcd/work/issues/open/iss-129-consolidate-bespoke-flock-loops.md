---
schema_version: 1
id: "iss-129"
slug: "consolidate-bespoke-flock-loops"
severity: "minor"
category: "tech-debt"
source: "impl-review"
found_during: "iss-101/102 reviews (2026-07-24 run queue, burst 3)"
found_at: "internal/fsutil/flock.go"
---

four bespoke LOCK_EX flock loops remain (memory/writer.go, intent/create.go, spec/store.go, history/store.go) now that fsutil.WithFileLock is the canonical inter-process lock primitive; consolidate them onto it (one-canonical-primitive) — pre-existing, no defect, pure debt