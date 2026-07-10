# Practice / MVP / tool extraction — 2026-07-09

A structured extraction pass over the verify-folder sources, asking of each
source: what practice does it warrant, what is the smallest unenforced enabler
(the MVP), and what would the enforcing tool be? This note is the durable
record and citation target for the run — the amended promotion-path ladder in
[`principles/README.md`](../../principles/README.md), the three new principle
files, the two principle extensions, and the issue seeds all trace here. It
builds directly on the [lineage review](2026-07-08-lineage-sources-vs-principles.md)
and the [principles SOTA mapping](2026-07-08-principles-sota-mapping.md), both
of which were re-read as first-class extraction sources.

## Method

Nine extractors ran in parallel, one per source or source cluster, each
articulating a full practice → MVP → tool ladder for every candidate with
every rung marked exists/partial/absent against the current record:

| Extractor | Source | Tier | Candidates |
|---|---|---|---|
| llm-verifier | kwok2026llmasaverifier (arXiv) | preprint | 7 |
| glm5-vibe | glm5team2026vibe (arXiv) | preprint | 8 |
| red-queen | iacob2026redqueen (arXiv) | preprint | 7 |
| ux-eval | Tan et al. (arXiv 2604.09581) | preprint | 7 |
| unknown-pdf | resolved to chesterman2026research (SSRN) | preprint | 6 |
| weng-harness | weng2026harness (practitioner post) | practitioner | 7 |
| new-sdlc | five-piece practitioner cluster | practitioner | 0 |
| confidential-quartet | four AI-generated confidential analyses in the local sources corpus | ai-generated | 8 |
| refs-bib-reclassify | the two 2026-07-08 research notes, reclassified | internal | 11 |

Total: **61 candidates**. The new-sdlc cluster yielded zero because its five
corpus entries are keyword stubs with no document text (see Housekeeping).
The confidential quartet was treated as **hypothesis-only** throughout: its
candidates enter the pipeline but carry a taint flag, and nothing they
support may be recorded without independent public warrant.

Two adversaries then ran: a **framework adversary** attacking the
practice/MVP/tool trichotomy itself (verdict below), and a **per-candidate
adversary** challenging each of the 61. Challenge outcomes:

| Outcome | n |
|---|---|
| sound | 49 |
| duplicate | 8 |
| wrong-layer (corrected, kept) | 3 |
| unsupported (kept as rejected record) | 1 |

A synthesis pass merged and reconciled: six of the eight duplicates were
cross-source twins folded into their stronger primary with both sources
credited (deterministic-gate-first → verifier-selects-gates-decide;
measured-oracle-accuracy → calibrate-the-judge; saturation-triggers-succession
and refresh-the-corpus → detector-lifecycle; badcase-corpus →
escapes-feed-the-corpus; held-out-corpus-for-detectors → bidirectional-proof).
The other two were duplicates of the *existing record* (itd-1, itd-50) and
are retained as rejected entries so the rejection is citable. 61 − 6 merges =
**55 surviving proposals**, tabulated below.

## Framework verdict: AMEND

The framework adversary ran six attacks on the trichotomy and returned
**amend** — five repairable defects, no fatal one.

**Attack 1 — boundary collapse: lands, with a smoking gun.** The corpus
itself contains the proof: badcase-corpus (glm5-vibe, filed at MVP) and
escapes-feed-the-corpus (red-queen, filed at practice) are the same idea —
verified escapes graduate into the detector's acceptance corpus, both
extending fix-the-detector's corpus clause — classified at different layers
by two passes under the same schema, each with a defensible rationale. The
practice/MVP boundary depends on how generously the existing principle is
read, not on the candidate. Worse, the MVP/tool boundary turns out to be
enforcement status, not artefact kind (checkable-anchor-criteria: a ten-line
tag-checking script is the MVP, the same check blocking a lifecycle
transition is the tool) — exactly the axis the existing promotion path
already owns. The trichotomy secretly encodes two axes
(claim-vs-mechanism, unenforced-vs-enforced) as one ladder.

