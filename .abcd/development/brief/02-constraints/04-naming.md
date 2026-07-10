# Maritime Naming Convention

Commands and abcd-owned directories use ship/voyage metaphors where natural.

| Path / command | Meaning |
|---|---|
| `/abcd:ahoy` | bare invocation — status + help: shows folder kind, install state, detected gaps, last install date. ZERO writes. |
| `/abcd:ahoy install` | mutating sub-verb — applies detected gaps (skeleton, config-change, history-store, marker-block, PATH symlink, version stamp). Centralised per-category approval. |
| `/abcd:ahoy uninstall` | reversible removal — strips the marker block from `CLAUDE.md` / `AGENTS.md` and the `/usr/local/bin/abcd` symlink if owned. Preserves `.abcd/`, `~/.abcd/`, and `hooks/hooks.json`. |
| `/abcd:ahoy dry-run` | read-only emit of the `DetectionResult` envelope as JSON. ZERO writes. Drives the Claude Code skill's two-pass approval protocol. |
| `/abcd:ahoy doctor` | read-only audit — detection envelope + cross-machine `history_audit.audit_repos()` gaps. ZERO writes. |
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
- `/abcd:reflect` — phase-retrospective surface (see itd-24). Not metaphor-mapped, soft register.
- `/abcd` (bare, top-level) — where-am-i status board (see itd-20). Not metaphor-mapped; the namespace root renders status + help (`status` / `help` are byte-identical aliases).

**Reserved meta-development commands** (later phases; named now to prevent collisions):

> **Note:** `/abcd:audit` appears in BOTH this table AND the "Metaphor exemptions" list above. The two listings encode two distinct contracts: the exemptions list says the verb is *exempt from the maritime convention*; this table says it is *reserved for a later-phase intent*. Both are true simultaneously, so both listings are kept.

| Path / command | Meaning |
|---|---|
| `/abcd:dredge` | cross-corpus synthesis (see itd-25 — a later phase). Maritime: dredging the seabed. Pairs with `lifeboat` as the cross-corpus counterpart to per-project rescue. |
| `/abcd:loot` | OSS-vendor-with-provenance — clone selected files from public repos into `vendor/<source>/`, record origin / licence / SHA / rationale in `.abcd/development/loot/<source>.md` (see itd-26 — a later phase). Maritime: raid the open ocean for outside cargo. Pairs with `dredge` (own corpus, salvage-frame) as the public-corpus counterpart (raid-frame). The pirate connotation is feature-not-bug — the verb itself prompts a licence-check reflex. |
| `/abcd:audit` | formal verification surface — hash-chain / Merkle audit trails, fidelity checks (see itd-16 — a later phase). Reserved; not metaphor-mapped, dignified register. |

Technical files (`config.json`, `corpus.json`, `rules.json`) are exempt — no metaphor needed.

**Metaphor-vs-exempt criterion (added post-audit 2026-05-07):** apply a maritime metaphor when the verb has a **natural maritime cognate that adds meaning** (e.g., `dredge` literally raises settled material; `loot` carries the licence-check reflex). Otherwise stay exempt — neutral verbs (`intent`, `capture`, `grill`, `audit`, `reflect`) signal meta-development surfaces and avoid stretched metaphors that obscure the verb's intent. Reach for a metaphor only when it teaches; not just because the convention exists.

**Bare-command-as-render discipline (added post-audit 2026-05-08):** every `/abcd:<verb>` command MUST treat the bare invocation (no args) as **status + help + render of current state** for that verb's namespace. Sub-verbs MUST earn their existence by doing something the bare invocation cannot — mutating state, taking a positional argument, scoping to a different time-axis or granularity, or performing an action distinct from rendering.

