---
schema_version: 1
id: "iss-2608300335037531"
slug: "manifest-determinism-and-outdir-contract-overclaim"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-183 third-round ruthless review, 2026-08-30"
found_at: ".abcd/development/readings/README.md, internal/core/reading/manifest.go, .abcd/development/brief/04-surfaces/README.md, internal/core/reading/assemble.go"
---

The charter, manifest.go and the brief surfaces index describe the manifest as timestamp-free and identical across two assemblies of one repository state, but it carries run_id, a minted timestamp plus entropy, so two dry-run assemblies of one HEAD produce different manifest hashes; state what holds — items and exclusions identical, the bundle byte-identical, the manifest differing only in run_id. Also the AssembleResult.OutDir comment claims a verbatim echo of an operator-supplied path while the front door now absolutises every relative --out against the working directory.
