# fn-82.5c — GL001 triage over shipped itd-27 / itd-28

Probe-gated, report-first triage of the 69 unbucketable GL001 tokens left
PARTIAL by fn-24 T8 R7. Seed inventory: `.work/issues.md` `2026-05-25` T8 entry.

## Mechanism probe (gating decision)

The `terminology_exceptions` mechanism named in the fn-24 follow-up (a
`code_identifiers` / `artefact_ids` allowlist that would let GL001 skip
API-shape and artefact-ID tokens without inflating the glossary with
non-domain entries) **is absent from the tree**. Probe:

```
$ grep -rq "terminology_exceptions" scripts/abcd/   # → ABSENT
```

The only surfaces that mention the string are this spec / task file. GL001's
sub-case-1 resolution (`scripts/abcd/_intent_lint/_intent.py::_check_gl001`)
clears a backtick token **only** when its lowercased text equals a glossary
`term:` display-name, an alias, or a forbidden synonym — there is no exception
allowlist consulted. Therefore, per the spec's Decision context, **the
gap-report branch is the expected primary outcome**; zero-unbucketed is not the
target a reviewer waits for, and no new mechanism is built here.

## Buckets

Every token is placed in exactly one of three buckets:

- **(B) canonical term file** — a genuine domain concept; a term file is added,
  which clears the GL001 token via display-name match.
- **(P) prose fix** — the backtick is not invoking a real term and can be
  reworded; limited to the two shipped intent files. **None taken** (see below).
- **(G) gap** — load-bearing but not a domain noun (schema/sidecar field name,
  CLI/sub-verb/skill/doc identifier, artefact ID, or third-party tool name).
  These need the absent `terminology_exceptions` allowlist; the gap is filed to
  `.work/issues.md`. **No broad edits, no mechanism built.**

### (B) Canonical term file — 1 token (cleared)

| token | action | clears |
|---|---|---|
| `disembark` | added `core/disembark.md` (`term: disembark`, `bounded_context: core`) — a genuine project-lifecycle surface, peer to `lifeboat`/`voyage`, referenced by `core/lifeboat.md` | ✓ GL001 no longer fires for `disembark` on itd-27 |

`blueprint` was a candidate in the fn-24 follow-up but is NOT a realized
concept: it has zero brief presence and appears only inside itd-27's seed-term
**enumeration** (line 62) as an example `core/` bounded-context name that was
never actually created. It is not a domain noun deserving a canonical file, so
it is bucketed **(G)** below rather than manufacturing a term for it.

### (P) Prose fix — 0 tokens

No prose fixes were taken. Every remaining flagged token is load-bearing where
it appears (a schema field name, an artefact ID naming a real spec/intent, a CLI
surface name, or a third-party tool name). Unbacktick-ing any of them would
either misrepresent a real identifier or rewrite a shipped intent's historical
record. The spec permits prose fixes but limits them to disjoint, non-broad
edits; the honest classification here is that none qualify.

### (G) Gap — 68 tokens (need `terminology_exceptions`, absent → filed)

Sub-classified. Each token below is currently flagged GL001 (warn) on itd-27
and/or itd-28; each needs an allowlist entry the tree does not provide.

**G1 — schema / sidecar / frontmatter field names (46).** Code shape, not a
bounded-context noun. Would belong in a `terminology_exceptions.code_identifiers`
allowlist.

- itd-27 (15): `aliases`, `bounded_context`, `definition`, `forbidden_synonyms`,
  `frozen_at`, `glossary_terms_used`, `grill_report_hash`, `grill_session_id`,
  `grilled_at`, `introduced_in`, `planning_attempt_id`, `prd_path`,
  `source_intent_hash`, `status`, `term`
- itd-28 (31): `allow_no_commit`, `backend`, `body_markdown`, `body_max_bytes`,
  `chat_id`, `dirty`, `findings`, `focus`, `generated_at`, `omitted_bytes`,
  `pinning`, `receipt_path`, `render_max_bytes`, `review_of_commit`,
  `review_type`, `reviewed_files`, `reviewer_model`, `reviewer_tool`,
  `sanitized_raw_artifact_sha256`, `session_id`, `spec_path`, `spec_sha256`,
  `staleness`, `summary`, `superseded_by`, `target_id`, `truncated`,
  `truncation_method`, `unpinned`, `verdict`, `worktree_sha256`

**G2 — CLI / sub-verb / skill / document identifiers (10).** Surface names, not
domain nouns.

- `grill`, `plan`, `refine`, `abcdgrill-skill`, `intent-grill-skill`,
  `intent-fidelity-reviewer`, `prd-archive`, `ready-for-agent`, `to-prd`,
  `press-release`

**G3 — artefact IDs (intent / spec IDs) (8).** Each names a real artefact.
Would belong in a `terminology_exceptions.artefact_ids` allowlist.

- itd-27: `itd-1`, `itd-24`, `itd-42`,
  `fn-3-strengthen-intent-stage-abcdgrill-skill`
- itd-28: `itd-1`, `itd-7`, `itd-13`,
  `fn-2-move-repoprompt-review-artifacts-into`, `fn-5-bsc`

(`itd-1` and `press-release` are shared across both intents — counted once in
the unique totals.)

**G4 — third-party tool / product names (3).** Not bounded-context nouns.

- `flock`, `gitleaks` — third-party tools
- `abcd` — the project's own product name (listing it as a glossary term has
  design implications; deferred to a dedicated decision, per fn-24 follow-up)

**G5 — stale aspirational context-name (1).**

- `blueprint` — an example `core/` context in itd-27's seed-term enumeration
  that was never realized as a term or a context. Neither a domain noun to
  define nor a real artefact; recorded here rather than manufacturing a term.

## Machine-checkable closure

After adding `core/disembark.md`, the GL001 check over itd-27/itd-28:

- itd-27: 30 unique GL001 tokens (was 31; `disembark` cleared)
- itd-28: 40 unique GL001 tokens (unchanged)
- combined unique: **68** (was 69)

All 68 remaining tokens are accounted for in bucket (G) above; the 1 cleared
token (`disembark`) is in bucket (B). 69 = 68 + 1 → **zero unbucketed**. The
remaining 68 are warn-severity residue whose closure requires the absent
`terminology_exceptions` mechanism, filed as a follow-up gap in `.work/issues.md`.

Verification command (reproduces the per-file unique counts and confirms every
flagged token appears in this report):

```bash
for f in itd-27-grill-skill-and-glossary itd-28-rp-reviews-into-flow; do
  bash scripts/lib/python_pick.sh -m scripts.abcd.lint \
    ".abcd/development/roadmap/intents/shipped/$f.md" --json \
  | python3 -c "import sys,json,re;d=json.load(sys.stdin);\
print(sorted({re.search(r'\`([^\`]+)\`',x['message']).group(1) for x in d['findings'] if x['code']=='GL001'}))"
done
```
