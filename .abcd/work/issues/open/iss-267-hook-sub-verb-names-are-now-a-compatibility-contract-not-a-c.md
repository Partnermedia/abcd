---
schema_version: 1
id: "iss-267"
slug: "hook-sub-verb-names-are-now-a-compatibility-contract-not-a-c"
severity: "minor"
category: "observation"
source: "user-observation"
found_during: "manual-capture"
---

Hook sub-verb names are now a compatibility contract, not a cushion. iss-266 makes every parent runnable, so an unknown sub-verb exits 2 instead of printing help at exit 0. On the Claude Code hook plane exit 2 is the BLOCKING status: exit 2 on UserPromptSubmit blocks the prompt, and exit 2 from 'guard hook' on PreToolUse blocks every Bash tool call -- and the PreToolUse wrapper in hooks/hooks.json treats 2 as a recognised code, so its 'FAILED TO RUN ... UNGUARDED' safety net does not fire. hooks.json ships with the plugin git clone while the binary is fetched from the latest release by hooks/bootstrap.sh, so the two can skew. No live instance today (every sub-verb hooks.json names exists at v0.5.1), and the fail-closed direction is intended. But a future rename of 'guard hook' or any 'hook <sub-verb>' must now carry an alias rather than rely on the old silent exit-0 cushion: against a skewed hooks.json/binary pair it would brick a session where it previously degraded silently. Found by the security review of the iss-266 fix.