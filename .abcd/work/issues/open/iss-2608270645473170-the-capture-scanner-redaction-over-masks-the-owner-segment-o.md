---
schema_version: 1
id: "iss-2608270645473170"
slug: "the-capture-scanner-redaction-over-masks-the-owner-segment-o"
severity: "nitpick"
category: "bug"
source: "agent-observation"
found_during: "capture-owner-overredaction-2026-08-27"
found_at: "internal/adapter/scanner"
---

The capture/scanner redaction over-masks the owner segment of a public 'owner/repo' string, replacing a legitimate public GitHub org/user name with [redacted-user] even when it is the very repository the record lives in. Observed 2026-08-27 when a captured issue body referencing this repo's own owner/repo slug had the owner masked. Not harmful (the slug still reads), but a false positive: the identity/path redaction treats an owner/repo owner as a sensitive identifier unconditionally. Consider exempting the repo's own owner (and public org names) from identity redaction, or narrowing the owner-segment match. Adjacent: iss-324 (scanner ispathsegmentbyte includes the slash in a third-party path).