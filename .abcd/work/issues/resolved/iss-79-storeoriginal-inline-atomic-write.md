---
schema_version: 1
id: "iss-79"
slug: "storeoriginal-inline-atomic-write"
severity: "minor"
category: "tech-debt"
source: "agent-finding"
found_during: "2026-07-12 /abcd:run iss-32 review"
resolution: "storeOriginal routed through fsutil.WriteFileAtomic behind an extended inline-atomic-write detector (os.O_EXCL+os.Rename); CreateTemp+Rename idiom gap (inject.go SaveState) filed as iss-82"
---

fifth divergent atomic write: internal/core/memory/ingest.go storeOriginal (~:806) is an inline temp(.memtmp,O_EXCL)+fsync+rename durable write that iss-32's consolidation left untouched — it is inline, not a named func, so TestNoNonCanonicalAtomicWritePrimitives (name-based) cannot catch it, and it was outside iss-32's named 4-copy corpus. It differs from canonical fsutil.WriteFileAtomic: no parent-dir fsync, no explicit chmod. Consolidate it onto fsutil.WriteFileAtomic(target, material.rawBytes, 0644) (keeping the pre-existing sources-dir symlink guard), OR extend the canonical-primitive detector to catch inline temp+rename sequences. Safe today (has its own symlink guard); this is a completeness/consistency follow-up.