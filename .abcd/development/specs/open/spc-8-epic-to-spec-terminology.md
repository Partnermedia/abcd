---
id: spc-8
slug: epic-to-spec-terminology
intent: itd-43
---
# epic-to-spec-terminology

## Summary

spc-8 delivers itd-43's implementable remainder: the **GL002 forbidden-synonym
lint gate**, built detector-first against the glossary, and the **live-prose
sweep** it drove — 19 flagged uses of `epic` as the concept-noun cleared to
`spec` across four record files, with the detector (not human judgment)
deciding what counted. **The spec stays open**: AC3 (the reviews-subsystem
`spec-review` token) is blocked on itd-28, which needs a maintainer-gated new
dependency; itd-43 stays `planned` until that lands or is re-scoped.

## Approach

Fix-the-detector, literally: the rule was armed before any prose changed, its
19 live-corpus findings were captured verbatim as the watched-fail evidence,
the sweep then drained exactly what it flagged, and a regression test pins the
live corpus at zero GL002 findings from here on.

- **The rule** (`internal/core/lint`): `checkForbiddenSynonyms` reads the
  glossary term files (`.abcd/development/brief/glossary/`) as the source of
  truth, flags an *enforced* synonym used as a standalone word in live prose,
  and errors if an enforced word is not actually declared in some term's
  `forbidden_synonyms` (the gate cannot drift from the glossary). Matching is
  case-insensitive with **explicit Unicode word boundaries**
  (`utf8.DecodeRune` + a word-rune class), deliberately not Go's ASCII-only
  regexp `\b`; a test proves `epics`, `epicenter`, and unicode-adjacent text
  do not match while bare and parenthetical uses do.
- **Enforcement is a deliberate subset** — `epic` only, for now. Most of the
  glossary's forbidden synonyms (`sprint`, `milestone`, `project`, `feature`)
  are common English words whose false-positive rate would sink the gate;
  each is opt-in via the rule config as the corpus is readied for it.
- **Exemption scope** (config in `.abcd/record-lint.json`, not silent skips):
  YAML frontmatter, code spans and fenced blocks, `allow_context` phrases
  (`epic-to-spec`, `epic->spec`, `epic→spec` — the rename's own name), and
  historical paths (`research/`, `decisions/`, dated `plans/`,
  `intents/shipped/` + `superseded/`, the glossary term files themselves).
  Live surfaces — brief body, roadmap, drafts/planned intents, open specs,
  docs, commands — are in scope.
- **Wiring:** registered in `.abcd/record-lint.json`, so it runs under
  `cmd/record-lint` inside `make preflight` and CI — the gate is a gate, not
  a claim.

## Milestones as delivered

1. GL002 rule + seven hermetic unit tests, watched-fail first (`23c0e86`,
   hardened by `a1e1467`).
2. The armed-detector sweep: 19 hits → 0 across
   `itd-34-three-intent-kinds.md` (11 lines), `itd-6-rp-mcp-only-integration.md`
   (6), `itd-48-intent-fidelity-reviewer-roles-2-3.md` (1 line, 2 tokens),
   `roadmap/phases/phase-0-substrate.md` (1) (`2a6e9cd`).
3. CHANGELOG entry for the new gate (`6c1a88a`).

## Acceptance-criteria satisfaction

AC as ordered in itd-43 → status and evidence:

1. **No live `epic`-as-noun reference remains** — gap-filled, detector-driven:
   GL002 armed pre-sweep flagged 19 lines (captured verbatim); post-sweep the
   corpus is at 0, pinned by `TestForbiddenSynonymsRealGlossary` (loads the
   real config + glossary over the real corpus). Historical records exempt
   per the AC's own carve-out.
2. **The glossary resolves to `spec` with `epic` forbidden** — met already:
   `brief/glossary/core/spec.md` carries `term: spec` and `epic` in
   `forbidden_synonyms`; no stale `epic.md` term file exists anywhere
   (verified by filesystem sweep). The intent's open question (rename vs
   stub) was answered historically — the rename happened with no stub.
3. **Reviews subsystem classifies against `spec-review`** — **BLOCKED on
   itd-28** (the native reviews subsystem does not exist in the Go tree; its
   intent needs a maintainer-gated new dependency). No epic-review token
   exists to rename and no spec-review emitter exists to test. This spec
   stays open on exactly this criterion.
4. **`issue.schema.json` uses `related_specs`** — satisfied-by-adjudication:
   no `issue.schema.json` exists in the Go tree (old-system reference); the
   native validator (`internal/core/capture/validate.go`) and capture engine
   already use `related_specs` exclusively, with no `related_epics`
   anywhere. Nothing dead was built.
5. **`internal/core/lint` raises no GL002 violation for `epic`** —
   gap-filled: the rule now exists (it did not before), is wired into the
   preflight/CI record-lint, and the live corpus passes at zero (same
   regression pin as AC1).

### Judgment call on record

Two of the 19 hits sat inside internal self-quotes (a working-log line and
spc-12's overview, quoted in `itd-48`). Both are abcd's own records, not
external citations, so they were swept for the single-word consistency this
intent exists to enforce — recorded here rather than silently.

Out-of-scope confirmations: the `spc-` ID prefix is untouched (the intent
renames the concept word, not the ID scheme); `epic_id`→`spec_id` was the
prior atomic change; no schema files were altered.
