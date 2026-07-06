# fn-82 R10 — docs/hygiene closure outcomes

Per-subpart evidence for the fn-82 task .8 docs/hygiene sweep. Each lettered
subpart records a fix, a verify-then-record outcome, or a record-absent outcome.

## (a) Reviewer-package READMEs — FIXED

Both packages lacked a `README.md` (directory-coverage convention). Added:

- `scripts/abcd/intent_fidelity_reviewer/README.md`
- `scripts/abcd/phase_audit_reviewer/README.md`

Content derived from each package's `__init__.py` façade docstring and `_cli.py`
CLI surface (verified against source, not paraphrased). Both state purpose,
submodule table, CLI surface, and the shared-machinery relationship (adr-9 "three
grains of one shape").

## (b) Stale `Autonomous/` + home-directory paths

- **`brief/06-delivery/01-build-sequence.md:19` — DEFERRED (tooling blocker).**
  The Phase 0 "read predecessors" bullet carries a live "Skim BOTH
  `~/ABCDevelopment/Autonomous/abcdZero/` … `abcdSubZero/`" instruction pointing
  at predecessor trees that no longer exist on disk (per the fn-18.4 mapping
  table, rows for `abcdZero`/`abcdSubZero`: "none on disk — historical
  predecessor"). The present-state rewrite (a skim of the two predecessor
  iterations for patterns, findings persisting at
  `.abcd/development/research/phase/0/predecessor-notes.md`; no dead home-dir
  paths; no change-narration) was authored and lint-verified clean, BUT any edit
  to a `brief/` file triggers the fn-75 brief-change derivation gate
  (`.pre-commit-config.yaml` `abcd-brief-derivation`, blocking). That gate is
  oracle-dependent and fail-closes with `oracle_unavailable` (exit 3) in an
  oracle-less autonomous session — it cannot be satisfied here without weakening
  a blocking safety gate. The build-sequence rewrite is therefore split out to a
  follow-up carrying only that one brief edit, to run where the oracle backend
  (RP/Codex) is available so the derivation gate can draw out and reconcile any
  implied intents/principles. Recorded as a follow-up; no home-dir dead paths are
  introduced or removed by this commit.

- **`.flow/tasks/fn-18-*.4.md`, `fn-24-*.6.md`, `fn-24-*.7.md` — VERIFIED, NO
  REWRITE (record).** These are completed-task Done records whose entire purpose
  is documenting the `Autonomous/` → `Apps/` path sweep (fn-18.4's
  per-row-verified mapping table) and the frozen `.flow/.backup-pre-1.0/`
  retention policy (fn-24.7's signpost). Their `Autonomous/` strings are the
  canonical audit record of that sweep — the LHS of a substitution table and the
  "these references are EXPECTED historical state" signpost. Rewriting them would
  erase the fn-18/fn-24 sweep history and directly contradict fn-24.7's explicit
  "future grep audits should not re-flag these" instruction. These are not stale
  drift; they are the intentional record. Left unchanged by design.

- **`04-surfaces/04-launch.md:16` `/Users/...` — VERIFIED, LEGITIMATE.** This is
  a regex example describing what the launch scanner's custom-regex layer
  hard-fails on (`/Users/...`, `/home/...`), not a real home-path reference.
  Left unchanged.

## (c) `04-surfaces/07-memory.md` B14 banner — VERIFIED CLEAN, NO CHANGE

The itd-36 status banner (line 3) reads present-state: the write core "shipped
via fn-38" and the lint family "shipped via fn-39" — consistent with the body.
No "Ships per itd-36" future-tense drift against a shipped body was found. The
`:73` "not yet built" phrase refers to itd-26 loot (a genuinely later,
unbuilt phase) — correct as written; not churned. The "future/inert at launch"
wording (lines 61/94/105) describes the adr-18 launch-vs-lifeboat gate design
state (the gate's real consumer is `/abcd:disembark`, not launch), not a tense
drift. Verify-then-fix outcome: verified clean, no edit.

## (d) CONTRIBUTING issue-recording → `/abcd:capture` — FIXED

Authoritative path resolved by `ls`: root `CONTRIBUTING.md` is the
contributor-workflow file (issue-recording, work organisation);
`docs/CONTRIBUTING.md` is the environment-setup file only (Python/venv/pre-commit)
and is linked from the root file. Edited the root `CONTRIBUTING.md`
issue-recording line to point at `/abcd:capture` (linked to
`commands/abcd/capture.md`) as the primary path, naming the underlying issue
ledger it writes to (`.abcd/development/activity/issues/`) — the structured store
the capture command actually appends to.

## (e) 6+ untracked `.flow/memory/` files — SCANNED, ALL DURABLE, COMMITTED

At HEAD there are **20** untracked files under `.flow/memory/` (the task text's
"6" predates later memory accumulation; the intent — scan/classify/commit
durable, never blanket-ignore — applies to all untracked memory files). All 20
live under the `bug/` and `knowledge/` tracks, which the root `.gitignore`
explicitly un-ignores as durable-record tracks (the `!.flow/memory/bug/` /
`!.flow/memory/knowledge/**` negations).

- **Secret/PII scan: CLEAN.** No emails, home-directory (`/Users/...`) paths,
  API keys, tokens (`ghp_`, `sk-`), or `-----BEGIN` blocks in any file.
- **Frontmatter sanity: CLEAN.** Every file carries `track:` + `title:`
  frontmatter — all are well-formed pipeline entries.
- **Classification: all 20 DURABLE.** They are bug-track and knowledge-track
  entries synthesized from recent task reviews — exactly the compounding record
  the memory pipeline persists. No non-durable file was found, so no per-file
  exclusion reason is required. Nothing under `.flow/memory/` is ignored or
  blanket-excluded.
- All 20 are committed by this task.

## (f) `tests/test_yaml.py` existence — VERIFIED PRESENT

The unverified issue-log claim is resolved: `tests/test_yaml.py` EXISTS (33 KB).
No edit or record-absent action needed; recorded present.
