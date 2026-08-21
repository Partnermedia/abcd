---
schema_version: 1
id: "iss-2608211432482363"
slug: "rfc-1-itd-26-one-sided-related-link"
severity: "nitpick"
category: "inconsistency"
source: "user-observation"
found_during: "bughunt-b/round-5"
found_at: ".abcd/development/roadmap/rfcs/rfc-1-pirate-mode-yolo-for-power-users.md"
resolution: "Add the reciprocal related_rfcs [rfc-1] to itd-26"
impact: internal
resolved_by:
  commit: "1c3ae21"
---

rfc-1 declares related_intents itd-26 but itd-26 has no reciprocal related_rfcs, a one-sided pair the RFC bidirectional convention expects (last unswept half after round-4 rfc-2/adr-43 fix)