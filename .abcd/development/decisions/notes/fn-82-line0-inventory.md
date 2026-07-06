# fn-82.3 — `line == 0` lint-emitter inventory (R3)

Every lint emitter that can produce a `line == 0` finding, its source, a
classification, and the intended action. `line == 0` findings are dropped
**unconditionally** by the `--since-staged` / `--since` scope filter
(`scripts/abcd/lint.py::_is_file_level_finding` → `_classify_and_filter`), so a
recoverable frontmatter finding must be promoted to a 1-based file line to
survive that filter.

The three classes (per the spec's Decision context):

- **frontmatter-recoverable** — the violation is (or names) a top-level
  frontmatter key, so `_yaml.frontmatter_key_lines` (already 1-based,
  leading-comment-tolerant) yields the file line. **Converted here.**
- **body-derived** — the finding's location is a body region, an external
  index, or a corpus/cross-file relationship, not a frontmatter key. The
  frontmatter-key→line technique does not apply. **Inventory-only, deferred.**
- **genuinely-file-level** — the violation is the *absence* of a whole section
  or an aggregate with no single source line. Stays `line == 0`, stays
  suppressed under `--since-staged`. **No action (by design).**

The 1-based-line contract every converted emitter honours (from
`.flow/memory` `bug/build-errors/yaml-parseerror-lines-are-slice-2026-06-26`):
`frontmatter_key_lines` returns **file-relative** 1-based lines that already
account for leading `<!-- -->` HTML-comment attribution before the opening
fence (the slice-relative / fence-offset trap). Callers fall back to line 1
when the key is absent or frontmatter parsing failed (fn-24 T2 round-8
contract).

## Inventory

| Code | Source (module:line) | Classification | Action |
|------|----------------------|----------------|--------|
| IL001 | `_intent_lint/_intent.py` `_check_schema` (`line=0`) | frontmatter-recoverable (present-key errors) / genuinely-file-level (missing-required, whose key is absent) | **Converted** — anchor present-key schema errors on the offending frontmatter key line (extracted from the `ValidationError.path` `$.key` or the `additional property 'key'` message); missing-required errors anchor at line 1 (frontmatter start), never a phantom absent-key line. |
| IL002 | `_intent_lint/_intent.py` `_check_il002` (missing-section branch, `line=0`) | genuinely-file-level | No action. Missing `## Acceptance Criteria` is a whole-section absence; there is no frontmatter key nor body line to anchor. The header-present / no-GWT branch already emits `line=header_line`. Stays suppressed under `--since-staged`. |
| IL003 | `_intent_lint/_intent.py` `_check_il003` (4 emit sites, `line=0`) | frontmatter-recoverable | **Converted** — `kind`-mismatch anchors on the `kind` key line; the three `spec_id` sites anchor on the `spec_id` key line. When the key is absent (e.g. `planned/`/`shipped/` missing `spec_id`), fall back to line 1 (the violation is the absence). This is the most-populated live subject (7 findings on the corpus). |
| IL013 | `_intent_lint/_intent.py` `_check_il013` (`line=0`) | frontmatter-recoverable | **Converted** — the violation *is* the `status:` key; anchor on `_key_line("status")`. The cleanest single-key subject (live on `itd-58`). |
| IL011 | `_intent_lint/_bundle.py` (2 sites) | body-derived (cross-file / phase-doc corpus relationship) | Inventory-only, deferred. The finding is about bundle-member membership spanning multiple phase docs — no single frontmatter key or line in the linted file. |
| IL014 | `_intent_lint/_bundle.py` (1 site) | body-derived (cross-file bundle relationship) | Inventory-only, deferred. Advisory (info) about an unscheduled bundle member; corpus-level, no anchoring key. |
| MQ002 | `_intent_lint/_memory.py` `_emit_at` (3 sites) | body-derived (external `.sources_index.json` coverage) | Inventory-only, deferred. Keyed on a source hash in an external memory index, not a frontmatter key of any linted file. |
| MQ003 | `_intent_lint/_memory.py` `_emit` (1 site) | body-derived (external index coverage-unavailable) | Inventory-only, deferred. Same external-index origin as MQ002. |
| PQ001–PQ006 | `scripts/abcd/lint_prompts.py` (13 emit sites) | body-derived (prompt structure); already line-carrying | No action needed. **Every** PQ emit site already passes a `line=` value — some frontmatter-key-derived (`key_lines.get('prompt_version', 1)`, `reads_untrusted_input`), the rest body-position-derived (`pv_line`, `cs_line`, `block_line`) or line-1 fallback. None emits an unconditional `line == 0`. Per the spec, PQ00x are body-derived and inventory-only; empirically they are already line-precise, so there is nothing to convert. The frontmatter-key→line technique for the two frontmatter-key PQ codes (PQ002/PQ006) is already applied at their emit sites. |
| TM003+ | `scripts/abcd/lint_terminology.py` | body-/frontmatter-derived; already line-carrying (fn-82.2) | No action needed. `lint_terminology.py` findings carry `err.line` natively (native `--json` emission, fn-82.2); the `_emit(...)` calls flagged by a naive scan are the **output/print** helper (`_emit(findings, json_out, text_blocks)`), not a per-finding line emitter. |

## Converted codes (named for R3 acceptance evidence)

The following emitters are converted from `line == 0` to line-precise
frontmatter emission in this task:

- **IL013** — `status:` key → `_key_line("status")`.
- **IL003** — `kind` / `spec_id` matrix violations → `_key_line("kind")` /
  `_key_line("spec_id")`, line-1 fallback on absent key.
- **IL001** — present-key schema errors → offending frontmatter key line;
  missing-required errors anchor at line 1 (frontmatter start).

## Frontmatter-recoverable subject for the test cases

**IL013** (`status:` key) is the primary conversion + test subject: the
violation is unambiguously a single named frontmatter key, so the
0-based-mark, fence-offset (leading `<!-- -->` comment), and leading-comment
test cases exercise a real emitter with a deterministic expected line. **IL003
`spec_id`** is the secondary subject (most-populated live emitter). A real
GR/GL-family frontmatter emitter that still emits `line == 0` does **not** exist
in the tree — the GL001 / GR005 / IL012 frontmatter emitters were already
promoted to line-precise by fn-18 T1 and fn-24 T2 (they consume
`frontmatter_key_lines` / `frontmatter_list_item_lines` today). The
frontmatter-recoverable IL-family emitters above are the genuine remaining
subjects; no synthetic subject is invented.

## Genuinely file-level findings (stay suppressed)

- **IL002** missing-`## Acceptance Criteria` branch — whole-section absence.
- **IL001** missing-required schema errors — the key is absent, so it anchors
  at frontmatter-start (line 1) rather than a phantom key line; a
  missing-required on a file with no staged frontmatter change stays dropped.

A regression fixture proves no NEW findings appear on an unrelated staged file
under `--since-staged` after the conversion.
