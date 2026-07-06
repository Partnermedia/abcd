---
id: itd-45
slug: phase-1-cleanup-before-phase-2
spec_id: null
kind: null
suggested_kind: standalone
reclassification_history: []
related_adrs: []
created: 2026-05-22
updated: 2026-05-22
---

# Sweep The Workshop Before Phase 2 — Cleanup Bundle From `.work/issues.md`

## Press Release

> **Before Phase 2 (capture) begins, abcd does a workshop sweep of the technical debt that accumulated through Phase 0 (substrate) and Phase 1 (ahoy).** A facilitator opens `/abcd:intent itd-45` and sees a single intent that consolidates the small, real, low-controversy cleanup items from `.work/issues.md` — each one too small to deserve its own intent on its own, all of them together enough to materially clean up the floor before the next phase's first stroke. The sweep closes lint false-positives that block legitimate commits, removes terminology residue the fn-7 / itd-43 sweep didn't catch, deletes stale paths and time-estimate language that violate the workspace standards, and unblocks two intents (itd-27, itd-28) that are mis-filed in `planned/` only because the cleanup hasn't run. None of this is glamorous; all of it is the kind of dust that compounds.
>
> "I came back to abcd after two weeks away and the first thing I did was scan `.work/issues.md` to remember what was outstanding," said Marcus, contributor. "Half of it was tiny stuff that wasn't blocking any one thing but was making every edit a little wobbly — a lint rule that flagged the word `Release` inside `## Press Release`, a pre-commit hook that hard-blocked any commit touching the old discipline files, intent files still sitting in `planned/` after their specs shipped. itd-45 swept all of it in one focused pass. Phase 2 started on a clean floor."

## Why This Matters

`.work/issues.md` is abcd's append-only ledger of small issues found mid-flight. The workspace standard is "every issue discovered during any work is recorded immediately… there are no low priorities" — which is right, but it also means the ledger grows continuously, and most entries individually are too small to justify their own intent. The workspace standard also asks: *at the end of a session, review the log and surface next actions.* This intent is the phase-boundary version of that review — a milestone scan that picks out the cleanup items that have accumulated since the last sweep, consolidates them into one intent, and ships them as a single focused pass before the next phase begins.

The pattern matters beyond this one instance. Phase-boundary cleanup intents are project-agnostic: any abcd project that ends a phase produces a ledger of small accumulated issues, and any of them can run a sweep intent at the boundary. The shape generalises (intent-driven retrospective on `.work/issues.md`); the contents are specific to abcdDev's Phase 0 → Phase 2 transit.

There is a second motivation specific to this boundary. Phase 1 (ahoy) introduces the first commands a contributor actually invokes. Phase 2 (capture) introduces the first lightweight ledger a product thinker uses mid-flight. Both phases ship surfaces that contributors and product thinkers will *touch* — which means lint false-positives, stale references, and mis-filed intents become first-contact friction in a way they were not when the surfaces were internal. The cost of *not* cleaning up before Phase 2 is that first-touch users hit the residue first.

## What's In Scope

The sweep covers items found in `.work/issues.md` between 2026-05-16 and 2026-05-22 that fit **all four** of the following filters:

1. **Small** — a single PR's worth of focused changes, not a multi-day effort.
2. **Low-controversy** — no architectural decision required; the fix shape is already specified in the issue note's "Suggested fix" field.
3. **Cross-cutting** — touches more than one file or area, so the unit of work is "the sweep" rather than "one fix in one place."
4. **Cleanup, not feature work** — removes friction, doesn't add capability.

Items that fit are grouped into five sub-sweeps below. The fifth sub-sweep (Ralph hardening) is the largest and most borderline; it could split out into its own intent if planning concludes the work is too big for the bundle.

### Sub-sweep A — lint hygiene (corpus-wide blockers)

- **GL002 forbidden-synonym false-positive on `## Press Release`** (`.work/issues.md` 2026-05-16). The matcher flags the literal `Release` inside the canonical section heading every intent carries. Fix: word-boundary aware matcher that skips ATX heading lines, or whitelist the exact phrase `Press Release`. Until this lands, the `drafts/`→`planned/` promotion gate is blocked corpus-wide on a false positive.
- **Pre-commit `abcd-lint` hook lacks CI's grandfathering** (`.work/issues.md` 2026-05-19). CI scopes `lint.py` to `base...head` changed files; the pre-commit hook does not, so any edit to a pre-fn-8 file (e.g. `itd-1`, `itd-37`) hard-blocks the commit on longstanding violations. Fix: give the pre-commit hook the same changed-lines / baseline scoping CI uses.

