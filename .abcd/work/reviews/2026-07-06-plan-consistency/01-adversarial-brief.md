# Adversarial Review — abcd-cli Design Brief

Specialist pass: red-team review of the brief's content. Base directory (all paths below are under it): `.abcd/development/brief/`.

Summary: **8 critical, 19 major, 12 minor** findings. The brief's two worst structural problems are (1) chapter 01-product and chapter 02-constraints describe a *pre-build design*, while chapters 04/05/06 describe a *partially-shipped system* with dozens of intents/specs (fn-30…fn-83, itd-42…itd-69) that 01-product says don't exist — so "adr-5: brief is current state" is violated by the brief itself; and (2) the repo-vs-workspace `.abcd/` scoping rule contradicts nearly every surface.

---

## CRITICAL

### C1. Launch `dry-run` semantics: press release demands hard-fail; launch surface forbids it
- **Files:** `01-product/01-press-release.md:32` vs `04-surfaces/04-launch.md:10,163`; perpetuated by `06-delivery/02-verification-matrix.md:27`
- **Evidence:** Press release AC: "**when** `/abcd:launch dry-run` runs, **then** preflight hard-fails with the offending file/line cited and no payload is written." Launch surface: "`/abcd:launch dry-run` — **report-only preview, always exit-0** (a preview never blocks)… running the *full* gate suite and **hard-failing** on a finding (exit non-zero) is the Phase-5 `ship` verb's behaviour, not dry-run's." Launch acceptance (line 163) re-asserts "still **exits 0**". Matrix row 27 ("Launch preflight | Deliberate PII fixture → hard-fail with preflight.md path") doesn't name the verb, so a test author following the press release will write the opposite test from one following the surface.
- **Why:** The product chapter's acceptance criterion is directly falsified by the surface contract. An implementer cannot satisfy both.
- **Fix:** Change press-release AC to `/abcd:launch ship` (or "prints the would-refuse finding, exits 0" for dry-run); add the verb to the matrix row.

### C2. Repo-scope `.abcd/` is simultaneously forbidden and required everywhere
- **Files:** `05-internals/03-configuration.md:212` vs `02-constraints/01-platform.md:7,15`, `01-product/02-context.md:37,45`, `04-surfaces/05-intent.md:269,288`, `04-surfaces/09-reflect.md:81-82`, `05-internals/10-in-session-dispatch.md:99`, `04-surfaces/04-launch.md:31`, `04-surfaces/05-intent.md:427` (`.abcd/coordination/`), `02-constraints/02-dependencies.md:21` (`.abcd/validation_disciplines/`)
- **Evidence:** Configuration: "**There is no repo-scope `.abcd/` — with four named exceptions**… `<repo>/.abcd/rules.json` plus `<repo>/.abcd/config.json`… Everything else a repo would need is workspace-scoped… wherever the rest of this file says 'repo' for those artefacts, read 'workspace'." Yet platform says "the design record (this brief…) lives at `.abcd/development/`" *in the dev repo* and "Lifeboat path: `.abcd/lifeboat/` (per-repo…)"; context says "`home` = current repo's `.abcd/lifeboat/`"; intent puts PRDs at `.abcd/intents/<itd-N>/prd.md`; reflect writes `.abcd/retrospectives/…` "a peer of `.abcd/intents/` and `.abcd/logbook/`"; in-session dispatch uses `.abcd/tmp/in-session/`; launch reads `.abcd/launch.allow`; corpus lives at `.abcd/corpus.json` (context:45) while the carve-out explicitly excludes `corpus.json` at repo scope (`04-surfaces/01-ahoy.md:61`) and ahoy uninstall says it preserves "`.abcd/` … (meta, config, **corpus**, rules…)" (`01-ahoy.md:291`).
- **Why:** The "read repo as workspace" patch note only covers 03-configuration.md itself; every other chapter still specifies repo-scope `.abcd/` trees. An implementer cannot determine where the lifeboat, logbook, PRDs, coordination locks, retrospectives, or corpus.json physically live for a standalone repo vs a repo-inside-workspace.
- **Fix:** Pick one scoping model, then sweep every `.abcd/` path in 01/02/04/05 chapters and state per-artefact scope in one table.

