---
id: itd-183
slug: the-cold-reading-sees-exactly-what-the-assembler-passes-posi
spec_id: spc-61
kind: bundle-member
suggested_kind: bundle-member
reclassification_history: []
builds_on: [itd-86]
severity: major
impact: additive
---

# The cold reading sees exactly what the assembler passes — positive inclusion, field projection, and a per-run manifest

Typed links: consumes
[adr-55](../../decisions/adrs/0055-the-construal-stands-in-the-record-its-history-does-not.md)
(every framing item on the include list cites the rule admitting it);
settles itd-86's central open question (blindness is structural, not
instructed).

## Press Release

> **Blindness becomes a property of the input, not a promise of the
> reader.** The location tiering is organisational, not an access control —
> nothing today prevents a reading reaching ledger content. The assembler
> closes that: it names what it includes (a record type added later,
> including one the instrument itself adds, is excluded by default rather
> than included by oversight), and it projects fields out of files rather
> than copying paths — a shipped intent holds both the claim record and its
> Audit Notes, and only the former may travel. Every run emits a manifest —
> what was passed, by path and field, hashed, with the run
> identifier — so a reader can judge contamination rather than accept a
> disclosure on trust. Invocation carries no free text: the operator
> supplies a position and a target state, and the reading's object and
> question come from its definition — there is no channel through which
> ledger content can travel in the framing of a request, because there is
> no prose input.

## What's In Scope

- **Include list:** the brief's `01-product/` and `02-constraints/`
  chapters — the framing section's construal statement included, the
  evidence chapter deliberately not (open questions and settled dead ends
  are deliberation, and a future brief-homed warm record must not walk in
  as "brief text"); `brief/glossary/`; `intents/shipped/` projected to
  press release, acceptance criteria, scope conditions, mechanism claim and
  `spec_id`; `intents/disciplines/`; `specs/`; the shipped tree (defined
  below). The chapter-level bound (03-evidence excluded as deliberation) was agreed at the maintainer review of 2026-08-28.
- **The include list is per-position** (maintainer readings design, 2026-08-28):
  assembly follows the invoked definition's object over the shared
  exclusion floor. The stated asymmetry: the widening reading excludes
  `intents/drafts/` and `intents/planned/` (they are the candidate set it
  is asked to widen); the entailment reading includes them (articulation
  precedes selection). The comparative reading's object is the widening
  reading's pre-admission output plus the selection-criteria discipline —
  within one cycle, before admission; a prior run's stored output stays
  read-blocked.
- Two assembler rules (ruled 2026-08-28): **no include may name a
  directory containing a record family** — "the shipped tree" is scoped
  to source, tests, documentation and configuration, `.abcd/` excluded
  wholesale, and record paths a reading legitimately needs named
  individually, so a family added later (the readings family included) is
  excluded by construction; and **a reading's object excludes the
  material whose state that reading exists to change** — the drafts
  asymmetry and the Audit-Notes exclusion are its instances, and the
  include list becomes derivable rather than remembered.
- **Exclude, and assert the exclusion:** `origin`; production mode; Audit
  Notes; scope-condition dispositions; `brief/03-evidence/`; `decisions/`;
  `roadmap/rfcs/`; `intents/superseded/`; `work/issues/` in every state,
  reading records and dispositions included; `plans/`; `research/notes/`;
  session transcripts; manifests; admission and selection grounds; the
  lapse log. Each exclusion item carries its detectable signal: frontmatter
  keys (`origin`, production mode) by key; `Audit Notes` by heading;
  directories by never appearing in the positive include walk; dispositions
  and grounds as record types in excluded paths, their fields dropped by
  projection. Prose-borne warmth inside an included chapter has no
  structural signal — that is what the chapter-level include bound and the
  glossary discipline carry, and it is disclosed as residue, not claimed as
  caught.
- **The manifest**, per run, on the render-without-writing idiom of
  `disembark plan` for dry runs; at ingest it is committed to the durable
  tier at `.abcd/development/readings/<run-id>/`, alongside the run
  record — a new record family (ruled 2026-08-28, superseding the earlier
  working-tier proposal: lifecycle selects the tier — a manifest is
  written once and never amended; a commit reference plus per-item hashes
  lets a third party re-run the assembler and diff the output, the
  strongest form of the checkability claim; and it carries paths, field
  names and hashes, never content, so nothing needs redaction). Manifests
  are warm on the next run: inside the read-block, excluded by rule.
