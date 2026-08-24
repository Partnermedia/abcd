---
schema_version: 1
id: "iss-2608241115201044"
slug: "sessionstart-exit2-advisory-invisible"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "sessionstart-hook-error-investigation"
found_at: "internal/surface/cli/cli.go"
---

SessionStart exit-2-as-advisory no longer reaches the user. abcd hook session-start returns exitError{Code: 2} whenever it has notices to print (internal/surface/cli/cli.go:1289), on the documented assumption that a non-zero SessionStart surfaces stderr without blocking. Claude Code v2.1.241 instead renders it as 'SessionStart:startup hook error' followed by a truncated echo of the hooks.json command string; the stderr notice itself never reaches the user. Every session-start notice is therefore an opaque error banner: transcript-capture gaps, staged-drain failures, the remaining-backlog count, binary skew, dogfood staleness, and the version transition. The data-loss-adjacent ones matter most - a user told that a session transcript was not captured now sees only a hook error, with no way to learn what it was.