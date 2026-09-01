---
schema_version: 1
id: "iss-2609012037433158"
slug: "ghsa-35fj-9w6f-7h62-cwe-319-cwe-311-isurl-admits-http-as-wel"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/ingest.go"
---

GHSA-35fj-9w6f-7h62 (CWE-319, CWE-311): isURL admits http as well as https and acquireSource fetches either, and the CheckRedirect of defaultFetch re-checks only the redirect count and the SSRF host guard, so a hop from https to plaintext http is followed — whereas update.go newUpdater refuses a redirect that left https. A man-in-the-middle on a plaintext source can rewrite its text and plant a License: header the store copies into provenance. The fix must refuse a non-https source in acquireSource before any fetcher runs, with a refusal that names the scheme, pin the scheme per hop in the redirect policy the way abcd update does, and move the http:// test fixtures (TestIngestRefusesSSRFTargets, TestIngestEncodesHiddenRunesInFetchedOrigin) to https:// so the SSRF guard stays the thing under test rather than passing on the scheme refusal.
