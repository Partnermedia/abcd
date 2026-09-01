---
schema_version: 1
id: "iss-2609012035100572"
slug: "ghsa-xq36-hcgf-9wrj-cwe-693-transcript-staging-is-unlocked-a"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/history/staging.go"
---

GHSA-xq36-hcgf-9wrj (CWE-693): transcript staging is unlocked and content-blind. history.Stage dedups on the session id alone, so a second SessionEnd for one session carrying different bytes is a no-op, and the drain then stores the stale copy and deletes the only copy of the newer transcript; and Stage takes no lock, so two concurrent SessionEnds for one session id both pass the listing check, write two .raw files, and Drain stores two records claiming the same session. Evidence: internal/core/history/staging.go Stage (the session-id-only loop and its no-lock doc comment), stagedFilename's nanosecond stamp, Drain's unlocked list-capture-remove; TestStageIsIdempotentPerSession pins the defect. The fix must establish: the stage handshake is exclusive (fsutil.WithFileLock on staging/.lock around list, content compare and write), identical bytes are a no-op, different bytes replace the staged copy at the existing path (last-writer-wins, the fresher end-of-session bytes being the ones worth keeping) with a Replaced signal the session-end hook reports instead of no-op, and Drain removes a staged file only if it still holds the bytes it captured, leaving a mid-drain replacement for the next pass.
