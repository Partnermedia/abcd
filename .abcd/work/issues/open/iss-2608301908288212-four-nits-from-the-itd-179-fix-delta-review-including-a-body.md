---
schema_version: 1
id: "iss-2608301908288212"
slug: "four-nits-from-the-itd-179-fix-delta-review-including-a-body"
severity: "nitpick"
category: "tech-debt"
source: "user-observation"
found_during: "itd-179-fix-delta-ruthless"
found_at: "internal/core/grounds/record.go"
---

four nits from the itd-179 fix delta review including a body line offset that is one ahead of the rendered body

Four nits from the itd-179 fix-delta ruthless review, settled at the ship commit.

1. `record.go` / `frontmatter.go` -- "body line N" counts from
   `frontmatter.Split`'s body, which keeps the blank line after the closing
   delimiter, while `capture`'s reader strips it. So the number is one ahead of
   the body a reader renders. The quoted opener text is the reliable locator;
   `Split`'s doc claim that writer and readers judge the same bytes is true of
   the SECTION and off by one line of the offset.
2. `record.go` and `intent/grounds.go` -- two doc claims say a bare body may be
   passed and comes back unchanged. No live caller passes one, and `Body` does
   not in fact take either: a text opening with a thematic break reads as a
   frontmatter opener and 67 bytes including the whole section are eaten.
   Unreachable today; the claim invites the caller that would reach it.
3. `record.go` -- the ambiguity refusal always says "`## Grounds` headings", but
   the pattern matches any depth, so a record with a `### Grounds` subsection is
   refused by a message naming a level it does not have.
4. `record.go` -- the frontmatter-only splice gains a blank line the whole-file
   form did not produce. Cosmetic; no real record has an empty body.
