---
schema_version: 1
id: "iss-2608300349309723"
slug: "itd-183-fourth-round-residue"
severity: "minor"
category: "inconsistency"
source: "impl-review"
found_during: "itd-183 fourth-round security review, 2026-08-30"
found_at: "internal/core/reading/project.go (sameRendering, excludedKeyInFirstBlock), internal/core/reading/assemble.go (--out refusals), internal/core/reading/status.go"
---

itd-183 fourth-round residue, all Low: sameRendering models only emphasis and code marks, so a heading carrying an HTML comment or tag, an entity, a link wrapper, a homoglyph, or nested in a blockquote or list item renders as the excluded title but slugs differently (strip tags and comments, decode common entities, unwrap links before slugging, or refuse on a slug substring match and correct the comment); top-level YAML spellings of an excluded key that the line pattern misses travel (explicit key, top-level or nested flow mapping, a double-quoted key with an escape); the two --out refusals interpolate the resolved absolute directory built from the operator's relative string, which scrubPaths cannot redact when cwd is not a prefix; status.go reads run-directory names under the assembler's own scratch, worth a charter sentence so it is not mistaken for an invariant-14 breach.
