# Terminology Glossary

<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->

This directory contains the canonical terminology glossary for the abcd project. Each term is a
Markdown file with YAML frontmatter that defines the term's meaning, bounded context, and usage
constraints.

---

## Format Specification

Every term file MUST begin with a YAML frontmatter block (`---` delimiters) conforming to
the terminology schema (`internal/core/schema/terminology.schema.json`). The body of the file (below the closing `---`)
provides narrative context, examples, and cross-references.

### Required Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `term` | string | Canonical lowercase name (no spaces) |
| `bounded_context` | string | Lowercase identifier matching the parent directory name. Current contexts: `core`, `interview`. New contexts are added by creating a new subdirectory — no schema edit needed. |
| `definition` | string (≥10 chars) | Precise, unambiguous definition |
| `aliases` | array | Acceptable alternative names |
| `forbidden_synonyms` | array | Words that MUST NOT substitute for this term |
| `status` | enum | `draft`, `stable`, or `deprecated` |
| `introduced_in` | string | Version or intent ID when this term was coined |

### Optional Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `starts_when` | string\|null | Condition/event that initiates this concept (lifecycle terms) |
| `ends_when` | string\|null | Condition/event that concludes this concept (lifecycle terms) |
| `not_to_be_confused_with` | string\|null | Related term that is commonly confused with this one |
| `versions` | array\|null | Version history records for the term definition |

### Constraints

- `forbidden_synonyms` and `aliases` MUST NOT overlap.
- When `status` is `stable` and either `starts_when` or `ends_when` is set, both must be
  present and non-null.
- `bounded_context` must match the name of the parent directory.

---

## Directory Layout

```
glossary/              ← .abcd/development/brief/glossary/ (the home, per adr-30)
├── README.md          ← this file (format spec + index)
├── _template.md       ← copy this when adding new terms
├── core/              ← terms that apply across all abcd bounded contexts
│   ├── brief.md
│   ├── intent.md
│   ├── spec.md
│   ├── oracle.md
│   ├── persona.md
│   ├── phase.md
│   ├── voyage.md
│   ├── transport.md
│   ├── lifeboat.md
│   └── disembark.md
└── interview/         ← terms specific to the grill/interview sub-verb
    ├── session.md
    └── embark.md
```

---

## Validation

The terminology lint (`internal/core/lint`) validates the term files against the
schema. Run it via the abcd CLI over the glossary directory to check every term:

```bash
abcd lint terminology .abcd/development/brief/glossary/
```

Or over a single term file:

```bash
abcd lint terminology .abcd/development/brief/glossary/core/brief.md
```

---

## Adding a New Term

### Manually

1. Copy `_template.md` to the appropriate subdirectory (`core/`, `interview/`, or a new context directory you create).
2. Fill in all required frontmatter fields.
3. Write a narrative body below the closing `---`.
4. Run the linter to verify the file passes validation.
5. Add the term to the index table below.

### Via `/abcd:intent grill` (glossary-aware mode)

When a project has this `terminology/` directory, `/abcd:intent grill` enters
**glossary-aware mode** and can write new term files inline during the grill session.

**How grill writes back:**

- When the interview surfaces a new noun the user wants to pin, grill offers to write it
  immediately (never batched to the end of the session — per Pocock's `/grill-with-docs`
  pattern). The user confirms the term name, bounded context, and definition before any write.
- New terms are written to `terminology/<bounded_context>/<term>.md` using **atomic write**
  (`<file>.tmp` + POSIX `rename(2)`). A `kill -9` mid-session cannot corrupt existing term files.
- All grill-written terms receive `status: draft` and `introduced_in: <current-intent-id>`.
- Grill also detects **forbidden synonyms** in the intent body and proposes canonical replacements
  (using canonical display names in body prose, qualified IDs in `glossary_terms_used` only).
- When a term exists in multiple bounded contexts, grill asks the user to disambiguate and
  optionally sets `contexts: [...]` on the intent frontmatter.

**Body-prose vs machine-field distinction:**

Grill NEVER writes qualified IDs (`core/persona`) into intent body prose. Body prose uses
canonical display names (`persona`). Qualified IDs appear only in machine fields
(`glossary_terms_used`, ADR cross-refs, lint output). The optional inline form
`[persona](glossary:core/persona)` is the only way a qualified ID may appear in body prose.

**ADR offers:**

Grill offers to draft an ADR only when all three Pocock clauses pass:
hard-to-reverse + surprising-without-context + real-trade-off. If any clause fails, no offer
is made. ADRs are written to `.abcd/development/decisions/adrs/` with atomic write.

The complete write-back protocol is a **design target** of `/abcd:intent grill`'s glossary-aware mode; no shipped file documents it yet.

---

## Term Index

### core/

| Term | Status | Definition (summary) |
|------|--------|----------------------|
| [brief](core/brief.md) | stable | Root document defining project purpose and constraints |
| [intent](core/intent.md) | stable | Press-release-shaped feature description before implementation |
| [oracle](core/oracle.md) | stable | AI model invoked to review or reason over artefacts |
| [persona](core/persona.md) | stable | Placeholder stakeholder character from the personas registry |
| [phase](core/phase.md) | stable | Discrete layer of the build sequence, grouping related specs |
| [spec](core/spec.md) | stable | Specced block of work implementing one or more intents |
| [voyage](core/voyage.md) | stable | Operations namespace at `~/.abcd/voyage/<source-root-sha>/` — the append-only record of disembark and embark runs against a source repository |
| [transport](core/transport.md) | stable | Mechanism by which context is delivered to an oracle |
| [lifeboat](core/lifeboat.md) | stable | Portable artefact packed by `/abcd:disembark` to transfer project knowledge across context boundaries; it always lands outside the source repository, at an operator-chosen destination |
| [disembark](core/disembark.md) | stable | The act of packing a lifeboat: distils a source project's settled artefacts, decisions, and configuration into a portable directory |

### interview/

| Term | Status | Definition (summary) |
|------|--------|----------------------|
| [session](interview/session.md) | stable | One interactive exchange in a grill sub-verb invocation |
| [embark](interview/embark.md) | stable | Opening move that initiates a grill session |

---

## Allowlist entries

The top-level `terminology_exclude_files` array in `.abcd/config.json`
lists files that must not be scanned for forbidden-synonym hits even when a future terminology
linter is wired up. Each entry below names the file, the controlling intent, and the reason it
is intentionally allowlisted.

- the issue-schema validator's **negative-test corpus** fixture (spc-18) — this fixture is the
  negative-test corpus for the
  issue-schema validator: it intentionally embeds strings that would otherwise be flagged as
  forbidden synonyms in order to assert that the validator rejects them. Linting the negative
  corpus for those same strings would produce a guaranteed false positive on every CI run.
  The entry is **inert documentation** — no `internal/core/lint` consumer reads
  `terminology_exclude_files` yet. Recording the allowlist artifact means the contract is
  pinned to a single named location before any consumer is wired in.

## Acknowledgements

Term file format adapted from [mattpocock/skills](https://github.com/mattpocock/skills) (MIT licence).
The `/abcd:intent grill` sub-verb's glossary capture pattern draws from the `grill-with-docs` skill
in that repository.