These two together are the highest-priority items in the bundle — both gate the routine work the next phase depends on.

### Sub-sweep B — terminology / schema residue not covered by itd-43

itd-43 covers the intent-prose terminology sweep (`epic`→`spec` across the intents tree, the canonical `terminology/core/epic.md`→`spec.md` rename, etc.). It does **not** cover the schema residue and test-fixture residue that fn-7 R1/R7 left PARTIAL.

- **`intent.schema.json` `epic_id` legacy-alias property** (`.work/issues.md` 2026-05-18, 2026-05-18 follow-up at line 254). `spec_id` is the canonical field and is already declared with `additionalProperties:false`; the `epic_id` alias is dead weight.
- **`prd.schema.json` `epic` property key** (`.work/issues.md` 2026-05-18 line 254). After fn-7.6, `commands/abcd/intent.md` says the PRD field is `spec` while the schema still validates `epic`. Doc/schema contradiction.
- **`tests/test_yaml.py` `epic_id` test fixtures** (`:350,380,383,642`).
- **`tests/abcd/test_issue_schema.py` negative-test `epic` identifiers** — necessary content for the negative tests, but outside fn-7's allowed-exception scope (#4). Fix: widen fn-7's exception list explicitly to cover schema negative-tests whose reason-to-exist is asserting the old key is rejected.

This sub-sweep coordinates with itd-43 — they should sequence as **B (schema/fixtures) before itd-43 (prose)**, because the canonical schema needs to settle before the prose sweep cites it.

### Sub-sweep C — lifecycle desync (intents mis-filed, paths stale)

- **itd-27 and itd-28 in `planned/` though specs (fn-3, fn-2) are done** (`.work/issues.md` 2026-05-16 line 91; 2026-05-17 line 246). Move depends on resolving GL002 blockers in the bodies (sub-sweep A) — the lint fix and the directory move must land together; doing the move alone leaves the intents lint-failing in `shipped/`.
- **Stale `~/ABCDevelopment/Autonomous/` paths in `.flow/tasks/fn-1-*`, `.flow/tasks/fn-2-*`, `.flow/specs/fn-2-*`** (`.work/issues.md` 2026-05-16 line 97). Repo layout is now `Apps/abcd/`, `Apps/idelphi/`. Sweep `.flow/` and rewrite.
- **Stale "cwd contains only an initial commit" prose in `brief/02-constraints/01-platform.md`** (`.work/issues.md` 2026-05-16 line 115). Repo is now fully built; the section misleads new contributors. Rewrite or drop.

### Sub-sweep D — workspace-standards violations

- **Time-estimate language in `brief/03-evidence/04-tradeoffs.md`** (`.work/issues.md` 2026-05-16 line 160). The "Jagged frontier" entry says "~5 min/agent at v1.0.0 lock"; the workspace `CLAUDE.md` forbids duration language. Reword to relative-cost statement.
- **itd-17 frontmatter missing `related_adrs: [adr-8]` reciprocal link** (`.work/issues.md` 2026-05-16 line 148). adr-8 already declares `related_intents: [itd-17, itd-28]`; the cross-reference contract `intent_lint.py` enforces expects the back-link.

### Sub-sweep E — Ralph & flowctl hardening (BORDERLINE — may split out)

This is the largest cluster. Each item is from a real Ralph run that lost work or wasted time; together they are the difference between Ralph being safe to run unattended and Ralph being a thing that needs babysitting. They are bundled here because they are all small and all the same shape (defensive fixes to the autonomous loop), but **planning may conclude this sub-sweep is too big for the cleanup bundle and split it into its own intent (provisional name: "Ralph infrastructure-failure resilience hardening").** That decision happens at `/abcd:intent plan itd-45`.

