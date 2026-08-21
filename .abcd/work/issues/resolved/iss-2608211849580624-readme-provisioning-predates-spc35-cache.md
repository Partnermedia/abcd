---
schema_version: 1
id: "iss-2608211849580624"
slug: "readme-provisioning-predates-spc35-cache"
severity: "minor"
category: "drift"
source: "user-observation"
found_during: "bughunt-round-6"
found_at: "README.md:146"
resolution: "README rewrites provisioning for the spc-35 cache and corrects the binary-meta remedy"
impact: fix
resolved_by:
  commit: "1395e72"
---

README provisioning section describes download-on-every-update, but spc-35 shipped a persistent CLAUDE_PLUGIN_DATA cache that provisions by verified copy; the .binary-meta deletion remedy is also a no-op on cache-provisioned roots