### C3. `meta.json` is abolished and load-bearing in the same file
- **Files:** `04-surfaces/01-ahoy.md:61` vs `01-ahoy.md:139,170-171,254,334-340`; `05-internals/03-configuration.md:5-16`; `06-delivery/02-verification-matrix.md:6`
- **Evidence:** Ahoy carve-out row: "repo install writes `<repo>/.abcd/config.json` and `<repo>/.abcd/rules.json` (**no `meta.json` / `corpus.json`** … setup metadata lives at `config.json["meta"]`)." Same file, detection step 3: "**`.abcd/` skeleton** — presence of `meta.json`, `config.json`, `rules.json`"; step 11: "read `meta.json.setup_version`"; install step 10: "Stamp `setup_version` + `setup_date` in `meta.json`." Configuration devotes a whole section to `.abcd/meta.json` "(unchanged from the original cut)". Matrix: "ahoy upgrade | … meta.json updated."
- **Why:** A single surface file specifies both "meta.json does not exist at repo scope" and "read/write meta.json" as steps of the same command. Unimplementable as written.
- **Fix:** If fn-16 moved setup metadata into `config.json["meta"]`, rewrite detection steps 3/11, install step 10, the configuration section, and the matrix row.

### C4. Out-of-scope chapter contradicts shipped/current status of at least six intents
- **Files:** `06-delivery/03-out-of-scope.md:53,57,60,65,66,68,69` vs `01-product/01-press-release.md:19`, `01-product/04-scope.md:17`, `05-internals/06-lint.md:104` (FS001), `04-surfaces/05-intent.md:370-404` (itd-50), `05-internals/03-configuration.md:58-114` (itd-53, itd-50), `05-internals/04-universal-patterns.md:186-211` (itd-52), `05-internals/07-memory.md:176-198` (itd-39), `01-product/03-mental-model.md:75` + `04-surfaces/05-intent.md:41` (itd-44 / fn-56)
- **Evidence:** Out-of-scope (self-declared "canonical for later-phase items") lists: "itd-29 — Autonomous-run resilience", "itd-39 — Scope-aware memory retrieval", "itd-44 — A fourth intent kind…", "itd-49 — Flow-state drift…", "itd-50 — The audit loop…", "itd-52 — abcd warns when you reach past it…", "itd-53 — A shipped intent no longer drifts…". Meanwhile: press release/scope present `/abcd:run` as "the itd-29 autonomous-run operator surface" in the current operator-internal command set; lint contract: "`FS001` … **Delivered** in fn-41 (itd-49)"; intent surface documents the itd-50 audit loop (fn-52) as live policy with config keys; configuration documents `review.autodrain` "(fn-43, itd-53)" as a current key with adr-16; universal-patterns § 9 documents itd-52's boundary with **shipped** doctor probes (fn-33, fn-42.2); the itd-44 `decision` capture verdict is written into the mental model and intent lifecycle as current behaviour.
- **Why:** The chapter the whole brief points at as "the canonical later-phase list" (press release:25, scope:29) declares later-phase things the rest of the brief documents as delivered. An implementer scoping the build from 06-delivery would delete shipped machinery or re-plan delivered work.
- **Fix:** Re-derive the list with the file's own enumeration command against actual lifecycle directories, and strike the shipped/phased-in entries (as done for itd-46/48).

### C5. 01-product scope ("thirteen intents, seven commands, 15 agents") is stale against the rest of the brief
- **Files:** `01-product/04-scope.md:19-23`, `01-product/03-mental-model.md:71` vs `06-delivery/03-out-of-scope.md:16-20`, `04-surfaces/README.md:14-15`, `02-constraints/04-naming.md:66-68`, `05-internals/01-agents.md:22`
- **Evidence:** Scope: "**Thirteen phased intents — ten standalone plus three disciplines:** itd-2, itd-3, itd-4, itd-6, itd-7, itd-27, itd-28, itd-34, itd-36, itd-40 … itd-1, itd-5, itd-37." Out-of-scope's own comment: "intents that have already left drafts/ … (e.g. itd-6 planned, itd-27/28 shipped, itd-37 a discipline, and **itd-20/24/63/69 planned under fn-83**)… and **itd-34,36,40,42 (later phased-in…)**." Surfaces README ships `/abcd` (itd-20) and `/abcd:reflect` (itd-24) as surfaces #8 and #9; naming's reserved vocabulary registers fn-83 artefacts (`reflection-composer`, `setup-wizard`); agents catalog marks `reflection-composer` "shipped — fn-83".
- **Why:** The chapter the README calls the entry point ("Start with 01-product…") describes a different, smaller project than chapters 04–06. itd-42, itd-20, itd-24, itd-63, itd-69 are phased/planned per 06-delivery but absent from the "thirteen"; the mental model's "The six phases target thirteen intents" is false by the brief's own accounting.
- **Fix:** Regenerate 04-scope's intent list from the phase docs (its own stated SSOT) instead of hand-counting.

