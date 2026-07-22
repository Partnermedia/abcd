---
schema_version: 1
id: "iss-36"
slug: "retired-name-banlist"
severity: "critical"
impact: internal
category: "drift"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: ".abcd/record-lint.json"
resolution: "Bans armed (d03e81a): 6 new banned_tokens + widened python-lint-names; 245 blockers flagged on first run (acceptance corpus exceeded — issue estimated ~50). Corpus drained behind the armed bans (6ca63da, 73 files) and a 4-segment prose review fixed ~30 truth defects the mechanical verify missed (14ee119). record-lint 0 blockers. logbook ban deliberately excluded — adjudication split to iss-56."
---

retired-name banlist population: roughly fifty stale references to retired locations and identifiers survive in the record, and the drift outlived a dedicated same-day consistency pass — development/activity (38 refs in the brief), .work/issues and bare .work/ as a root path (11 refs), scripts/abcd/, .abcd/logbook, .flow/, .github/workflows/lint.yml and lint-corpus.yml, --since-staged, and the python-lint-names pattern narrowly missing bare intent_lint (used by intents/README.md and adr-34). Detector (per retire-the-name and fix-the-detector): add each retired spelling to record-lint banned_tokens with allow_context for genuinely historical passages, widen the intent_lint pattern, then drain the references BEHIND the armed bans — never ahead of them. Acceptance corpus: the reference counts above; the bans are proven when they flag all ~50 on first run and the count ratchets to zero.