---
schema_version: 1
id: "iss-2608270655490511"
slug: "launch-bundle-scanner-skips-the-lock-extension-as-if-binary"
severity: "nitpick"
category: "security"
source: "agent-finding"
found_during: "security-cut-agent-flagged-siblings-2026-08-27"
found_at: "internal/adapter/scanner/scanner.go"
---

launch bundle scanner skips the .lock extension as if binary, but lockfiles are text and can carry secrets, so an included .lock ships unscanned (same class as the .svg gap the launch-payload fix closed, but pre-existing and out of that scope). Flagged by the launch-payload fix agent.