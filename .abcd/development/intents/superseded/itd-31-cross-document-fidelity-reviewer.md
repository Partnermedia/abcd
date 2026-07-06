---
id: itd-31
slug: cross-document-fidelity-reviewer
spec_id: null
kind: standalone
bundle: null
suggested_kind: null
reclassification_history:
  - { date: 2026-05-07, from: bundle-member, to: standalone, reason: "Bundle dissolved when role-by-verb split landed (per command-structure-review round 2): three intent-review roles now have three distinct verbs (review/consistency/shape) under /abcd:intent, dissolving the unified-/abcd:audit-surface premise that the tier-0-audit-substrate bundle rested on. itd-32 superseded by this intent." }
  - { date: 2026-05-27, from: "standalone", to: "superseded", reason: "absorbed by itd-48" }
kind_at_supersession: standalone
superseded_by: itd-48
---

# The Brief and Every Intent Stay Consistent — Or Abcd Tells You Why They Don't

## Press Release

> **abcd ships `/abcd:intent consistency` — a Carmack-level *cross-document* role on top of the existing `intent-fidelity-reviewer`** that compares every intent against every other intent and against the brief, surfacing terminology drift, premise contradictions, scope leakage, sequencing impossibilities, naming reservation conflicts, schema/state contradictions, reference rot, and acknowledgement gaps. Where the `intent-fidelity-reviewer` checks one delivered reality against one shipped intent (its *single-document role*, surfaced via `/abcd:intent review`), this intent extends the same agent with a *cross-document role* surfaced via `/abcd:intent consistency`. The agent count stays 15; what grows is the reviewer's responsibilities. The mechanical part runs as `intent_lint.py --cross-doc` in CI (catches reference rot, naming collisions, schema contradictions, missing acknowledgements). The judgment part runs as a Carmack-level oracle review at every disembark, every `/abcd:intent plan`, and on demand. The verb is polymorphic on arg presence: bare scans the whole corpus; `/abcd:intent consistency <itd-N>` scans one intent's relationship with the rest. Findings land in `.abcd/logbook/audit/consistency-<ts>/report.{json,md}` with R1–R7 categorisation; severity drives whether the finding blocks promotion or just warns.
>
> "I'd written 30 intents over six months and didn't realise three of them silently disagreed about what 'session' meant," said Carol, product lead. "The first consistency review found it in two minutes. The lint catches reference rot every commit; the oracle review at disembark catches the harder semantic drift. Nothing slips."

## Why This Matters

The 2026-05-07 brief audit found 40 inconsistencies across the brief and 27 intents — 8 blockers, 13 warnings, 5 info. The user fixed all of them in one sweep, but several patterns stood out:

1. **Brief-edit ownership has no audit trail.** Every intent that touches the brief lists "brief edits" in scope. There's no canonical mechanism for tracking which edits an in-flight intent owes vs. has shipped.
2. **The "paper-only agent gets new responsibilities" pattern is unbounded.** Four intents committed `intent-fidelity-reviewer` to new responsibilities; the reviewer's documentation in five locations went stale.
3. **Cross-cutting registries have no SSOT enforcement.** Verdict enums, lint codes, agent catalog, spc-N space, glossary, persona registry — every consumer either re-states or extends in isolation.
4. **Terminology drift around 4 nouns** (`session`, `review`, `context`, `intent`).
5. **Reservation conflicts** (`spc-N` epic-ID space, command names).

These are not "bad intents" — they're failure modes of *coordination across many intents over time*. The intent-fidelity-reviewer (per itd-1) handles "did reality match this one intent?" but cannot handle "do these intents agree with each other and with the brief?". This intent fills that gap.

## What's In Scope

### The audit categories (R1–R7)

The audit categories are **empirical**, derived from the 2026-05-07 sweep. Future findings extend them; the initial set is:

