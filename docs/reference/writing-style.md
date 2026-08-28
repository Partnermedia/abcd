# Writing style

The canonical style guide for prose committed to this repository — one
reference page holding every writing rule, which other surfaces point at
rather than restate.

Every rule carries an **enforcement** label, and that label is a fact about
shipped machinery, never an aspiration:

- **machine-enforced**: A shipped `abcd docs lint` rule checks it; the label
  names the rule id.
- **review**: Humans (and reviewing agents) check it. A rule stays labelled
  `review` until its lint ships, however firmly it is agreed.

The colon and semicolon casing rules below are staged in the development
record as `itd-141`; until that lint ships, those rules are `review`. The
list-item em-dash rule is machine-enforced as a banned token, with the one
residual case its line-based pattern cannot see noted in its row.

## Language

| Rule | Enforcement |
|---|---|
| User-facing prose is British English. | review (the `spelling/*` docs-lint family flags common US spellings as advisory warns; the label stays `review` because warns gate nothing) |
| Code-side text is US English: identifiers, code comments, strings, flags, JSON keys. | review |

The split runs at the code boundary, not the file boundary; a British-English
reference page documenting a `--color` flag spells the flag as the code does. <!-- docs-lint: allow — the flag name is code-side text -->

## Tense

| Rule | Enforcement |
|---|---|
| Docs state present reality: what is, never what was or what is planned. History lives in git; rationale lives in the development record. | review |
| Change-narration tokens ("previously", "formerly", and their kin) are blocked in `docs/` and `README.md`. <!-- docs-lint: allow — this row documents the banned tokens themselves --> | machine-enforced (the `present_tense/*` docs-lint family) |

## Page types

| Rule | Enforcement |
|---|---|
| Documentation follows [Diátaxis](https://diataxis.fr/): one type per page (tutorial, how-to, reference, explanation), and each `docs/` folder holds one type. | review |

## Punctuation

| Rule | Enforcement |
|---|---|
| Em dashes are allowed in running prose. | n/a (permission, not a check) |
| Em dashes are not used inside list items: A pivot in a list item takes a colon instead. | machine-enforced (`punctuation/em-dash-in-list-item`); an em dash on a list item's wrapped continuation line is the residual `review` case until the `itd-141` lint ships |
| After a colon: A capital letter. | review (lint staged in `itd-141`) |
| After a semicolon: lower case (the semicolon joins clauses of one sentence). | review (lint staged in `itd-141`) |

The staged lints mask code spans and fenced blocks; a `--flag: value` example
never trips a prose rule.

## Structure

Checks that already ship in `abcd docs lint`, listed here so the page states
the whole machine-enforced surface:

| Rule | Enforcement |
|---|---|
| Relative links resolve. | machine-enforced (`links_resolve`) |
| No stray root-level markdown beyond the allowlisted set. | machine-enforced (`stray_root_docs`) |
| Citations are well-formed, crosswalked, and within source policy. | machine-enforced (the `citation_*` rules) |
| User-facing prose stays host-agnostic; naming a specific tool is confined to attribution. | machine-enforced (the `harness/*` rules) |
| Reserved names the repository has banned do not appear (e.g. a record file naming itself). | machine-enforced (the verb-managed `names/*` rules) |

## Escapes

A line that legitimately trips a **banned-token** rule — the `present_tense/*`,
`punctuation/*`, `spelling/*`, `harness/*` and `names/*` families — carries
`<!-- docs-lint: allow -->` with a reason (a quoted title keeping its original
spelling, a code-side name, attribution); the escape is deliberate and
reviewable, never a default. The other
machine-enforced rules have no line escape: `links_resolve`, `stray_root_docs`
and the `citation_*` rules are satisfied by fixing the link, the file placement,
or the citation itself, not by annotating the line.
