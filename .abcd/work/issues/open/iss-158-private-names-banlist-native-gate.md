---
schema_version: 1
id: "iss-158"
slug: "private-names-banlist-native-gate"
severity: "major"
category: "future-work-seed"
source: "agent-finding"
found_during: "managed-repo NEXT.md privacy-leak investigation 2026-07-29"
found_at: "internal/adapter/scanner"
---

no commit-time gate for user-private identifiers: the canonical home for this is itd-74 (name-banlist, in drafts/) — its private-banlist layer (local untracked file + committed guard hook, refuse by entry-key only, never echoing the value) is exactly the gate this incident needed. This issue is NOT a new primitive: it records the field evidence that itd-74's scope must include machine identifiers (hostnames, tailnet IPs/prefixes, device names), not only harness/project names, and that the incident strengthens the case for promoting itd-74 out of drafts (a maintainer decision). Generic patterns (iss-154/157) catch shapes; only the banlist catches this user's actual names. Prior art at skill level: the confidential-sources sync-banlist guard; dogfood prototype already live per itd-74's own record. Complements, does not replace, iss-154.