# Plan-Consistency Review — Consolidated Summary

**Scope:** the entire `.abcd/` and `docs/` trees of this repo (213 files, ~23k lines of markdown), reviewed before the abcd-cli build begins. This is a **plan review, not an implementation review** — findings of the form "referenced tooling does not exist in this repo" are context (the Go implementation comes later), not defects. What counts against the plan: internal contradictions, stale internal records, and enforcement designs with holes independent of implementation.

**Method:** six specialist review passes, run independently and cross-verified — adversarial brief review, terminology/glossary audit, cross-corpus consistency, standards/hygiene audit, link/structure integrity, and a structure-vs-state-of-the-art evaluation. Reports are the numbered siblings of this file. Headline claims were spot-verified against the files before consolidation.

**Standing decision recorded during this review:** `brief/glossary/` becomes the single source of truth (SST) for terminology (see § 5).

---

## Verdict

The brief's *skeleton* is genuinely strong — in several dimensions ahead of state of the art — but the *content* is not currently a buildable plan. The corpus has two authorship eras (the predecessor `abcdDev` design vs the 2026-07-06 Go-rebuild decisions) that were never reconciled, and the brief's own governing rule (adr-5: "the brief is the current state, always") turns every stale passage into a formal defect rather than tolerable history. Before building, the plan needs one reconciliation pass (Phase 0.5) driven by the findings below — most are mechanical once the root cause is named.

## 1. The root cause everything else hangs off

`work/CONTEXT.md` says the copied `development/` record "is the starting spec, not current truth, until Phase 0.5 reconciles it — do not treat it as authoritative." But **nothing inside `development/` carries that marker** — the brief calls itself "canonical, current" in three places, and all 20 ADRs sit at `status: accepted`. A reader (human or agent) entering via the brief has no way to discover it is suspended. Two consequences:

- The eight 2026-07-06 decisions in `work/DECISIONS.md` (Go rewrite, drop flow-next/RepoPrompt/Ralph/codex, single repo, Cobra) **never graduated to ADRs**, despite that file's own graduation rule — so the ADR corpus formally asserts the old architecture, and the brief still *mandates* flow-next and RepoPrompt as dependencies (`02-constraints/02-dependencies.md`) and a dev→public mirror the decisions killed (`04-surfaces/04-launch.md`, phase-5). The `deprecated` ADR state exists in the lifecycle spec and is used zero times.
- The brief never locks implementation language or runtime at all (`02-constraints/01-platform.md` locks paths, not toolchain), while ~40 references specify Python artefacts (`intent_lint.py`, `personas.py`, `harness.py`) as contracts.

**Fix shape:** either write the Go-rebuild ADR set and mark voided ADRs `deprecated`, or stamp a not-yet-authoritative banner into `brief/README.md` — and add the language/toolchain lock to platform constraints.

## 2. Adversarial findings — contradictions an implementer cannot resolve

The red-team pass ([`01-adversarial-brief.md`](01-adversarial-brief.md)) found **8 critical, 19 major, 12 minor** defects. The criticals, all evidence-verified:

| # | Contradiction | Where |
|---|---|---|
| C1 | Press release AC: `launch dry-run` **hard-fails** on PII. Launch surface: dry-run is "report-only, **always exit-0**"; hard-fail is `ship`'s behaviour | `01-press-release.md:32` vs `04-launch.md:10` |
| C2 | "There is no repo-scope `.abcd/` (four exceptions)" vs repo-scope `.abcd/` paths required by nearly every surface (lifeboat, PRDs, logbook, retrospectives, corpus.json) | `03-configuration.md:212` vs 8+ files |
| C3 | `meta.json` abolished ("no meta.json at repo scope") and load-bearing (read/written in 4 steps) **in the same file** | `01-ahoy.md:61` vs `:139,170,254` |
| C4 | The canonical out-of-scope list declares later-phase at least six intents the rest of the brief documents as **delivered** (itd-39/44/49/50/52/53) | `03-out-of-scope.md` vs lint/config/intent chapters |
| C5 | Entry-point scope chapter says "thirteen intents, seven commands, 15 agents" — chapters 04–06 say otherwise on all three counts | `04-scope.md` vs surfaces/agents |
| C6 | `/abcd:reflect` is "Reserved — later phase" in constraints and a **shipped surface** in 04-surfaces/agents | `04-naming.md:27` vs `09-reflect.md` |
| C7 | No language/runtime lock anywhere; all concrete contracts are Python in a repo premised as a Go CLI | `01-platform.md` |
| C8 | `/abcd:intent plan` specified twice in one file: §1/§2 with no PRD step; §5 as a 10-step freeze that **refuses without a PRD**. Following §1 builds a verb GR002 blocks 100% of the time | `05-intent.md:120` vs `:296` |

