---
schema_version: 1
id: "iss-174"
slug: "rules-override-withholds-bundled-default-upgrades"
severity: "minor"
category: "security"
source: "review-followup"
found_during: "iss-156 adversarial review 2026-07-30"
found_at: "internal/core/rules/rules.go"
---

mergeDomain replaces a domain's recall and rules arrays wholesale, so a repo that overrides either array silently withholds every later security upgrade to the bundled defaults: internal/core/rules/rules.go copies the override array over the base one instead of unioning, with no schema bump and no notice, so a repo that pinned PII recall or PII rules before iss-156 keeps the old set and never sees the new network keywords or the never-commit-network-identifiers rule line. Per-field merge is the documented and wanted behaviour for state, but for security-bearing arrays the quiet outcome is a stale ruleset that looks current. Options to weigh: union rather than replace for recall/aliases (additive, no loss), a distinct replace-vs-extend marker in the override, or at minimum a loud diagnostic in abcd rules naming which bundled entries an override is withholding. First activated by the iss-156 PII upgrade, which is why it is captured now.