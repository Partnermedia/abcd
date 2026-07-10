# `/abcd:intent` — Press-Release Intent Capture

abcd uses **intents** in press-release format (Amazon working-backwards) as the unit of forward-looking *user-facing* planning. Intents capture *what user-facing capability exists once shipped*, written in present tense as if already delivered. This is engineered to discipline product clarity before scope creep — a reader of an intent thinks like a product person first, an engineer second.

Plumbing work (adapters, agents, harness, scaffolding) lives in this brief, not in intents — see [`01-product/03-mental-model.md`](../01-product/03-mental-model.md) for the rationale.

Intents live at `.abcd/development/intents/{drafts,planned,shipped,disciplines,superseded}/`. Five directories encode lifecycle position. Three are for press-release-shaped intents (the standalone + bundle-member kinds); one is for disciplines (no press release, no spec); one is for intents killed by reclassification.

- **`drafts/`** — press-release-shaped intent captured but no native spec yet. Bench of ideas / forward-looking work. Cheap to draft and discard.
- **`planned/`** — a committed capability, scoped into a roadmap phase and awaiting its Go build. Its `spec_id` is `null` (unscheduled) or points at a `spc-N` once the spec layer schedules it (Phase 4). The native spec store ([adr-26](../../decisions/adrs/0026-native-spec-layer-ccpm-backend.md)) is the scheduling home. Bundle-member intents in `planned/` share a `spec_id` with their bundle-mates.
- **`shipped/`** — a capability built in Go. Empty until the Go build ships capabilities into it (Phase 1 onward). The intent's "Audit Notes" section holds drift findings (per-criterion verdicts: `MET`, `MET_WITH_CONCERNS`, `NOT_MET`, `INCONCLUSIVE`) once `intent-fidelity-reviewer`'s Role 1 has run on it via `/abcd:intent review <itd-N>`; an intent moves here when its linked spec closes.
- **`disciplines/`** — discipline-kind intents (cross-cutting rules with no user moment). They never get a native spec of their own; instead they impose acceptance gates that every *other* spec inherits and is checked against. Disciplines have no `status` frontmatter — presence in this directory IS the active state. Superseded disciplines move to `superseded/`.
- **`superseded/`** — intents killed by reclassification or absorption (e.g., when a smaller intent is folded into a larger one, or a discipline is replaced by a stricter successor). The file records `superseded_by: <itd-N>` (the successor) AND `kind_at_supersession: <original-kind>` (what shape the intent had when retired — standalone vs bundle-member vs discipline). Preserved as historical record; never deleted.

