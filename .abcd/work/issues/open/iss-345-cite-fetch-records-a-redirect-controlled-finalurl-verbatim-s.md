---
schema_version: 1
id: "iss-345"
slug: "cite-fetch-records-a-redirect-controlled-finalurl-verbatim-s"
severity: "minor"
category: "security"
source: "agent-finding"
found_during: "bughunt-round-2"
found_at: "internal/core/cite/fetch.go"
---

cite fetch records a redirect-controlled FinalURL verbatim, so C1, bidi and zero-width runes in a hostile Location query reach abcd docs lint --json and the committed citations baseline raw; the termsafe doc claim that encoding/json escapes control characters is false for those classes
## Evidence

- `internal/core/cite/fetch.go:204` assigns `out.FinalURL = resp.Request.URL.String()`. `url.Parse` rejects C0/DEL but preserves raw non-ASCII in `RawQuery`, and `String()` writes it verbatim — so the live payload set from a hostile redirect is C1 (2-byte U+009B), bidi (U+202E) and zero-width (U+200B).
- Reproduced end-to-end with `newHTTPChecker` against an httptest 302 whose Location query carried all three: `Status=="ok"`, raw bytes in `FinalURL`, `lint.SaveBaseline` accepted them (`Baseline.validate` checks only non-empty), and `abcd docs lint --json` (internal/surface/cli/cli.go:445 render) marshals them raw — `encoding/json` escapes only <0x20 plus U+2028/9.
- The false invariant is stated at `internal/termsafe/termsafe.go:22-23` and duplicated at `internal/core/lifeboat/coverage.go:149`.
- Refuter verdict: CONFIRMED substantive (minor, security). Remedy: percent-encode surviving non-printing runes at the fetch boundary + a `Baseline.validate` rung + correct both doc claims; a blanket JSON-path sanitizer was rejected (would corrupt a committed record that must round-trip). Distinct from iss-259/iss-264 (terminal text renders).
