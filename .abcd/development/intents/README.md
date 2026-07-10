# abcd Intents

Intent specifications for the abcd plugin.

---

## What's an Intent?

An **intent** captures *what user-facing capability exists once shipped* — written as an Amazon working-backwards press release, not an engineering feature spec.

**Why press releases instead of feature specs:** Feature specs are engineering-shaped from the start (Problem → Design → Tasks). Press releases are user-experience-shaped (what *exists for the user* once shipped). Forcing intent capture in user-facing language disciplines product clarity before engineering scope.

**Plumbing work doesn't get intents.** Adapters, agents, harness, scaffolding — these have no user moment, and forcing press-release format on them produces strained or mis-targeted prose. Plumbing lives in the brief at `.abcd/development/brief/README.md`. See brief § 1.5 for the full three-layer model (brief = whole picture, intents = user-facing why, specs = how).

This is a **codified abcd principle**: intent capture is press-release-shaped. abcd ships projects with this convention pre-baked.

---

## Intent IDs

Intent IDs follow the pattern `itd-N` (unpadded, mirrors the native spec `spc-N` convention). Filenames: `itd-N-<slug>.md`.

`itd` reads as "intent" and pairs visually with the native spec `spc-N`.

**IDs are capture-stable.** An intent keeps its `itd-N` for life — IDs are assigned in capture order and never renumbered. Sequencing is *not* encoded in the ID; it lives in the phase docs at [`../phases/`](../roadmap/phases), whose `## Scope` sections are the single source of truth for which intents a phase bundles (see [adr-9](../decisions/adrs/0009-phase-as-product-layer.md)).

**Why unpadded:** abcd anticipates intent counts that would exceed any practical padding budget. Unpadded matches `spc-N` visually, avoids the future migration, and reads naturally in prose ("itd-7 spawned itd-19"). Lexical-vs-numeric sort is handled at tool layer (the record lint, registries, dashboards) rather than via filename padding.

---

## Three Intent Kinds

