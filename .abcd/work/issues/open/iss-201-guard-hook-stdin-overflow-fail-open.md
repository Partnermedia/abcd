---
schema_version: 1
id: "iss-201"
slug: "guard-hook-stdin-overflow-fail-open"
severity: "major"
category: "bug"
source: "user-observation"
found_during: "bug-hunt loop round 9 (state issue #197), security hunt angle + independent adversarial verification"
found_at: "internal/surface/cli/guard.go:154 (hook verb stdin read), guard.go:159-161 (fail-open routing)"
---

abcd guard hook reads its PreToolUse stdin payload via a bare LimitReader with no overflow check, so a payload padded past the 1 MiB cap (e.g. a blocked command with a trailing shell comment full of junk) is silently truncated, breaks JSON parsing, and routes to the fail-open-loud path (exit 1, command runs) with a diagnostic that misleadingly blames the host ('not readable JSON') instead of reporting the actual cap overflow; the sibling check verb's guardCandidate already reads cap+1 specifically to detect and refuse this. Independently verified: threshold is sharp and deterministic (cap-1 blocks, cap+1 fails open), always a loud (not silent) fail-open, exploitation cost bounded by needing >1 MiB of attacker-influenced content inside one tool_input