- **R1 — Terminology drift**: same noun used with different meanings across docs. Mechanical detection via glossary cross-reference (consumes `.abcd/development/foundation/terminology/`); oracle judgement for un-glossed nouns.
- **R2 — Schema/state contradictions**: intent fields the template doesn't acknowledge; status mismatches; sequencing impossibilities (intent X depends on itd-Y before it has shipped). Mechanical (lint).
- **R3 — Reservation conflicts**: spc-N collisions, command-name reservations (per `02-constraints/04-naming.md`), duplicated agent/skill names. Mechanical.
- **R4 — Premise drift**: brief makes a claim subsequent intents have superseded without amending the brief. Oracle judgement.
- **R5 — Scope leakage**: intent X "in scope" overlaps intent Y "in scope"; "out of scope" of X is "in scope" of Y. Oracle judgement.
- **R6 — Reference rot**: cross-references to nonexistent files / sections / IDs / agents. Mechanical.
- **R7 — Acknowledgement gaps**: external source cited in intent but missing from README / Acknowledgements. Mechanical (regex over intent References sections + diff against README).

### The two halves of implementation

**Mechanical half** (`intent_lint.py --cross-doc`, CI-integrated):
- R3 spc-N collisions (compare every intent's `epic_id` against `.flow/epics/`)
- R3 command-name reservations (parse `04-naming.md` reserved list, scan all intents for unauthorised use)
- R6 reference rot (every `link` resolves; every `itd-N` / `spc-N` / `iss-N` / file-path / line-number cited resolves)
- R7 acknowledgement aggregation (every intent's References section's external citations appear in README's Acknowledgements; CI fails if missing)
- R2 schema (intent frontmatter against `intent.schema.json`)
- R2 sequencing (every intent's declared dependencies are scoped to the same phase or an earlier one)
- Lint codes `XD001`–`XD007` per `06-lint.md` namespace.

**Judgment half** (`/abcd:intent consistency`, oracle-driven):
- R1 terminology drift (where un-glossed nouns appear in 3+ docs with potentially-different meanings)
- R4 premise drift (oracle reads brief + all intents in two passes — shallow + deep — and surfaces architectural contradictions)
- R5 scope leakage (oracle compares "in scope" / "out of scope" sections across overlapping intents)
- Output: `.abcd/logbook/audit/consistency-<utc-ts>/report.{json,md}` (matches transparency invariant: JSON internal, MD render).
- Triggered: at every disembark Phase 0, at every `/abcd:intent plan`, and on demand.

### `/abcd:audit` is bare-verb, single-purpose

`/abcd:audit` is reserved per `02-constraints/04-naming.md` for compliance / hash-chain integrity (per itd-16; bare verb, single purpose). This intent's surface lives under `/abcd:intent`, not under `/abcd:audit` — the cross-document review is intent-corpus content fidelity, which belongs in the intent namespace alongside `/abcd:intent review` (Role 1) and `/abcd:intent shape` (Role 3). Each verb means one thing: `audit` = compliance; `review` = single-doc fidelity; `consistency` = cross-doc fidelity; `shape` = kind classification. The earlier proposal to bundle all audit roles under `/abcd:audit` (itd-32) was superseded by this split (see frontmatter `reclassification_history`).

### Brief edits

- **`05-internals/01-agents.md`**: extend the `intent-fidelity-reviewer` row to declare its second (cross-document) role alongside its existing single-document role. Document the two roles' distinct inputs/outputs/triggers in the same row (or as two adjacent sub-rows under the agent's name). Agent count stays at **15** — this is an extension of an existing agent, not a new one.
- **`05-internals/06-lint.md`**: register `XD001`–`XD007` codes (already reserved).
- **`02-constraints/04-naming.md`**: confirm `/abcd:audit` is reserved for compliance / hash-chain only (bare verb; sub-verbs may be added later if compliance scope expands).
- **`04-surfaces/02-disembark.md` § Phase 0**: add `/abcd:intent consistency` to the disembark prerequisites.

### Composition with lint codes added 2026-05-08

itd-31's Role 2 (cross-document fidelity) coexists with several lint families introduced by ideas 1-4. The cross-document role does NOT duplicate these mechanical checks — those run in `intent_lint.py` at plan-review. Role 2's job is the *judgement* layer over the corpus: terminology drift across multiple intents (R1), premise drift in the brief (R4), scope leakage between intents (R5). The mechanical lint codes Role 2 should be aware of (and cross-reference in its findings):

- **`MQ001` / `MQ002`** (memory quotation budgets, per itd-36) — if a memory page violates per-page or cumulative quotation, that's mechanical; Role 2's interest is whether the *pattern* of violations across pages indicates a corpus-wide source-overuse problem.
- **`MS001` / `MS002`** (memory source-class lints, per itd-36) — `MS002` (cross-class without weighting note) and Role 2's R1 (terminology drift) intersect: a page mixing `external_pdf` and `session_memory` without weighting may produce R1 findings if the source classes use the same terms with different meanings.
- **`ML001`** (memory licence missing, per itd-36) — mechanical; Role 2's R7 (acknowledgement gap) intersects when an `external_*` source is cited but missing from the README acknowledgements.
- **`MG001`–`MG004`** (modification grammar, per itd-37) — `MG001`–`MG003` are mechanical (concreteness lint); `MG004` is the boilerplate-detection verdict (Role 1's discipline check, not Role 2). Role 2's interest: does the cross-epic `modification_grammar_<domain>` synthesis show that multiple epics' Modification Grammar sections describe the same pattern with conflicting rules? That's R4-shaped (premise drift across the discipline application).
- **`VR001`** (vocabulary-registration, per itd-37 Ripple) — mechanical; Role 2's R1 (terminology drift) intersects when a term is registered but the registration's meaning conflicts with another intent's usage.
- **`SD001`** (bare-command-as-render discipline, per the 2026-05-08 surface-discipline sweep) — mechanical; Role 2's R5 (scope leakage) intersects when a sub-verb's existence is structurally redundant with another command's bare invocation.

**Role 2 is plan-review-trigger-aware (added 2026-05-08 per idea-4 Frontier Awareness).** When Frontier Awareness ships, Role 2 acquires a new trigger surface — runs at `/flow-next:plan-review` (in addition to its existing on-demand and at-disembark triggers) to evaluate whether tasks assigned in the epic spec fall within the assigned agent's declared `capability_scope` (per itd-5's extension). The semantic judgement ("is THIS task within agent X's frontier?") is genuinely Role 2 cross-document territory — same shape as itd-31's existing R3 (premise drift) or R5 (scope leakage). The static set-membership check stays in `intent_lint.py`; Role 2 owns the borderline judgement.

## What's Out of Scope

- **Replacing `intent-fidelity-reviewer`'s single-document role**: this intent *adds* a second (cross-document) role to the same agent. The original single-document role (delivered-reality vs shipped-intent acceptance) is preserved unchanged. The two roles run on different triggers and write to different outputs; they are siblings within the same agent's responsibilities, not replacements.
- **Adding a 16th agent to the catalog**: explicitly rejected. abcd's agent count grows by user-facing responsibility, not by audit subtype. The reviewer's two roles share its prompt scaffolding, its oracle backend resolution, its receipts. Splitting them into two agents would duplicate prompt-quality infrastructure for no operational benefit.
- **Auto-fixing inconsistencies**: the review reports findings; humans (or sub-verbs like `/abcd:intent grill`) resolve them.
- **Cross-corpus consistency**: this audit is single-repo (this repo's brief + intents only). Cross-corpus synthesis is itd-25 (`/abcd:dredge`), different category.
- **Hash-chain / Merkle audit**: itd-16's job (`/abcd:audit` — bare verb, compliance / tamper-evidence).
- **Real-time consistency checks during intent authoring**: too expensive; defer if there's demand.

## Acceptance Criteria

- **Given** a project with 30 committed intents and a brief, **when** the user runs `/abcd:intent consistency`, **then** a report is written to `.abcd/logbook/audit/consistency-<ts>/report.{json,md}` containing findings categorised by R1–R7, each with severity (blocker / warn / info), file references, and a recommended resolution.
- **Given** an intent whose `epic_id` collides with another intent's `epic_id`, **when** `intent_lint.py --cross-doc` runs in CI, **then** XD003 (reservation conflict) is emitted as a blocker and CI fails.
- **Given** an intent's References section cites an external source (e.g., `mattpocock/skills (MIT)`), **when** the README's Acknowledgements section does NOT include that source, **then** XD007 (acknowledgement gap) is emitted as a warn.
- **Given** an intent uses a glossary-defined noun in a meaning that conflicts with the glossary entry, **when** `/abcd:intent consistency` runs, **then** R1 (terminology drift) is reported with the conflicting locations cited.
- **Given** the brief makes a claim subsequent intents have superseded, **when** the oracle audit runs, **then** R4 (premise drift) is reported with the brief location and the superseding intents listed.
- **Given** `/abcd:disembark to <path>` runs, **when** Phase 0 reaches the audit step, **then** `/abcd:intent consistency` runs automatically and any blocker-severity findings refuse the disembark with clear remediation guidance.
- **Given** `/abcd:intent plan <itd-N>` runs, **when** the promotion gate executes, **then** `intent_lint.py --cross-doc` runs against the candidate intent and any blocker-severity finding refuses promotion.
- **Given** the audit categories are extended (a new R8 lands), **when** the brief is updated, **then** `06-lint.md` records the new lint code numbers AND `04-surfaces/05-intent.md` (or wherever the audit role is documented) reflects the new category.

## Open Questions

- **Two-pass vs one-pass oracle audit**: the 2026-05-07 manual sweep used two passes (shallow + deep) — is that the right shape for the automated version, or one combined pass?
- **Cost cap for oracle judgement half**: oracle audit over 30+ intents at deepest setting can be expensive. Token budget? Auto-downgrade if budget tight?
- **Mechanical vs judgement boundary**: some categories (R2 sequencing) sit in between — the lint can detect "intent A declares a dependency on intent B" but only the oracle can judge "is this dependency real or accidental?". Where exactly does mechanical end?
- **Output retention**: do consistency reports accumulate in `.abcd/logbook/audit/` indefinitely, or is there a retention policy (last N, last N days, prune on epic close)?
- **Integration with `/abcd:intent grill`** (itd-27): when grill is sharpening an intent, should it also surface relevant consistency findings for that intent? Probably yes; defer to plan-review.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/. (Ironic in this case: the cross-document reviewer will be retroactively run against this intent by itself once shipped.)_

## References

- 2026-05-07 manual audit (this session): produced the R1–R7 categorisation. Findings live in `.specstory/` history.
- Coordinates with: `itd-1` (acceptance gates), `itd-16` (hash-chain audit, sibling top-level verb `/abcd:audit`), `itd-27` (`/abcd:intent grill` sub-verb — consumes glossary the same way), `itd-34` (three intent kinds — Role 3 / shape classification surfaces under `/abcd:intent shape`).
- Supersedes `itd-32` (audit-role taxonomy): the unified-`/abcd:audit`-surface premise that itd-32 rested on dissolved when the three review roles got three distinct verbs (review/consistency/shape under `/abcd:intent`). itd-32 was retired via supersession on 2026-05-07 (file moved to `superseded/itd-32-audit-role-taxonomy.md`; frontmatter records `superseded_by: itd-31`).
- Builds on: `intent-fidelity-reviewer` (per itd-1). This intent extends the same agent with a second (cross-document) role; agent count stays at 15.
- `06-lint.md` reserves `XD001`–`XD007` for this intent's lint codes.
- Coordinates with: `itd-42` (coherence-aware grill) — different register: this intent reviews delivered documents for drift; itd-42 grills an intent for coherence before it is planned.
