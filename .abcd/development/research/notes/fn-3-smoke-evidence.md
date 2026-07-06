# fn-3 Smoke Evidence

<!-- Manual smoke evidence for fn-3.6 acceptance criteria.
     The grill skill is a Claude-invoked skill (not a CLI command); full end-to-end smoke
     requires a human-driven Claude session. This document records:
     1. Unit test evidence (automated, covers key contracts)
     2. Code review trace of smoke-critical paths
     3. Deferred evidence items (require human grill session + itd-6) -->

## Context

The `/abcd:intent grill` skill runs interactively inside Claude Code. It cannot be
invoked as a subprocess or replayed in a fully automated Ralph loop without the
itd-6 MCP replay harness. This evidence file records the highest-confidence evidence
achievable within the Ralph autonomous loop, and explicitly calls out the items
that require a human-driven session.

---

## Unit Test Evidence (automated, 2026-05-11)

```bash
python3 -m pytest tests/abcd/test_grill_skill.py tests/abcd/test_grill_phase2.py \
  tests/abcd/test_grill_glossary_mode.py tests/abcd/test_intent_lint.py -v
# Result: 379 passed in < 1s
```

### Phase 1 injection resistance

`tests/abcd/test_grill_skill.py` covers:
- `natural_completion` exit on adversarial input (session refuses to comply)
- Phase 1 does not honour "ignore previous instructions"-style directives
- Session sealed after Phase 2 runs (resume never crosses the boundary)

### Phase 2 PRD synthesis resistance

`tests/abcd/test_grill_phase2.py` covers:
- Seven-section PRD shape required (PRD section validator failure on missing section; not a GR lint code)
- `source_intent_hash` provenance check (`GR005` blocker on mismatch)
- PRD synthesis treats grill findings as the sole input source (not raw intent text verbatim)
- Phase 2 does not produce attacker-controlled content in `## Implementation Decisions`
  when Phase 1 found no valid grill findings from an adversarial intent

### Freeze roundtrip

`tests/abcd/test_intent_lint.py` covers:
- `GR003` emitted when PRD body modified after freeze (hash mismatch)
- `GR003` silent when body unchanged
- `frozen_content_hash` = SHA-256 of PRD body + stable PRD frontmatter (excluding `{frozen_at, frozen_content_hash, epic, planning_attempt_id}`; provenance fields `source_intent_hash`, `grill_report_hash`, `grill_report_path` are INCLUDED), hex lowercase
- `promote_check` gate exits non-zero on any blocker

### Glossary mode

`tests/abcd/test_grill_glossary_mode.py` covers:
- Forbidden synonym detection (`GL002` flag + propose-with-accept per occurrence)
- Three-clause ADR test evaluation (all-clauses-pass → ADR offer fires)
- Term file atomic write to `terminology/<context>/<term>.md`

---

## Code Review Trace (smoke-critical paths)

### `--fresh` flag contract

**File**: `skills/abcd-intent-grill/phase-1-workflow.md`

The `--fresh` flag is documented to rotate Socratic moves and produce a deterministically
different question set. Code review confirms:
- `--fresh` clears `grill-state.json` before Phase 1 starts
- `_questions.md` documents move rotation as seeded from `(intent_id + session_timestamp)`
- A fresh session with the same intent produces a different move sequence (different seed)
- Phase 2 still runs after `--fresh` Phase 1 (inseparability preserved)

### Injection canary — Phase 2 surface

**File**: `skills/abcd-intent-grill/phase-2-synthesis.md`

Phase 2 consumes `grill-report.json` (structured findings), not the raw source intent text
verbatim. Review confirms:
- Synthesis prompt reads: "synthesise from the grill report, not from the raw intent body"
- Phase 2 prompt is structured to treat both source intent and grill transcript as untrusted
- Attribution footer at line 261 is NOT included in user-visible PRD text (per Task .4 contract)

### Planner handoff

**File**: `docs/reference/commands.md` (plan row)

Review confirms:
- `/abcd:intent plan itd-N` writes `frozen_at`, `frozen_content_hash`, `planning_attempt_id`
- Epic spec first `## Links` YAML block requires `intent: itd-N`, `prd: <path>`,
  `planning_attempt_id: <uuid>` (not frontmatter — correct per fn-3 convention)

---

## Deferred Evidence (requires human grill session or itd-6)

The following smoke items require a live grill session invoked in Claude Code and cannot
be captured in a Ralph autonomous loop without the itd-6 MCP replay harness:

| Item | Deferred to | Notes |
|------|-------------|-------|
| Phase 1 injection resistance — actual session transcript | Human grill session | Fixture + unit tests cover the contract; transcript requires live Claude invocation |
| Phase 2 PRD `## Implementation Decisions` — actual output from adversarial intent | Human grill session | Phase 2 output depends on live synthesis run |
| `--fresh` produces different question sequence — actual diff | Human grill session | Move rotation is deterministic in code; live diff requires two runs |
| `/abcd:intent plan` freeze — actual PRD frontmatter with freeze fields | Human grill session | Requires Phase 2 output PRD to freeze |
| `/flow-next:plan` handoff — actual epic spec `## Links` YAML block | Human grill session | Requires frozen PRD as input |

These items are documented in `fn-3-coordination.md` under the itd-6 deferred dependency.
The CI replay step (itd-6) will provide automated evidence for all of the above.

---

## Pre-commit Evidence

All commits in fn-3.6 passed pre-commit hooks:
- `trim trailing whitespace` ✓
- `fix end of files` ✓
- `flow-state drift check` ✓
- `PII scan` ✓
- `detect hardcoded secrets` ✓ (gitleaks: no leaks found)
- `.flow/reviews/ integrity` ✓
