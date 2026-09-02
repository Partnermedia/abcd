---
schema_version: 1
id: "iss-2609020543544634"
slug: "a-here-document-delimiter-that-is-not-letter-led-fell-to-the"
severity: "major"
category: "security"
source: "review-followup"
found_during: "autonomous-run-2026-09-01"
origin: researcher-authored
production_mode: hand-written
found_at: "internal/core/guard/tokenize.go"
---

A here-document delimiter that is not letter-led fell to the arithmetic-shift branch, so the DOCUMENT was tokenized as commands and one apostrophe in the body raised ErrUnparsableCommand — which the pre-tool-use hook maps to fail-OPEN, leaving any hazard on a later line to run unguarded. bash accepts a WORD after <<: 'cat <<20', 'cat <<$D' and the '<<20' of 'let mask=1<<20' are all here-documents (the shell tokenizes the redirection before the let builtin sees arithmetic), and each of those spellings reached exit 2 / NOT CHECKED on v0.7.0. Evidence symbol: isDelimStart (internal/core/guard/tokenize.go), consulted at the '<<' branch of tokenize. It also falsified the claim both front doors make — that a quote inside a here-document body is document text and never unparsable. The fix must read the conservative superset of delimiter words (letter, digit, underscore, $-led) so the body is data, with the enclosing arithmetic context left as the sole shift discriminator.
