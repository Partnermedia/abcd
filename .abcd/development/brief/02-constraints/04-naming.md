# Maritime Naming Convention

Commands and abcd-owned directories use ship/voyage metaphors where natural.

| Path / command | Meaning |
|---|---|
| `/abcd:ahoy` | bare invocation — status + help: shows folder kind, install state, detected gaps, last install date. ZERO writes. |
| `/abcd:ahoy install` | mutating sub-verb — applies detected gaps (skeleton, config-change, history-store, marker-block, PATH symlink, version stamp). Centralised per-category approval. |
| `/abcd:ahoy uninstall` | reversible removal — strips the marker block from `CLAUDE.md` / `AGENTS.md` and the `/usr/local/bin/abcd` symlink if owned. Preserves `.abcd/`, `~/.abcd/`, and `hooks/hooks.json`. |
| `/abcd:ahoy dry-run` | read-only emit of the `DetectionResult` envelope as JSON. ZERO writes. Drives the Claude Code skill's two-pass approval protocol. |
| `/abcd:ahoy doctor` | read-only audit — detection envelope + cross-machine `history_audit.audit_workspaces()` gaps. ZERO writes. |
| `/abcd:disembark` | leave the ship → pack a lifeboat for the journey |
| `/abcd:embark` | board a new ship → unpack the lifeboat |
| `/abcd:launch` | put the (cleaned) ship to sea publicly |
| `/abcd:dredge` | cross-corpus synthesis — surface latent patterns from accumulated captures (see itd-25 — a later phase). Maritime: dredging the seabed for what's settled. Pairs with `lifeboat` (per-project rescue) as the cross-corpus counterpart (latent-value rescue). |
| `.abcd/lifeboat/` | the portable artefact (rescue from a sinking project) |
| `.abcd/logbook/` | record of voyages — logs, state, reports per command run |

