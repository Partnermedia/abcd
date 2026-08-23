---
schema_version: 1
id: "iss-2608210934566224"
slug: "missed-transcript-capture-recovery-sweep"
severity: "major"
category: "future-work-seed"
source: "user-observation"
found_during: "plugin-update post-mortem 2026-08-21"
---

Session-end transcript capture is best-effort and its loss is silent: a cancelled or killed SessionEnd hook (update-then-quit, crash, SIGKILL) leaves no trace that a session was never captured into the history store. Add a recovery sweep — at session start or in ahoy doctor — that compares harness transcripts against the history store index and reports (or captures) the gap, turning silent loss into a caught-on-next-start notice. abcd history capture already ingests retroactively.

Raised minor -> major on 2026-08-23. The sweep was seeded as a nicety against
rare events — update-then-quit, crash, SIGKILL. `iss-2608230817034768` shows the
loss is not rare but systematic: redaction cost rises at roughly 0.7 s per MB
against a shutdown budget somewhere near 1.5 s, so past a few MB loss is the
norm rather than a risk. Nine of this one repo's ended transcripts are already
absent from its store.

That also makes this the detector the capture fix has to land against, since no
fix to the budget can be watched fail without it — and the sweep is what would
pin the budget itself, which observational data cannot. Distinguishing "never
ended" from "ended and lost" is the whole difficulty: three of that repo's
apparent absences were simply sessions still running, and two more predate the
store's first record, which is exactly the ambiguity a sweep has to resolve
before it can report anything trustworthy.