---
schema_version: 1
id: "iss-2609012039116200"
slug: "install-accepts-relative-in-checkout-or-world-writable-data-dir-as-cache"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/ahoy/data_dir.go"
---

Sub-finding of GHSA-4q78-ccfv-f374 that does not need the design decision: `ahoy install` uses any CLAUDE_PLUGIN_DATA value as the verified cache, including one that is relative, lies inside the repository being installed, or is world-writable. `internal/core/ahoy/data_dir.go:pluginDataDir` returns the env value unexamined and `owned_copy.go:ownedCopySourceReady` plus `apply.go:installOwnedEntry` copy from it. The harness's persistent per-plugin data directory (spc-35) is always absolute, never inside a project checkout, and never world-writable, so refusing those shapes is a strict hardening of the current contract: a relative value resolves against the checkout the verb runs in, an in-checkout value blesses committed bytes as the owned PATH binary, and a world-writable cache lets any local user swap both artefact and record. The fix must establish that such a data dir is ignored with a note naming why, that install then degrades to the existing loud pinned-symlink path exactly as when no cache exists, and that a harness-shaped data dir still installs the owned copy; the display-only readers of the same variable (`vintage.go:readPinnedTag`, the skew notice) make no trust decision and are unchanged.
