---
schema_version: 1
id: "iss-231"
slug: "guard-warn-tier-silent-on-hook"
severity: "major"
category: "bug"
source: "impl-review"
found_during: "iss-200 Option-D design security review (2026-08-15)"
found_at: "internal/surface/cli/guard.go:VerdictWarn case"
---

guard warn tier is silent on the live PreToolUse hook: internal/surface/cli/guard.go maps VerdictWarn to Fprintln(stderr, message) + return nil = exit 0, but the failOpen path's own comment (same function) documents that a pre-tool-use hook exiting 0 has its stderr DISCARDED — which is exactly why failOpen uses exit 1 to be loud. So every warn-tier entry (e.g. git-reset-hard) writes a message nobody ever sees and the command runs as if allowed. The warn tier is effectively invisible. Fix direction: map VerdictWarn to exit 1 (the same non-blocking-but-loud status failOpen already relies on), so a warn both runs and surfaces its message. Load-bearing for iss-200's Option-D design, whose 'warn the middle with context' model is impossible while warn is silent; also a standalone correctness bug independent of iss-200.