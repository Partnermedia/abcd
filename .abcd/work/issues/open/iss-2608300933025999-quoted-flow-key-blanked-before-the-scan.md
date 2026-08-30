---
schema_version: 1
id: "iss-2608300933025999"
slug: "quoted-flow-key-blanked-before-the-scan"
severity: "major"
category: "security"
source: "impl-review"
found_during: "itd-183 eighth-round ruthless review, 2026-08-30"
found_at: "internal/core/reading/project.go (blankQuoted, excludedKeyInFirstBlock, rawHTMLHeading cache)"
---

The quote blanker blanks a quoted flow KEY as though it were a scalar — {"origin": x} and {a: 1, 'origin': x} sit after a brace or comma, so the span is blanked before the flow scan can read it and the excluded key travels under a manifest asserting refusal (pre-existing on both sides of the eighth-round commit, in the function it rewrote; the flow-key pattern's own quoted-key alternative never fires). Skip a quoted span whose remainder begins with optional whitespace and a colon — a quoted scalar followed by a colon is a key. Also: excludedKeyInFirstBlock still opens its block at the first --- found anywhere, so the frontmatter-less false-refusal class persists on the key path (take the range from firstBlockRange); the escape-aware double-quote alternative admits an invalid-YAML shape the previous commit refused (disclose); a heading whose text follows a blank line inside the element, or carries a line break element, is neither refused nor in the residue list; the per-element regexp cache is unbounded by input-supplied tag names (moot in a one-shot CLI, worth a note for a long-lived transport).
