# Terminology / Glossary Consistency Audit

Specialist pass: terminology and glossary consistency. **Corpus:** `.abcd/`, `docs/`, `AGENTS.md`, `README.md`, `CONTRIBUTING.md`. Glossary at `.abcd/development/brief/glossary/` (core/ 10 files, distribution/ 3, interview/ 2, plus `README.md`, `_template.md`).

**Framing caveat (affects everything below):** `.abcd/work/CONTEXT.md:29-33` states the copied `.abcd/development/` record "still describes the *old* architecture … It is the starting spec, not current truth, until Phase 0.5 reconciles it." This repo is the Phase-0 Go rebuild; the Python tooling the glossary/lint docs describe lives in the frozen sibling `abcdDev` repo. Many findings below are "inherited spec vs. this repo's reality" gaps that Phase 0.5 is supposed to close — they are still real drift in the committed record.

## (A) Glossary integrity

1. **README index is out of date — 4 of 15 entries unindexed.** `glossary/README.md` "Term Index" lists 9 core terms + 2 interview terms. Missing: `core/disembark.md` (exists, stable, not indexed) and the **entire `distribution/` context** (`end-user.md`, `release.md`, `version.md`). The "Directory Layout" diagram (README lines ~85-102) likewise omits `disembark.md` and `distribution/`, and the format-spec table says "Current contexts: `core`, `interview`" — `distribution` absent. No phantom entries (everything indexed exists).
2. **README self-identifies as a different directory.** The layout diagram is rooted at `terminology/`, and validation commands reference `python3 scripts/abcd/lint_terminology.py .abcd/development/foundation/terminology/` — the actual directory is `.abcd/development/brief/glossary/`, and neither `scripts/abcd/lint_terminology.py` nor `scripts/abcd/schemas/terminology.schema.json` exists in this repo. adr-7 and adr-11 also use the old `terminology/core/…` paths (historical records, tolerable — but the glossary's own README is live documentation).
3. **Dead references in README:** `.abcd/config.json` (`terminology_exclude_files` allowlist — file does not exist here; the README itself admits the entry is "inert documentation — no `scripts/abcd/lint.py` consumer reads it yet"); `tests/abcd/test_issue_schema.py` (no `tests/` dir); `skills/abcd-intent-grill/phase-1-glossary-mode.md` (no `skills/` dir).
4. **Template compliance:** all 15 entries carry the full required frontmatter (`term`, `bounded_context`, `definition`, `aliases`, `forbidden_synonyms`, `status`, `introduced_in`) and all template body sections (narrative, When to use, When NOT to use, Examples, Related terms). Deviations:
   - **`core/lifeboat.md` has invalid YAML**: `starts_when` is a plain scalar beginning with a backtick, which is a YAML reserved indicator. This would fail frontmatter parse (`TM003` in the described lint). Same field style is fine elsewhere (voyage/session use plain prose).
   - **Two file shapes coexist**: core/interview files open with an HTML comment *before* the `---` frontmatter; distribution files open with `---` directly. The README spec says files "MUST begin with a YAML frontmatter block". One of the two shapes is non-conformant, depending on parser tolerance.
   - `voyage.md` and `session.md` add a `## Lifecycle` section absent from `_template.md` (harmless extension); `lifeboat.md` has lifecycle fields but no such section (inconsistent presentation).
   - `introduced_in` values: `phase-1`, `itd-67`, and `adr-9` (phase.md) — spec says "Version or intent ID"; `adr-9` is neither.
5. **Cross-file alias↔forbidden collisions (no lint covers these — `TM010` is per-file only):**
   - `distribution/release.md` `aliases: [… "snapshot"]` vs `core/disembark.md` and `core/lifeboat.md` `forbidden_synonyms: [… "snapshot" …]`.
   - `distribution/end-user.md` `aliases: [… "user"]` vs `core/persona.md` `forbidden_synonyms: ["user", …]`.
   The distribution bodies acknowledge the tension in prose, but the machine fields directly conflict, and `GL002` as specified ("intent body uses a value listed in **any** term's forbidden_synonyms" — blocker) is context-blind: writing the canonical distribution alias "snapshot"/"user" in an intent body would mechanically fire a blocker.
6. **Two competing "canonical glossaries".** `glossary/README.md` calls itself "the canonical terminology glossary"; `02-constraints/04-naming.md § Vocabulary-registration` says "Terms are registered in **this glossary file** (`02-constraints/04-naming.md`)" and hosts a ~60-row Reserved Vocabulary table. Two stores both claiming term-registry authority, with no cross-reference between them. *(Resolved by the review-time decision: `glossary/` is the SST — see `00-summary.md` § 5.)*

## (B) Per-term verdicts (problem terms only)

**spec** — residual forbidden synonym + phantom reference.
- `.abcd/development/brief/04-surfaces/05-intent.md:167`: "Auto-running the reviewer off that queue is still deferred (no **epic** currently owns…" — unexempt prose use of the `core/spec` forbidden synonym in a path the `TM002` contract itself names in-scope (`.abcd/development/brief/**/*.md`).
- `.abcd/development/intents/planned/itd-24-reflect-command.md:11`: `glossary_terms_used: […, core/epic, …]` — qualified ID for a term file that no longer exists (renamed to `core/spec` per adr-11) → dangling reference (`GL001` class) in a *planned* (live) intent.
- Frontmatter `surface_history.reason` fields in `disciplines/itd-1-acceptance-gates.md:11` ("applies to every other epic") and `drafts/itd-30…:8` ("its epic depends on") still say epic — exempt from TM002 (frontmatter not scanned) but visible drift.

**intent** — definition contradicts corpus layout; forbidden list collides with first-class concepts.
- `glossary/core/intent.md` body: "Intent files live under `.abcd/development/roadmap/intents/`" — actual location is `.abcd/development/intents/{drafts,planned,shipped,disciplines,superseded}` (roadmap/ contains only phases/ and rfcs/). The stale `roadmap/intents` phrasing also appears in `AGENTS.md:34` and `.abcd/README.md:9`.
- `forbidden_synonyms: ["ticket", "story", "issue", "requirement"]` — but "issue" is a first-class abcd concept (`iss-N`, the `/abcd:capture` issue ledger, 387 corpus hits for capture) with **no glossary entry**, and "User Stories" is a mandated PRD section (`06-delivery/02-verification-matrix.md:63` "User Stories ≥5"). A mechanical GL002 scan cannot distinguish these; the definition set is internally at war with the corpus.

**lifeboat / disembark** — glossary forbids "snapshot"; the brief uses it routinely for exactly this artefact.
- `01-product/02-context.md:37`: "`.abcd/lifeboat/` in any repo is the latest **disembark snapshot**, regenerable from current state."
- `04-surfaces/03-embark.md:18`: "that repo's latest **disembark snapshot**"; `04-surfaces/02-disembark.md:148`: "when the new **snapshot** lands at `<path>` … never … stale snapshots"; `00-meta.md:32`: "**Disembark snapshots** … provide the audit-traceable history chain."
- These are the forbidden-synonym usage the entries prohibit ("Do not use 'export', 'backup', or 'snapshot'"). Plus the lifeboat.md YAML defect (A.4) and disembark missing from the README index (A.1).

**embark** — two meanings, only the minority sense glossaried.
- `interview/embark.md` defines embark solely as "the opening move of a grill session".
- The dominant corpus sense (~most of 382 hits) is the `/abcd:embark` **command** that unpacks a lifeboat (`04-surfaces/03-embark.md:1` "`/abcd:embark` — Unpack a Lifeboat"; naming.md:13 "board a new ship → unpack the lifeboat"). `core/lifeboat.md` even warns the two are "distinct" — but no `core/embark` entry exists, so GL003 (cross-context collision) can never fire and the glossary actively defines the wrong primary sense. Forbidden "start"/"begin" are ubiquitous ordinary words (mechanical-lint hazard).

**session** — the polysemy adr-7 used to justify bounded contexts is still unresolved.
- adr-7 Decision 3: "`session` means a grill interview session in the intent stage and an agent-runtime session in the oracle layer — two different concepts sharing one word." Only `interview/session.md` exists; the agent-runtime/host sense has no entry yet dominates whole documents: `05-internals/10-in-session-dispatch.md` ("the host session's `Task` tool", `backend="in_session"`), `04-surfaces/03-embark.md:5` "hunt the **originating session**", plus an operator-internal command literally named `session` (`04-surfaces/README.md:3`) and SpecStory transcript sessions (`05-internals/03-configuration.md`). At least three live senses; one defined.

**oracle** — new-architecture prose uses the forbidden vocabulary.
- `oracle.md` forbids "LLM … in workflow documentation". `AGENTS.md:49`: "**LLM review**/agent work is delegated to the host; native/CLI/API/MCP **oracles** are opt-in adapters" — the same sentence uses both, with "LLM review" naming the oracle role. `.abcd/work/CONTEXT.md:26`: "host-delegated **LLM**". Benign uses of "LLM classifier" (`04-surfaces/05-intent.md:39,117,197`) and "AI"/"model" throughout would all trip a mechanical GL002.

**phase** — usage consistent, but its forbidden list bans two words the corpus needs.
- "milestone" is a *defined load-bearing concept* — the end condition of a phase (`roadmap/README.md:9-10`, `phase.md` body itself) — yet is a forbidden synonym of both `phase` and `spec` and has no glossary entry of its own.
- "version"/"release" are now canonical distribution-context terms; phase.md's forbidden list predates the distribution context and GL002 is context-blind (see A.5).

**transport** — minor polysemy: `10-in-session-dispatch.md:56` labels the dispatch `backend` field "The *transport*" (agent-dispatch sense) vs the glossary's oracle-delivery sense; forbidden "prompt" collides with a whole subsystem vocabulary (prompt quality, `prompt_router_hook.py`, `lint_prompts.py` — `PQ` codes).

**voyage** — second, unglossaried sense: `.abcd/development/voyage/` is a *provenance directory* for embark/disembark history (`03-embark.md:98-102`), distinct from voyage-as-lifecycle. Forbidden "project" appears ~everywhere in ordinary sense (mechanical hazard only).

**Clean:** `brief`, `persona` (usage — the alias conflict is end-user's), `end-user`, `release`, `version` (the `abcd version` CLI verb and Makefile VERSION stamping match the distribution definition).

## (C) Missing-term candidates, ranked by load-bearing weight (word-boundary occurrence counts across corpus)

| Rank | Term | Count | Why it needs an entry |
|---|---|---|---|
| 1 | **grill** | 426 | Names the entire `interview` bounded context; both existing interview entries define themselves *in terms of* "grill session"; itd-27/adr-7 built on it. The glossary's own anchor noun is undefined. |
| 2 | **surface** | 430 | Core architecture noun with ≥3 senses: user-facing command surface (`04-surfaces/`), operator-internal surface, and the Go package `internal/surface/`. AGENTS.md boundary rules depend on it. |
| 3 | **PRD** | 222 | The grill Phase-2 artefact; subject of the entire freeze contract (adr-7), GR002-GR005, `prd.schema.json`. Undefined. |
| 4 | **audit** | 836 | Heavily polysemous: reserved `/abcd:audit` verb, phase audit, fidelity audit (itd-1), workflow audit (zizmor), hash-chain audit. Highest-count undefined noun; senses actively colliding. |
| 5 | **gate** | 512 | Acceptance gates, promotion gate, pre-push gate, review-queue gate, fn-76 validation gate, safety gate (itd-62). No definition anywhere. |
| 6 | **session** (agent-runtime sense) / **embark** (command sense) | 594/382 | Not new words — missing *second bounded-context entries* the existing design (adr-7 Decision 3, GL003) explicitly anticipates. |
| 7 | **memory** | 639 | Three documented senses (`.abcd/memory/` substrate, `/abcd:memory` surface, legacy root `memory/` snapshot — `05-internals/03-configuration.md:225-243`) and a naming.md enum, but no glossary term. |
| 8 | **fn-N / itd-N / iss-N / adr-N** | 1271/1942/6/480 | The identifier grammar underpinning every cross-reference. Partially described inside spec.md/intent.md bodies; the grammar itself (and iss-N entirely) is unregistered. |
| 9 | **harness** | 239 | "LLM harnesses", "harness Protocol", host-harness portability — a boundary concept for the host-delegation architecture. Undefined. |
| 10 | **milestone** | 57 | Low count but structurally critical: defined concept (phase end condition) that is simultaneously a forbidden synonym of two terms — only its own entry resolves that. |
| 11 | **ahoy / launch / capture / reflect** | 354/480/387/72 | Four of nine user-facing surfaces; `disembark`/`embark` got glossary entries, these didn't. `launch`'s nautical-not-run sense is pinned only in naming.md:19. |
| 12 | **adapter** | 67 | Has an entire brief chapter (`05-internals/02-adapters.md`, semantic-role naming rule) and appears in AGENTS.md boundaries; no term file. |
| 13 | **flow / flow-next** | 678 | External-boundary vocabulary (`.flow/`, flowctl, flow-state) that adr-11 treats as a preserved boundary; deserves a boundary-marking entry. |
| 14 | **task** | 298 | The `fn-N.M` sub-unit of a spec; mentioned in spec.md's definition but unregistered — the exact "task drift" a terminology audit worries about is unpoliced. |
| 15 | Lower tier: **backend** (255), **payload** (200, launch payload), **manifest** (169), **logbook** (147, in naming.md table), **sub-verb** (147), **dev-sync** (125), **loot** (130) / **dredge** (75) (reserved verbs, naming.md only), **marker block** (57), **checkpoint** (50 — lifeboat.md says it "has flow-control connotations" but nothing defines it), **bare invocation** (25). |

## (D) Enforcement assessment

**What is specified (extensively):**
- `05-internals/06-lint.md` — a full lint contract: `GL001-GL005` (glossary), `GR001-GR005` (grill), `TM002-TM011` (terminology schema + the `epic` prose scan), `VR001` (vocabulary registration), `SD001`, `XD001-XD007`, with severities, `.abcd/config.json` overrides, and four trigger points: "**Pre-commit hook** … runs `scripts/abcd/lint.py --since-staged`. Fails the commit on any `severity: blocker` finding"; "**`/abcd:intent plan` promotion gate**: runs `intent_lint.py --promote-check` … Blocks on `severity: blocker`"; "**CI workflow** (`.github/workflows/lint.yml`) … over only the lines changed in the PR"; "**Full-corpus promote-backlink workflow** (`.github/workflows/lint-corpus.yml`…)".
- `adr-7` Decision 2 — "**Cite-or-fail lint enforcement** … a promotion blocker, not advisory … advisory lint produces 'glossary drift' within two sprints".
- `04-naming.md § Vocabulary-registration` — "HARD from the start … `intent_lint.py` blocks at plan-review on missing registrations. Lint code (reserved): `VR001`."
- `adr-11` — "`forbidden_synonyms: [epic]` … turns the rename into a *guarded* invariant: `lint_terminology.py` flags any future regression."

**What actually exists in this repo: none of it.**
- No `scripts/` directory at all; no `intent_lint.py`, `lint_terminology.py`, `lint_prompts.py`, `lint.py`, no `terminology.schema.json`.
- `.github/workflows/` contains only `ci.yml` — gofmt/build/vet/test (Go), gitleaks, zizmor. **Zero markdown or terminology linting.** No `lint.yml`, no `lint-corpus.yml`.
- No `.pre-commit-config.yaml`. The only hook is `.githooks/pre-push` → `make preflight` (Go build/vet/test/race).
- No `.abcd/config.json`, no `.flow/`, no `tests/`, no `skills/`, no `agents/`.

**Verdict:** In `abcd-cli`, terminology enforcement is **specified in inherited documentation but entirely non-existent operationally** — neither automated nor even manually runnable (the tools aren't in the tree). The lint contract's dozens of "**Delivered** in fn-NN" claims describe the frozen Python sibling repo, and are false as statements about this repo until Phase 0.5 reconciliation (which `CONTEXT.md` explicitly flags). *(Plan-review reframe: the missing tooling is expected at this stage; what counts against the plan is the design gaps below.)* Even taken as future spec, it has designed-in gaps: `GL002` is context-blind (conflicts with the distribution aliases, A.5); `TM002` prose enforcement covers only the single synonym `epic` ("broad multi-synonym prose enforcement is future work"); the `terminology_exclude_files` allowlist is self-described "inert"; `XD` mechanical codes are deferred; and nothing lints cross-file alias↔forbidden collisions.

## (E) Naming-convention compliance (`02-constraints/04-naming.md`)

**Compliant:**
- The maritime table's nine user-facing surfaces exactly match `04-surfaces/README.md`'s table (ahoy, disembark, embark, launch, intent, capture, memory, `/abcd`, reflect); reserved verbs (dredge, loot, audit, reflect) are consistently marked later-phase in both.
- `launch` nautical-sense discipline: no usage of "launch" meaning "run a program" found in the corpus.
- Bare-command-as-render: surface docs comply (e.g., `02-disembark.md:139` bare invocation "never mutates state"); earned sub-verbs match the naming.md list.
- Technical-file exemption respected (`config.json`, `rules.json`, etc.).

**Non-compliant / drifted:**
1. **The vocabulary-registration table is riddled with dead paths in this repo**: `scripts/abcd/reflect_writer.py`, `scripts/abcd/setup_wizard/`, `scripts/abcd/schemas/task_classes.json` ("a cross-check test fails if this table and the JSON diverge" — neither table-consumer nor JSON exists), `.abcd/config.json` (row at naming.md:105 names it as an existing file), `.flow/` reviews paths. The registration requirement itself ("MUST be registered … `intent_lint.py` blocks") is unenforceable here.
2. **Glossary dualism** (see A.6): naming.md § "Glossary location" declares *itself* the registration target while `brief/glossary/` claims canonical status; no rule says which wins for a given term. *(Resolved: glossary/ is the SST.)*
3. **`brief/README.md` layout tree omits `glossary/`** entirely (only `.abcd/development/README.md` links it) — a directory-coverage/indexing gap for the glossary's host directory.
4. **Stale `roadmap/intents` phrasing** in `AGENTS.md:34` and `.abcd/README.md:9` vs the actual `development/intents/` layout — a naming/location claim the corpus contradicts.
5. Borderline: itd-29's operator surface plans a `run status` sub-verb (`04-surfaces/README.md:3`) — "status" names what bare renders, the exact shape `SD001` bans; `/abcd` documents `status`/`help` as "byte-identical aliases" of bare, which the discipline text doesn't explicitly permit either.
6. The Go code (`internal/`, `cmd/`) currently implements only `version` + bare status board — too early to violate the maritime convention; no conflicts there.

**Highest-value fixes, in order:** (1) reindex `glossary/README.md` (add disembark + distribution context, fix the `terminology/`→`glossary/` self-description and dead script paths); (2) fix `core/lifeboat.md` YAML; (3) resolve the two alias↔forbidden cross-context collisions (snapshot, user) and the context-blind GL002 spec; (4) add second-context entries for `embark` (command) and `session` (agent-runtime) plus a `grill` entry; (5) sweep the four "snapshot"-for-lifeboat and one "epic" prose residues; (6) fix `core/intent.md`'s `roadmap/intents/` path (and AGENTS.md/.abcd/README.md echoes); (7) decide the naming.md-vs-glossary registry question at Phase 0.5 *(decided: glossary/ wins)*, and state plainly in `06-lint.md` that no lint tooling exists in this repo yet.