**Earned sub-verbs** (do something bare cannot): `/abcd:intent "<text>"` (canonical create — bare quoted text mutates: creates draft, per spc-30/itd-46); `/abcd:intent plan <itd-N>` (action: promotion); `/abcd:intent grill <itd-N>` (action: adversarial interview); `/abcd:capture list --open|--resolved|--wontfix|--all` (filtered query — distinct from default-open render); `/abcd:capture promote <iss-N>` (action: issue → intent-new interview + four-field back-link, live as of spc-30/itd-46); `/abcd:audit chain` and `/abcd:audit lifeboat` (different application targets); `/abcd:memory ingest <source>` (mutates: adds source); `/abcd:memory ask <question>` (action: query); `/abcd:oracle ask <prompt>` (action: invoke oracle).

**Forbidden sub-verbs** (collapse to bare): `<verb> show`, `<verb> stats`, `<verb> list` (plain, no filters), `<verb> view`. These name what bare already does. Lint code (reserved): `SD001` — sub-verb names what bare renders.

**Rationale:** the bare convention is what gives abcd its discoverability ("type the verb, see what it does"). Sub-verbs that just rename "show me the state" obscure the discoverability instead of enhancing it. The discipline rules `show`/`stats`/`list`/`view` out of the namespace at design time, not in review.

**Brief-is-current-state discipline (per [adr-5](../../decisions/adrs/0005-brief-is-current-state.md)):** the brief reflects the project's *current* state. No version label on the brief; no `archive/<NN>/` directory inside `brief/`; no version-changelog blobs in `brief/README.md`. History lives in `git log brief/`; inflection-point rationale lives in [`../../decisions/adrs/`](../../decisions/adrs); forensic snapshots come from `/abcd:disembark` (logged in `voyage/disembark/history.jsonl`, per [adr-4](../../decisions/adrs/0004-lifeboat-as-regenerable-output.md)).

## Vocabulary-registration requirement (HARD from the start)

Every term introduced in any spec's `## Modification Grammar > Ripple > Vocabulary delta` sub-bullet (per itd-37) MUST be registered in this glossary file in the same spec. The vocabulary-registration lint blocks at plan-review on missing registrations. Lint code (reserved): `VR001` — vocabulary delta term not registered in glossary.

**Why hard from the start, not soft.** A discipline that ships with "soft initially, hard once stable" is structurally weaker than itd-1 (acceptance gates) and itd-5 (prompt-quality additions), both of which ship hard from day one. Cost of hard enforcement: ~30 seconds per new term. Cost of soft enforcement: vocabulary drift compounds; the cross-document fidelity reviewer (Role 2) finds drift post-hoc that should have been blocked at design time.

