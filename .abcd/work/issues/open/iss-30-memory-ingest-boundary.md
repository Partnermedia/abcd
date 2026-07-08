---
schema_version: 1
id: "iss-30"
slug: "memory-ingest-boundary"
severity: "major"
category: "bug"
source: "agent-finding"
found_during: "2026-07-08 multi-agent review"
found_at: "internal/core/memory/ingest.go"
---

memory ingest input-boundary defects: HTTP status is never checked so 404/500 error pages are silently ingested as source content (internal/core/memory/ingest.go:558-575); tilde expansion mangles ~user paths into home+user concatenations (ingest.go:579-584); a --keep-original failure after the page write reports total failure although pages and registry were durably mutated (ingest.go:301-311); CRLF pages are accepted by parseFrontmatter but rejected by splitFileFrontmatter so hashes and summaries silently degrade (yaml.go:558-591); the URL-ingest success path, content-type handling, PDF extraction, and original-storage are untested, as are YAML block scalars and double-quoted escapes. Detector: an ingest-boundary test suite — fetch status matrix, content-type matrix, CRLF round-trip, tilde cases, partial-failure reporting, parser-parity cases. Acceptance corpus: the six instances above.