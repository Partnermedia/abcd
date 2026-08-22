---
schema_version: 1
id: "iss-2608220150157510"
slug: "changelog-per-change-entries-are-the-top-merge-contention-site"
severity: "major"
category: "process"
source: "user-observation"
found_during: "abcdev-site session close-out 2026-08-22"
found_at: "CHANGELOG.md"
---

CHANGELOG.md is the repo's most contended file (205 changes since July): every user-facing change appends an entry near the top, so any two in-flight branches conflict there — and it keeps biting on real merges. The SOTA fix is per-change fragments (the towncrier/changesets pattern, which is also the shape the ledger already uses): each change commits one fragment file with a collision-proof name, and the release cut — which already owns CHANGELOG writes via launch ship (adr-37) — assembles the dated section and deletes the fragments. Zero contention by construction; the changelog-driven release flow is unchanged, only its input format widens