Every intent has a `kind` declared in frontmatter, set at `/abcd:intent plan` time. Three kinds exist (see [`brief/04-surfaces/05-intent.md § 1`](../brief/04-surfaces/05-intent.md#1-intent-ids-kinds-and-lifecycle) for the canonical reference):

| `kind` | Has press release? | Lives in | Maps to | Examples |
|---|---|---|---|---|
| `standalone` | Yes | `drafts/` → `planned/` → `shipped/` | One spec (1:1) | itd-3, itd-4, itd-7, most of the corpus |
| `bundle-member` | Yes | Same as standalone, with `bundle: <id>` linking members | Shared spec with bundle-mates (N:1) | (no live bundles — `tier-0-audit-substrate` dissolved 2026-05-07; itd-31 promoted to standalone, itd-32 superseded) |
| `discipline` | **No** — uses `## Rule` instead | `disciplines/` | No spec; imposes acceptance gates on every other spec | itd-1 (AC gate), itd-5 (prompt-quality) |

Standalone is the default (~60% of the corpus). Bundle-members ship together as one spec. Disciplines are cross-cutting rules with no user moment of their own; they apply to every other spec as inherited acceptance gates.

The kind is **project-agnostic** — application projects (e.g., a macOS app under abcd) produce their own disciplines (privacy-impact review, accessibility passes, code-style conventions). The three kinds are a property of the intent framework, not of abcd's particular subject matter.

**The persisted `kind` enum stays three-valued.** The capture-time classifier has a *fourth* verdict, `decision` (a standing infrastructure choice — "we use Postgres"), but `decision` is **never a persisted `kind`** and never enters this lifecycle: a confirmed `decision` routes to the existing ADR store (`../../decisions/adrs/`, `adr-N-<slug>.md`), not to a draft. There is deliberately **no `intents/decisions/` directory**. See [itd-44](drafts/itd-44-fourth-intent-kind-decision.md) (spc-56 thin adoption) and `brief/04-surfaces/05-intent.md § 1`.

---

## Lifecycle Directories

| Directory | Status / role | Meaning |
|---|---|---|
| `drafts/` | 📝 Draft | Press-release-shaped intent captured but no native spec yet. Bench of ideas / forward-looking work. Cheap to draft and discard. |
| `planned/` | 📅 Planned | A committed capability awaiting its Go build — scheduled into a roadmap phase, or committed-but-unscheduled awaiting sequencing (the two axes are orthogonal, per [adr-34](../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md)). `spec_id` is `null` until the native spec layer schedules it (Phase 4), then points at a `spc-N`; bundle-member intents share a `spec_id` with their bundle-mates. |
| `shipped/` | ✅ Shipped | Linked spec closed; `intent-fidelity-reviewer` ran. The intent's "Audit Notes" section contains per-criterion verdicts (per the itd-1 discipline) plus a three-bucket prose audit. |
| `disciplines/` | 📐 Active rule | Discipline-kind intents. Never get a native spec of their own; instead they impose acceptance gates that every *other* spec inherits and is checked against. **No `status` frontmatter field — the directory IS the state.** |
| `superseded/` | 🗄️ Superseded | Intents killed by reclassification or absorption. The file records `superseded_by: <itd-N>` (the successor) AND `kind_at_supersession: <original-kind>` (what shape the intent had when retired). Preserved as historical record. |

There is no `active/` state — "active" is implicit (a planned intent's linked spec is currently in flight in the native spec store; an active discipline is any intent in `disciplines/`).

**Directory location is the single source of truth for lifecycle state across all kinds.** Standalone and bundle-member intents derive state from `drafts/` / `planned/` / `shipped/`; disciplines derive state from `disciplines/` / `superseded/`. No intent has a `status` field that could disagree with its directory; the record lint enforces the contract.

---

## Lifecycle (mostly automatic, with deliberate manual steps)

```
1. /abcd:intent "<free-text>"
   ├─ Interview captures press release (headline + persona quote + scope)
   ├─ Assigns next itd-N ID (capture-stable — never renumbered)
   ├─ Requires `## Acceptance Criteria` section (per the itd-1 discipline) — refuses to write if missing/malformed
   ├─ LLM classifier writes advisory `suggested_kind` (default: standalone)
   └─ Writes intents/drafts/itd-N-<slug>.md (no spec-plan call yet)

2. /abcd:intent plan <itd-N> [<itd-M>...]    (when ready to commit to work)
   ├─ Lints acceptance criteria; refuses promotion if missing/malformed
   ├─ Reads suggested_kind + cross-references; proposes a kind
   ├─ User confirms or overrides; binding `kind:` is written
   │
   ├─ standalone (single intent ID):
   │     ├─ Plans the native spec (plan + plan-review)
   │     ├─ Injects bidirectional link (spec.intent = itd-N; intent.spec_id = spc-N)
   │     └─ drafts/ → planned/
   │
   ├─ bundle-member (multiple intent IDs in one plan call):
   │     ├─ Plans one native spec with all intents as joint input
   │     ├─ spec.intent = [itd-A, itd-B]; each intent.spec_id = spc-N; each intent.bundle = <bundle-id>
   │     └─ All members: drafts/ → planned/
   │
   └─ discipline (single intent, kind chosen explicitly):
         ├─ NO spec-plan call
         ├─ Registers acceptance gates in .abcd/disciplines/<itd-N>.json
         ├─ Plan-review on the discipline's `## Rule` for sanity
         └─ drafts/ → disciplines/ (active state encoded by directory location; no `status:` field)

3. /abcd:intent ship <itd-N>          (standalone + bundle only — disciplines never ship)
   ├─ If intent is in drafts/: runs full pipeline (plan + plan-review first)
   ├─ Runs the native spec work
   └─ Spec continues in the native spec store

4. Spec marked done in the native spec store   (standalone + bundle: work complete, automatic from here)
   ├─ intent_lifecycle_hook detects status change via the intent: link
   ├─ planned/ → shipped/ (bundles: all members move together)
   └─ Triggers intent-fidelity-reviewer agent (single-document role)
       └─ Per-criterion verdicts (MET / MET_WITH_CONCERNS / NOT_MET / INCONCLUSIVE)
           plus three-bucket prose audit (honoured / diverged / missing)
           appended to intent file's "Audit Notes" section.
           For bundles, review runs per-intent against the same delivered reality.

5. /abcd:intent reclassify <itd-N> --kind <new-kind> [--reason <text>]
   ├─ Records reclassification_history entry (date + from-kind + to-kind + reason)
   ├─ Moves the file between directories as the new kind dictates
   └─ --kind superseded --by <itd-M> is the supersession path:
        ├─ file moves to superseded/
        ├─ frontmatter gains superseded_by: itd-M (the successor)
        └─ frontmatter gains kind_at_supersession: <original-kind>
            (preserves the shape the intent had when retired —
             standalone vs bundle-member vs discipline change the
             meaning of "superseded")

Continuously: intent-fidelity-reviewer's shape-classification role suggests
              reclassifications based on cross-reference patterns, scope overlap,
              and supersession candidates. User accepts via /abcd:intent reclassify.
```

**Manual overrides:**

- `/abcd:intent ship` can force-move planned→shipped if the hook missed (standalone + bundle only)
- `/abcd:intent link <itd-N> <spc-N>` for retroactive linking of pre-existing specs
- `/abcd:intent review <itd-N>` to manually re-run the fidelity reviewer at any time (Role 1 — single-doc fidelity)
- `/abcd:intent reclassify <itd-N> --kind <new-kind>` for late kind changes

**No `/abcd:intent move`** — file location follows verb side-effects, not user intervention.

---

## Bidirectional Link Convention

| File | Frontmatter field |
|---|---|
| `intents/{drafts,planned,shipped}/itd-N-<slug>.md` | `spec_id: spc-N` (scalar, or `null` when in drafts/ — **never a list**) |
| the native spec `spc-N-<slug>` | `intent: itd-N` (or list — one spec may consume several intents; that is the bundle direction) |

Both directions present once `/abcd:intent plan` runs. The record lint (pre-commit + CI) verifies they agree, and rejects a list-valued `spec_id`.

**Split-the-intent doctrine.** An intent is the unit of consumption: it is implemented by exactly one spec. Work too big for one spec decomposes into *tasks inside* that spec; an intent containing two separately verifiable promises is two intents — split it (precedent: the launch PRD's Tier A/B split into itd-67 and itd-72). This keeps the close hook singular, coverage computable, and doneness unambiguous — an intent can never be half-consumed and called done.

---

## Press Release Format

Every intent file uses this template (spc-3 fields shown; all new fields are optional — pre-existing intents without them remain valid):

```markdown
---
id: itd-N
slug: <kebab-case-slug>
# NOTE: no `status:` field — directory location is the canonical lifecycle state.
#   See brief/04-surfaces/05-intent.md § 6 for the lint rule and rationale.
kind: null               # set by /abcd:intent plan: "standalone" | "bundle-member" | "discipline"
spec_id: null
# spc-3 fields (optional; additive — pre-existing intents valid without them):
contexts: null           # [list] of bounded-context IDs; required when term has cross-context collision
glossary_terms_used: null  # [list] of qualified <context>/<term> IDs; auto-populated by grill skill
warrants_assumed: null   # [list] of Toulmin warrants assumed (not made explicit in AC)
grilled_at: null         # ISO8601 UTC; set by grill skill at Phase 1 completion
grill_session_id: null   # UUIDv4; set by grill skill
grilled_intent_hash: null  # SHA-256 of intent at grill time (intent_source_hash recipe)
prd_path: null           # relative path to PRD (e.g. .abcd/intents/itd-N/prd.md); set by grill Phase 2
prd_grandfathered: null  # true = pre-spc-3 planned intent; GR002+GL005 suppressed-as-info
---

# <Headline — what user-facing capability exists>

## Press Release

> **abcd ships with <capability>.** <2-4 sentences describing what users can now do, in present tense as if shipped.>
>
> "<Customer quote — picked from personas.json>," said <persona> <role>.

## Why This Matters

<1-2 paragraphs explaining the underlying user need.>

## What's In Scope

- <Bullet>

## What's Out of Scope

- <Bullet — preventing scope creep>

## Acceptance Criteria

> _Required (per itd-1). At least one Given-When-Then bullet. Hard-blocked at /abcd:intent plan time if missing or malformed._

- **Given** <preconditions>, **when** <user/system action>, **then** <observable outcome>.

## Prior Art

> _Required. Positions the intent against the existing corpus: what it builds on, what almost covers it, why it is nonetheless its own intent. At least one resolvable reference (sibling intent, brief section, principle, ADR, or external source); "none found — searched <where>" is a valid entry, an empty section is not._

- <Reference + one line on the relation>

## Open Questions

- <Bullet — anything not yet decided>

## Audit Notes

<Empty until intent moves to shipped/. intent-fidelity-reviewer populates this with per-criterion verdicts plus a three-bucket prose audit comparing delivered reality to the press release above.>
```

---

## PRD Freeze Contract (spc-3)

When `/abcd:intent plan <itd-N>` runs, the PRD at `prd_path` is **frozen**:

1. `frozen_content_hash` is computed from the PRD body + stable frontmatter fields (provenance fields INCLUDED; `frozen_at`, `frozen_content_hash`, `spec`, `planning_attempt_id` EXCLUDED).
2. `frozen_at`, `frozen_content_hash`, `planning_attempt_id` are written atomically to the PRD.
3. Mutating the frozen PRD after promotion triggers `GR003` (blocker lint).

The freeze is **non-self-referential**: re-computing the hash on the frozen PRD (excluding the freeze fields) yields the same value. Mutating `body_markdown` or any included frontmatter field changes the hash; mutating `frozen_at`, `frozen_content_hash`, `spec`, or `planning_attempt_id` does NOT change the hash.

**Hash recipes** — two distinct recipes documented in `prd.schema.json`:
- `intent_source_hash` recipe: used at grill time and plan-time provenance check (hashes the parent intent).
- `frozen_content_hash` recipe: used at freeze time (hashes the PRD body + stable provenance fields).

---

## Customer Quotes — Persona Convention

Customer quotes use placeholder personas from `.abcd/development/personas.json` (Alice, Bob, Carol, ... — a fixed alphabetical sequence). Selection is **by role, never by name**: the intent's audience picks the role; the role's registered name is used. Every persona is they/them.

This is a discipline ([`disciplines/itd-79-persona-registry.md`](disciplines/itd-79-persona-registry.md), enforced by the `persona_registry` record-lint rule): never use real names in press releases (PII), but never use generic "a hypothetical user" language (loses voice). Named personas keep quotes grounded without leaking real-world identifiers.

---

## Sequencing — see `phases/`

Which intents a phase bundles, and in what order phases ship, is **not recorded here.** Sequencing lives in the phase docs at [`../phases/`](../roadmap/phases) — each phase doc's `## Scope` section is the single source of truth (per [adr-9](../decisions/adrs/0009-phase-as-product-layer.md)). An intent listed in no phase doc's `## Scope` is implicitly **unscheduled** — valid for `drafts/` and `planned/` alike: a draft is uncommitted, an unscheduled planned intent is committed but awaiting sequencing. The invariant runs one way only: any intent a phase `## Scope` names is committed by definition and lives in `planned/` (or `disciplines/`) — see [adr-34](../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md).

This README describes the intent corpus by *lifecycle state* (the directory listings below); it deliberately does not duplicate the phase→intent mapping.

---

## Bundles

Active bundles (sets of intents that ship as one shared spec via multi-arg `/abcd:intent plan`):

| Bundle ID | Members | Why a bundle |
|---|---|---|
| ~~`tier-0-audit-substrate`~~ (dissolved 2026-05-07) | ~~itd-31 + itd-32~~ | The bundle premise (unified `/abcd:audit` surface bundling all review/audit roles into one verb's subverbs) was dissolved when the round-2 command-structure review split the three intent-fidelity-reviewer roles into three distinct verbs under `/abcd:intent` (review/consistency/shape). itd-31 promoted to standalone; itd-32 superseded by itd-31. |

Bundles are declared in member intents' frontmatter (`bundle: <bundle-id>`); membership is bidirectional (verified by the record lint). When a bundle's shared spec closes, all member intents move from `planned/` to `shipped/` together.

**Note on cross-phase bundle attempts:** the `intent-capture-discipline` bundle (itd-27 + itd-30) was retired. The bundle invariant requires *one shared spec shipped together* — and per [adr-9](../decisions/adrs/0009-phase-as-product-layer.md) all bundle members must belong to the same phase. itd-27 has a plan-reviewed spec (`spc-3`); itd-30 is unscheduled. Both intents were reclassified to `standalone`; if itd-30 is later picked up, its spec can depend on or extend `spc-3` for shared interview/lint/persona-registry plumbing without needing the bundle declaration.

---

## Drafts

Captured intents that haven't been promoted to native specs yet. Each standalone or bundle-member intent moves to `planned/` once the user runs `/abcd:intent plan <itd-N>`; discipline-kind intents move to `disciplines/`. For the sequencing view — which phase bundles which intents — see [`../phases/`](../roadmap/phases); this directory listing is the raw filesystem view.

```
drafts/
├── itd-8-with-code-bundling.md
├── itd-9-schema-migration.md
├── itd-10-purge-uninstall.md
├── itd-11-pass-b-pitfall-mitigation.md
├── itd-12-work-adapter-weighting.md
├── itd-13-scheduled-dev-sync.md
├── itd-14-prompt-registry-versioning.md
├── itd-15-self-dogfooded-sota-audit.md
├── itd-16-hash-chain-merkle-audit.md
├── itd-17-model-effectiveness-tracking.md
├── itd-18-permission-template-per-project-type.md
├── itd-19-stage-aware-behaviour.md
├── itd-21-no-lifeboat-scaffolding.md
├── itd-22-opencode-portability.md
├── itd-23-spec-kit-interop.md
├── itd-25-dredge-cross-corpus-synthesist.md
├── itd-26-loot-oss-vendor.md
├── itd-30-design-fictions-as-intent-format.md
├── (itd-31 superseded — moved to superseded/itd-31-cross-document-fidelity-reviewer.md; absorbed by itd-48)
├── (itd-32 superseded — moved to superseded/itd-32-audit-role-taxonomy.md)
├── itd-33-agent-communication-infrastructure.md
├── itd-35-lifeboat-integrity-audit.md           (sibling to itd-16)
├── itd-39-scope-aware-memory-retrieval.md        (extends itd-3 recall hook to memory)
├── itd-41-phase-negotiator.md                    (Socratic phase-proposer, per adr-10)
├── itd-44-fourth-intent-kind-decision.md
├── itd-47-oracle-gates-autonomous-mode.md
├── itd-51-harness-adoption-readiness-rubric.md
├── itd-55-first-principles-foundations-auditor.md
├── itd-57-manual-hold-sentinel.md
├── itd-59-autonomous-worker-transcript-capture.md
├── itd-60-doc-fidelity-anti-drift.md
├── itd-61-brief-change-derivation.md
├── itd-62-pluggable-safety-gate.md
├── itd-64-benchmark-driven-config-optimization.md
├── itd-70-launch-release-retention-newest-per-line.md
├── itd-73-derived-versioning.md
├── itd-74-name-banlist.md
└── itd-75-cli-eval-harness.md
```

---

## Disciplines

Active discipline-kind intents (cross-cutting rules with no user moment of their own; impose acceptance gates that every other spec inherits). They never get a native spec — they ARE the rule, not a feature being built. **No `status` frontmatter field — presence in `disciplines/` IS the active state; supersession moves to `superseded/`.**

```
disciplines/
├── itd-1-acceptance-gates.md         (## Acceptance Criteria gate on every intent + spec)
├── itd-5-prompt-quality-additions.md (prompt_version + self-improvement + injection canaries + capability_scope static)
├── itd-37-modification-grammar.md    (## Modification Grammar gate on every spec — Naur's Modification axis + Ripple sub-axis)
└── itd-79-persona-registry.md        (persona names from personas.json, selected by role — persona_registry lint gate)
```

See [`brief/04-surfaces/05-intent.md § 1`](../brief/04-surfaces/05-intent.md#1-intent-ids-kinds-and-lifecycle) "Discipline format" for the template (no press release; uses `## Rule` + `## Why` + `## Acceptance Criteria` instead; no `status` field).

**Discipline subtypes** (e.g. methodology / documentation / audit) are deferred — see the revisit triggers in the brief. For now each discipline declares a free-text `kind_notes` field describing what kind of rule it is.

---

## Planned

`planned/` holds the committed capabilities awaiting their Go build — some scheduled into a roadmap phase, others committed but not yet sequenced (per [adr-34](../decisions/adrs/0034-lifecycle-and-scheduling-orthogonal.md)). Their `spec_id` is `null` until the spec layer schedules them (Phase 4).

```
planned/
├── itd-2-in-session-subagent-dispatch.md
├── itd-3-modular-rules-loader.md
├── itd-4-issue-capture.md
├── itd-6-rp-mcp-only-integration.md
├── itd-7-rp-workspace-portability.md
├── itd-20-top-level-abcd-dispatcher.md
├── itd-24-reflect-command.md
├── itd-27-grill-skill-and-glossary.md
├── itd-28-rp-reviews-into-flow.md
├── itd-29-autonomous-run-resilience.md
├── itd-34-three-intent-kinds.md
├── itd-36-memory-unification.md
├── itd-40-folder-classification.md
├── itd-42-coherence-aware-grill.md
├── itd-43-epic-to-spec-terminology.md
├── itd-46-abcd-intent-quoted-text-create-symmetric.md
├── itd-48-intent-fidelity-reviewer-roles-2-3.md
├── itd-49-flow-state-drift-detector.md
├── itd-50-loop-toward-acceptance.md
├── itd-53-review-queue-auto-drain-fidelity-gate.md
├── itd-58-session-reviewer-verdict-ingestion.md
├── itd-63-setup-wizard-explains-installs.md
├── itd-65-launch-preflight-gate-suite.md
├── itd-66-launch-payload-render-parity.md
├── itd-67-installable-versioned-plugin.md
├── itd-69-plugin-metadata-lockstep-update.md
└── itd-72-launch-ship-tier-b-publishing.md
```

---

## Superseded

Two intents currently superseded.

```
superseded/
├── itd-31-cross-document-fidelity-reviewer.md  (superseded by itd-48 on 2026-05-27; kind_at_supersession: standalone; absorbed by itd-48 which owns Roles 2 + 3)
└── itd-32-audit-role-taxonomy.md  (superseded by itd-31 on 2026-05-07; kind_at_supersession: bundle-member; tier-0-audit-substrate bundle dissolved)
```

Intents move here when they are killed by reclassification or absorption — e.g., when a smaller intent is folded into a larger one (`/abcd:intent reclassify <itd-N> --kind superseded --by <itd-M>`), or when a discipline is replaced by a stricter successor. Each superseded intent records two fields:

- **`superseded_by: <itd-M>`** — the successor intent
- **`kind_at_supersession: <original-kind>`** — what shape the intent had when retired (`standalone`, `bundle-member`, or `discipline`)

The original-kind preservation matters: "superseded" means different things depending on what the intent *was*. A superseded standalone is a retired capability. A superseded bundle-member is a retired half of a coupled pair. A superseded discipline is a retired rule that was inherited by every other spec. Without `kind_at_supersession`, future archaeology has to reconstruct the original shape from `reclassification_history` — which exists but is harder to query.

Files in `superseded/` are preserved as historical record; never deleted.

---

## Shipped

`shipped/` holds capabilities built in Go. It is empty until the Go build ships them (Phase 1 onward); an intent moves here automatically when its linked spec closes and `intent-fidelity-reviewer` has run.