There is no `active/` state — "active" is implicit (a planned intent's linked spec is currently in flight in the native spec store; an active discipline is any intent in `disciplines/`).

## 1. Intent IDs, kinds, and lifecycle

### Intent IDs

- **ID format:** `itd-N` (unpadded — e.g., `itd-1`, `itd-15`). Mirrors the spec store's `spc-N` format. Filenames: `itd-N-<slug>.md`. Lexical-vs-numeric sort handled at the tool layer (`internal/core/lint`, registries) rather than via filename padding.
- **The low IDs (itd-1..itd-7) reflect an early one-time rebase to ordering signal.** Intents created since are capture-stable, picking up at itd-27+. ID number is *not* an execution-order guarantee — the canonical build order is the phase plan at [`roadmap/phases/`](../../roadmap/phases/README.md).
- **An intent carries no release or sequencing field.** Per [adr-9](../../decisions/adrs/0009-phase-as-product-layer.md), an intent's sequencing is its *phase membership*, recorded editorially in the owning phase doc's `## Scope` — not in intent frontmatter. An intent not yet listed in any phase doc is implicitly unscheduled (a `drafts/` bench item). "Which release" is an output of completing phases, never an input stamped on a draft; the former `target_release` field was removed for this reason.

### Intent kinds (per [`01-product/03-mental-model.md`](../01-product/03-mental-model.md))

Every intent has a `kind` declared in frontmatter, set at `/abcd:intent plan` time. Three kinds:

| `kind` | Has press release? | Lives in | Maps to | Examples |
|---|---|---|---|---|
| `standalone` | Yes | `drafts/` → `planned/` → `shipped/` | One spec (1:1) | itd-3, itd-4, itd-7, most of the corpus |
| `bundle-member` | Yes | Same as standalone, with `bundle: <id>` linking members | Shared spec with bundle-mates (N:1) | (no live bundles in any phase — see [`intents/README.md`](../../intents/README.md#bundles) for retired bundle history) |
| `discipline` | **No** — uses `## Rule` instead | `disciplines/` | No spec; imposes acceptance gates on every other spec | itd-1 (AC gate), itd-5 (prompt-quality) |

**`kind` is binding once set at plan time.** Late changes go through `/abcd:intent reclassify <itd-N> --kind <new-kind>`, which records the change in the intent's frontmatter `reclassification_history` and surfaces it for reviewer review.

**Two distinct history fields, two distinct concerns:** `reclassification_history` records *kind* transitions (standalone ↔ bundle-member ↔ discipline ↔ superseded). `surface_history` records *surface-shape* transitions where the kind is unchanged but the user-facing surface form changes (e.g., skill → sub-verb, top-level command → sub-verb of another command, command → flag). Both are append-only; both have the same `{ date, from, to, reason }` shape. Worked example: itd-27 was always `kind: standalone`, but its surface shifted from a top-level skill (`/abcd:grill`) to a sub-verb of `/abcd:intent` on 2026-05-07 — that's a `surface_history` entry, not a `reclassification_history` entry, because the kind is unchanged. The two fields together preserve a complete audit trail of an intent's evolution without overloading either.

**`suggested_kind` is advisory.** At capture time, the LLM classifier writes a soft hint (`suggested_kind: <kind>`) to frontmatter. This is informational only — the binding decision is at plan time.

**A fourth capture verdict, `decision`, routes to the ADR store (itd-44).** The capture-time classifier (`intent classify-capture-kind`) can also emit `decision` — a *standing infrastructure choice* (no user moment, not a per-artefact rule, e.g. "we use Postgres"). `decision` is a **capture verdict only, never a persisted `kind`**: the `kind` / `kind_at_supersession` enums stay three-valued (`standalone` / `bundle-member` / `discipline`). A confirmed `decision` DIVERTS capture to the existing ADR store (`.abcd/development/decisions/adrs/`, `adr-N-<slug>.md`) instead of writing an intent draft — no spec, no lifecycle directory, no `intents/decisions/`. The verdict is advisory: capture confirms "capture as an ADR?" or overrides to a normal draft carrying a *plannable* `suggested_kind` (`null`/`standalone`, never `decision`). `suggested_kind: decision` validates in the schema but is REFUSED by `plan_single`/`reclassify` ("decisions are not plannable"). See [itd-44](../../intents/drafts/itd-44-fourth-intent-kind-decision.md) and the [ADR store README](../../decisions/adrs/README.md).

**Bundle invariant: all members of a bundle MUST belong to the same phase.** A bundle ships as one shared spec; one spec belongs to one phase. Cross-phase bundles are structurally impossible — the shared spec cannot live in two phases at once. `/abcd:intent plan <itd-A> <itd-B> ...` (multi-arg, kind=bundle-member) hard-blocks promotion when the proposed members are scoped to different phases. The only resolutions are: (a) re-scope all members into the same phase before re-running plan, or (b) downgrade one or more members to `kind: standalone` so they ship independently. Worked example: the `intent-capture-discipline` bundle (itd-27 + itd-30) was retired on 2026-05-07 precisely because the two members were scoped to different phases — both intents reclassified to standalone (see their `reclassification_history` entries for the full reasoning). Lint code: `IL011` per [`05-internals/06-lint.md`](../05-internals/06-lint.md).

### Discipline format

Discipline-kind intents skip the press release entirely (no user moment to describe; no customer quote to attribute). They also have no `status` field — the directory IS the state (`disciplines/` = active; `superseded/` = retired). The format is:

```markdown
---
id: itd-N
slug: <kebab-case>
kind: discipline
kind_notes: "<free-text describing what kind of discipline this is — e.g.,
              'cross-cutting acceptance-criteria gate, applied via lint and auditor'>"
---

# <Headline — what rule this imposes on every spec>

## Rule

<One-paragraph statement of the rule, in present tense.>

## Why

<2-3 paragraphs: the failure mode this prevents; the cost of not having the rule;
the prior art (other projects' versions of this discipline).>

## What's In Scope

<Bullets — what every other spec must do/include because of this rule.>

## What's Out of Scope

<Bullets — what this discipline does NOT cover.>

## Acceptance Criteria

> _Required even for disciplines (per the itd-1 discipline itself). At least one
> Given-When-Then bullet describing how the rule is checked._

- **Given** <preconditions>, **when** <spec event>, **then** <gate behaviour>.

## Audit Notes

<Empty until first spec ships under this discipline. The shape-classification
role of intent-fidelity-reviewer populates findings here.>

## References

<Citations to brief sections, related intents, prior art.>
```

**Discipline subtypes come later.** The `kind_notes` field is free-text deliberately. This is a **deferred** capability — Discipline subtype taxonomy is NOT a live Role 3 `suggestion_type` (the three live types are `kind_change`, `bundle`, `supersession`). The subtype taxonomy (e.g., a closed enum of `methodology` / `documentation` / `audit` / `convention`) moves from free-text to formal enum when ANY of the following:

1. **Three or more disciplines exist** with `kind_notes` describing similar shapes.
2. **A user is confused** about which existing discipline a finding belongs to.
3. **`intent-fidelity-reviewer`'s shape-classification role cannot describe a proposed discipline** without more constraint than `kind_notes` provides.
4. **Cross-project usage** — once abcd is in three or more projects, comparing disciplines across them needs a shared subtype vocabulary.

Until then, `kind_notes` is the free-text descriptor.

### Lifecycle

- **The low IDs (itd-1..itd-7) reflect an early one-time rebase to ordering signal.** Intents created since are capture-stable, picking up at itd-27+. ID number is *not* an execution-order guarantee — the canonical build order is the phase plan at [`roadmap/phases/`](../../roadmap/phases/README.md).
- **An intent carries no release or sequencing field.** Per [adr-9](../../decisions/adrs/0009-phase-as-product-layer.md), an intent's sequencing is its *phase membership*, recorded editorially in the owning phase doc's `## Scope` — not in intent frontmatter. An intent not yet listed in any phase doc is implicitly unscheduled (a `drafts/` bench item). "Which release" is an output of completing phases, never an input stamped on a draft; the former `target_release` field was removed for this reason.
- **Lifecycle (automated, not user-managed):**

```
1. /abcd:intent "<free-text>"   (canonical bare quoted create)
   ├─ Interview captures press release (headline, what user does, customer quote, scope)
   ├─ Picks persona at random from .abcd/development/personas.json for the customer quote
   ├─ Assigns next itd-N ID (capture-stable once assigned)
   ├─ Requires `## Acceptance Criteria` section (per itd-1) — refuses to write if missing/malformed
   ├─ LLM classifier writes advisory `suggested_kind` to frontmatter (default: standalone)
   └─ Writes intents/drafts/itd-N-<slug>.md (no spec created yet)

2. /abcd:intent plan <itd-N> [<itd-M>...]    (when ready to commit to work)
   ├─ Reads prd_path (produced by the grill → PRD flow); refuses if null — no PRD yet, grill first
   │   (GR002 blocker; suppressed-as-info for prd_grandfathered intents). See § 5 for the full freeze sequence.
   ├─ Runs internal/core/lint: refuses to promote if `## Acceptance Criteria` is missing/malformed
   ├─ Reads suggested_kind + cross-references; proposes a kind (standalone / bundle-member / discipline)
   ├─ User confirms or overrides the kind (binding); written to `kind:` frontmatter
   │
   ├─ IF kind = standalone (single intent ID passed):
   │     ├─ Creates a stub spec in the native spec store for the intent
   │     ├─ Runs plan-review (Carmack-style review of the stub)
   │     ├─ Injects bidirectional link (spec.intent: itd-N; intent.spec_id: spc-N)
   │     └─ Moves intents/drafts/itd-N-*.md → intents/planned/itd-N-*.md
   │
   ├─ IF kind = bundle-member (multiple intent IDs passed):
   │     ├─ Creates one shared spec with all intents as joint input
   │     ├─ Runs plan-review for the combined plan
   │     ├─ Injects bidirectional link to shared spc-N (spec.intent: [itd-A, itd-B];
   │     │   each intent.spec_id: spc-N; each intent.bundle: <bundle-id>)
   │     └─ Moves all members from drafts/ → planned/
   │
   └─ IF kind = discipline (single intent ID, kind chosen explicitly):
         ├─ NO spec created — disciplines don't get their own spec
         ├─ Registers the rule in .abcd/disciplines/<itd-N>.json (acceptance gates that
         │   every subsequent spec inherits in its plan-review check)
         ├─ Runs plan-review on the discipline's `## Rule` itself for sanity
         └─ Moves intents/drafts/itd-N-*.md → intents/disciplines/itd-N-*.md
             (no `status` field — the disciplines/ directory IS the active state)

3. /abcd:intent ship <itd-N>          (kick off implementation; standalone + bundle only)
   ├─ If intent is in drafts/: runs full pipeline (plan + plan-review first)
   ├─ Kicks off implementation via the native ship/run seam
   └─ Returns; spec continues in the native spec store
   (Disciplines have no `ship` step — they are continuously active once in disciplines/.)

4. Spec marked done in the native spec store   (standalone + bundle: work complete)
   ├─ native spec-store `spec close` close-hook (spc-36) → intent lifecycle reconcile (spc-28)
   └─ Moves intents/planned/itd-N-*.md → intents/shipped/itd-N-*.md (+ enqueues a review)
       (For bundles, all member intents move together when the shared spec closes.)

   Then, as a separate MANUAL step — run /abcd:intent review <itd-N>:
   └─ intent-fidelity-reviewer agent (single-document role / Role 1 itd-1 pass)
       └─ Compares as-shipped reality against original press release + acceptance criteria
           └─ Per-criterion verdicts (MET / MET_WITH_CONCERNS / NOT_MET / INCONCLUSIVE)
               written to the intent file's "Audit Notes" section (verdict of record),
               plus a per-run audit/review-<ts>/ logbook report.
               For bundles, review runs per-intent (each member's acceptance criteria
               checked separately against the same delivered reality).
   (spc-12 ships only this MANUAL review surface. spc-28 then shipped the on-close hook
    that moves the intent planned → shipped and QUEUES a review on that transition.
    Auto-running the reviewer off that queue is still deferred (no epic currently owns
    it; spc-6 disowned auto-firing). Until then, the `## Audit Notes` of a freshly
    shipped intent stays empty until /abcd:intent review is run by hand.)

5. /abcd:intent reclassify <itd-N> --kind <new-kind>     (late reclassification)
   ├─ Records the change in intent.reclassification_history (date + from-kind + to-kind + reason)
   ├─ Moves the file between directories (e.g., drafts/ → disciplines/) as needed
   ├─ For supersession: --kind superseded --by <itd-M> moves the file to superseded/,
   │   writes superseded_by: itd-M, AND captures the original kind in
   │   kind_at_supersession: <original-kind> (so future readers know what shape the
   │   intent had when it was retired — standalone vs bundle-member vs discipline
   │   change the meaning of "superseded")
   └─ Triggers intent-fidelity-reviewer (shape-classification role) to verify the new kind fits

On demand: intent-fidelity-reviewer (shape-classification role) scans the corpus when
              the user runs /abcd:intent shape (spc-29 ships only the on-demand surface).
              Bare /abcd:intent (status+help) surfaces the latest cached shape suggestions
              in its summary — bare invocation never runs a fresh scan. The user accepts
              via /abcd:intent reclassify; declined suggestions become entries in the
              intent's Audit Notes for future review. (Deferred follow-up per
              .abcd/work/issues/ [spc-29 follow-up]: scheduled / pre-commit shape scanning —
              the shape(...) function's mode="pre_commit" parameter is preserved as a seam
              but no hook invokes it.)
```

## 2. Subcommands

| Subcommand | Purpose | File movement |
|---|---|---|
| `/abcd:intent` (no args) | Help + status: lists intents grouped by directory (drafts / planned / shipped / disciplines / superseded), shows commands, surfaces shape suggestions from `intent-fidelity-reviewer`, suggests next actions for any intent in flight | — |
| `/abcd:intent "<free-text>"` | **Canonical create** (spc-30/itd-46): a leading quoted seed is the canonical create entry. Interview-driven capture (press release with persona quote + acceptance criteria); assigns `itd-N`; LLM classifier writes advisory `suggested_kind`. A leading quote always creates — never falls through to bare render | writes to `drafts/itd-N-<slug>.md` (no spec created) |
| `/abcd:intent refine <itd-N>` | Interactive refinement of an existing intent (sharpen press release, fill open questions, update scope, add/edit acceptance criteria) | (stays in current state) |
| `/abcd:intent grill <itd-N>` | Socratic adversarial interview that stress-tests an intent for vagueness, missing acceptance, hidden assumptions before planning. Glossary-aware once `terminology/` exists. `--brief-section <id>` flag for stress-testing a brief section instead. (per itd-27) | (stays in current state) |
| `/abcd:intent plan <itd-N> [<itd-M>...]` | Validates the frozen PRD exists (refuses if `prd_path` is null — the grill → PRD flow produces it first; GR002 blocker, suppressed-as-info for `prd_grandfathered`); lints acceptance criteria; proposes `kind` (standalone / bundle-member / discipline) based on `suggested_kind` + cross-references; user confirms or overrides; binds `kind:`; routes to the kind-specific lifecycle path (see § 5 freeze sequence) | varies by kind: `drafts/` → `planned/` (standalone, bundle) or `drafts/` → `disciplines/` (discipline) |
| `/abcd:intent ship <itd-N>` | Kicks off implementation via the native ship/run seam. Only valid for `kind: standalone` or `kind: bundle-member`. If intent is still in `drafts/`, runs the full pipeline first (plan + plan-review). On spec completion the lifecycle hook completes the move automatically; this command can also force-move if hook missed. **Disciplines have no `ship` step** — they are continuously active once in `disciplines/`. | (eventually) → `shipped/` (or no-op for disciplines) |
| `/abcd:intent review <itd-N>` | **Role 1 — single-document fidelity.** Compares the intent's press release + acceptance criteria against delivered reality (code, configs, docs, tests). Per-criterion verdicts (`MET` / `MET_WITH_CONCERNS` / `NOT_MET` / `INCONCLUSIVE`) appended to the intent's `## Audit Notes`. Aligns with the spec store's `plan-review` / `impl-review` / `completion-review` vocabulary — same operation shape (adversarial second opinion), different opponent (press release vs engineering spec). spc-12 ships this **manual** verb; spc-28 shipped the on-close hook (move `planned → shipped` + queue a review), but auto-running the reviewer off that queue is still deferred (no spec currently owns it; spc-6 disowned auto-firing). | (stays) |
| `/abcd:intent consistency [<itd-N>]` | **Role 2 — cross-document fidelity.** Surfaces five judgement categories (terminology drift, premise contradictions, scope leakage, sequencing impossibilities, naming conflicts) across briefs + intents. **Bare** scans the whole corpus; **with `<itd-N>`** narrows to one intent's relationship with the rest. Findings land in `.abcd/logbook/audit/consistency-<ts>/report.{json,md}`. Live as of spc-29 (judgement half + on-demand verb); mechanical-half categories and pre-commit hook are deferred follow-ups (see `.abcd/work/issues/` `[spc-29 follow-up]` entries). | (stays) |
| `/abcd:intent shape [<itd-N>]` | **Role 3 — kind classification.** Examines whether an intent's declared `kind` (the noun) still fits the corpus. Surfaces *suggested* reclassifications across three live types: `kind_change`, `bundle`, `supersession`. **Bare** scans the corpus; **with `<itd-N>`** checks one intent. Pairs with `reclassify` (action verb that commits a `shape` finding). On-demand only as of spc-29; findings land in `.abcd/logbook/audit/shape-<ts>/report.{json,md}`. Concurrency via `flock(2)` on `.abcd/coordination/shape.lock` (see § 7). Scheduled / continuous invocation is a deferred follow-up (see `.abcd/work/issues/` `[spc-29 follow-up]`). | (stays) |
| `/abcd:intent reclassify <itd-N> --kind <new-kind> [--reason <text>]` | Late reclassification (e.g., a standalone intent realised to be a bundle-member; a draft realised to be a discipline; a shipped intent superseded by a later one). Records `reclassification_history` entry; moves the file between directories as the new kind dictates. `--kind superseded --by <itd-M>` is the supersession path: the file moves to `superseded/`, frontmatter records `superseded_by: itd-M` AND `kind_at_supersession: <original-kind>` so future readers know what shape the intent had when retired. | varies by destination kind |
| `/abcd:intent link <itd-N> <spc-N>` | Manual bidirectional link — used if the auto-link missed (rare) or for retroactive linking of pre-existing specs | (no move; updates frontmatter) |

**No aggregator verb.** A `check` subverb that runs `review` + `consistency` + `shape` together is *not* provided — the three primitives have very different runtime costs (review is code+oracle expensive; consistency is corpus-wide expensive; shape is cheap on demand). Bundling them produces a slow verb users avoid. Release-readiness is `/abcd:launch`'s pre-flight job. (Note: a future scheduled / pre-commit shape leg is recorded as a **deferred follow-up** under `.abcd/work/issues/` `[spc-29 follow-up]`; spc-29 ships only the on-demand surface.)

**Bare-command-as-help is a universal abcd convention** — every command (`/abcd:ahoy`, `/abcd:disembark`, `/abcd:embark`, `/abcd:launch`, `/abcd:intent`, `/abcd:capture`) shows status + suggested next actions when invoked without args. Provides discoverability without forcing the user to remember subcommand names.

## 3. Press-release format (standalone + bundle-member kinds)

Standalone and bundle-member intents use this template:

```markdown
---
id: itd-N
slug: <kebab-case>
# NOTE: no `status:` field. Lifecycle state is encoded by directory location only
#   (drafts/ | planned/ | shipped/ | disciplines/ | superseded/) — uniform across all kinds.
#   Per the 2026-05-08 directive: "directory IS the state, no cached mirror."
kind: null               # set by /abcd:intent plan: "standalone" | "bundle-member"
suggested_kind: null     # advisory, written by capture-time LLM classifier; can be ignored
bundle: null             # for kind: bundle-member, the bundle ID
spec_id: null            # or spc-N (set by /abcd:intent plan)
reclassification_history: []   # appended to by /abcd:intent reclassify (kind changes only)
surface_history: []            # appended when an intent's user-facing surface shape changes (e.g., skill → sub-verb, top-level command → sub-verb, command → flag) WITHOUT changing kind. Distinct from reclassification_history. Schema: { date, from, to, reason }. Hand-edited or written by future tooling.
---

# <Headline — what user-facing capability exists>

## Press Release

> **abcd ships with <capability>.** <2-4 sentences in present tense.>
>
> "<Customer quote>," said <persona> <role>.

## Why This Matters
## What's In Scope
## What's Out of Scope

## Acceptance Criteria      # Required (per the itd-1 discipline); Given-When-Then bullets

## Open Questions
## Audit Notes               # populated by `/abcd:intent review` (manual Role 1 run)
```

Discipline-kind intents use a different template — see § 1 "Discipline format" above.

## 4. Persona registry

See [`01-product/05-personas.md`](../01-product/05-personas.md) for the canonical persona registry (SSOT). The intent-create flow (`/abcd:intent "<text>"`) calls the native persona picker to pick a persona for the customer quote in each press release; the codified abcd principle (no real names, no "hypothetical user") lives in the canonical persona file.

## 5. Frontmatter fields (spc-3 additions)

spc-3 adds the following optional frontmatter fields to intent files. All are additive — pre-existing intents without them remain valid (schema: `intent.schema.json`).

| Field | Type | When set | Purpose |
|-------|------|----------|---------|
| `contexts` | `[list]` | optional; required when a cited term has cross-context collision | Bounded contexts this intent references; used by GL003 to resolve cross-context ambiguity |
| `glossary_terms_used` | `[list]` | auto-populated by grill skill | Qualified `<context>/<term>` IDs cited in the intent body. Machine-readable only — body prose uses canonical display names, not qualified IDs |
| `warrants_assumed` | `[list]` | optional; populated by grill skill | Toulmin warrants surfaced during grill that the author chose to assume rather than make explicit in acceptance criteria |
| `grilled_at` | ISO8601 | set by grill skill | UTC timestamp of Phase 1 grill completion |
| `grill_session_id` | UUID | set by grill skill | UUIDv4 of the Phase 1 grill session that produced the latest grill report |
| `grilled_intent_hash` | SHA-256 | set by grill skill | Hash of the intent at grill time (intent_source_hash recipe). Copied to PRD as `source_intent_hash`. Used by `/abcd:intent plan` to detect intent-edited-after-grill |
| `prd_path` | string or null | set by grill skill Phase 2 | Relative path to the PRD at `.abcd/intents/<itd-N>/prd.md`. Null until grilled |
| `prd_grandfathered` | bool or null | set by one-shot migration | True for pre-spc-3 planned intents. Suppresses GR002 and GL005 as info-only (not blocker). Cleared when intent is regrilled |

### Term ID semantics — machine vs body prose

**Machine-readable fields** (qualified `<context>/<term>` form REQUIRED):
- `glossary_terms_used` frontmatter field
- PRD frontmatter `glossary_terms_used`
- Grill report `glossary_candidates`
- Lint output and `internal/core/lint` JSON output
- Schema enforcement

**Body prose** (canonical display name only): intent body markdown uses the term's canonical display name (e.g., `persona`, not `core/persona`). Lint extractor for GL005 knows both shapes. Optional explicit citation in body: `[persona](glossary:core/persona)` is supported; lint treats the link target as authoritative when present.

### `/abcd:intent grill` as the PRD-producing sub-verb

`/abcd:intent grill <itd-N>` runs two phases over a single session context:

1. **Phase 1 (interactive)**: Socratic adversarial interview. Produces a grill-report at `.abcd/logbook/grill/<ts>-<itd-N>/grill-report.json`. Writes `grill_session_id`, `grilled_at`, `grilled_intent_hash`, `glossary_terms_used` back to intent frontmatter.
2. **Phase 2 (silent synthesis)**: Consumes the sharpened intent + glossary citations + grill findings. Produces the PRD at `.abcd/intents/<itd-N>/prd.md` with all required frontmatter including `source_intent_hash`, `grill_report_path`, `grill_report_hash`. Sets `prd_path` on the intent.

The PRD is a **frozen contract** artefact (not a session log). It lives at a per-intent path, not under logbook/.

### `/abcd:intent plan` as the PRD-validating + freezing sub-verb

`/abcd:intent plan <itd-N>` runs the ordered freeze sequence:

1. Reads `prd_path` from intent frontmatter; refuses if null (no PRD yet).
2. Validates PRD file exists, non-empty, passes section and frontmatter validators.
3. **Provenance verification**: computes `current_intent_hash` (intent_source_hash recipe); verifies it matches PRD's `source_intent_hash`; verifies PRD's `grill_report_hash` against on-disk report. Refuses on any mismatch.
4. **Draft+deprecated term stabilisation**: surfaces any cited term with `status: draft` or `status: deprecated`; user resolves before promotion continues.
5. Computes `frozen_content_hash` (frozen_content_hash recipe; provenance fields INCLUDED to prevent tampering).
6. Writes `frozen_at`, `frozen_content_hash`, `planning_attempt_id` to PRD (atomic).
7. Writes durable attempt journal at `.abcd/intents/<itd-N>/.planning-attempt.json`.
8. Passes intent + frozen PRD to the native plan step as primary context.
9. Writes `## Links` block to the new spec (atomic; idempotent).
10. Writes `spec: spc-N` back to PRD frontmatter.

Both the press-release intent and the frozen PRD are immutable input artefacts post-promotion. The press release is the elevator pitch; the PRD is the AI-consumption contract.

## 6. Acceptance gates and bidirectional link verification

`internal/core/lint` (cross-cutting) runs pre-commit and at `/abcd:intent plan` time. It verifies:

- **Acceptance criteria present and well-formed** (per the itd-1 discipline): every intent in `drafts/`, `planned/`, and `disciplines/` has a `## Acceptance Criteria` section with at least one Given-When-Then bullet. Intents cannot be promoted from `drafts/` → `planned/` (or `drafts/` → `disciplines/`) without this. Hard block.
- **`kind` is set on intents in `planned/`, `shipped/`, `disciplines/`, and `superseded/`.** Intents in `drafts/` may have `kind: null` (binding decision is at plan time). Lint blocks promotion from `drafts/` if `kind` cannot be inferred + confirmed.
- **`kind: bundle-member` requires a `bundle:` field** pointing to a bundle ID; lint verifies that *all* members of a bundle reference the same bundle ID, and that bundles are bidirectional in their members' frontmatter. **Exception for superseded bundle-members:** intents in `superseded/` with `kind_at_supersession: bundle-member` carry `bundle: null` AND `bundle_at_supersession: <bundle-id>` (preserves the bundle the intent was part of when retired, while signalling the bundle is no longer active). Lint enforces this exception.
- **Bundle invariant: all members belong to the same phase.** `/abcd:intent plan <itd-A> <itd-B> ...` (multi-arg, kind=bundle-member) hard-blocks promotion when the proposed members are scoped to different phases. Lint code `IL011`. Resolution: re-scope into one phase or downgrade one member to `kind: standalone`. See § 1 "Bundle invariant" for the canonical statement and the worked example (`intent-capture-discipline` retirement on 2026-05-07).
- **`surface_history` entries are well-formed.** Every entry must include `date` (ISO YYYY-MM-DD), `from` (free-form surface descriptor), `to`, and `reason` (non-empty). Lint code `IL012` (severity: warn — it's an audit trail, not a gate). See itd-27's `surface_history` (skill → sub-verb on 2026-05-07) for a worked example.
- **`kind: discipline` lives only in `disciplines/` or `superseded/`.** Discipline-kind intents in `drafts/` are an error (caught at plan time when the user picks a kind; rare).
- **No intent has a `status` field — across any kind.** Lifecycle state is encoded by directory location only (`drafts/` / `planned/` / `shipped/` / `disciplines/` / `superseded/`). The 2026-05-08 directive removed the cached-mirror option: directory IS the state, no exceptions. Lint hard-blocks any frontmatter containing a `status:` key (lint code: `IL013`, severity: blocker; templates and existing files were stripped in the 2026-05-08 sweep). The historical `status: draft | planned | shipped` field on standalone/bundle-member intents has been retired; uniform "directory is canonical" applies to all kinds.
- Every intent in `drafts/` has `spec_id: null` (drafts have no plan yet).
- Every intent in `planned/` has non-null `spec_id` pointing to an existing native-spec-store `<spec_id>-*.md` whose frontmatter `intent` field matches the intent's `id` (or contains the intent's `id` as one of a list, for bundle-member intents).
- Every intent in `shipped/` has a closed-status linked spec (or `spec_id: null` + a `manual_ship_reason` field for the no-spec case).
- Discipline-kind intents have `spec_id: null` always (disciplines never get a spec; this is structurally enforced).
- **Every intent in `superseded/` has both `superseded_by: <itd-M>` AND `kind_at_supersession: <original-kind>`.** The first names the successor; the second preserves what shape the intent had when it was retired (standalone vs bundle-member vs discipline change the meaning of "superseded"). Both are required. If `kind_at_supersession: bundle-member`, the intent ALSO carries `bundle_at_supersession: <bundle-id>` — preserving the bundle membership at retirement time even though the active `bundle:` field is `null`.
- No intent ID collisions; no spec referencing a non-existent intent ID.
- File location matches `kind` frontmatter (drift between dir and field flagged).
- For intents promoted from issues (per itd-4): bidirectional `related_issues` ↔ `related_intents` linkage holds. Implemented by spc-23 (intent-fidelity-reviewer `--issue-drift` mode).

Drift triggers a warning, not a block (since spec-store state may legitimately lag intent state during work in progress). Acceptance-criteria absence and kind/directory mismatch are hard blocks (the whole point of the itd-1 discipline is to force the AC discipline; the kind/directory contract makes the lifecycle navigable).

## 7. The `intent-fidelity-reviewer` agent (three roles, three verbs)

`intent-fidelity-reviewer` is a single agent in the catalog (per `05-internals/01-agents.md`) that owns three roles. Roles share the agent's prompt scaffolding, oracle backend resolution, and receipts; they differ in what they review, when they run, where findings land, and **which subverb users invoke them through**. Each role has its own dedicated verb — no role-by-kind dispatch, no hidden-state forking.

The verb `review` is chosen for Role 1 to align with the spec store's review vocabulary (`plan-review`, `impl-review`, `completion-review`); `audit` is reserved for the top-level `/abcd:audit` (compliance / hash-chain integrity). Each verb means one thing.

### Role 1 — single-document fidelity → `/abcd:intent review <itd-N>`

`/abcd:intent review <itd-N>` is the **manual** Role 1 surface for `kind: standalone` and `kind: bundle-member` intents. Compares:
- **Intent press release + acceptance criteria** ("what user-facing capability exists, plus the verifiable bar")
- **Delivered reality** (current state of the source repo — code, configs, docs, tests)

This is product-tier review. The opponent is the codebase. Distinct from the spec store's `completion-review` (engineering-tier — code vs spec) — both can pass or fail independently, and disagreement between them is signal: the spec may have mistranslated the press release.

**Two passes, two destinations.** Role 1 judges two document kinds and writes each verdict to its own destination:
- the **itd-1 acceptance pass** (a shipped intent) writes per-criterion verdicts into that intent's `## Audit Notes` (the verdict of record) **plus** a per-run `audit/review-<ts>/` logbook report;
- the **itd-37 `MG004` pass** (a native spec's `## Modification Grammar`) writes its `PASS` / `FAIL` verdict to an `audit/spec-mg-<ts>/` logbook receipt — native specs have no `## Audit Notes` section, so the verdict cannot land in-file.

**What spc-12 ships.** spc-12 ships the **discipline-judgement subset** of `/abcd:intent review` — the itd-1 per-criterion acceptance verdicts and the itd-37 `MG004` boilerplate check, with their writers and receipts. The broader **press-release prose review** (the `honoured` / `diverged` / `missing` buckets below) and other prose/terminology/PRD-fidelity outputs are **deferred** to a later spec. spc-12 also ships the **manual** review surface; spc-28 then shipped the on-close hook (move `planned → shipped` + queue a review on that transition). Auto-running the reviewer off that queue is still deferred — no spec currently owns it; spc-6 disowned auto-firing.

**What spc-23 ships.** spc-23 ships the `--issue-drift` role — a corpus-wide bidirectional cross-reference walk between shipped intents and the `iss-N` ledger (per itd-4). Receipts land at `.abcd/logbook/audit/issue-drift-<ts>/`. Default exit 0 with warnings to stderr; `--strict` opts into exit 1 for CI gates. See the native spec `spc-23-intent-fidelity-reviewer-extension`.

Outputs findings with two layers:

**Per-criterion verdicts** (per the itd-1 discipline) — for every Given-When-Then bullet in the acceptance section:
- `[MET]` — verified
- `[MET_WITH_CONCERNS]` — partially observed
- `[NOT_MET]` — divergence (with explicit "what was delivered vs what was promised")
- `[INCONCLUSIVE]` — could not verify

**Press-release prose review** (three buckets — `honoured` / `diverged` / `missing`):
- **honoured** — capabilities the press release promised that exist as described
- **diverged** — capabilities present but materially different from the press release
- **missing** — capabilities the press release promised that aren't observable in delivered reality

Findings appended to the intent's `## Audit Notes` section. Manual re-run via `/abcd:intent review <itd-N>` available at any time. **Overall verdict rollup:** any `NOT_MET` → overall `NOT_MET`; any `INCONCLUSIVE` without `NOT_MET` → overall `INCONCLUSIVE`; any `MET_WITH_CONCERNS` without `NOT_MET`/`INCONCLUSIVE` → overall `MET_WITH_CONCERNS`; else `MET`. (Note: the verdict tags read as audit-shaped judgement labels even though the verb is `review` — the two registers are deliberately distinct: `review` is the verb / process; verdict tags are the output shape.)

For bundle-member intents, this role runs *per intent* against the same delivered reality (each member's acceptance criteria checked separately).

#### The audit loop — record-only vs loop-to-acceptance (itd-50 / spc-52)

Role 1 records per-criterion verdicts; **itd-50 adds the POLICY that decides what happens to a recorded verdict.** The policy rides the review-queue drainer on the run seam (adr-27) — it is NEVER in the pure on-close lifecycle hook (the lifecycle close stays a pure data function; the mode logic lives in the drainer/policy layer).

**Three audit-loop modes, facilitator-elected per intent** via the `audit_mode` frontmatter key:

- **`record-only`** (the default, and today's behaviour) — a `NOT_MET` is written to `## Audit Notes`; no re-work is triggered. An **absent** `audit_mode` key resolves to `record-only` (additive, spc-28/spc-43-compatible).
- **`loop-to-acceptance`** — a `NOT_MET` re-opens the linked work and iterates against the same acceptance criteria until they read `MET`, bounded by `audit_budget` (the spec-grain SHIP/NEEDS_WORK fix-loop lifted to the intent grain). See `05-internals/03-configuration.md` for the `audit_mode` / `audit_budget` keys, the default budget (`3`), and the fail-closed rule for a malformed/zero/negative budget.

**Full state-coverage table** (every Family-2 rollup maps to a defined action — no dead-ends; the loop trigger is `NOT_MET` only):

| Reviewer rollup | Loop action (`loop-to-acceptance`) |
|---|---|
| all `MET` | Succeed → manual-verification gate. `audit_outcome=MET`. |
| `MET_WITH_CONCERNS` (no `NOT_MET`) | Proceed to the gate (concerns are advisory; the product thinker sees them). Does NOT consume budget. `audit_outcome=MET`. |
| `NOT_MET` (budget remaining) | Re-open linked work (re-enqueue at the queue layer), `audit_iterations_used += 1`. |
| `NOT_MET` (budget exhausted) | Terminate `UNACHIEVABLE` (budget-exhausted) → replan invitation. |
| `NOT_MET` + reviewer judges criteria unmeetable | Terminate `UNACHIEVABLE` (reviewer-impossible) **early**, before budget exhaustion → replan invitation. |
| `INCONCLUSIVE` | Fail-closed: recorded as today, no iterate, no summons, no replan; never flips to `UNACHIEVABLE`. |

`UNACHIEVABLE` is an intent-level **rollup** terminal the policy layer writes — it is **NOT** a per-criterion verdict (`ACCEPTANCE_VERDICTS = {MET, MET_WITH_CONCERNS, NOT_MET, INCONCLUSIVE}` is unchanged) and is recognised at the `Overall:` / rollup parse layer only.

**The `UNACHIEVABLE` replan surface (no rollback).** A terminal `UNACHIEVABLE` writes a `why-unachievable` explanation + a **replan invitation** block (naming both the product thinker and the facilitator) into the intent's `## Audit Notes`. The intent **stays in `shipped/`** — its `spec_id` / `kind` / directory + delivered artifacts are byte-untouched (a `drafts/` move would break the lifecycle invariant + spc-48 lint and read as a partial un-ship). The invitation seeds `/abcd:intent grill`; no machine authors a replan, nothing is auto-rolled-back.

**Gated manual verification + verification receipt (R5).** The manual-verification invitation renders **only** when the machine rollup is acceptance-eligible (all `MET`, or `MET_WITH_CONCERNS` with no `NOT_MET`) — if any criterion is not `MET`/concerns, the loop or replan invitation runs first; the product thinker is never asked to hand-test something the audit already knows is broken. The sign-off is recorded as a **verification receipt distinct from the machine verdict of record** — a separate JSON artifact under `.abcd/logbook/audit/verify-<ts>/receipt.json`, never merged into `## Audit Notes`:

```json
{ "intent_id": "itd-N", "machine_rollup": "MET", "state": "offered",
  "justification": "...optional...", "recorded_by_role": "product thinker",
  "ts": "20260614T…Z" }
```

Receipt **states**: `offered` (the gate opened — the drainer stamps this on a `MET` loop outcome), `accepted` (the product thinker confirms the intention is delivered), `rejected_wrong_criteria` (every criterion passes but the *why* is not delivered).

**`rejected_wrong_criteria` → replan, NOT a synthetic `NOT_MET`.** When the machine says `MET` but the product thinker judges the criteria themselves were wrong, the defect is the *criteria*, not the code — so the rejection routes to the **same replan surface** as `UNACHIEVABLE` (one writer, two entry points), carrying the rejection justification into the seeded grill. It does **not** write a `NOT_MET` (which would re-loop the implementation against criteria that already pass) and does **not** move the intent.

### Role 2 — cross-document fidelity → `/abcd:intent consistency [<itd-N>]`

Introduced by itd-48 (which superseded itd-31). The opponent is *other documents*: compares the brief and every intent against each other (and against the brief itself), surfacing the five live judgement categories — **terminology drift, premise contradictions, scope leakage, sequencing impossibilities, naming conflicts**. No spec-store analogue — the spec store reviews one artefact at a time; corpus-wide consistency is pure abcd ground.

spc-29 ships the judgement half on demand via `/abcd:intent consistency` (Carmack-level oracle review).

**Deferred follow-up** (recorded in `.abcd/work/issues/` under `[spc-29 follow-up]`): the mechanical-half lint categories — schema/state contradictions, reference rot, acknowledgement gaps — were originally planned as `internal/core/lint --cross-doc` codes `XD002`/`XD006`/`XD007` per `05-internals/06-lint.md`; the lint-code half is deferred to a follow-up intent. Pre-commit hook wiring that would let `/abcd:intent consistency` findings block commits is also deferred.

**Polymorphic on arg presence (same operation, narrowed scope):** bare = scan the whole corpus; with `<itd-N>` = scan one intent's relationship with the rest. This is *not* the forbidden hidden-state dispatch — the operation is identical; the arg just narrows scope (like `git log` vs `git log <path>`).

Findings land in `.abcd/logbook/audit/consistency-<ts>/report.{json,md}`.

### Role 3 — kind classification → `/abcd:intent shape [<itd-N>]`

Introduced alongside the three intent kinds (per itd-34). The opponent is the *kind taxonomy*: examines whether each intent's declared `kind` (the noun in frontmatter) still fits the corpus. The verb `shape` matches the taxonomy noun and pairs cleanly with `reclassify` (the action verb that commits a `shape` finding).

spc-29 ships the on-demand surface only. **Bare** scans the corpus; **with `<itd-N>`** checks one intent. Findings appear in `/abcd:intent` status output (passive surface) and as a separate report at `.abcd/logbook/audit/shape-<ts>/report.{json,md}`. The user accepts a suggestion via `/abcd:intent reclassify`; declined suggestions are logged for future review (so the reviewer doesn't re-surface the same suggestion every run).

**Concurrency contract** (between any future scheduled invocation and on-demand `shape`):

```
.abcd/coordination/shape.lock  (file lock via flock(2))
```

- On-demand `/abcd:intent shape` acquires **blocking with 60s timeout**; on timeout, reports "background run in progress; try again or use `--wait`".
- The intent's `## Audit Notes` section is updated atomically (read full file → modify in memory → write via `.tmp` + `rename(2)`) by the on-demand path.

Mirrors the file-claim pattern itd-33 will introduce in a later phase but is far simpler — single lock per logbook subdirectory, no agent identity, no heartbeat.

The three live `suggestion_type` values this role produces:

- **`kind_change`** — a 1:1 reclassification between `standalone` and `discipline`. Example: "intent X has no user moment in its press release (the customer quote describes a process, not a feature); consider `kind: discipline`."
- **`bundle`** — "intents X and Y reference each other in scope/references and target the same release; consider `kind: bundle-member` with shared bundle ID."
- **`supersession`** — "intent X's scope is fully covered by intent Y; consider `kind: superseded --by Y`."

> **Deferred follow-up** (recorded in `.abcd/work/issues/` under `[spc-29 follow-up]`): pre-commit hook wiring for continuous shape scanning, and the `shape(...)` function's `mode="pre_commit"` parameter is preserved as a seam but no hook invokes it. Discipline subtype clustering ("once enough disciplines exist, surfaces 'three disciplines have similar `kind_notes`; consider formalising a subtype'") was named in earlier itd-34 drafts and is *not* shipped by spc-29 — it is not a live suggestion type.

### Logbook layout

All review/audit verbs write to `.abcd/logbook/audit/<sub-tier>-<ts>/`. The directory's name (`audit/`) reflects "this is the on-disk audit trail" regardless of which verb produced it; the sub-tier prefix names the verb. Verb-to-sub-tier mapping:

- `/abcd:intent review <itd-N>` → `audit/review-<ts>/` (Role 1 itd-1 acceptance pass, single-document fidelity per itd-1)
- `/abcd:intent review` (MG004 boilerplate pass) → `audit/spec-mg-<ts>/` (Role 1 itd-37 `MG004` check on a native spec's `## Modification Grammar`; one per-run batch receipt, one `results[]` entry per spec — native specs have no `## Audit Notes` section, so the verdict lands here, per itd-37)
- `/abcd:intent consistency` → `audit/consistency-<ts>/` (Role 2, cross-document fidelity per itd-48, which superseded itd-31)
- `/abcd:intent shape` → `audit/shape-<ts>/` (Role 3, shape classification per itd-34)
- `/abcd:audit chain` → `audit/chain-<ts>/` (conversation/edit-history Merkle, default application per itd-16)
- `/abcd:audit lifeboat <path>` → `audit/lifeboat-<ts>/` (lifeboat-artefact integrity per itd-35)

`chain` and `lifeboat` are sub-verbs of `/abcd:audit` (the umbrella verb); `review`, `consistency`, `shape` are sub-verbs of `/abcd:intent`. The `review` verb produces *two* sub-tiers — its itd-1 acceptance pass writes `review-<ts>/` (alongside the verdict-of-record in the intent's `## Audit Notes`) and its itd-37 `MG004` pass writes `spec-mg-<ts>/` — because Role 1 judges two document kinds (a shipped intent and a native spec) and only the intent has an in-file `## Audit Notes` section. Bare `/abcd:audit` and bare `/abcd:intent` are status+help only per the universal bare-command-as-help convention. Layout codified in `05-internals/04-universal-patterns.md § 6`.

## 8. Reports

`intent-report.{json,md}` in `.abcd/logbook/intent/<timestamp>/` — one report per `/abcd:intent` invocation. JSON has full detail; MD has skim summary.
