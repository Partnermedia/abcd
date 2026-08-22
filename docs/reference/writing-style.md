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

The machine-checkable subset of the punctuation rules below is staged in the
development record as `itd-141`; until that lint ships, those rules are
`review`.

## Language

| Rule | Enforcement |
|---|---|
| User-facing prose is British English. | review |
| Code-side text is US English: identifiers, code comments, strings, flags, JSON keys. | review |

The split runs at the code boundary, not the file boundary; a British-English
reference page documenting a `--color` flag spells the flag as the code does.

## Tense

| Rule | Enforcement |
|---|---|
| Docs state present reality: what is, never what was or what is planned. History lives in git; rationale lives in the development record. | review |
| Change-narration tokens ("previously", "formerly", and their kin) are blocked in `docs/` and `README.md`. | machine-enforced (the `present_tense/*` docs-lint family) |

## Page types

| Rule | Enforcement |
|---|---|
| Documentation follows [Diátaxis](https://diataxis.fr/): one type per page (tutorial, how-to, reference, explanation), and each `docs/` folder holds one type. | review |

## Punctuation

| Rule | Enforcement |
|---|---|
| Em dashes are allowed in running prose. | n/a (permission, not a check) |
| Em dashes are not used inside list items: A pivot in a list item takes a colon instead. | review (lint staged in `itd-141`) |
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

## Escapes

A line that legitimately breaks a machine-enforced rule carries
`<!-- docs-lint: allow -->` with a reason; the escape is deliberate and
reviewable, never a default.