**The canary is hand-maintained counts**: every one in the brief is wrong or contested (15 vs 16 agents in seven places; 6 vs 7 vs 9 commands; "thirteen intents"; "13 more later-phase items" vs ~30). The brief's own out-of-scope chapter states the correct policy — *derive, don't count* — and then violates it. Apply that policy globally.

Also notable: the verification matrix doesn't cover surfaces 8–9, reviewer Roles 2/3, or most post-fn-30 machinery, and several load-bearing gates use unverifiable language ("sufficient verdict", "faithful subset", "no surprises") with no operationalisation — the round-trip fidelity floor is an admitted open question yet "faithful subset" is a matrix row someone must pass/fail.

## 3. Mechanical consistency (cheap to fix, high trust payoff)

- **70 broken relative links + 20 broken anchors** ([`05-link-structure.md`](05-link-structure.md)). Three systemic causes: the `intents/` tree moved out of `roadmap/` (~15 links + 54 prose references), predecessor-repo trees (`scripts/`, `.flow/`, `tests/`, `agents/`) referenced from a repo that doesn't have them, and renumbered section headings (the `05-intent.md §6→§7` renumber alone broke five cross-references).
- **`intents/README.md` is the worst file in the corpus**: its lifecycle listings say 1 planned (actual: 8) and 2 shipped (actual: 17), list shipped intents as drafts, and omit 15 actual drafts — which makes ~56 intent files unreachable from any index. adr-3 makes directories the truth; regenerate the listings from disk.
- Numbering is otherwise clean (adr-1..20 gapless; itd gaps 38/68/71 — 38 documented, 68/71 unexplained; one recorded renumbering, itd-45→49, contradicts the "never renumbered" rule with no rule amendment).
- Hygiene ([`04-standards-hygiene.md`](04-standards-hygiene.md)): no PII, no credentials, no email leaks anywhere (the home-directory-pattern hits are all scanner specs). Violations of house standards: template-mandated `created:`/`updated:` frontmatter on ~70 intents (git-inferable), three time-estimate leaks ("~5 min/agent" — already self-recorded as a known issue in itd-45, which sits unexecuted in drafts/), and pervasive historical framing ("was originally proposed…") inside a present-state brief.
- `docs/` (the Diátaxis skeleton) is the **cleanest area of the corpus** — internally consistent, accurate, correct boundaries. One watch-item: `explanation/README.md` names "the abstraction boundary", a concept CONTEXT.md lists as old-architecture pending reconciliation.

## 4. Structure evaluation

Full assessment in [`06-sota-structure.md`](06-sota-structure.md).

**SOTA compliance.** The taxonomy (meta → product → constraints → evidence → surfaces → internals → delivery → glossary) is a defensible superset of PRD + arc42 + spec-kit, and genuinely *ahead* of SOTA in eight places: the lint-enforced bounded-context glossary, uniform Given-When-Then acceptance at every altitude, the verification matrix as a first-class chapter, derivable out-of-scope, the regenerable/append-only/compounding-curated artefact taxonomy, the self-governing 00-meta + adr-5 doctrine, lint codes as requirement IDs, and its epistemics (Naur recovery-humility, verified-absence notes). Missing vs SOTA: **success metrics** (the biggest hole), a **risk register**, a **consolidated security/threat model**, an **operational/migration slot**, and the **FAQ half** of PR/FAQ. Two systemic weaknesses: delivery-status prose fused into the current-state contract (`06-lint.md` is the worst case), and empty "evidence" chapters normalising empty canonical files.

**Genericity.** The *skeleton* is generic — every chapter maps cleanly to a web app, library, or pipeline. The *flesh* is not: "surfaces" as written assumes a command-per-surface CLI; the brief↔lifeboat shape contract in 00-meta couples the template to an abcd product feature (the one structural, not just textual, leakage); 03-evidence's population mechanism is abcd machinery; 05-internals file names are this product's subsystems. A generic template keeps the chapter slots, defines "surface" abstractly with per-project-type examples, and moves everything abcd-specific into example content. Two-digit numbered *file* prefixes: keep, with a retire-never-reuse rule (slot 7 was already recycled once). Numbered *section headings*: drop — they are the source of the broken anchors.

