---
schema_version: 1
id: "iss-2608291814565929"
slug: "memory-store-dir-guard-walks-two-copies"
severity: "major"
category: "security"
source: "impl-review"
found_during: "ultra-v0.6.8-followup"
found_at: "internal/core/memory/writer.go"
---

ultra-v0.6.8 C8: safeMemoryDir in internal/core/memory/writer.go duplicates validatedMemoryDir's segment walk (same Lstat, same symlink/non-dir refusal, same UnsafeStorePathError); they differ only in the IsNotExist arm (create vs return absent). Because safeMemoryDir returns an empty path when the store is absent, bare.go and lint.go each carry an if !present { mem = Dir(root) } fallback with a comment explaining why. This is the GHSA-72rp-qxm2-r8vq guard, so two walkers means a hardening fix to one leaves the other open. Fix: one walker memoryDir(repoRoot, create) (string, bool, error) that always returns Dir(repoRoot); both wrappers become one-liners and the call-site fallbacks disappear.
