---
schema_version: 1
id: "iss-2608241115201044"
slug: "sessionstart-exit2-advisory-invisible"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "sessionstart-hook-error-investigation"
found_at: "internal/surface/cli/cli.go"
resolution: "abcd hook session-start exits 0 and delivers its notices where they are read. It returned exit 2 on the belief that a non-zero SessionStart puts stderr in front of the human; the harness renders an opaque startup hook error banner and drops the text, so every notice arrived as an error with no content. The notice text goes to stderr and only a CONSTANT plus a count goes to stdout, naming abcd history staged and abcd ahoy. That split matters: SessionStart stdout is injected into the session context, and an adversarial review demonstrated a directive payload reaching context through meta.setup_version in the TRACKED .abcd/config.json when the text was briefly routed there. bootstrap.sh's own comment had already recorded the reasoning. The shell half of the same premise is deferred as iss-2608251011427187."
impact: fix
---

SessionStart exit-2-as-advisory no longer reaches the user. abcd hook session-start returns exitError{Code: 2} whenever it has notices to print (internal/surface/cli/cli.go:1289), on the documented assumption that a non-zero SessionStart surfaces stderr without blocking. Claude Code v2.1.241 instead renders it as 'SessionStart:startup hook error' followed by a truncated echo of the hooks.json command string; the stderr notice itself never reaches the user. Every session-start notice is therefore an opaque error banner: transcript-capture gaps, staged-drain failures, the remaining-backlog count, binary skew, dogfood staleness, and the version transition. The data-loss-adjacent ones matter most - a user told that a session transcript was not captured now sees only a hook error, with no way to learn what it was.