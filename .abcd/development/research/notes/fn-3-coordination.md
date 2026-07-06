# fn-3 Cross-Repo Coordination Note

<!-- Committed research note (round-1 P1 #11: NOT .work/issues.md — .work/ is gitignored).
     Documents cross-repo interactions discovered during fn-3 implementation.
     Use this file for any cross-repo coordination needs that must survive git history. -->

## Scope

Epic `fn-3-strengthen-intent-stage-abcdgrill-skill` is implemented entirely in this repo
(`abcdDev`). No code changes are required in sibling repos during the fn-3 implementation
window. The cross-repo interactions documented here are deferred dependencies, not current
blockers.

## Deferred Dependencies

### itd-6 (MCP Replay Harness)

The grill skill ships canary fixtures and manual smoke evidence for Phase 1 and Phase 2
injection resistance (per acceptance criteria in fn-3.6). **Automated CI replay-mode
verification of these fixtures is owned by itd-6**, which delivers the MCP replay harness
that fn-3 currently lacks.

- **Current state (fn-3 ship):** Fixtures committed; smoke evidence at
  `.abcd/development/research/notes/fn-3-smoke-evidence.md` — unit test coverage documented
  for automated contracts; live grill session evidence deferred to itd-6.
- **itd-6 deliverable:** CI step that re-runs grill in replay mode and asserts hash identity
  for `determinism-baseline.json` and `determinism-prd-baseline.md`; CI step that replays
  the injection-canary session and asserts non-compliance output.
- **Cross-reference:** itd-6 epic (when specced) must depend on fn-3's fixture corpus.

### itd-6 (Determinism Hash Enforcement)

`tests/fixtures/grill/determinism-baseline.json` and `tests/fixtures/grill/determinism-prd-baseline.md`
ship with `"enforcing": false` and `PLACEHOLDER` hash values. The CI step that makes these
enforcing (fills actual hashes, flips `enforcing: true`) is owned by itd-6.

### Future Sibling-Repo Skills

The grill skill (`skills/abcd-intent-grill/`) is self-contained in `abcdDev`. If a sibling
project (e.g., `idelphiDev`) installs `abcd`, it inherits the grill skill via the plugin
mechanism with no fn-3-specific coordination needed.

## No Current Blockers

All fn-3 tasks (.1 through .6) are implementable in `abcdDev` alone. No changes to sibling
repos, shared infrastructure, or external systems are required before fn-3 ships.

## Attribution Audit Status (fn-3.6 acceptance)

Every file under `skills/abcd-intent-grill/` and `.abcd/development/foundation/terminology/`
that derives from Pocock must contain an attribution line. Audit completed 2026-05-11:

- [x] `skills/abcd-intent-grill/SKILL.md` — line 7: `<!-- Adapted from mattpocock/skills (MIT). Core loop pattern derived from /grill-me and /grill-with-docs. Phase 2 PRD shape adapted from /to-prd. See ACKNOWLEDGEMENTS.md for full attribution. -->`
- [x] `skills/abcd-intent-grill/phase-2-synthesis.md` — lines 3–5: `<!-- Adapted from mattpocock/skills/skills/engineering/to-prd/SKILL.md (MIT licence). Attribution: Matt Pocock, to-prd skill. See ACKNOWLEDGEMENTS.md. -->` + footer at line 261 (attribution in prompt footer, NOT in user-visible PRD text)
- [x] Terminology seed files (core/, interview/) — each file carries `<!-- Adapted from mattpocock/skills (MIT). See README Acknowledgements. -->` comment header; DDD bounded-context structure attributed in `ACKNOWLEDGEMENTS.md`
- [x] PRD synthesis prompt footer — attribution at `skills/abcd-intent-grill/phase-2-synthesis.md:261`, outside PRD body (per Task .4 contract)

## Smoke Evidence

Unit test suite confirms core contracts (run 2026-05-11, all passing):

```
python3 -m pytest tests/abcd/test_grill_skill.py tests/abcd/test_grill_phase2.py \
  tests/abcd/test_grill_glossary_mode.py tests/abcd/test_intent_lint.py -v
# 312 + 67 = 379 tests passed in < 1s
```

Contracts verified by unit tests:
- **Phase 1 injection resistance**: `tests/abcd/test_grill_skill.py` — `natural_completion` exit
  on adversarial input; session refuses to comply
- **Phase 2 PRD synthesis**: `tests/abcd/test_grill_phase2.py` — seven-section shape (PRD
  section validator failure on missing section), `GR005` (source_intent_hash mismatch) blocker
- **Glossary mode**: `tests/abcd/test_grill_glossary_mode.py` — forbidden synonym detection,
  three-clause ADR test, term file atomic write
- **Freeze roundtrip + lint codes GL001–GL005, GR001–GR005**: `tests/abcd/test_intent_lint.py`
  — `GR003` hash-mismatch detection, promote-check gate, severity overrides

Full replay-mode canary (deterministic hash identity across invocations, adversarial session
replay asserting non-compliance output in CI) is deferred to itd-6 MCP replay harness.
