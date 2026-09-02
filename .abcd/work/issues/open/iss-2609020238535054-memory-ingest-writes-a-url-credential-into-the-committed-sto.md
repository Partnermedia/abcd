---
schema_version: 1
id: "iss-2609020238535054"
slug: "memory-ingest-writes-a-url-credential-into-the-committed-sto"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/memory/ingest.go"
---

memory ingest writes a URL credential into the committed store: materialFromFetched sets both the page origin and the page title from the fetched final URL verbatim, so basic-auth userinfo (user:pass@) and credential-shaped query keys survive into the page frontmatter's source.citation, into .sources_index.json (origin and consumers.memory.citation) and into the IngestResult.Citation the --json render prints. redactText cannot see an opaque basic-auth password, so the scanner backstop does not catch it. The six 'fetch failed for %s' sites in acquireSource, defaultFetch and readFetchedResponse echo the raw source the same way, as do the byte-cap refusal and the redirect-count refusal. CWE-522/CWE-532. The fix must clear the userinfo and drop the credential-shaped query keys (token, api_key, apikey, access_token, key, password, secret, signature; case-insensitive) where the final URL becomes the stored origin and title, and route every fetch-failure message through the redacting renderer.