**Attack 2 — missing layers: lands for records, partially for
verification.** Record-shaped candidates degenerate the ladder:
record-scope-facts-not-theory's practice statement and MVP are the same
paragraph; ladder-onboarding "tops out at MVP"; adr-argument-fields' practice
rung is unstated anywhere. The framework had no home for documentation
deliverables except calling Markdown an MVP, and improvised a "rung genuinely
absent" escape hatch three times. Verification artefacts (anchor corpora,
calibration fixtures, labelled pairs) got stuffed into MVP but are really
evidence attached to a rung — the same corpus serves both the MVP script and
the tool gate.

**Attack 3 — conflation: lands terminologically, glances conceptually.**
The framework's layer-1 definition used "discipline" for the unenforced
statement, while abcd's promotion path defines a discipline as the *enforced*
form — layer 3. The first sentence contradicted the doctrine it claimed to
translate. Conceptually, "practice" does mix values, definitional refinements
of proven/done, and detailed procedures — but splitting these would triple
the taxonomy for little gain; a tightened definition suffices.

**Attack 4 — rival framings: partially lands.** Policy/mechanism plus a
maturity axis on mechanism fits the corpus strictly better — it explains
every wobble (schema-parity-first-gate enters at tool purely because abcd
already holds the policy and half the mechanism; the layer label records the
repo-relative delta and rots as the repo evolves). Wardley evolution maps
genesis→product onto script→core-absorption, which script-first-mvp already
encodes. But no rival is worth switching to wholesale: the trichotomy's real
value is not the taxonomy but the **intake discipline** — articulate the full
ladder, mark absent rungs, never fabricate — which demonstrably produced
honest extractions (61 candidates with explicit absent-rung annotations and
one self-declared weakest candidate).

**Attack 5 — redundancy: lands as stated, with a twist.** The trichotomy is
very nearly promotion-path ∘ script-first-mvp: practice = the README's
not-yet-enforced layer, tool = discipline-kind intent or core absorption,
MVP = script-first-mvp's script rung generalised to conventions and file
formats. What it adds: the explicit MVP rung between principle and
discipline, and the two intake rules. But it also *contradicted* the
promotion path on one point — the README files "a value or convention" at
the principle layer while the trichotomy filed every convention at MVP.
Adopting it as a third standalone doctrine file would therefore both
duplicate and contradict existing doctrine — a one-canonical-primitive
violation at doctrine level. The fix is landing it as an amendment that
unifies, never a sibling.

**Attack 6 — the tool trap: lands decisively against the phrasing.** "The
extension of the MVP that makes the practice better" is refuted by roughly a
third of the corpus the trichotomy classified: the most anchor-accurate
reviewer drove the worst producer outcomes (88%-accurate reviewer → 21.8%
writer success vs 80%-accurate → 40.5%); tools saturate and must be archived;
detectors are themselves debt with false-positive budgets and kill criteria;
tools get gamed and deform the practices they serve. The ladder as stated was
monotonic ascent; the sources say the top rung is where gaming, ossification,
and decay live, so it needs a downward arrow.

**Net.** A useful intake schema with five repairable defects: a proven
instability at the practice/MVP seam, a repo-relative layer label
masquerading as an intrinsic type, one terminological self-contradiction, a
falsified apex claim, and a doctrine-level duplication risk. Rejection would
discard an intake discipline that verifiably produced honest, well-grounded
extractions.

### The eight amendments

1. **Layer = entry rung**, explicitly repo-relative and dated: the rung where
   the candidate's actionable delta sits given the current record, with every
   rung marked exists/partial/absent as schema, not habit.
2. **Key candidates by the practice they serve**, so cross-source duplicates
   merge mechanically instead of being classified apart (the
   badcase-corpus/escapes-feed-the-corpus split is the proof case).
3. **Strike "or discipline" from the layer-1 definition** — a discipline is
   the enforced form and belongs to layer 3; layer 1 is the normative
   statement that survives any mechanism.
4. **Rewrite the tool definition**: not "makes the practice better" but
   "makes the practice enforced or cheap, at the price of becoming a
   maintained artefact with its own lifecycle — false-positive budget, gaming
   surface, saturation, kill criterion" — and add the downward arrow (demote
   to advisory on stale calibration, archive to regression-only on
   saturation), warranted by the corpus's own judge-the-gate-by-the-work,
   detector-lifecycle, and saturation evidence.
