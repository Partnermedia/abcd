---
schema_version: 1
id: "iss-2608241845109280"
slug: "abcd-site-build-refuses-a-committed-record-so-the-site-rende"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "v0.6.4 release PR 2026-08-24"
found_at: ".abcd/work/issues/resolved/iss-2608241347321757-resolving-an-issue-is-a-convention-nobody-can-fail-so-the-le.md"
---

abcd site build refuses a committed record, so the site render fails on main and nothing catches it before CI. .abcd/work/issues/resolved/iss-2608241347321757-...md:52 carries a four-space indented code block, and the renderer supports a fixed subset only: 'unsupported markdown construct: indented code block: fence code with backticks so it renders as a copyable command — the site renders a fixed subset and never passes text through unrendered'. The binary is behaving as designed; the record is non-conforming. It reached main because the site-screenshots workflow is path-filtered and the ci changes classifier stands jobs down for a pull request confined to .abcd/, so PR 489 never ran a site build, and the failure only surfaced on the first later PR touching a watched path (the v0.6.4 release PR 497). Two gaps, not one: the record's markdown is wrong, and record-lint does not check records against the renderer's supported subset even though every record under .abcd/work/issues/ is site-rendered. The second is the detector; the first is its acceptance corpus. Note screenshots is not a required status check, so this never blocked a merge.