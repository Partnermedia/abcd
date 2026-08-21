---
schema_version: 1
id: "iss-2608210934566224"
slug: "missed-transcript-capture-recovery-sweep"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "plugin-update post-mortem 2026-08-21"
---

Session-end transcript capture is best-effort and its loss is silent: a cancelled or killed SessionEnd hook (update-then-quit, crash, SIGKILL) leaves no trace that a session was never captured into the history store. Add a recovery sweep — at session start or in ahoy doctor — that compares harness transcripts against the history store index and reports (or captures) the gap, turning silent loss into a caught-on-next-start notice. abcd history capture already ingests retroactively.