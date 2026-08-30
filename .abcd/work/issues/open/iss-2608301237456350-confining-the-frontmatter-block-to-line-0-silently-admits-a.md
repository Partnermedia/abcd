---
schema_version: 1
id: "iss-2608301237456350"
slug: "confining-the-frontmatter-block-to-line-0-silently-admits-a"
severity: "minor"
category: "security"
source: "user-observation"
found_during: "itd-183-round-9-security"
found_at: "internal/core/reading/project.go"
---

confining the frontmatter block to line 0 silently admits a delimited block carrying an excluded key that the manifest still asserts was excluded

Found by the round-9 adversarial security review of build/itd-183.
REGRESSION INTRODUCED BY 2225d6cb, and a DELIBERATE one -- flagged as a
documented trade to re-confirm, not as a defect.

`firstBlockRange` refuses to open a block unless line 0 starts with `---`
(BOM allowed):

    if !strings.HasPrefix(frontmatter.TrimBOM(lines[0]), "---") { return 0, 0, false }

So an admitted .md whose first line is blank, a space, or a preamble, followed
by `---` / `origin: <value>` / `---`, reaches the bundle intact. Verified
admitted on HEAD and refused on the parent for a leading blank line, a leading
`<!-- x -->`, and a leading space.

Why it is defensible: that region is body prose to `frontmatter.Fields`,
`Duplicates` and `site.StripFrontmatter` alike, so it would have travelled as
prose either way. The parent's behaviour was a REFUSAL, not a redaction.

Why it still bites: the blind reader is not this binary. An LLM reading the
bundle reads `\n---\norigin: ...\n---\n` as frontmatter, and the manifest tells
it that key was excluded. The commit's claim "Agreeing loses no key" is scoped
to this binary's own readers, not to the reader the bundle exists for.

FOR THE FACILITATOR -- two acceptable resolutions:
(a) keep the change and restate the claim honestly in the commit message and
    the charter: "a block preceded by anything is body prose to this binary and
    travels"; or
(b) open the block at the first NON-BLANK line rather than at byte 0 (skipping
    blank/whitespace-only lines and the BOM). That recovers the leading-blank
    and leading-space cases without reintroducing the thematic-break false
    refusal the change was made to close, since a `---` after real prose still
    opens nothing.