5. **Land as an amendment to the promotion path** in
   [`principles/README.md`](../../principles/README.md), cross-referenced
   from [`script-first-mvp`](../../principles/script-first-mvp.md) — never a
   third doctrine file. Resolve the convention collision explicitly: an
   unenforced convention remains a principle-layer artefact in the record;
   "MVP" names the enabling artefact beneath it, not a third governance
   category.
6. **Degenerate-ladder clause** for record-shaped candidates: documentation
   restructures, scope paragraphs, and preambles may declare principle = MVP
   or tops-out-at-MVP without penalty.
7. **Evidence artefacts are attachments to a rung, never rungs** —
   acceptance corpora, calibration fixtures, labelled pairs, baselines.
8. **Keep the two intake rules verbatim** — articulate the full ladder for
   every candidate; never fabricate an absent rung. They are the framework's
   demonstrated value and survive every attack.

### The canonical amended ladder

> principle (the normative statement — value, definition of proven/done/good,
> or rule of action — that survives any particular mechanism) → enabling
> convention, script, or file format (the MVP: the smallest unenforced
> enabler) → discipline-kind intent or core absorption (the tool: makes the
> practice enforced or cheap, at the price of becoming a maintained artefact
> with its own lifecycle — false-positive budget, gaming surface, saturation,
> kill criterion — demotable to advisory on stale calibration, archivable to
> regression-only on saturation). A candidate's layer is its ENTRY RUNG —
> repo-relative and dated, the rung where the actionable delta sits given the
> current record, with every rung marked exists/partial/absent. Evidence
> artefacts (acceptance corpora, calibration fixtures, baselines) are
> ATTACHMENTS to a rung, never rungs. Record-shaped work may declare a
> degenerate ladder (principle = MVP, or tops out at MVP). Intake rules:
> articulate the full ladder for every candidate; never fabricate an absent
> rung.

### Boundary cases (the adversary's evidence file)

