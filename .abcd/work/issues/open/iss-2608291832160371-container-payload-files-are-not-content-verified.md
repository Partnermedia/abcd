---
schema_version: 1
id: "iss-2608291832160371"
slug: "container-payload-files-are-not-content-verified"
severity: "major"
category: "security"
source: "impl-review"
found_during: "v0.6.9-security-pass"
found_at: "internal/adapter/scanner/scanner.go"
---

GHSA-9wv7 residual: compressed/container payload files (.zip .gz .tgz ...) cannot be content-verified by the byte scan; they are reported as container_unverified, not refused. Decide whether the launch gate should refuse containers in the payload or decode known formats under the cap and scan entries.
