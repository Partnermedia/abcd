---
schema_version: 1
id: "iss-24"
slug: "changelog-fragments"
severity: "minor"
category: "future-work-seed"
source: "user-observation"
found_during: "sources-ingest session 2026-07-08"
---

changelog fragments for abcd-managed repos: per-change entries land as files in a fragments directory (towncrier pattern) and are assembled into CHANGELOG.md at release, so concurrent PRs never conflict on the Unreleased block; candidate abcd verb/discipline — assembly could hook the release gate