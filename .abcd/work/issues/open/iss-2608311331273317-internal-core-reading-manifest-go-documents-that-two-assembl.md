---
schema_version: 1
id: "iss-2608311331273317"
slug: "internal-core-reading-manifest-go-documents-that-two-assembl"
severity: "minor"
category: "future-work-seed"
source: "agent-observation"
found_during: "manual-capture"
origin: researcher-authored
production_mode: hand-written
---

internal/core/reading/manifest.go documents that two assemblies of one repository state at one commit produce manifests differing in run_id and in nothing else. Nothing asserts it. spc-65 scopes the manifest to two weaker properties (no timestamp-shaped key or scalar, item paths in lexicographic order) and excludes it from the byte comparison, so a nondeterminism confined to the manifest — a changed content hash, an item recorded at a different field — is invisible to the amnesia eval. The cheap closure is a manifest comparison modulo run_id in the same eval.