### C6. `/abcd:reflect` and the top-level `/abcd` are simultaneously "reserved, later phase" and "shipped surface"
- **Files:** `02-constraints/04-naming.md:27,38` vs `04-surfaces/README.md:14-15`, `04-surfaces/08-abcd.md`, `04-surfaces/09-reflect.md`, `05-internals/01-agents.md:22`
- **Evidence:** Naming (twice — exemptions list and "Reserved meta-development commands" table): "`/abcd:reflect` | major-milestone retrospectives (see itd-24 — **a later phase**). **Reserved**". Surfaces README row 9 ships it; 09-reflect.md documents built behaviour ("owned by task `fn-83-operator-surfaces-manifest-lockstep.3`", writer refusals, tests); agents catalog: "`reflection-composer` | shipped — fn-83". The `/abcd` board (itd-20) is likewise a live surface (08-abcd.md, "proven by two tests") but appears nowhere in naming's command tables, scope's command list, or the press release.
- **Why:** The constraints chapter — "hard locked decisions" — says a shipped command doesn't exist yet.
- **Fix:** Move reflect out of naming's reserved table into the live table; register `/abcd` (bare board) in naming and scope.

### C7. Implementation language/runtime is never specified — and everything concrete contradicts a "Go CLI"
- **Files:** `02-constraints/01-platform.md` (whole file), `05-internals/03-configuration.md:363-375`, `06-delivery/01-build-sequence.md:22-23`, plus ~40 references to `scripts/abcd/*.py`
- **Evidence:** The platform chapter ("locked decisions about *where* abcd runs") locks repos and paths but never states a language, runtime, or minimum toolchain. Every concrete implementation reference is Python + bash: "`scripts/abcd-cli` — bash wrapper → `abcd_cli.py`… `scripts/abcd/` — **Python support package**"; "Define **harness.py** interface"; `personas.py`, `intent_lint.py`, `reflect_writer.py`, `flock(2)`, `os.rename`. The project premise for this review ("abcd-cli is a Go CLI about to be built") appears nowhere in the brief; the brief describes a **Claude Code plugin** ("abcd is a Claude Code plugin", `01-product/01-press-release.md:5`).
- **Why:** Either the premise is wrong or the brief is: if the build target is a Go CLI, the entire internals chapter (Python module names as contracts, `.py` paths baked into the glossary and lint contract) is unusable as a spec; if it's the Python plugin, the constraints chapter still fails to lock the runtime (Node? Python version? macOS-only pieces are noted only for RP). Either way, a "before we build" brief must state the language/runtime and doesn't.
- **Fix:** Add a language/runtime/toolchain lock to `02-constraints/01-platform.md`; if the project pivoted to Go, the pervasive `.py` contract paths must be renamed or marked as legacy.

### C8. `/abcd:intent plan` has two incompatible specifications (with vs. without mandatory PRD/grill)
- **Files:** `04-surfaces/05-intent.md:120-144` (§1 lifecycle) and `:200` (§2 table) and `06-delivery/02-verification-matrix.md:34` vs `04-surfaces/05-intent.md:294-306` (§5) and `05-internals/06-lint.md:56` (GR002)
- **Evidence:** §1 lifecycle step 2 and the §2 subcommand row specify plan as: lint AC → propose kind → call `/flow-next:plan` → link → move — **no PRD anywhere**. Matrix row 34 tests exactly that flow. §5 specifies plan as a 10-step freeze sequence beginning "1. Reads `prd_path` from intent frontmatter; **refuses if null (no PRD yet)**", making grill (the only PRD producer) a hard prerequisite; GR002 is "blocker for newly-promoted".
- **Why:** Two mutually exclusive flows for the same verb in the same file. An implementer following §1/§2/the matrix builds a plan verb that GR002 then blocks 100% of the time (or that skips freeze/provenance entirely). Also unresolved: `/abcd:intent ship <itd-N>` from `drafts/` "runs full pipeline (plan + plan-review first)" — with the §5 rules this must refuse for any ungrilled draft, which no text acknowledges.
- **Fix:** Fold the PRD gate into the §1 lifecycle diagram and §2 table (or mark §5 as the superseding flow); update matrix row 34; specify ship-from-drafts behaviour for ungrilled intents.

---

## MAJOR

