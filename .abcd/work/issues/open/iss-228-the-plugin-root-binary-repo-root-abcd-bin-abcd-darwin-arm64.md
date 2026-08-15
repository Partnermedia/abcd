---
schema_version: 1
id: "iss-228"
slug: "the-plugin-root-binary-repo-root-abcd-bin-abcd-darwin-arm64"
severity: "minor"
category: "observation"
source: "agent-observation"
found_during: "ahoy install dogfood"
found_at: "internal/core/ahoy"
---

The plugin-root binary (repo-root abcd -> bin/abcd-darwin-arm64) sat a month stale after iss-171 merged, so the ahoy skill's first resolution rung reported pre-iss-171 gaps (wrong target, missing --bin-dir) with no staleness signal; the skill and detection have no guard that warns when the plugin-root binary predates the source tip in a source checkout