- **Ralph counts infrastructure failures as task attempts** (`.work/issues.md` 2026-05-18 line 260). Network outage burns the per-task attempt budget on `ConnectionRefused` iterations that never reached the model. Auto-blocks healthy tasks. Reports a halted-by-outage run as `completion_reason=NO_WORK` / `promise=COMPLETE`.
- **Ralph block handler dumps raw API-error JSON into the task `.md` Done summary** (`.work/issues.md` 2026-05-18 line 274). ~8KB of throwaway JSON in a human-meaningful prose field; would survive in git history if committed before notice.
- **`ralph-guard.py` `codex exec` block is substring-matched** (`.work/issues.md` 2026-05-18 line 330). False-positives on impl-reviews of Codex-related tasks; reviews of fn-11 oracle-cascade tasks structurally hit the block because their summaries contain `codex exec`.
- **`ralph-guard.py` `PROTECTED_FILE_PATTERNS` matches plugin `hooks/hooks.json`** (`.work/issues.md` 2026-05-21 line 404). Blocks legitimate `Write` calls to the abcd plugin's hooks manifest because the pattern matches with `endswith`. Anchor or drop.
- **flowctl `rp` review path leaks RepoPrompt context tabs** (`.work/issues.md` 2026-05-18 line 320). ~50 stale contexts accumulated; suspected cause of `-32600 Invalid Request` stalls. flowctl has create verbs but no delete verb.
- **flowctl `rp` review path has no input-size budget guard** (`.work/issues.md` 2026-05-18 line 286). Codex 1 MiB cap overflow on wide-scope reviews; the existing `FLOW_CODEX_EMBED_MAX_BYTES` budget logic is wired only into the `codex`/`copilot` embed paths, not `rp`.

## What's Out Of Scope

The following items in `.work/issues.md` are explicitly **not** part of itd-45 — they need their own intents or already have one:

- **itd-43 (`epic`→spec terminology sweep)** — already a draft intent; itd-45 sub-sweep B coordinates with it but does not absorb it.
- **`oracle.py` cascade orchestrator gap and itd-6 cascade spec** (`.work/issues.md` 2026-05-16 lines 191, 197). Feature work, not cleanup; depends on itd-2's in-session leg.
- **`intent-fidelity-reviewer` Roles 2 and 3** (`.work/issues.md` 2026-05-16 line 178). Owned by itd-31 (Role 2 cross-doc) and itd-34 (Role 3 kind classification) when those plan into specs; itd-45 does not anticipate them.
- **fn-12 oracle-backed gates un-satisfiable in autonomous mode** (`.work/issues.md` 2026-05-19 line 360). Structural; needs either an `_build_cli_oracle` Codex-leg extension (fn-11/itd-6 scope) or an `lifeboat-oracle` agent spec. Not a sweep item.
- **Flow-state drift detector** (`.work/issues.md` 2026-05-17 line 240). New feature with its own intent; explicitly deferred from fn-6 scope.
- **`lint_terminology.py` native `--codes`/`--json`** (`.work/issues.md` 2026-05-18 line 300). A "later spec" already named by fn-8 T10; itd-45 does not preempt.
- **Reviews-index supersession Rule 1 root data inconsistency** (`.work/issues.md` 2026-05-18 line 32). The narrow checker fix landed in fn-7.3; the root data inconsistency (focus-string drift across a supersession link in fn-4's committed sidecars) is left as-is per fn-7's "sidecars are historical records, not edited" boundary, and itd-45 inherits that boundary.
- **fn-5 plan-review round-27 residuals** (`.work/issues.md` 2026-05-16 line 137). Already scoped to the respective fn-5 task work via `/flow-next:work`.
- **Upstream flowctl `/tmp/*.md` collision** (`.work/issues.md` 2026-05-20 line 369). Local patch already shipped under `abcdDev/scripts/local-patches/`; upstream fix is flow-next's, not abcd's. No-op once upstream lands.
- **itd-42 `grilled_intent_hash` stale** (`.work/issues.md` 2026-05-18 line 306). Fix is to regrill itd-42 before promotion — handled by `/abcd:intent grill itd-42`, not by a sweep.
- **Pyright `StrEnum` import** (`.work/issues.md` 2026-05-18 line 314). Harness env config, not real type bugs. Out of repo scope.
- **`/abcd:intent new` removal** (`.work/issues.md` 2026-05-22 line 413). Design decision in this session; needs its own intent or its own spec, not a sweep item — it changes a public command surface.
- **itd-44 fourth-intent-kind-decision** — separate draft, deferred pending itd-39.

## Acceptance Criteria

- *Given* a contributor running `lint.py` over the intent corpus, *when* GL002 evaluates an intent body, *then* the literal `Release` inside the `## Press Release` heading does not fire as a forbidden synonym; only genuine occurrences in body prose do.
- *Given* a contributor staging an edit to a pre-fn-8 file (e.g. `itd-1`, `itd-37`), *when* the `abcd-lint` pre-commit hook runs, *then* it gates only on findings introduced by the staged change, matching CI's `base...head` behaviour — pre-existing violations in unmodified lines no longer block the commit.
- *Given* the intent corpus after the sweep, *when* a reviewer greps the intents tree, *then* `epic` literals remain only where itd-43 has not yet landed; the schema residue (`intent.schema.json` `epic_id` alias, `prd.schema.json` `epic` key, `test_yaml.py` fixtures) is gone and `additionalProperties:false` validates cleanly.
- *Given* a reviewer surveys `intents/planned/`, *when* they list the files there, *then* itd-27 and itd-28 are no longer present — they have been moved to `intents/shipped/` and their bodies lint clean.
- *Given* a contributor reads `brief/02-constraints/01-platform.md`, *when* they reach the "This directory IS abcdDev" section, *then* the prose describes the current built repo, not the original fn-1 scaffolding moment.
- *Given* a contributor reads `brief/03-evidence/04-tradeoffs.md`, *when* they reach the "Jagged frontier" entry, *then* there is no duration language ("~5 min/agent" or similar) — only relative-cost language.
- *Given* `intent_lint.py` scans the corpus, *when* it checks adr-8's reciprocal-cross-reference contract, *then* itd-17's frontmatter declares `related_adrs: [adr-8]`.
- *Given* `.flow/tasks/` and `.flow/specs/` are searched for the legacy `Autonomous/` path, *when* the search runs, *then* zero hits return — the prose has been rewritten to the current `Apps/` layout.
- *Given* the Ralph sub-sweep E lands (or is split out and landed separately), *when* a Ralph run encounters a `ConnectionRefused` iteration with `total_tokens:0`, *then* it does not consume the per-task attempt budget; the loop backs off; the run does not falsely report `COMPLETE`. (If sub-sweep E splits out, this criterion moves with it.)

## Open Questions

- **Should sub-sweep E (Ralph hardening) be part of itd-45 or its own intent?** The items are bundle-eligible by size but cohesive in a way that argues for a focused name ("Ralph infrastructure-failure resilience"). Decide at `/abcd:intent plan itd-45`.
- **Sub-sweep ordering vs. itd-43.** Sub-sweep B (schema/fixture residue) should run before itd-43 (prose sweep) so the schema settles first. But itd-43 has not yet been planned into a spec. Does itd-45 wait for itd-43 to plan, or does itd-45's spec sequence itself ahead of itd-43? Best resolved by planning itd-43 first, then itd-45 can declare its dependency cleanly.
- **Sub-sweep A coordination.** The GL002 fix and the pre-commit grandfathering fix are independent but both touch the lint surface; they could land as one PR or two. Defer to the spec.
- **itd-27 / itd-28 lint fixes.** Their bodies use glossary-forbidden synonyms (`project`, `AI`, `issue`, `User`, `archive`, `Release`) per fn-6's `## Why the intent move is deferred`. Are those fixes part of itd-45 sub-sweep C (the directory move), or do they belong to a separate "glossary cleanup of shipped intents" pass? The current draft folds them into C since the move and the lint fix must land together.
- **`.work/issues.md` ledger hygiene after the sweep.** When itd-45 lands, do the swept entries get marked resolved in place (with a back-link to the merging PR), archived to a `.work/issues-archive/` directory, or left untouched (the PR provides the historical record)? The workspace standard doesn't specify a resolution convention. Decide at `/abcd:intent plan itd-45`.

## Related

- **`.work/issues.md`** — source ledger for every item in scope; entries dated 2026-05-16 through 2026-05-22.
- **itd-43** (`epic`→spec terminology sweep) — coordinates with sub-sweep B; sequences before itd-45 if both plan in close succession.
- **itd-44** (fourth-intent-kind-decision) — sibling draft; not in scope, deferred pending itd-39.
- **adr-9** (phase as a product-reflection layer) — frames why a Phase 1 → Phase 2 boundary is a natural sweep moment.
- **fn-6** (Phase 1 reconciliation) — itd-45 is the second pass at reconciliation: fn-6 fixed the structural desync; itd-45 picks up the small accumulated dust fn-6's `## Boundaries / non-goals` left behind.
- **fn-7** (`epic`→spec terminology sweep code/schemas/brief) — itd-45 sub-sweep B picks up the R1/R7 PARTIAL residue fn-7 could not close without widening its own scope.
- **fn-8** (discipline enforcement machinery) — provides the lint substrate (`intent_lint.py`, `lint_prompts.py`, `lint.py` orchestrator) that sub-sweep A modifies.
- **Phase 1 — ahoy** (`roadmap/phases/phase-1-ahoy.md`) — the phase whose work generated most of the entries in scope.
- **Phase 2 — capture** (`roadmap/phases/phase-2-capture.md`) — the phase itd-45 clears the floor for.
