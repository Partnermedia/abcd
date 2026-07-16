# Agent prompt changelog

Per [itd-5](../.abcd/development/intents/disciplines/itd-5-prompt-quality-additions.md),
every `agents/*.md` prompt carries a `prompt_version` and a corresponding entry
here recording the bump rationale (and, at `1.0.0` lock, the self-improvement
pre-flight outcome and calibration-corpus delta).

**Version band.** Agents ship in the `0.x` band until they clear their calibration
corpus — `0.x` means "shipped and wired, honestly unmeasured"; `1.0.0` means
"measured against a corpus and locked" (itd-81's amendment to itd-5, which governs
over the brief's earlier `1.0.0`-at-close expectation). The four M6 synthesis
agents below ship at `0.1.0`: wired to their `abcd disembark` verbs, unmeasured.

## 0.1.0 — 2026-07-16 (itd-88 M6 — synthesis agents)

### principle-distiller 0.1.0

First entry. Host-delegated distiller behind `abcd disembark principles
<lifeboat-dir> --principles-json`. Reads a packed lifeboat's ADRs, intents,
resolved issues, and graveyard findings; emits `principles.json` with each
principle citing a record id, a graveyard finding id, or a packed lifeboat path
(cite-or-be-dropped over `R ∪ F ∪ P`). Carries `reads_untrusted_input: true`,
`capability_scope.task_classes: [principle_distillation]`, and an injection-canary
fixture. Unmeasured (no corpus yet); no self-improvement pre-flight run.

### graveyard-interpreter 0.1.0

First entry. Host-delegated interpreter behind `abcd disembark graveyard
<lifeboat-dir> --lessons-json`. Reads the sealed `graveyard/archaeology.json` and
`graveyard/abandoned.json`; emits the graveyard **lessons** schema (no `mode`, no
`prompt_version` field — the pre-M6 lessons schema), each lesson citing a live
layer-1/2 finding id (cite-or-be-dropped over the finding-id set). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes:
[cross_document_audit]`, and an injection-canary fixture. Unmeasured; no
self-improvement pre-flight run.

### press-release-composer 0.1.0

First entry. Host-delegated composer behind `abcd disembark press-release
<lifeboat-dir> --press-release-json`. Reads the packed brief, spine, and
`principles.json`; emits a single `press-release.json` document that must cite at
least one path in `brief/**`, `rescue/spine.md`, or `principles.json`
(whole-document refusal if it cites nothing resolvable). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes: [surface_render]`,
and an injection-canary fixture. Unmeasured; no self-improvement pre-flight run.

### lifeboat-oracle 0.1.0

First entry. Host-delegated auditor behind `abcd disembark oracle <lifeboat-dir>
<source-repo> --oracle-json`. Reads the packed lifeboat corpus against its source
repo; emits an `oracle` audit carrying a registered verdict (`SHIP` / `NEEDS_WORK`
/ `MAJOR_RETHINK` — out-of-enum refuses the whole payload) and findings that each
cite a packed lifeboat path (cite-or-be-dropped over the packed-path set). Carries
`reads_untrusted_input: true`, `capability_scope.task_classes: [oracle_review,
audit]`, and an injection-canary fixture. Unmeasured; no self-improvement pre-flight
run.