**Terminology consistency.** Not currently consistent ([`02-terminology.md`](02-terminology.md)): "snapshot" is a forbidden synonym of *lifeboat* yet the brief says "disembark snapshot" routinely; "epic" (forbidden per adr-11) survives in prose and in a live intent's `glossary_terms_used`; `core/brief.md` calls the brief "immutable once approved", contradicting adr-5; "embark", "session", "voyage", and "transport" each carry a second undefined sense; cross-context alias↔forbidden collisions exist (`snapshot`, `user`). The glossary's own anchor noun **grill** (426 uses) is undefined, as are `surface` (430), `gate` (512), `audit` (836), `PRD` (222), and `milestone` — the last simultaneously a defined concept and a forbidden synonym of two terms.

**Enforcement.** Designed extensively — GL/GR/TM/VR/SD lint families with severities, pre-commit + promotion-gate + CI trigger points, cite-or-fail per adr-7. As a plan review, the absence of the tooling here is not a defect — but the enforcement *design* has holes that survive the reframe: GL002 is context-blind (writing the canonical distribution alias "user" or "snapshot" would fire a blocker); no code checks cross-file alias↔forbidden collisions; prose synonym enforcement (TM002) covers only the single word "epic"; the reference-rot code (XD006) that would have caught the 90 broken links is deferred; and two registries both claim to be canonical.

## 5. Glossary as single source of truth — decided; plan edits it entails

1. `02-constraints/04-naming.md` keeps conventions only (maritime metaphor criterion, bare-command-as-render, SD001); its ~60-row Reserved Vocabulary table folds into per-term glossary files; **VR001 retargets to the glossary**.
2. Rewrite `glossary/README.md` for this repo: it currently describes itself as `terminology/`, indexes 11 of 15 terms (missing `core/disembark` and the whole `distribution/` context), and points its validation commands at a predecessor path.
3. Fix defects in the store itself: `core/lifeboat.md` has invalid YAML (backtick-leading scalar); 11 of 15 files put an HTML comment *above* the frontmatter, violating the README's own "MUST begin with frontmatter" rule (the `distribution/` files show the compliant shape); `core/brief.md` and `core/intent.md` bodies contradict the brief's operating model.
4. Add the missing load-bearing terms (`grill`, `surface`, `gate`, `audit`, `PRD`, `milestone`) and the second-context entries the bounded-context design already anticipates (`core/embark` for the command, an agent-runtime `session`).
5. Extend the lint design for a single registry: a cross-context alias↔forbidden collision check, and context-aware GL002.

## 6. Ranked pre-build actions

1. **Reconcile the two eras** — Go-rebuild ADRs written, voided ADRs marked `deprecated`, or a suspension banner on the brief (root cause; makes everything else legible).
2. **Resolve the 8 critical brief contradictions** (§2 table) — each blocks an implementer.
3. **Regenerate every derived artefact from its declared SSOT**: `04-scope.md` from phase docs, `03-out-of-scope.md` from lifecycle dirs, `intents/README.md` listings from disk, verification matrix rows from surface Acceptance sections. Delete hand-maintained counts.
4. **Execute the glossary-as-SST consolidation** (§5).
5. **Link sweep**: `roadmap/intents/` → `intents/` (54 refs), unnumber section headings, promote XD006 (reference-rot lint) from deferred to first-shipped.
6. **Separate contract from delivery status** in brief chapters (strip "Delivered in fn-N" narration to a sidecar or the spec layer) — the largest single source of drift-by-construction before the template is reused.
7. **Add the missing template slots** before freezing the skeleton: success metrics, risks + security posture (committed, not gitignored), operational/migration, product FAQ.
8. Smaller hygiene: drop `created:`/`updated:` frontmatter (or record an explicit tooling carve-out), execute the already-drafted itd-45 cleanup intent, kill the PLACEHOLDER banners on canonical files, add the four missing brief-chapter READMEs, document or close the itd-68/71 numbering gaps.

## Report index

| File | Specialist pass |
|---|---|
| [`01-adversarial-brief.md`](01-adversarial-brief.md) | Red-team review of the brief's content (contradictions, ambiguity, gaps, untestable requirements) |
| [`02-terminology.md`](02-terminology.md) | Glossary integrity, per-term usage consistency, missing terms, enforcement design, naming compliance |
| [`03-cross-corpus-consistency.md`](03-cross-corpus-consistency.md) | Drift between brief, ADRs, intents, roadmap, and work-state (F0–F25) |
| [`04-standards-hygiene.md`](04-standards-hygiene.md) | House documentation standards: SSOT, git-inferable info, PII, time estimates, present-state, coverage |
| [`05-link-structure.md`](05-link-structure.md) | Broken links/anchors, index-vs-disk diffs, numbering, naming, orphans |
| [`06-sota-structure.md`](06-sota-structure.md) | Brief structure vs state of the art; genericity as a template; ranked structural recommendations |