- **The evals that guard this assembler** — the read-block eval
  (sentinel-planted warm content must never appear in output) and the
  amnesia eval (the same state assembled twice is byte-identical) — are
  separate instruments with their own claims and verdicts, split to their
  own drafts (the read-block eval and the amnesia eval) per the maintainer review of 2026-08-28.

## Resolutions implemented (2026-08-27; ratified 2026-08-28)

- **The reader-side boundary is owned here.** Passing curated input to a
  reader that can open the repository is instructed blindness with extra
  steps. The invocation constructs a context-isolated reading: the
  assembled input is the reading's working set and no repository path
  enters its context. This is the enforcement half of
  invariant 15 (adopted 2026-08-28); the definitions assert the posture, this
  intent builds the wall.
- **"The shipped tree" is defined**: the delivered artefact — source,
  tests, `docs/`, the root README — excluding `.abcd/**`, `agents/`,
  `evals/`, and the instrument's own package. The reading never receives
  its own definition, the eval that guards it, or the assembler's include
  table.
- **Manifest warmth has one ground**: the manifest's *content* is cold (it
  enumerates passed cold items only, hashed, no timestamps), which is why
  committing it for reader audit is safe; the manifest as *evidence* is
  warm (it reveals run timing and target selection), which is why it sits
  inside the read-block and no future reading receives it. Both halves are
  stated, neither is traded away.
- **Comparative candidate ids — superseded by M8 (the maintainer rulings of 2026-08-28):** the
  ruled interface carries position and target state only. The ids-only
  candidate list proposed 2026-08-27 is retained solely as the recorded fallback, if testing against the four definitions shows the evaluative
  position needs candidates at invocation: restricted to the intent
  family, shape-validated, never prose.

## What's Out of Scope

- The reading definitions and the output contract — sibling intents.
- Preventing timing and target-selection leakage: the operator still
  chooses when to run a reading and what to point it at, which can carry
  information. The manifest and session records make this visible after the
  fact; they do not prevent it. Disclosed residue, not a gap to be closed
  here.

## Scope Conditions

None stated.

## Acceptance Criteria

- **Given** a repository state, **when** the assembler runs, **then** its
  output contains no field on the exclusion list.
- **Given** a new record type added under `.abcd/development/`, **when**
  the assembler runs, **then** that type is absent from the output without
  any change to the assembler.
- **Given** the invocation interface, **when** it is inspected, **then**
  it accepts a position and a target state, and nothing else — no
  free-text argument anywhere (ruled, maintainer 2026-08-28). Whether the evaluative position needs its candidates or criteria at
  invocation is tested against the four definitions before the interface
  freezes; the anticipated revision, if the need is shown, is a
  shape-validated record-id list, never prose.
- **Given** an assembler run, **when** the manifest is emitted, **then**
  every item passed appears with its path, its field where projection
  occurred, and a hash — and a reader can determine that a
  named excluded field was not passed.
- **Given** a reading invocation, **when** its context is constructed,
  **then** it contains the assembled input and no repository path — the
  blindness is enforced by construction, not instructed.

## Rulings (2026-08-27, as revised 2026-08-28)

- **Manifest home: the durable tier** (revised by the maintainer rulings
  of 2026-08-28, decision log). Written at ingest to
  `.abcd/development/readings/<run-id>/`, alongside the run record — a
  new record family, excluded by this assembler like every record
  family; the local render remains for
  dry runs. Grounds: the manifest's purpose is that a reader can judge
  contamination rather than accept a disclosure on trust, and a gitignored
  manifest delivers that only on the machine that ran it. It is safe to
  commit because it enumerates the passed (cold) items only.
- **No free text at any position — ruled as M8 (the maintainer rulings of 2026-08-28),
  stricter than proposed:** position and target state only. The ids-only
  comparative argument proposed here is superseded; it survives only as
  the recorded fallback if the evaluative position proves to need candidates
  at invocation.

