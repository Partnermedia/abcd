# fn-82 .1 — IL008 corpus sweep + intent-lint dead-helper decision

Records two operational decisions made when landing IL008 (missing `## Press
Release` presence lint): the corpus-sweep outcome per firing intent, and the
fate of three production-dead body-analysis helpers.

## IL008 corpus sweep (fixed vs. grandfathered)

The IL008 sweep runs over the real intent corpus scope (`drafts/`, `planned/`,
`disciplines/`, `shipped/`; only `superseded/` excluded). The corpus walk
(`_load_corpus_intents`) skips `README.md` / `_template.md`, so those are not
in scope even though a raw `intents/**/*.md` glob would surface them.

At the time IL008 landed, exactly two intents fired (both WARNING severity, so
neither blocks CI):

| Intent | Location | Effective kind | Outcome | Rationale |
|--------|----------|----------------|---------|-----------|
| `itd-44-fourth-intent-kind-decision` | `drafts/` | `null` | **grandfathered** | Delivered-thin decision record (a `decision` verdict routed to the ADR store, per `fn-56`). It is a design/decision record, not a press-release-shaped intent; synthesising a press release would be fabrication, not signal. |
| `itd-69-plugin-metadata-lockstep-update` | `drafts/` | `standalone` | **grandfathered** | A minimal principle stub auto-derived by the brief-change derivation gate (`itd-61` / `fn-75`). It genuinely should carry a press release once fleshed out, but the stub has no press-release content to write yet. |

Both are pre-existing drafts predating IL008. WARNING severity means the sweep
surfaces them without blocking; they are recorded here rather than force-fixed
with invented prose.

`itd-27` and `itd-28` (the two shipped intents the spec flagged as a possible
IL008 target, serialized before `fn-82 .5`'s GL001 prose fixes) already carry a
canonical `## Press Release` section and are `kind: standalone` — they do NOT
fire IL008. No shipped-file edit was required for them, so there is no
`## Press Release`-adding edit for `.5` to land on top of.

## Dead-helper decision (ordered second, per the spec)

Three body-analysis helpers in `_intent_lint/_substrate.py` were
production-dead (no non-test caller) but retained a test:

- `_strip_code_regions` — a newline-collapsing strip variant; imported into
  `_intent.py` but never called (only named in a comment).
- `_mask_press_release_heading` — canonical-heading blanker.
- `_strip_heading_lines` — ATX-heading blanker.

**Decision: deleted, with their tests.** IL008's fence-awareness reuses the
existing `_fence_state_lines` + `_PRESS_RELEASE_HEADING_RE` pair (the same
infrastructure IL007 and GL002 use), so it does not resurrect
`_strip_code_regions` (the deprecated line-collapsing variant explicitly
warned against in the `_intent.py` `prose_body` comment). With no production
and no IL008 use, all three are removed. The now-unused module-level
`_ATX_HEADING_RE` (only `_strip_heading_lines` used it) was removed too. The
packageization-parity census (`test_intent_lint_packageization_parity.py`) drops
`_mask_press_release_heading` from its expected getattr surface accordingly.