**Sense disambiguation:** `/abcd:launch` uses the *nautical* sense (a ship's first entry to water, i.e. the public maiden voyage of the cleaned `*Dev` repo) — not the generic software sense of "run a program".

**Metaphor exemptions** — meta-development surfaces, not voyage steps:

- `/abcd:intent` — product-framing surface (press-release-shaped roadmap capture). "Intent" does semantic work the brief depends on; no maritime word carries that meaning.
- `/abcd:capture` — issue-capture surface (see itd-4). "Capture" is deliberately neutral so the verb doesn't pre-commit to whether a finding is bug, nitpick, or systemic pattern; the synthesist (in a later phase) decides later.
- `/abcd:intent grill` — Socratic-questioning sub-verb (see itd-27). "Grill" is borrowed register from prior art ([Pocock skills](https://github.com/mattpocock/skills)) and signals adversarial interrogation directly; no maritime word carries that meaning. Lives as a sub-verb of `/abcd:intent` — its mid-session glossary writes and per-session logbook output are command-shaped (per `05-internals/08-skills.md`).
- `/abcd:audit` — formal verification surface (see itd-16 — a later phase). Reserved; not metaphor-mapped, dignified register.
- `/abcd:reflect` — major-milestone retrospective (see itd-24 — a later phase). Reserved; not metaphor-mapped, soft register.

**Reserved meta-development commands** (later phases; named now to prevent collisions):

> **Note:** `/abcd:audit` and `/abcd:reflect` appear in BOTH this table AND the "Metaphor exemptions" list above. The two listings encode two distinct contracts: the exemptions list says these verbs are *exempt from the maritime convention*; this table says they are *reserved for later-phase intents*. Both are true simultaneously, so both listings are kept.

| Path / command | Meaning |
|---|---|
| `/abcd:dredge` | cross-corpus synthesis (see itd-25 — a later phase). Maritime: dredging the seabed. Pairs with `lifeboat` as the cross-corpus counterpart to per-project rescue. |
| `/abcd:loot` | OSS-vendor-with-provenance — clone selected files from public repos into `vendor/<source>/`, record origin / licence / SHA / rationale in `.abcd/development/activity/loot/<source>.md` (see itd-26 — a later phase). Maritime: raid the open ocean for outside cargo. Pairs with `dredge` (own corpus, salvage-frame) as the public-corpus counterpart (raid-frame). The pirate connotation is feature-not-bug — the verb itself prompts a licence-check reflex. |
| `/abcd:audit` | formal verification surface — hash-chain / Merkle audit trails, fidelity checks (see itd-16 — a later phase). Reserved; not metaphor-mapped, dignified register. |
| `/abcd:reflect` | major-milestone retrospectives (see itd-24 — a later phase). Reserved; not metaphor-mapped, soft register. |

Technical files (`meta.json`, `config.json`, `corpus.json`, `rules.json`) are exempt — no metaphor needed.

**Metaphor-vs-exempt criterion (added post-audit 2026-05-07):** apply a maritime metaphor when the verb has a **natural maritime cognate that adds meaning** (e.g., `dredge` literally raises settled material; `loot` carries the licence-check reflex). Otherwise stay exempt — neutral verbs (`intent`, `capture`, `grill`, `audit`, `reflect`) signal meta-development surfaces and avoid stretched metaphors that obscure the verb's intent. Reach for a metaphor only when it teaches; not just because the convention exists.

**Bare-command-as-render discipline (added post-audit 2026-05-08):** every `/abcd:<verb>` command MUST treat the bare invocation (no args) as **status + help + render of current state** for that verb's namespace. Sub-verbs MUST earn their existence by doing something the bare invocation cannot — mutating state, taking a positional argument, scoping to a different time-axis or granularity, or performing an action distinct from rendering.

**Earned sub-verbs** (do something bare cannot): `/abcd:intent "<text>"` (canonical create — bare quoted text mutates: creates draft, per fn-30/itd-46); `/abcd:intent plan <itd-N>` (action: promotion); `/abcd:intent grill <itd-N>` (action: adversarial interview); `/abcd:capture list --open|--resolved|--wontfix|--all` (filtered query — distinct from default-open render); `/abcd:capture promote <iss-N>` (action: issue → intent-new interview + four-field back-link, live as of fn-30/itd-46); `/abcd:audit chain` and `/abcd:audit lifeboat` (different application targets); `/abcd:memory ingest <source>` (mutates: adds source); `/abcd:memory ask <question>` (action: query); `/abcd:oracle ask <prompt>` (action: invoke cascade).

**Forbidden sub-verbs** (collapse to bare): `<verb> show`, `<verb> stats`, `<verb> list` (plain, no filters), `<verb> view`. These name what bare already does. Lint code (reserved): `SD001` — sub-verb names what bare renders.

**Rationale:** the bare convention is what gives abcd its discoverability ("type the verb, see what it does"). Sub-verbs that just rename "show me the state" obscure the discoverability instead of enhancing it. The discipline rules `show`/`stats`/`list`/`view` out of the namespace at design time, not in review.

**Brief-is-current-state discipline (per [adr-5](../../decisions/adrs/adr-5-brief-is-current-state.md)):** the brief reflects the project's *current* state. No version label on the brief; no `archive/<NN>/` directory inside `brief/`; no version-changelog blobs in `brief/README.md`. History lives in `git log brief/`; inflection-point rationale lives in [`../../decisions/adrs/`](../../decisions/adrs); forensic snapshots come from `/abcd:disembark` (logged in `voyage/disembark/history.jsonl`, per [adr-4](../../decisions/adrs/adr-4-lifeboat-as-regenerable-output.md)).

## Vocabulary-registration requirement (HARD from the start)

Every term introduced in any spec's `## Modification Grammar > Ripple > Vocabulary delta` sub-bullet (per itd-37) MUST be registered in this glossary file in the same spec. `intent_lint.py` blocks at plan-review on missing registrations. Lint code (reserved): `VR001` — vocabulary delta term not registered in glossary.

**Why hard from the start, not soft.** A discipline that ships with "soft initially, hard once stable" is structurally weaker than itd-1 (acceptance gates) and itd-5 (prompt-quality additions), both of which ship hard from day one. Cost of hard enforcement: ~30 seconds per new term. Cost of soft enforcement: vocabulary drift compounds; the cross-document fidelity reviewer (Role 2) finds drift post-hoc that should have been blocked at design time.

**Glossary location.** Terms are registered in this file (`02-constraints/04-naming.md`) under either the maritime convention table (if the term is a `/abcd:<verb>` command or `.abcd/<directory>` artefact with a maritime cognate), the metaphor-exemptions list (if it's a meta-development surface), or a new "Reserved vocabulary" section below. The reviewer Role 2 cross-document audit verifies registration on every plan-review.

**Reserved vocabulary** (controlled enums, PR-to-extend):

| Term | Type | Source |
|---|---|---|
| `phase retrospective` | The five-section README (`went well` / `could improve` / `lessons learned` / `decisions made` / `metrics`) written by `/abcd:reflect <phase-id>` to `.abcd/retrospectives/<phase-id>/README.md`. Phase-grained only (the intent form was dropped per the itd-24 grill). Composed by the `reflection-composer` agent from the fn-66 phase-audit receipt; rendered/written by the deterministic writer `scripts/abcd/reflect_writer.py`. Links to the phase doc + audit report + member specs only (SSOT — no body duplication). | fn-83 spec + `fn-83-operator-surfaces-manifest-lockstep.3` (itd-24) |
| `reflection-composer` | The 16th catalog agent: composes phase-retrospective prose from a seeded single-pass interview grounded in the fn-66 phase-audit receipt's per-bullet acceptance verdicts. Dispatched by `/abcd:reflect`. `capability_scope.task_classes: [surface_render]`. | fn-83 spec + `fn-83-operator-surfaces-manifest-lockstep.3` (itd-24) |
| `setup-wizard` | The display-only surface (`scripts/abcd/setup_wizard/`) that explains a missing external dependency when the fn-76 validation gate fails closed: four fixed-order elements (tool name + version floor / requiring capability / what fails without it / exact install step), sourced from the gate's typed `MissingToolPayload` (single source) with a curated blurb registry for prose only. NEVER weakens the gate — declining stays fail-closed and records a logbook decline. NOT a top-level command in v1 (rendered through the gate CLI + a standalone `explain` entrypoint). | fn-83 spec + `fn-83-operator-surfaces-manifest-lockstep.4` (itd-63) |
| `JSON sidecar` | The canonical `review.json` file written into each per-review directory under `.flow/reviews/`. Consumers MUST read the JSON sidecar; the rendered `.md` is derived. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.1`; schema at `docs/reference/review-schema.md` |
| `MD render` | The derived `review.md` file rendered mechanically from the JSON sidecar (front-matter from metadata, prose from `body_markdown`, "## Findings" from `findings[]`). Not canonical; consumers read the JSON sidecar. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.1` |
| `write-time sanitiser` | The Stage 1 sanitiser applied to review body text before writing the JSON sidecar and rendered MD. Strips absolute paths, API keys, and PII patterns. Implemented in `scripts/abcd/_review_lib.py::sanitise_text()`. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.3` |
| `Stage 1` (write-time sanitiser) | Write-time mutation pass: applied before writing `body_markdown` and before hashing the raw artifact. Strips absolute paths and secret/PII patterns. The only component that mutates content. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.3` |
| `Stage 2` (detect-and-block) | Pre-commit/CI gitleaks gate: `verify_reviews.py` runs `gitleaks protect --staged --redact=100` and **blocks the commit** if secrets survive Stage 1. Detection only — does not rewrite files. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.3` |
| `staleness` (review-freshness) | A review is stale when the files listed in `reviewed_files` have changed since `review_of_commit` (detectable when `pinning: "commit"`). Staleness signals that a re-review may be needed. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.4` |
| `body cap` | The `body_max_bytes` field: maximum byte length of `body_markdown` in the JSON sidecar before truncation applies. Active cap value stored in `review.json` at generation time. Replaces the legacy `summary cap` term. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.1` |
| `render cap` | The `render_max_bytes` field: maximum byte length of the rendered `review.md`. May be smaller than `body cap` since MD adds structure overhead. Active cap value stored in `review.json` at generation time. Replaces the legacy `summary cap` term. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.1` |
| `staging directory` | A `.staging-<NNNN>/` sibling directory used during atomic per-review directory writes. The writer creates the staging directory, populates it, then renames it to the final `<NNNN>-<slug>-<ref>/` name — POSIX rename is atomic within a single filesystem. Staging directories are gitignored. | fn-2 spec + `fn-2-move-repoprompt-review-artifacts-into.1` |
| `decision class` ∈ `{intent, RFC, ADR}` | Roadmap-record classifier — three decision-record surfaces with distinct lifecycles (forward-user-facing / forward-contested / retrospective-settled) | adr-1, adr-5; see [`../../decisions/README.md`](../../decisions/README.md) |
| `kind` ∈ `{standalone, bundle-member, discipline}` | Persisted intent kind classifier — stays three-valued; `decision` is NEVER a persisted `kind` (it has no lifecycle directory under thin adoption) | itd-34 |
| `capture verdict` ∈ `{standalone, bundle-member, discipline, decision}` | Capture-TIME classifier verdict (`intent classify-capture-kind`) — the first three mirror the persisted `kind`; `decision` is capture-only: it routes a confirmed standing infrastructure choice to the existing ADR store (`adr-N`), never to a persisted `kind` or the intent lifecycle. Admitted by `suggested_kind` (advisory hint) but REFUSED by `plan_single`/`reclassify`. | itd-44 (fn-56) |
| `source.class` ∈ `{session_memory, external_pdf, external_transcript, external_article, oracle_review, work_notes, issue_ledger, dredge_synthesis, spec_modification_grammar, modification_grammar}` | Memory page source class | itd-36 |
| Lifecycle classes ∈ `{regenerable, append-only, compounding-curated}` | Artefact lifecycle taxonomy | `05-internals/04-universal-patterns.md § 8` |
| Review verdicts ∈ `{SHIP, NEEDS_WORK, MAJOR_RETHINK}` | Carmack-style review verdicts | `05-internals/01-agents.md § Verdict-tag protocol` |
| Criterion verdicts ∈ `{MET, MET_WITH_CONCERNS, NOT_MET, INCONCLUSIVE}` | Per-criterion intent acceptance | itd-1 |
| `task_classes` (capability_scope tokens) ∈ `{oracle_review, intent_review, spec_planning, code_rescue, principle_distillation, lifeboat_packing, audit, lint, surface_render, cross_document_audit}` | Closed enum, agent frontmatter. Machine-readable source of truth: [`scripts/abcd/schemas/task_classes.json`](../../../../scripts/abcd/schemas/task_classes.json) — a cross-check test fails if this table and the JSON diverge. PR-to-extend. | itd-5 extension (idea-4) — lean ~10 tokens drawn from current abcd surfaces |
| `frozen_content_hash` | SHA-256 hex string written to PRD frontmatter by `/abcd:intent plan` at freeze time. Non-self-referential: provenance fields included, operational fields excluded. Recipe documented in `prd.schema.json`. | fn-3 task .5 |
| `intent_source_hash` | SHA-256 hex string computed over the parent intent's body + stable frontmatter. Written by grill skill as `grilled_intent_hash`; copied to PRD as `source_intent_hash`. Recipe documented in `prd.schema.json`. | fn-3 task .5 |
| `planning_attempt_id` | UUIDv4 written to PRD frontmatter and the durable attempt journal by `/abcd:intent plan`. Used by GR004 to detect stale planning attempts. | fn-3 task .5 |
| `prd_grandfathered` | Boolean frontmatter field on pre-fn-3 planned intents. When `true`, GR002 and GL005 are suppressed-as-info (not blocker). Cleared on regrill. | fn-3 task .5 |
| Grandfather migration | One-shot sweep at fn-3 ship time: appends `prd_path: null` + `prd_grandfathered: true` to every intent in `planned/` that predates the PRD requirement. | fn-3 task .5 |
| `promote-check mode` | `intent_lint.py --promote-check <intent.md>` evaluates the intent as if it were already in `planned/`, firing GR002 and GL005 at planned-state severity for pre-flight checks. | fn-3 task .5 |
| `abandon-attempt mode` | `intent_lint.py --abandon-attempt <itd-N>` (also `/abcd:intent plan --abandon-attempt <itd-N>`) clears a stale planning attempt: removes the attempt journal and sets `planning_attempt_id: null` in the PRD. Freeze fields (`frozen_at`, `frozen_content_hash`, `spec`) are preserved — the PRD itself is not un-frozen. Remediation for GR004. | fn-3 task .5 |
| `failure_mode_tag` ∈ `{hallucination, scope_drift, stale_context, under_specification_blindness, format_violation}` | Closed enum, in the later-phase Frontier Awareness intent | idea-4, a later phase |
| `work_item.type` ∈ `{intent_promotion, spec_task, command_run}` | Coordination claim unit (a later phase) | itd-33, a later phase |
| Claim primitives ∈ `{take, yield, escalate}` | Coordination conflict-resolution verbs (a later phase; cooperative-checkpoint semantics, no mid-flight abort) | itd-33, a later phase |
| Escalation choice ∈ `{wait_then_swap, swap_now, sequence, keep_both}` | Human-resolved escalation outcomes (a later phase) | itd-33, a later phase |
| `release.outcome` ∈ `{completed, abandoned}` | Audit-log outcome on claim release (a later phase) | itd-33, a later phase |
| `claim.status` ∈ `{active, paused, released}` | Three-state claim lifecycle; `paused` interlocks with itd-29's pause/resume/rewind (a later phase) | itd-33, a later phase |
| `recall` | Keyword-list field on each rules.json domain — natural-language phrases matched word-boundary against user prompts to trigger rule injection. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `domain` | Uppercase grouping key in rules.json (e.g. `COMMITTING`, `DOCUMENTATION`). Domains carry state (`active` / `dormant`), recall keywords, and rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `dormant` | rules.json domain state value (opposite of `active`). Recall-match injection is skipped for dormant domains; star-command `*<DOMAIN>` still activates them. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `active` | rules.json domain state value (default). Recall-match injection fires when prompt matches recall keywords. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `*<DOMAIN>` | Leading-anchored uppercase prompt prefix that explicitly activates a domain regardless of recall match. Regex `(?:^|\s)\*([A-Z][A-Z0-9_]*)(?=$|\s)`. Hyphen/period/slash following the domain name fails the boundary (no activation). Multi-star activates multiple domains in left-to-right order. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.5` |
| `force_refresh_every_n` | `.abcd/config.json` field under the `rules.` namespace. Integer; default 5. Every N prompts, `prompt_router_hook.py` forces a full re-inject regardless of dedup signature match (compaction recovery). | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.4` |
| `.abcd/config.json` | Repo-scope config file. Named in the in-repo carve-out at `05-internals/03-configuration.md:182` alongside `<repo>/.abcd/rules.json` and `.specstory/cli/config.toml`. Reads include `rules.force_refresh_every_n` and `docs.target`. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.8` |
| `rules.json` | Repo-scope rule overrides file at `<repo>/.abcd/rules.json`. Named in the in-repo carve-out at `05-internals/03-configuration.md:182`. JSON Schema 2020-12; schema at `scripts/abcd/schemas/rules.schema.json`. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `COMMITTING` | Default plugin-bundled domain. Recall keywords trigger injection of commit-discipline rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `DOCUMENTATION` | Default plugin-bundled domain. Recall keywords trigger injection of documentation-discipline rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `ROADMAP` | Default plugin-bundled domain. Recall keywords trigger injection of roadmap/intent rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `ISSUES` | Default plugin-bundled domain. Recall keywords trigger injection of issue-tracking rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `INTENTS` | Default plugin-bundled domain. Recall keywords trigger injection of intent-capture rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `LIFEBOAT` | Default plugin-bundled domain. Recall keywords trigger injection of lifeboat/disembark rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `PII` | Default plugin-bundled domain. Recall keywords trigger injection of PII-protection rules. | fn-14 spec + `fn-14-modular-rules-loader-prompt-router-hook.1` |
| `managed-workspace` | Folder kind: holds repo-shaped subdirs, has an ABCD marker block, no `.git/` at its own root. Per brief `04-surfaces/01-ahoy.md:104-127`. | fn-15 spec + `fn-15-folder-classification-workspacesjson.3` |
| `managed-repo` | Folder kind: has an ABCD marker block and a `.git/` directory. | fn-15 spec + `fn-15-folder-classification-workspacesjson.3` |
| `unmanaged-workspace` | Folder kind: no marker block, holds repo-shaped subdirs. | fn-15 spec + `fn-15-folder-classification-workspacesjson.3` |
| `unmanaged-repo` | Folder kind: a git repo without abcd management; bare `/abcd:ahoy` offers install to adopt. | fn-15 spec + `fn-15-folder-classification-workspacesjson.3` |
| `unmanaged-folder` | Folder kind: no marker block, no `.git/`, no repo-shaped subdirs. | fn-15 spec + `fn-15-folder-classification-workspacesjson.3` |
| `root_commit` | Immutable repo identity key in `index.json`, computed via `git rev-list --max-parents=0 HEAD`. Survives rename, remote move, and GitHub-handle change. | fn-15 spec + `fn-15-folder-classification-workspacesjson.2` |
| `workspaces.json` | User-scope registry at `~/.abcd/workspaces.json` listing every workspace and repo abcd manages on this machine. Keyed on mutable `name`. Per brief `05-internals/03-configuration.md:135-167`. | fn-15 spec + `fn-15-folder-classification-workspacesjson.4` |
| `index.json` | History-store registry at `~/.abcd/history/index.json` recording each repo's identity + lineage. Keyed on immutable `root_commit`. | fn-15 spec + `fn-15-folder-classification-workspacesjson.5` |
| `aliases` | Array of prior names a repo has had (e.g., renamed on GitHub). Recorded in per-root-sha `meta.json`. | fn-15 spec + `fn-15-folder-classification-workspacesjson.5` |
| `supersedes` | Lineage cross-ref in `index.json` repo entry: this entry was re-founded from another root-sha. | fn-15 spec + `fn-15-folder-classification-workspacesjson.5` |
| `superseded_by` | Lineage cross-ref in `index.json` repo entry: this entry was superseded by another root-sha (re-founding produces a new entry). | fn-15 spec + `fn-15-folder-classification-workspacesjson.5` |
| `output_dir` | TOML field in `<repo>/.specstory/cli/config.toml [local_sync]` block; absolute path to the per-root-sha SpecStory output directory under `~/.abcd/history/<root-sha>/specstory/`. | fn-15 spec + `fn-15-folder-classification-workspacesjson.6` |
| `closed-review tombstone` | Durable provenance record `{context_id, window, workspace, safe_key, closed_at}` written by `close-tab` under `.flow/.rp-review-tabs/.tombstones/`, keyed by `_rp_safe_token(context_id)`. Survives the receipt deletion so orphan-after-close untitled chats remain GC-reapable. | fn-68 spec + `fn-68-rp-closed-review-tombstone-gc-durable.1` |
