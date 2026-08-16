# Brief-first planning prep — first full run (2026-08-16)

An unattended pre-pass over the draft corpus, run ahead of the human planning
interviews: per-draft planning briefs plus independent SOTA fit-challenges for
the freshly filed plannable drafts, and a supersession triage over the older
bench. Nothing was planned, no acceptance criteria were confirmed, and no
draft's routing was filed — every output is a proposal in the local work tier
(`.abcd/.work.local/scratch/planning-briefs/`), consumed by the interviews.

## What ran

- **Classification** of all 55 drafts (real press release + real
  Given/When/Then vs seeded placeholder vs no criteria at all).
- **Planning briefs** for the seven fresh plannable drafts (itd-106, itd-107,
  itd-108, itd-109, itd-113, itd-114, itd-115): summary-back with per-AC
  provenance, itd-84 hand-run as an ungraded proposal, proposed acceptance
  criteria, open-question resolutions, blocks-planning flags.
- **Independent fit-challenges** for the same seven, by a separate pass per
  draft (evaluator outside the loop), testing the declared SOTA path against
  `prefer-sota` / `sota-per-intent` and the recorded conventions.
- **Triage** of the 47 older drafts against the current record (planned/,
  shipped/, superseded/, ADRs, the decision log).

## Yield

- Fresh drafts: **no FILE-AS-IS survived the pre-pass** — all seven briefs
  proposed SPLIT (itd-109 with a documented HOLD trigger). The two drafts that
  carried an in-file verdict (itd-114 FILE-AS-IS, itd-115 no-reversal) were
  both overturned by deeper record-reading: the table alone is not the gate;
  reading the records it touches is.
- Every existing acceptance-criteria bullet across the fresh drafts proved
  agent-seeded and unconfirmed — the itd-1 interview walk is not a formality.
- Older bench (47 drafts): 11 PLANNABLE (itd-9, 10, 18, 21, 39, 41, 55, 83,
  87, 90, 91 — itd-39 the strongest), 2 LIKELY-HANDLED (itd-11 superseded by
  shipped itd-88; itd-85 delivered but never shipped out of drafts/, iss-240),
  14 STALE-NEEDS-REWORK, 14 BLOCKED, 6 SEED. Roughly two-thirds of the bench
  cannot be planned as written.
- Defects captured during the run: iss-236 (launch `local_username` false
  positive), iss-237 (brief drift: unbuilt doc-fidelity internals described as
  current), iss-238 (`/abcd:audit` reservation contradiction), iss-239
  (dangling pre-rebuild spec ids; a delivered-claim the tree cannot back),
  iss-240 (itd-85 unshipped), iss-241 (prompt baselines for deleted agents),
  iss-242 (`intent --json` has no per-intent list mode), iss-243
  (`intent_sota` lint never built; 9 of 107 intents declare SOTA).

## Learnings, routed

1. **The pre-pass earns its keep in the record-reading, not the table.** Both
   verdict changes and all the strong reversal flags came from reading the
   touched records (ADRs, resolved issues, the decision log), which a
   capture-time hand-run tends to skip. → Codified as the unattended pre-pass
   paragraph in the `/abcd:intent` surface page (this change).
2. **Independent fit-challenges disagree productively.** Brief and challenge
   converged on the verdict for itd-114 from different evidence (a recorded-id
   reversal vs a dropped exact-registry alternative); neither alone carried
   both. Keep the two passes separate.
3. **Missing SOTA declarations are systemic, not per-draft** — the template
   slot and lint were never built. → iss-243; interviews should treat a
   missing `## SOTA` as a standing agenda item until the detector exists.
4. **Fresh drafts drift within days.** Two of the six cited issue states that
   had changed by prep time (iss-186, iss-205); one cited a resolved issue as
   a live prerequisite. → Factual references in drafts are re-verified at the
   interview; the two clear cases were corrected in this change.
5. **The verb cannot enumerate its own corpus.** Classification required
   shell-level scanning of 55 files. → iss-242 (per-intent list mode with
   `ac_state`).
6. **Name intents by id in run prompts.** A prompt referring to "the
   merge-queue intent" by description cost a resolution pass; the intent
   (itd-115) was filed by a parallel session mid-run. Ids are the only stable
   handle across concurrent work.
7. **Triage before prepping the bench.** Two-thirds of the older drafts are
   stale, blocked, or seeds; preparing briefs for them ahead of a rework pass
   would have been waste. Supersession triage is the cheap filter that should
   precede any bulk prep.

## Boundaries held

The seven itd-84 hand-runs remain **ungraded**: grading into the
decomposition-calibration note happens only when the human confirms each
routing at the interview. `abcd intent plan` was not run for any draft, and no
proposed criteria were written into any draft.
