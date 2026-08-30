---
schema_version: 1
id: "iss-2608300258519825"
slug: "heading-and-key-floor-spellings-and-fence-blind-verifier"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 second-round reviews, 2026-08-30"
found_at: "internal/core/reading/project.go (redactExcluded, verifyRedaction, excludedKeyLineRe)"
---

The exclusion floor's heading half matches the title the section scanner reports and neither the redactor nor the verifier normalises an ATX closing sequence or recognises a setext heading, so an excluded section travels under '## Audit Notes ##', '## Open Questions #' or an underlined heading while the manifest asserts it refused; the key half misses a quoted key ("origin":) and a block scalar with an internal blank line; and the verifier is not fence-aware while the redactor is, so a fenced example of the record template in any admitted markdown refuses every assembly. Normalise closing hashes and detect setext in both places, accept a quoted key, skip blank lines inside a block, and verify over the same fence-aware section scan the redactor uses.
