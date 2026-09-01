---
schema_version: 1
id: "iss-2608301421380392"
slug: "the-raw-heading-soft-bound-does-not-match-a-crlf-blank-line"
severity: "major"
category: "security"
source: "user-observation"
found_during: "itd-183-round-10-ruthless"
found_at: "internal/core/reading/project.go"
---

the raw heading soft bound does not match a CRLF blank line so an unclosed heading in a CRLF file reads its title over the rest of the document and travels

Found by the round-10 adversarial ruthless review of build/itd-183. NOT
introduced by this branch -- `rawHeadingBoundRe` carries the same blank-line
alternative on the round-9 parent and before it. Recorded rather than chased
under the ruling that round 10 closes only what the branch introduced.

The soft bound on a raw heading's text is `\n[ \t]*\n`, which a CRLF blank line
does not match: its bytes are `\r\n\r\n`, and the `\r` falls outside the class.
So an element that is never closed has no bound at all, its title is read over
the rest of the document, and the slug names something else:

```
<h2>Audit Notes<CR><LF>
<CR><LF>
<sentinel><CR><LF>
```

Verified admitted; the byte-identical document with LF line endings refuses.

This is the same class round 10 closed for YAML keys -- a carriage return is
whitespace to every reader of the file and is not in the scan's own class -- and
it stays open on the heading side. Remedy: `\n[ \t\r]*\n`, or a normalisation
of line endings before the heading scan. Seed material for the exclusion floor's
own intent.
