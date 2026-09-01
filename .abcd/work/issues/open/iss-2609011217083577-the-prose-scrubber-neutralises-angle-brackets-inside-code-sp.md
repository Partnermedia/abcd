---
schema_version: 1
id: "iss-2609011217083577"
slug: "the-prose-scrubber-neutralises-angle-brackets-inside-code-sp"
severity: "major"
category: "ux"
source: "user-observation"
found_during: "the v0.7.0 cut, changelog ingest"
found_at: "internal/termsafe/prose.go"
origin: researcher-authored
production_mode: hand-written
---

The prose scrubber neutralises angle brackets inside code spans, so a documented
placeholder ships as a shell redirect. `cleanProse`
(`internal/termsafe/prose.go:83`) rewrites every HTML opener by inserting a space
after the `<`, which is correct and load-bearing for prose: it stops an injected
`<script>` or `<!--` reaching a rendered surface as live markup, and the comment
case folds into the same rule. It does not exempt code spans, where CommonMark
never parses HTML in the first place, so the neutralisation buys nothing there
and corrupts the content instead.

Observed while ingesting the v0.7.0 changelog. The composer emitted a code span
holding the invocation with an angle-bracket placeholder:

```
abcd reading ingest --reading-json <path>
```

and `launch ship --changelog-json` wrote a space in after the opener:

```
abcd reading ingest --reading-json < path>
```

The payload on disk is correct; the mutation happens inside the ingest. The
damage is not cosmetic: `<path>` is an inert placeholder, while `< path>` is a
valid shell input redirection from a file named `path`. A reader who copies the
documented invocation out of the release record runs something that means
something else, and fails in a way that does not point back at the changelog.

`-->` has the same shape (rewritten to `-- >` on the next line), so any prose
documenting an HTML comment or an arrow token is exposed the same way.

## Grounds

- deferred: not fixed at the point of discovery, because the finding surfaced
  mid-release-cut and the fix touches a primitive that every store-before-commit
  redactor routes through; the v0.7.0 line was reworded to avoid the placeholder
  instead, which leaves the defect intact and recorded rather than half-fixed
  under release pressure

## Candidate remedies

Not decided; the choice wants a maintainer.

- Skip the neutralisation inside backtick spans, parsing the span boundaries
  first. Most faithful, and the most code: the scrubber gains a tokenizer, and a
  malformed or unbalanced span has to fail closed rather than open.
- Substitute a lookalike that cannot open a tag, such as the full-width `＜`, or
  an HTML entity where the destination is known to render markdown. Cheap, but it
  silently alters a string a reader may copy, which is the same class of defect
  as the one it replaces.
- Leave the primitive alone and refuse at ingest instead: a changelog payload
  carrying `<` inside a code span is rejected with a message telling the composer
  to reword. Fails closed, adds no rendering knowledge to the scrubber, and keeps
  the correction where the prose is authored.

The third fits the existing shape best. `launch ship --changelog-json` already
re-derives the cut and proves the prose against it, so a payload refusal joins
the completeness bijection rather than forming a new gate, and the composer is
already the component that owns wording.
