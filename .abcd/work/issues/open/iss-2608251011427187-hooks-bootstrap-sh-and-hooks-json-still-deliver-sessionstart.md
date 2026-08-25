---
schema_version: 1
id: "iss-2608251011427187"
slug: "hooks-bootstrap-sh-and-hooks-json-still-deliver-sessionstart"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "adversarial review of the v0.6.6 priority sweep 2026-08-25"
found_at: "hooks/bootstrap.sh"
---

hooks/bootstrap.sh and hooks.json still deliver SessionStart notices by a non-zero exit, which the harness renders as an opaque banner with the text dropped. iss-2608241115201044 fixed the binary half: abcd hook session-start now exits 0, writes its notice text to stderr and only a CONSTANT to stdout. The shell half is unchanged and inconsistent with it. bootstrap.sh's notice() exits 2 on a REPORTED CONDITION including a successful install, and hooks.json's SessionStart chain propagates that ($s) plus a trailing exit 2 on the missing-binary branch. Under the same premise those messages are dropped too, so a fresh install's success line and the missing-binary complaint are both delivered by the mechanism the binary half abandoned. bootstrap.sh's own comment states the reasoning that makes the binary fix correct — 'a SessionStart hook's stdout becomes model context' — and an adversarial review demonstrated that concern concretely, planting a directive payload in a TRACKED config value that reached context when notices were briefly routed to stdout. So the shell must NOT simply move its text to stdout either. The open question is what channel a shell hook has left: a constant on stdout plus text on stderr mirrors the binary, but bootstrap.sh's exit code also governs the install path, and changing it touches the most safety-critical surface in the repo. Deliberately deferred rather than bundled into the binary fix.