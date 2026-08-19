---
schema_version: 1
id: "iss-291"
slug: "the-v0-6-0-tag-is-permanently-unpublishable-its-roll-branch"
severity: "minor"
category: "process"
source: "user-observation"
found_during: "v0.6.0 release attempt"
---

The v0.6.0 tag is permanently unpublishable: its roll branch carried no receipts commit, the receipt gate fail-closed at release time, and the re-release path builds only from the tagged commit whose frozen tree can never contain receipts naming its own content commit. Every push to main re-attempts and fails until a newer dated heading supersedes it. The tag stays as a record (anti-tag-move); v0.6.1 is the recovery