### M1. Agent count: 15 vs 16 in seven places
- **Files:** `05-internals/01-agents.md:3` ("declares **16 agents**", table has 16 rows) vs `05-internals/README.md:7` ("15-agent catalog"), `01-product/01-press-release.md:19` ("15 agents"), `01-product/04-scope.md:25` ("15 agents, 11 adapters"), `05-internals/05-prompt-quality.md:3` ("15 agents = 15 prompt `.md` files"), `05-internals/07-memory.md:157` ("**The agent count stays at 15**"), `05-internals/03-configuration.md:346` ("agents/ # 15 agents" — the tree lists 15 files, missing `reflection-composer.md`), `03-configuration.md:415` ("agents/<name>.md (×15)").
- **Fix:** One canonical count in 01-agents; everything else should say "see 01-agents" (the brief's own SSOT rule).

### M2. Top-level command count: six vs seven vs nine
- **Files:** `05-internals/08-skills.md:19` ("abcd ships **six top-level commands**: `/abcd:ahoy`, `/abcd:disembark`, `/abcd:embark`, `/abcd:launch`, `/abcd:intent`, `/abcd:capture`" — omits `/abcd:memory`) vs `01-product/04-scope.md:9-15` (seven, incl. memory) vs `04-surfaces/README.md:5-15` (nine, incl. `/abcd` and `/abcd:reflect`). Also `04-surfaces/05-intent.md:210` enumerates the bare-as-help convention over six commands, omitting memory/abcd/reflect, and `05-internals/03-configuration.md:326-328` command tree omits `abcd.md` and `reflect.md` while `08-abcd.md:9` requires `commands/abcd.md` and `09-reflect.md:110` requires `commands/abcd/reflect.md`.
- **Fix:** Make 04-surfaces/README the SSOT and delete counts elsewhere.

### M3. Presidio: hard launch dependency vs explicitly NOT the launch gate engine
- **Files:** `02-constraints/02-dependencies.md:10` ("**`presidio`** — PII scanner (**hard dependency, hard-fail in launch preflight**)") and `06-delivery/01-build-sequence.md:111` ("Scan stack (gitleaks + **Presidio** + custom regex…)") vs `04-surfaces/04-launch.md:15` ("PII scan… via the in-repo `src/pii.py` engine… (**Presidio is a recommended dependency / doctor-probe** in the broader design, **not** the wired ship-gate engine…)").
- **Why:** The dependencies chapter (constraints tier) asserts the exact thing the launch surface, citing fn-64 C1a, denies.
- **Fix:** Update 02-dependencies and build-sequence §10 to the fn-64 reality.

### M4. Verification matrix says the dev repo gets version-bumped; launch surface (adr-19) forbids it
- **Files:** `06-delivery/02-verification-matrix.md:30` ("Launch version bump | Default patch bump; `--version` override; **both repos updated**; `marketplace.json` updated") vs `04-surfaces/04-launch.md:97-102` ("Leaves the **dev** manifests UNVERSIONED… the dev repo's committed manifests carry no version") and `:130-131` ("the dev repo is never tagged").
- **Fix:** Matrix row → "public snapshot only; dev manifests unversioned (adr-19)".

### M5. Verification matrix says the hook is "registered" at install; ahoy says install NEVER writes hooks
- **Files:** `06-delivery/02-verification-matrix.md:9` ("ahoy rules-loader install | … `prompt_router_hook.py` **registered**") vs `04-surfaces/01-ahoy.md:249-253` ("**Hook registration** (`plugin-owned`, VERIFY-ONLY per fn-16 T1)… Install **NEVER writes** `hooks.json`").
- **Fix:** Matrix row → "hooks.json verified present (verify-only)".

### M6. ADR store has three competing canonical locations
- **Files:** `.abcd/development/decisions/adrs/` in `01-product/03-mental-model.md:75`, `04-surfaces/05-intent.md:41`, `glossary/README.md:125`; `.abcd/development/research/adr/` in `03-evidence/03-open-questions.md:28` ("**`.abcd/development/research/adr/`** — Architecture Decision Records"), `03-evidence/04-tradeoffs.md:67`, `05-internals/03-configuration.md:420`; `docs/development/decisions/adrs/` as embark's canonical **target** in `04-surfaces/03-embark.md:60,118`.
- **Why:** The itd-44 `decision` verdict routes captures "to the existing ADR store" — which of the three? Embark scaffolds ADRs into `docs/` (public, launch-included) while abcd's own live ADRs sit under `.abcd/` (launch-excluded) — a visibility-relevant divergence no text reconciles.
- **Fix:** Declare one store; fix 03-evidence and 03-configuration references; state embark's ADR target explicitly and why it differs (if intentional).

### M7. The glossary/terminology store has three competing homes
- **Files:** `glossary/` (this chapter) vs `glossary/README.md:75,81` (lint commands target `.abcd/development/foundation/terminology/…`) and `04-surfaces/02-disembark.md:111` ("terminology.md # rendered from `.abcd/development/foundation/terminology/<context>/<term>.md`") vs `02-constraints/04-naming.md:56-60` ("Every term … MUST be registered in **this glossary file**… Terms are registered in this file (`02-constraints/04-naming.md`)").
- **Why:** VR001 requires registration in `04-naming.md`; GL001–GL005/TM lint validate `terminology/` term files; the brief chapter is named `glossary/` but its own README calls the root `terminology/` and points the linter at a path that isn't this directory. An implementer cannot say which artefact "the glossary" is.
- **Fix:** One store, one name; make VR001 and the GL/TM families reference the same location; rename the chapter or the paths. *(Resolved by the review-time decision: `glossary/` is the SST — see `00-summary.md` § 5.)*

### M8. Glossary README omits the `distribution/` context and the `disembark` term it ships
- **Files:** `glossary/README.md:22` ("Current contexts: `core`, `interview`"), `:49-66` (layout lists core/ + interview/ only, and core/ list omits `disembark.md`), `:131-153` (Term Index has no distribution section and no disembark row) vs existing files `glossary/core/disembark.md`, `glossary/distribution/{end-user,release,version}.md`.
- **Fix:** Regenerate layout + index; update the "current contexts" sentence.

### M9. Every core/interview term file violates the glossary's own "MUST begin with frontmatter" rule
- **Files:** `glossary/README.md:13-14` ("Every term file MUST **begin** with a YAML frontmatter block (`---` delimiters)") vs `glossary/_template.md:1-2` and all `core/*.md`, `interview/*.md` (line 1 is `<!-- Adapted from mattpocock/skills (MIT)… -->`, frontmatter starts line 2). The `distribution/*` files comply.
- **Why:** Either the terminology linter (TM003 anchors "at line 1") rejects the template and 11 shipped term files, or the linter tolerates leading comments and the spec is wrong. Copying `_template.md` as instructed produces a non-conforming file.
- **Fix:** Move the attribution comment into the body (as distribution/ files do) or amend the spec.

### M10. `core/brief.md` declares the brief "immutable once approved" — contradicting adr-5 and the lifecycle taxonomy
- **Files:** `glossary/core/brief.md:19-20` ("written by a human stakeholder and **treated as immutable once approved**"; also "lives at the root of the `.abcd/` hierarchy" — actual: `.abcd/development/brief/`) vs `README.md:3` ("The brief reflects the project's *current* state"), `05-internals/04-universal-patterns.md:172` (Compounding-curated examples: "…**the brief itself**").
- **Fix:** Rewrite the term body to the current-state model.

### M11. `core/intent.md` declares intents "never edited after plan" — contradicting refine, Audit Notes, reclassification
- **Files:** `glossary/core/intent.md:19-21` ("frozen at promotion time and is an immutable input artefact — it is **never edited** after `/abcd:intent plan` is run") vs `04-surfaces/05-intent.md:198` (`refine` — "Interactive refinement of an existing intent… (stays in current state)"), `:366` (Role 1 verdicts "appended to the intent's `## Audit Notes`"), `:392` (`UNACHIEVABLE` "writes a … replan invitation block … into the intent's `## Audit Notes`"), `:35` (reclassify appends `reclassification_history`).
- **Why:** Note `05-intent.md:307` itself says "Both the press-release intent and the frozen PRD are immutable input artefacts post-promotion" while the same file writes into shipped intents — the immutability claim needs scoping (body-immutable? frontmatter+AuditNotes mutable?) everywhere it appears.
- **Fix:** Define exactly which regions of an intent file are frozen and which are append-only, once.

### M12. Invariant 2 ("never blocks") is falsified by the fail-closed security gates
- **Files:** `02-constraints/03-invariants.md:11` ("**MCP-preferred, structural-fallback** — every external-tool call has a configured backend AND a structural fallback. **Never blocks.**") vs `04-surfaces/04-launch.md:14` ("gitleaks >= 8.18.0; **absent/older = fail-closed, never a regex fallback**"), `02-constraints/02-dependencies.md:13-22` ("declining still **fails closed**").
- **Why:** gitleaks/presidio are external tools; the invariants chapter says any architecture that blocks is "wrong even if it works", while the security design (correctly) mandates blocking. The invariant as stated would license an implementer to add a fallback that weakens the gate.
- **Fix:** Scope invariant 2 to oracle/LLM/plugin capability calls and explicitly carve out security scanners as fail-closed.

### M13. 00-meta's brief↔lifeboat skeleton contract has no implementation anywhere
- **Files:** `00-meta.md:24-28` ("**`/abcd:ahoy`** copies an empty version of this skeleton into a fresh repo… **The mapping table between brief and lifeboat sections is the contract; round-trip tests catch divergence.**") vs `04-surfaces/01-ahoy.md` (detection/apply passes contain no brief-skeleton copy; the "`.abcd/` skeleton" gap is `meta.json, config.json, rules.json` only — line 139), `04-surfaces/03-embark.md:37` (embark writes the amended press release to a **single** `.abcd/development/brief/README.md`, not the numbered-folder skeleton), and no mapping table exists in `02-disembark.md § 5` or anywhere else; the verification matrix has no round-trip shape test.
- **Why:** The declared design contract ("one canonical skeleton, used three ways") is not reflected in either consuming surface, and the artefact called "the contract" (the mapping table) does not exist.
- **Fix:** Either add the skeleton-copy step to ahoy + the mapping table to 02-disembark, or delete the claim from 00-meta.

### M14. Systematic dead cross-reference: "05-intent.md § 6" for the reviewer (it's § 7)
- **Files:** `04-surfaces/README.md:11`, `05-internals/01-agents.md:20,28,32`, `05-internals/04-universal-patterns.md:99`, and `04-surfaces/05-intent.md:204` itself ("Concurrency via `flock(2)`… (see § 6)") all point at "`05-intent.md § 6` / anchor `#6-the-intent-fidelity-reviewer-agent-three-roles-three-verbs`" — but the reviewer section is **`## 7`** (`05-intent.md:331`); § 6 is "Acceptance gates and bidirectional link verification". The flock contract cited by universal-patterns as "§ 6" actually lives under § 7 Role 3.
- **Fix:** Renumber or fix all five references (and consider heading anchors instead of numbers, since this file clearly gained a section).

### M15. Verification matrix does not cover surfaces 8 and 9, Roles 2/3, or most post-fn-30 machinery
- **File:** `06-delivery/02-verification-matrix.md` (whole file)
- **Evidence (absence):** No rows for the `/abcd` board (itd-20: six sections, known-state lines, zero-writes, 5s flowctl timeout), `/abcd:reflect` (refusal table, five-section template), `/abcd:intent consistency` / `shape` (Roles 2–3, shipped per fn-29), the audit loop (`audit_mode`/`audit_budget`/`UNACHIEVABLE`), `review.autodrain`, the RC/FS/TM/PA/PR lint families, the in-session wire protocol (fence, sentinel exit 120, stale request_id), oracle_send receipts, the setup wizard/validation gate, launch bump-tier auto-detection and retention pruning, folder classification (five kinds) and `workspaces.json`/`index.json` registries, or ahoy's "uninstall→install byte-identical round-trip" acceptance (`01-ahoy.md:294-298` says this "is an acceptance criterion, not just prose" — but no matrix row exists).
- **Why:** The matrix is the brief's stated "test coverage across surfaces" (`04-surfaces/README.md:25`); it covers roughly the pre-fn-30 design and nothing shipped since.
- **Fix:** Regenerate rows from each surface's Acceptance section; the matrix should reference them rather than restate stale summaries.

### M16. Unverifiable acceptance language in load-bearing gates
- **Files:** `01-product/01-press-release.md:29` / `04-surfaces/02-disembark.md:132,140` ("oracle audit returns a **'sufficient' verdict with specific findings (not vague approval)**"), `06-delivery/02-verification-matrix.md:16` ("has specific findings, not vague approval"), `:25` ("new repo memory+ADRs **faithful subset**"), `04-surfaces/04-launch.md:164` ("lists exactly the include/exclude payload manifest… **with no surprises**").
- **Why:** "Sufficient", "specific", "faithful subset", "no surprises" have no operationalisation; the press release itself admits (Open Question 3) there is no round-trip fidelity floor — yet "faithful subset" is a matrix row someone must pass/fail. Also note `sufficient` is not a member of any registered verdict enum (`{SHIP, NEEDS_WORK, MAJOR_RETHINK}` / `{MET,…}` in `02-constraints/04-naming.md:83-84`) — the lifeboat-oracle's verdict vocabulary is unregistered.
- **Fix:** Define the lifeboat-oracle verdict enum and a minimal mechanical proxy (e.g. "≥N findings each citing a file", schema-validated); either quantify "faithful subset" or demote the row to the oracle gate.

### M17. `dev-sync` is a hard dependency of disembark/capture but is absent from every command inventory
- **Files:** `04-surfaces/02-disembark.md:19-25` (Phase 0 of disembark runs `abcd dev-sync`), `04-surfaces/06-capture.md:59` (migration on "first run of `abcd dev-sync work`"), `05-internals/03-configuration.md:282-287` (triggers) vs `01-product/04-scope.md:17` (operator-internal set = `deps-check`, `ralph-up`, `session`, `run`), `05-internals/03-configuration.md:328` (commands tree: "deps-check, ralph-up, session").
- **Why:** dev-sync is invoked as `abcd dev-sync` (the PATH-symlinked CLI), but the PATH symlink default is **no** for public repos (`01-ahoy.md:245-246`) — leaving disembark's Phase 0 dependent on a binary the default install may not put on PATH, and the verb itself unregistered in the surface/operator taxonomy.
- **Fix:** Register dev-sync (operator-internal CLI verb), and specify how disembark invokes it when the symlink was declined.

### M18. Grill skill artefact both exists and doesn't
- **Files:** `05-internals/06-lint.md:90` (TM002 in-scope path set includes "**`skills/abcd-intent-grill/SKILL.md`**"), `glossary/README.md:127` ("See `skills/abcd-intent-grill/phase-1-glossary-mode.md` for the complete write-back protocol") vs `05-internals/08-skills.md:9` ("**abcd ships zero skills.** An earlier version… proposed `/abcd:grill` as a skill; it has since been promoted to a sub-verb"), `05-internals/03-configuration.md:331-345` (skills tree lists `abcd-ahoy` … `abcd-capture`, `commit-attribution`, `secrets-and-pii` — no `abcd-intent-grill`, and also no `abcd-memory`).
- **Why:** Two files hard-reference a `skills/abcd-intent-grill/` directory the plugin layout and skills chapter say doesn't exist. ("Plugin-runtime workflow file" may be intended, but then the layout tree is wrong.)
- **Fix:** Add `abcd-intent-grill/` (and `abcd-memory/`) to the layout, or fix the TM002 path set and glossary README reference.

### M19. 03-configuration's intent tree omits `disciplines/` and `superseded/`
- **Files:** `05-internals/03-configuration.md:401-405` (intents/: "drafts/ … planned/ … shipped/" only) vs `04-surfaces/05-intent.md:7` ("Intents live at `…/intents/{drafts,planned,shipped,disciplines,superseded}/`. **Five directories** encode lifecycle position") and everything downstream (IL003/IL005/IL006 lint codes assume the five-dir model).
- **Fix:** Add the two directories to the layout tree.

---

## MINOR

### m1. Press release: "and 13 more later-phase items" — actual list has ~30
- `01-product/01-press-release.md:25` names six examples "and 13 more later-phase items"; `06-delivery/03-out-of-scope.md:36-72` lists ~30 bullets. The out-of-scope file's own rule ("It is **not** hand-counted… rather than maintaining a total that re-drifts") is violated by the press release. Fix: drop the count.

### m2. Phantom `/abcd:oracle` command in the naming glossary
- `02-constraints/04-naming.md:46` lists "`/abcd:oracle ask <prompt>` (action: invoke cascade)" among earned sub-verbs. No `/abcd:oracle` command exists in any surface list, the reserved table, or the internals (oracle is plumbing). The vocabulary-registration file references an unregistered surface — the exact drift VR001/SD lint exists to prevent. Fix: delete or reserve it.

### m3. "Four-file carve-out" that enumerates two files / four exceptions of which two aren't `.abcd/`
- `04-surfaces/01-ahoy.md:61` ("**four-file carve-out** … writes `<repo>/.abcd/config.json` and `<repo>/.abcd/rules.json`"); `05-internals/03-configuration.md:212` counts the CLAUDE.md marker block and `.specstory/cli/config.toml` as two of the "four named exceptions" to a rule about `.abcd/`. Fix: say "two `.abcd/` files plus two non-`.abcd/` in-repo artefacts".

### m4. `abcd init --json` used at install step 3, PATH symlink created at step 8
- `04-surfaces/01-ahoy.md:204-207` vs `:244-248`. How the `abcd` binary is invoked before its symlink exists (plugin-root path? `${ABCD_PLUGIN_ROOT}/scripts/abcd-cli`?) is unspecified. Fix: name the invocation path for pre-symlink steps.

### m5. "Voyage" has three incompatible senses
- Glossary (`glossary/core/voyage.md`): "complete project lifecycle… ends when… final release tag is cut"; `02-constraints/04-naming.md:17`: "`.abcd/logbook/` | **record of voyages** — logs, state, reports **per command run**" (a command run = a voyage?); `.abcd/development/voyage/` = embark/disembark provenance only (`03-embark.md § 7`). This is the exact "terminology drift" failure mode `03-mental-model.md:104` warns about. Fix: align the naming-table gloss and the voyage/ dir description with the glossary term (or split terms).

### m6. Glossary `interview/embark` vs the `/abcd:embark` command; `core/disembark.md` cross-links the wrong counterpart
- `glossary/interview/embark.md` defines embark as a grill opening move; there is **no** `core/embark` term for the top-level command. `glossary/core/disembark.md` ("When to use"): "It is the counterpart to [embark](../interview/embark.md)'s inbound opening" — linking the pack command to the *grill* term instead of the unpack command. A GL003-style cross-context collision manufactured inside the glossary itself. Fix: add `core/embark`, repoint disembark's link.

### m7. `distribution/release.md` invents launch behaviour
- `glossary/distribution/release.md` example: "`launch ship` refuses a no-change release unless `--force` is passed." `04-surfaces/04-launch.md:9` defines flags `--mode`, `--version`, `--allow-dirty`, `--allow-doc-warnings` — no `--force`, no no-change refusal. Fix: correct the example or add the behaviour to the launch surface.

### m8. Two "Phase 0"s and two phase numbering axes plus a third in-command "Phase 0"
- Product Phase 0 (Substrate), plumbing "Phase 0 — Foundation" (`06-delivery/01-build-sequence.md:13`), and disembark's internal "PHASE 0 — DEV-SYNC" (`02-disembark.md:19`). The two-axis note (`01-build-sequence.md:7`) helps, but `05-internals/02-adapters.md` labels adapters "(phase 4…)" (product axis) while `01-build-sequence.md § 5` calls the same work "Phase 2 — Settled-artefact adapters" (plumbing axis) with no axis marker at either site. Fix: prefix every phase mention with its axis ("product phase 4" / "plumbing phase 2"), and rename disembark's internal stage ("stage 0", not "phase").

### m9. Matrix "Agent dedup" row references `.abcd/memory/pitfalls.md`, which doesn't fit the page model
- `06-delivery/02-verification-matrix.md:13` vs `05-internals/07-memory.md:27` (pages are `<type>_<domain>_<slug>.md`); `.flow/memory/pitfalls.md` is flow-next's, and `memory/pitfalls.md` is the legacy snapshot (`03-configuration.md:313`). Fix: name the actual dedup source.

### m10. Press-release Open Question vs disembark: transcript density "will be measured" vs "resolved in Phase 0"
- `01-product/01-press-release.md:40` ("Phase 0 sampling… **measures** actual density **before** Pass B's design locks") vs `04-surfaces/02-disembark.md:67` ("transcript signal density **resolved in Phase 0** (`research/phase/0/transcript-sampling.md`)") and the concrete 15%-gate in `01-build-sequence.md:86`. The Open Question is stale. Fix: close or reword it.

### m11. `07-memory.md` internal § reference ambiguity and an unregistered `recall` sub-verb
- `05-internals/07-memory.md:141` ("regenerable index (per **§ 8** lifecycle taxonomy)") — this file's § 8 is "Cross-cutting integration"; the taxonomy is `04-universal-patterns.md § 8`. Also `:196`: "`abcd memory recall [keyword]`" — a diagnostic verb absent from the surface contract (`04-surfaces/07-memory.md` sub-verbs: bare/ingest/ask/lint) and from itd-39's later-phase status. Also `09-provenance-substrate.md:124` (References: registry "is regenerable per the lifecycle taxonomy") contradicts its own § 3 ("DURABLE… only PARTIALLY regenerable (adr-13; **this supersedes** the blanket 'regenerable' classification)"). Fix: correct the § pointers, mark `recall` as itd-39/later-phase, fix the stale References bullet.

### m12. Documentation-hygiene violations: historical asides throughout a "present-state" brief
- Examples: `04-surfaces/README.md:19` ("`/abcd:grill` **was originally proposed** as a user-facing skill…"), `05-internals/README.md:13` ("Slot 7 **was previously reserved** for `07-audits.md` — retired when…"), `05-internals/03-configuration.md:212` ("**Earlier drafts of this brief located**…"), `02-constraints/04-naming.md:31` (a note explaining why two tables both exist), `04-surfaces/08-abcd.md:44-54` (a whole section about a stub that never existed). The workspace standards mandate "Present-state only… No historical design notes, ever". Also `03-evidence/` is described in `README.md:15` as "placeholders for now" while `04-tradeoffs.md` is populated ("Status: PARTIAL"). Fix: sweep per the project's own rule; history to git/ADRs.

---

## Cross-cutting observations for the fix pass

1. **The brief has two authorship eras and no reconciliation.** 01-product, 02-constraints (except naming's glossary table), 03-evidence, and most of 06-delivery describe the pre-build design; 04-surfaces (esp. 05-intent, 08, 09), 05-internals, and the naming reserved-vocabulary table have been maintained through ~fn-83. adr-5 ("brief is current state") makes the stale halves *bugs*, not history. The single highest-leverage fix is regenerating 01-product/04-scope, 06-delivery/03-out-of-scope, and 06-delivery/02-verification-matrix from the phase docs / lifecycle directories they claim as SSOT.
2. **Counts are the canary.** Every hand-maintained count in the brief is wrong or contested (agents 15/16, commands 6/7/9, "thirteen intents", "13 more later-phase items", "four-file carve-out"). The brief's own out-of-scope file already states the correct policy ("derive… rather than maintaining a total that re-drifts") — apply it globally.
3. **`/abcd:intent consistency` (Role 2) would catch most of this.** Findings C1, C4–C6, M1–M8 are precisely its five judgement categories (terminology drift, premise contradictions, scope leakage, sequencing impossibilities, naming conflicts). Running the fn-29 verb over the brief before build, and wiring the deferred XD mechanical codes (reference rot would catch M14), is the systemic remedy the brief already designed for itself.