Twelve classification wobbles ground the amendments; the sharpest:
badcase-corpus vs escapes-feed-the-corpus (identical content, different
layers — amendment 1/2); schema-parity-first-gate (enters at tool solely
because the practice already exists here — the label is repo-relative);
record-scope-facts-not-theory, ladder-onboarding, adr-argument-fields (the
three degenerate ladders — amendment 6); checkable-anchor-criteria (MVP/tool
boundary = enforcement status, the promotion path's own axis);
context-scoped-successor-bans (the direct convention collision with the
README — amendment 5); validator-in-the-loop (its tool rung exists as iss-39
while its MVP is absent — the ladder runs backwards);
provenance-in-the-same-change and preregister-the-intent (all rungs exist —
articulation is ceremony, useful only as lineage evidence);
artifact-level-ai-role (the schema's demand for a practice rung invited a
strained one rather than preventing it).

### Relation to existing doctrine, and the recommendation

The trichotomy is neither duplicate nor orthogonal — a refinement that
unifies the two existing doctrines and contradicts one at a seam.
script-first-mvp governs the MVP→tool edge for capabilities; the promotion
path governs the practice→tool edge for norms and had no intermediate rung.
The trichotomy's genuine addition is the explicit enabling-artefact rung
between principle and discipline, plus the two intake rules neither doctrine
states. Recommendation, honoured in this batch: land the ladder as an
amendment to the promotion-path paragraph in `principles/README.md` with a
cross-reference from `script-first-mvp` naming it the capability-specific
instance of the MVP→tool edge — **one canonical ladder, never a third
doctrine file** — incorporating all eight amendments.

## Proposals by disposition

All 55 surviving proposals. Layer is the entry rung (repo-relative, dated
2026-07-09). Sources give public CSL keys; "lineage review" and "SOTA
mapping" are the two 2026-07-08 notes linked above; tainted evidence says
"confidential corpus (AI-generated)".

### Recorded (three new principles, two principle extensions, README amendment)

The framework ladder itself lands as the promotion-path amendment in
`principles/README.md` plus the `script-first-mvp` cross-reference. Ten
proposals land in principle files:

| Proposal | Rung | Kind | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|---|
| verifier-selects-gates-decide | practice | new principle | high | An LLM verifier slots above the deterministic gate, never in place of it: it selects among candidates and monitors progress; the deterministic gate stays the verdict of record, and every declared LLM gate names its deterministic partner. | kwok2026llmasaverifier, chesterman2026research | Sound; deterministic-gate-first merged in as the reproducibility warrant (~78% verifier pairwise ceiling = a verifier-only gate passes ~1 in 5 wrong candidates). |
| evaluate-at-the-user-surface | practice | new principle | high | An evaluator consumes the system through the same surface its users do; privileged access paths make it structurally blind to exactly the failures users hit. | arXiv 2604.09581 | Sound — called the strongest candidate in its source; iss-31/iss-48 verified open as instances. |
| evaluator-outside-the-loop | practice | new principle | high | The gate that judges a change and the permission layer that bounds it live outside the surface the change may edit; a change touching its own gate is split out for outside review. | weng2026harness | Sound — generalises ratchet-not-big-bang's baseline special case; quotes verified against the live post. |
| detector-lifecycle | practice | extends fix-the-detector | high | Detectors are themselves debt: each carries a false-positive budget, a kill criterion, a correlation audit, and pruning when the guarded class disappears. | SOTA mapping, lineage review, iacob2026redqueen, glm5team2026vibe | Sound; umbrella for the decay theme — saturation-triggers-succession and refresh-the-corpus folded in. |
| anchored-detector-succession | practice | extends fix-the-detector | high | A detector is never replaced or loosened on argument alone: challenger and incumbent score on the same fixed anchor corpus; strict improvement promotes, ties keep the incumbent. | iacob2026redqueen | Sound — fix-the-detector verified silent on replacement. |
| escapes-feed-the-corpus | practice | extends fix-the-detector | high | A detector's demonstrated escapes become adversarial fixtures in its successor's objective; every verified defect graduates into the corpus. | iacob2026redqueen, glm5team2026vibe | Sound; badcase-corpus merged in with the issue-to-fixture graduation MVP. |
| held-out-proof-for-detectors | practice | extends fix-the-detector | medium | For judgment-shaped detectors, flagging every founding instance is training evidence, not proof; adequacy is demonstrated on held-out instances. | iacob2026redqueen | Sound — mildly corrects fix-the-detector's "proven" clause. |
| freeze-the-criterion-within-a-pass | practice | extends fix-the-detector | medium | A detector changes only between passes, never during one; the forced re-audit is lazy. | iacob2026redqueen | Sound — governs the timing of mutation, distinct from the rest of the family. |
| verdicts-carry-their-detector | practice | extends fix-the-detector | medium | Every persisted verdict names the detector version that produced it; on detector change, dependent verdicts are erased and recomputed, never rescaled. | iacob2026redqueen | Sound — the gap in ratchet-not-big-bang verified; family-filed with derived-artifacts-pin-inputs. |
| bidirectional-proof | practice | extends guards-prove-themselves | high | Every acceptance corpus carries negative (ok:) cases and is curated as key examples — minimal, distinct-facet, boundary-inclusive. | SOTA mapping, lineage review, weng2026harness | Sound; held-out-corpus-for-detectors absorbed; adversary directed text amendments, not a twelfth file. Its ratchet-not-big-bang prune clause rides with the deferred set. |

### Captured as seeds

Filed via the native capture surface; each names its practice and detector.
Pri is the extraction pipeline's priority and is independent of the captured
issues' severity field. Throughout this note, the run's "practice" vocabulary
names the same rung the principles README calls "principle":

| Proposal | Rung | Pri | Seed | Statement |
|---|---|---|---|---|
| wiring-status-vocabulary | mvp | high | iss-50 | Decided/Implemented as the fixed value set for brief rows' wiring-status field — the state whose absence produced the empty-`shipped/` incident. Sources: lineage review. Adversary: sound — "one of the most immediately actionable candidates". |
| context-scoped-successor-bans | mvp | high | iss-51 | banned_tokens entries carry a successor and name their context; allow-contexts narrow and enumerable. Corrects iss-36 before implementation. Sources: lineage review, SOTA mapping. Adversary: sound — the live conflict with iss-36 verified. |
| tier-travels-with-the-source | mvp | high | iss-52 | Every corpus source carries an epistemic tier surfaced at every consult; AI-generated material never launders into reviewed authority. Sources: chesterman2026research. Adversary: sound — governs this pipeline's own taint handling. |
| ears-unwanted-behaviour-rows | mvp | medium | iss-53 | Every mutating verb's brief row carries at least one If/Then unwanted-behaviour clause (EARS). Sources: lineage review. Adversary: sound — spec-side complement to iss-34. |
| adr-argument-fields | mvp | low | iss-54 | ADR frontmatter selectively gains related_principles, a confidence qualifier, and revisit_when — three fields, not fourteen. Sources: lineage review. Adversary: sound — praised for adversarial posture toward its own source. |

iss-55 (corpus-stub-backfill) is housekeeping, not a proposal — see below.

### Deferred (parked until an instance arises)

Per reality-is-filable's states-on-observed-need: these are sound but bind
mechanisms or duties abcd does not yet operate; each activates when its
instance appears. 31 proposals, by family.

**Instrumented judgment** (activates with the first live LLM judge —
itd-75/itd-1 territory):

| Proposal | Rung | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|
| calibrate-the-judge | practice | high | An LLM judge counts as a gate only after its verdicts are measured against a human-labelled sample, recalibrated on model/prompt change. | glm5team2026vibe, kwok2026llmasaverifier | Sound; measured-oracle-accuracy merged in (tie rate, correlated-bias warning). The one high-priority deferral: no LLM gate exists yet to bind. |
| judge-the-gate-by-the-work | practice | medium | A gate over adaptive producers is judged by the work it induces and its gaming resistance, not standalone accuracy. | iacob2026redqueen | Sound with a bound: 88%-accurate reviewer → 21.8% writer vs 80% → 40.5%; anchor-accuracy floor must be in the filed text. |
| standardized-instruments-for-llm-judgment | practice | medium | LLM-judge scores anchor to a fixed validated instrument, never ad-hoc per-run scales. | arXiv 2604.09581 | Sound — file as the coordinated instrumented-judgment family. |
| graded-scores-to-rank-discrete-verdicts-to-gate | practice | medium | Rank with graded scores, gate with discrete verdicts; discrete ranking collapses hedged belief into ties. | kwok2026llmasaverifier | Sound — 88/100 ties from a discrete judge on a provably unequal pair. |
| decompose-compound-verdicts | practice | medium | Compound verification questions bias toward the salient factor; decompose into per-criterion passes when stakes warrant the cost. | kwok2026llmasaverifier | Sound — cost-proportionality clause mandatory (~2–3-point lift at C-times cost). |
| evidence-outlives-the-claim | practice | medium | A verification claim counts only if it ran against the canonical target and its artefacts persist; evidence-free verdicts downgrade. | kwok2026llmasaverifier | Sound — carrier of no-self-certified-done's surviving intuition. |
| counterbalanced-comparison | mvp | low | Pairwise oracle runs structurally cancel slot bias (slot-swap, ring pass, averaged). | kwok2026llmasaverifier | Sound — MVP entry verified correct. |
| progress-score-early-warning | tool | low | A graded verifier score over the trajectory prefix flags stall/drift in long runs — advisory telemetry, never a gate. | kwok2026llmasaverifier | Sound — itd-29's operator surface is the exact seam. |

**Gate and judgment values:**

| Proposal | Rung | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|
| abundance-shifts-gates-to-judgment | practice | medium | When production gets cheap, the judgment gate strengthens; a generator without a paired evaluator is half a feature. | chesterman2026research | Sound — value statement covered by nothing existing. |
| grounded-not-declared | practice | medium | Verification grounding is a ladder (declared → executed → perceptual); a gamed gate is fixed by moving grounding up a level, never by penalising instances. | glm5team2026vibe | Sound — adds the fix's direction to fix-the-detector. |
| attribute-the-failure | practice | medium | Every gate failure carries a cause class (defect / environment / staleness) before anyone acts; environment failures never count as defects or retry into silence. | glm5team2026vibe | Sound — the misclassification-as-suppression bound must be in the text. |
| humans-move-up-the-stack | practice | medium | Every new mechanical gate declares where the human decision point relocates; oversight moves up, never silently deletes. | weng2026harness | Sound — decorates the same promotion-path paragraph this run amends. |

**Agent-run conventions** (activate as the autonomous-run surface lands —
itd-29/itd-59 territory):

| Proposal | Rung | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|
| typed-refusal-perimeter | practice | medium | Prohibited surfaces are declared up front; contact terminates the run with a typed, machine-readable refusal. | arXiv 2604.09581 | Sound — real delta over unrecognized-input-never-writes. |
| declared-editable-surface | practice | medium | Work invited onto config/record/harness files gets an explicit editable-surface declaration; outside edits are refused and logged. | weng2026harness | Sound — cross-reference typed-refusal-perimeter at filing. |
| durable-delegation-artifacts | practice | medium | Delegated and background work writes outputs and status to durable working-tier files as it runs; transient-only results are already lost. | weng2026harness | Sound — resolves a verified conflict between the tier conventions and the delegation rule. |
| itemized-merge-not-rewrite | practice | medium | Accumulated orientation artefacts grow by item-level operations, never wholesale regeneration. | weng2026harness | Sound with mandatory scope bound (composes with iss-38's generation remedy for derived files). |
| negative-results-are-record | practice | medium | Abandoning a hypothesis produces a filed outcome, not silence — the residual is abandonment-without-successor plus the mandatory why. | weng2026harness | Sound — the filed text must name the residual precisely. |
| runtime-outcome-checklist-sync | mvp | medium | A run decomposes its task into 2–6 observable outcome states kept as a live checklist, verified before terminating. | arXiv 2604.09581 | Sound — domain transfer accepted (drift mechanism is horizon length). |
| stepwise-friction-map | mvp | medium | A fixed friction-tag vocabulary emitted at the moment of friction, not in post-hoc review. | arXiv 2604.09581 | Wrong-layer upheld and applied: corrected from tool to MVP (two-rung skip over undelivered itd-59). |
| docs-informed-evaluation | mvp | medium | Eval agents get the relevant docs/ how-to as roadmap and record which doc they used — every behavioural eval doubles as a docs test. | arXiv 2604.09581 | Wrong-layer upheld and applied: corrected to MVP, seeding a new intent. |
| validator-in-the-loop | mvp | medium | A generator of record artifacts runs the validator on its own output before finishing. | glm5team2026vibe | Sound — tool rung verified open as iss-39; the middle rung is the delta. |
| no-lossy-round-trips | practice | medium | At machine-to-machine seams, pass the canonical structured representation; never emit a human rendering and re-parse it. | glm5team2026vibe | Sound — consumer-side rule absent; producer architecture exists. |
| fold-the-history | mvp | low | Long-running skills fold stale tool output and re-orient from the durable record past a token threshold. | glm5team2026vibe | Wrong-layer upheld and applied: corrected from practice to MVP; one model's constants rejected as doctrine. |

**Record clauses** (text amendments awaiting their next principles pass):

| Proposal | Rung | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|
| exception-set-disclosure | practice | medium | An enforcement claim states its blocking semantics and its exception set; silence on exceptions overstates. | lineage review, SOTA mapping | Sound — clean one-clause amendment to enforcement-claims-are-facts. |
| checker-adapts-to-artifact | practice | medium | Automate validation without changing the specification: the checker adapts to the artefact, never the artefact to the checker. | lineage review | Sound — filed form corrected to a bounds clause on fix-the-detector. |
| tacit-residue-stopping-rule | practice | medium | reality-is-filable gains a stopping rule: taxonomy defects are filable in principle; tacit residue is not a defect — stop filing. | lineage review, SOTA mapping | Sound — the layer split upheld. |
| event-capture-beside-taxonomy | practice | medium | Capture the event even when the state lags: an append-only layer saves the fact while the taxonomy fix is pending, never instead of it. | SOTA mapping | Sound — the while-not-instead reconciliation goes in the amendment text. |
| record-scope-facts-not-theory | practice | medium | The record claims fact-currency, never theory-transmission; the working theory is rebuilt each session, and gates exist because that rebuild is fallible. | lineage review | Sound — filed form corrected to a README scoping preamble; a degenerate-ladder boundary case. |

**Lineage corroborations** (all rungs exist; no artefact beyond a citation —
the corpus ledger carries the edges):

| Proposal | Rung | Pri | Statement | Sources | Adversary |
|---|---|---|---|---|---|
| provenance-in-the-same-change | practice | low | Independent corroboration of the same-change attribution practice. | chesterman2026research | Sound as already-covered; constrained to spawn no artefact. |
| preregister-the-intent | practice | low | Grounds intents-precede-surfaces in the preregistration/model-card literature. | chesterman2026research | Sound as already-covered; lineage grounding only. |
| reasoning-traces-explain-why | mvp | low | Corroborates itd-59's premise that passive logs record what, not why. | arXiv 2604.09581 | Sound as already-covered; a citation on itd-59, nothing more. |

### Tainted hypotheses (await independent verification)

Six proposals rest wholly or partly on the four AI-generated confidential
analyses in the local sources corpus. Per tier-travels-with-the-source's own
logic, AI-generated material supports nothing on its own: each survives the
adversary only because an independent public lineage exists, and any future
filing must cite that lineage in place of the AI evidence. None is recorded
in this batch.

| Proposal | Rung | Pri | Statement | Independent lineage | Adversary |
|---|---|---|---|---|---|
| checkable-anchor-criteria | practice | medium | Acceptance criteria carry a machine-checkable anchor or an explicit human-judgment tag — at intake, not downstream. | EARS, specification-by-example | Sound; fit corrected to extends-itd-1 (G-W-T prose can still be uncheckable). |
| forbidden-transition-unrepresentable | practice | medium | Order-of-work invariants are structural: the forbidden shortcut has no representation. | make-illegal-states-unrepresentable | Sound — structural dual of reality-is-filable, covered by nothing else. |
| absent-guard-refuses | practice | medium | A missing guard is a denial, never a silent bypass: the guarded operation refuses when the guard cannot be established. | established fail-closed security practice | Sound — neither guards-prove-themselves nor unrecognized-input-never-writes covers unavailability. |
| derived-artifacts-pin-inputs | practice | medium | Derived artefacts record a content hash of their inputs; consumers refuse on post-derivation change. | build-system content addressing | Sound — names the root cause behind iss-38 and iss-47. |
| schema-parity-first-gate | tool | medium | Cross-surface contract parity proven by a byte-comparable golden fixture as the first CI gate. | golden-file testing | Sound — the adversary's own exhibit that layer is a repo-relative entry rung; concrete detector shape for iss-44. |
| ladder-onboarding | mvp | low | Onboarding docs as one dependency-ordered ladder with a five-minute first win. | progressive disclosure | Sound but barely — rated low by the challenge. |

All six sources: confidential corpus (AI-generated).

### Rejected

| Proposal | Reason |
|---|---|
| artifact-level-ai-role | Unsupported: the evidence warrants generic AI-use disclosure, which the Assisted-by trailer and ACKNOWLEDGEMENTS conventions already implement; the per-artifact-granularity delta is speculative. Source: chesterman2026research. |
| inconclusive-is-first-class | Duplicate of the existing record: itd-1 already fixes the four-verdict vocabulary with INCONCLUSIVE first-class. AI-generated evidence carries zero weight. Source: confidential corpus (AI-generated). |
| no-self-certified-done | Duplicate of the existing record: itd-1's intent-fidelity reviewer plus itd-50's budget-bounded loop with an UNACHIEVABLE terminal. Surviving fragments rehomed into evidence-outlives-the-claim. Source: confidential corpus (AI-generated). |

## Housekeeping

- **Six public corpus entries are keyword stubs** with no document text: the
  five practitioner SDLC-cluster pieces (which yielded zero candidates for
  exactly that reason) and weng2026harness (whose evidence was verified
  against the live post instead — the ledger copy must be backfilled before
  its candidates are cited downstream). Filed as iss-55; the backfill runs
  through the ingest path so tier and provenance travel with the text.
- **The unidentified PDF resolved to chesterman2026research** — the SSRN
  research-integrity working paper, already registered in the corpus. The
  on-disk PDF attaches to that entry corpus-side; no new registration.

## Cross-references

- [Lineage bibliography vs the ten principles](2026-07-08-lineage-sources-vs-principles.md)
  and [Principles vs state of the art](2026-07-08-principles-sota-mapping.md) —
  the two notes the refs-bib-reclassify extractor re-read; eleven of the 61
  candidates are their findings re-expressed as ladders.
- [`principles/README.md`](../../principles/README.md) — the amended
  promotion path this run produced.
- [`principles/script-first-mvp.md`](../../principles/script-first-mvp.md) —
  the capability-specific instance of the MVP→tool edge.