**Glossary location.** Terms are registered in this file (`02-constraints/04-naming.md`) under either the maritime convention table (if the term is a `/abcd:<verb>` command or `.abcd/<directory>` artefact with a maritime cognate), the metaphor-exemptions list (if it's a meta-development surface), or a new "Reserved vocabulary" section below. The reviewer Role 2 cross-document audit verifies registration on every plan-review.

**Reserved vocabulary** (controlled enums, PR-to-extend):

| Term | Type | Source |
|---|---|---|
| `phase retrospective` | The five-section README (`went well` / `could improve` / `lessons learned` / `decisions made` / `metrics`) written by `/abcd:reflect <phase-id>` to `.abcd/retrospectives/<phase-id>/README.md`. Phase-grained only (the intent form was dropped per the itd-24 grill). Composed by the `reflection-composer` agent from the spc-66 phase-audit receipt; rendered/written by the deterministic reflect writer. Links to the phase doc + audit report + member specs only (SSOT — no body duplication). | spc-83 spec + `spc-83-operator-surfaces-manifest-lockstep.3` (itd-24) |
| `reflection-composer` | The 16th catalog agent: composes phase-retrospective prose from a seeded single-pass interview grounded in the spc-66 phase-audit receipt's per-bullet acceptance verdicts. Dispatched by `/abcd:reflect`. `capability_scope.task_classes: [surface_render]`. | spc-83 spec + `spc-83-operator-surfaces-manifest-lockstep.3` (itd-24) |
| `setup-wizard` | The display-only surface (in the Go binary, `internal/core/...`) that explains a missing external dependency when the spc-76 validation gate fails closed: four fixed-order elements (tool name + version floor / requiring capability / what fails without it / exact install step), sourced from the gate's typed `MissingToolPayload` (single source) with a curated blurb registry for prose only. NEVER weakens the gate — declining stays fail-closed and records a logbook decline. NOT a top-level command in v1 (rendered through the gate CLI + a standalone `explain` entrypoint). | spc-83 spec + `spc-83-operator-surfaces-manifest-lockstep.4` (itd-63) |
| `JSON sidecar` | The canonical `review.json` file written into each per-review directory in the review store. Consumers MUST read the JSON sidecar; the rendered `.md` is derived. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.1`; schema at `docs/reference/review-schema.md` |
| `MD render` | The derived `review.md` file rendered mechanically from the JSON sidecar (front-matter from metadata, prose from `body_markdown`, "## Findings" from `findings[]`). Not canonical; consumers read the JSON sidecar. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.1` |
| `write-time sanitiser` | The Stage 1 sanitiser applied to review body text before writing the JSON sidecar and rendered MD. Strips absolute paths, API keys, and PII patterns. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.3` |
| `Stage 1` (write-time sanitiser) | Write-time mutation pass: applied before writing `body_markdown` and before hashing the raw artifact. Strips absolute paths and secret/PII patterns. The only component that mutates content. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.3` |
| `Stage 2` (detect-and-block) | Pre-commit/CI secret-scan gate: the verifier runs the configured secret scan over the staged content and **blocks the commit** if secrets survive Stage 1. Detection only — does not rewrite files. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.3` |
| `staleness` (review-freshness) | A review is stale when the files listed in `reviewed_files` have changed since `review_of_commit` (detectable when `pinning: "commit"`). Staleness signals that a re-review may be needed. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.4` |
| `body cap` | The `body_max_bytes` field: maximum byte length of `body_markdown` in the JSON sidecar before truncation applies. Active cap value stored in `review.json` at generation time. Replaces the legacy `summary cap` term. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.1` |
| `render cap` | The `render_max_bytes` field: maximum byte length of the rendered `review.md`. May be smaller than `body cap` since MD adds structure overhead. Active cap value stored in `review.json` at generation time. Replaces the legacy `summary cap` term. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.1` |
| `staging directory` | A `.staging-<NNNN>/` sibling directory used during atomic per-review directory writes. The writer creates the staging directory, populates it, then renames it to the final `<NNNN>-<slug>-<ref>/` name — POSIX rename is atomic within a single filesystem. Staging directories are gitignored. | spc-2 spec + `spc-2-move-repoprompt-review-artifacts-into.1` |
| `decision class` ∈ `{intent, RFC, ADR}` | Roadmap-record classifier — three decision-record surfaces with distinct lifecycles (forward-user-facing / forward-contested / retrospective-settled) | adr-1, adr-5; see [`../../decisions/README.md`](../../decisions/README.md) |
| `kind` ∈ `{standalone, bundle-member, discipline}` | Persisted intent kind classifier — stays three-valued; `decision` is NEVER a persisted `kind` (it has no lifecycle directory under thin adoption) | itd-34 |
| `capture verdict` ∈ `{standalone, bundle-member, discipline, decision}` | Capture-TIME classifier verdict (`intent classify-capture-kind`) — the first three mirror the persisted `kind`; `decision` is capture-only: it routes a confirmed standing infrastructure choice to the existing ADR store (`adr-N`), never to a persisted `kind` or the intent lifecycle. Admitted by `suggested_kind` (advisory hint) but REFUSED by `plan_single`/`reclassify`. | itd-44 (spc-56) |
| `source.class` ∈ `{session_memory, external_pdf, external_transcript, external_article, oracle_review, work_notes, issue_ledger, dredge_synthesis, spec_modification_grammar, modification_grammar}` | Memory page source class | itd-36 |
| Lifecycle classes ∈ `{regenerable, append-only, compounding-curated}` | Artefact lifecycle taxonomy | `05-internals/04-universal-patterns.md § 8` |
| Review verdicts ∈ `{SHIP, NEEDS_WORK, MAJOR_RETHINK}` | Carmack-style review verdicts | `05-internals/01-agents.md § Verdict-tag protocol` |
| Criterion verdicts ∈ `{MET, MET_WITH_CONCERNS, NOT_MET, INCONCLUSIVE}` | Per-criterion intent acceptance | itd-1 |
| `task_classes` (capability_scope tokens) ∈ `{oracle_review, intent_review, spec_planning, code_rescue, principle_distillation, lifeboat_packing, audit, lint, surface_render, cross_document_audit}` | Closed enum, agent frontmatter. Machine-readable source of truth: the Go binary's `task_classes` schema (`internal/core/...`) — a cross-check test fails if this table and the schema diverge. PR-to-extend. | itd-5 extension (idea-4) — lean ~10 tokens drawn from current abcd surfaces |
| `frozen_content_hash` | SHA-256 hex string written to PRD frontmatter by `/abcd:intent plan` at freeze time. Non-self-referential: provenance fields included, operational fields excluded. Recipe documented in `prd.schema.json`. | spc-3 task .5 |
| `intent_source_hash` | SHA-256 hex string computed over the parent intent's body + stable frontmatter. Written by grill skill as `grilled_intent_hash`; copied to PRD as `source_intent_hash`. Recipe documented in `prd.schema.json`. | spc-3 task .5 |
| `planning_attempt_id` | UUIDv4 written to PRD frontmatter and the durable attempt journal by `/abcd:intent plan`. Used by GR004 to detect stale planning attempts. | spc-3 task .5 |
| `prd_grandfathered` | Boolean frontmatter field on pre-spc-3 planned intents. When `true`, GR002 and GL005 are suppressed-as-info (not blocker). Cleared on regrill. | spc-3 task .5 |
| Grandfather migration | One-shot sweep at spc-3 ship time: appends `prd_path: null` + `prd_grandfathered: true` to every intent in `planned/` that predates the PRD requirement. | spc-3 task .5 |
| `promote-check mode` | The intent lint's `--promote-check <intent.md>` mode evaluates the intent as if it were already in `planned/`, firing GR002 and GL005 at planned-state severity for pre-flight checks. | spc-3 task .5 |
| `abandon-attempt mode` | The intent lint's `--abandon-attempt <itd-N>` mode (also `/abcd:intent plan --abandon-attempt <itd-N>`) clears a stale planning attempt: removes the attempt journal and sets `planning_attempt_id: null` in the PRD. Freeze fields (`frozen_at`, `frozen_content_hash`, `spec`) are preserved — the PRD itself is not un-frozen. Remediation for GR004. | spc-3 task .5 |
| `failure_mode_tag` ∈ `{hallucination, scope_drift, stale_context, under_specification_blindness, format_violation}` | Closed enum, in the later-phase Frontier Awareness intent | idea-4, a later phase |
| `work_item.type` ∈ `{intent_promotion, spec_task, command_run}` | Coordination claim unit (a later phase) | itd-33, a later phase |
| Claim primitives ∈ `{take, yield, escalate}` | Coordination conflict-resolution verbs (a later phase; cooperative-checkpoint semantics, no mid-flight abort) | itd-33, a later phase |
| Escalation choice ∈ `{wait_then_swap, swap_now, sequence, keep_both}` | Human-resolved escalation outcomes (a later phase) | itd-33, a later phase |
| `release.outcome` ∈ `{completed, abandoned}` | Audit-log outcome on claim release (a later phase) | itd-33, a later phase |
| `claim.status` ∈ `{active, paused, released}` | Three-state claim lifecycle; `paused` interlocks with itd-29's pause/resume/rewind (a later phase) | itd-33, a later phase |
| `recall` | Keyword-list field on each rules.json domain — natural-language phrases matched word-boundary against user prompts to trigger rule injection. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `domain` | Uppercase grouping key in rules.json (e.g. `COMMITTING`, `DOCUMENTATION`). Domains carry state (`active` / `dormant`), recall keywords, and rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `dormant` | rules.json domain state value (opposite of `active`). Recall-match injection is skipped for dormant domains; star-command `*<DOMAIN>` still activates them. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `active` | rules.json domain state value (default). Recall-match injection fires when prompt matches recall keywords. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `*<DOMAIN>` | Leading-anchored uppercase prompt prefix that explicitly activates a domain regardless of recall match. Regex `(?:^|\s)\*([A-Z][A-Z0-9_]*)(?=$|\s)`. Hyphen/period/slash following the domain name fails the boundary (no activation). Multi-star activates multiple domains in left-to-right order. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.5` |
| `force_refresh_every_n` | `.abcd/config.json` field under the `rules.` namespace. Integer; default 5. Every N prompts, the prompt-router hook forces a full re-inject regardless of dedup signature match (compaction recovery). | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.4` |
| `.abcd/config.json` | Repo-scope config file. Named in the in-repo carve-out in `05-internals/03-configuration.md` § The two `.abcd/` scopes, alongside `<repo>/.abcd/rules.json` and `.specstory/cli/config.toml`. Reads include `rules.force_refresh_every_n` and `docs.target`. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.8` |
| `rules.json` | Repo-scope rule overrides file at `<repo>/.abcd/rules.json`. Named in the in-repo carve-out in `05-internals/03-configuration.md` § The two `.abcd/` scopes. JSON Schema 2020-12; schema in the Go binary (`internal/core/...`). | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `COMMITTING` | Default plugin-bundled domain. Recall keywords trigger injection of commit-discipline rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `DOCUMENTATION` | Default plugin-bundled domain. Recall keywords trigger injection of documentation-discipline rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `ROADMAP` | Default plugin-bundled domain. Recall keywords trigger injection of roadmap/intent rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `ISSUES` | Default plugin-bundled domain. Recall keywords trigger injection of issue-tracking rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `INTENTS` | Default plugin-bundled domain. Recall keywords trigger injection of intent-capture rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `LIFEBOAT` | Default plugin-bundled domain. Recall keywords trigger injection of lifeboat/disembark rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `PII` | Default plugin-bundled domain. Recall keywords trigger injection of PII-protection rules. | spc-14 spec + `spc-14-modular-rules-loader-prompt-router-hook.1` |
| `managed-repo` | Folder kind: a git repo abcd already manages — has an ABCD marker block (or an in-tree `.abcd/`, or an `index.json` entry) and a `.git/` directory. Per brief `04-surfaces/01-ahoy.md` § 0 Folder classification. | spc-15 spec + `spc-15-folder-classification-workspacesjson.3` |
| `unmanaged-repo` | Folder kind: a git repo without abcd management; bare `/abcd:ahoy` offers install to adopt. | spc-15 spec + `spc-15-folder-classification-workspacesjson.3` |
| `unmanaged-folder` | Folder kind: not a git repo and no abcd markers; nothing to act on. | spc-15 spec + `spc-15-folder-classification-workspacesjson.3` |
| `root_commit` | Immutable repo identity key in `index.json`, computed via `git rev-list --max-parents=0 HEAD`. Survives rename, remote move, and GitHub-handle change. | spc-15 spec + `spc-15-folder-classification-workspacesjson.2` |
| `index.json` | History-store registry at `~/.abcd/history/index.json` recording each repo's identity + lineage. The **sole user-scope registry** (abcd is single-repo, adr-28 — there is no `workspaces.json`). Keyed on immutable `root_commit`. | spc-15 spec + `spc-15-folder-classification-workspacesjson.5` |
| `aliases` | Array of prior names a repo has had (e.g., renamed on GitHub). Recorded in per-root-sha `meta.json`. | spc-15 spec + `spc-15-folder-classification-workspacesjson.5` |
| `supersedes` | Lineage cross-ref in `index.json` repo entry: this entry was re-founded from another root-sha. | spc-15 spec + `spc-15-folder-classification-workspacesjson.5` |
| `superseded_by` | Lineage cross-ref in `index.json` repo entry: this entry was superseded by another root-sha (re-founding produces a new entry). | spc-15 spec + `spc-15-folder-classification-workspacesjson.5` |
