# Lineage bibliography vs the ten principles — 2026-07-08

The intellectual-lineage bibliography (21 entries: theory of programming,
decision-record methodology, requirements and acceptance criteria, domain
language) was reviewed one source per research agent, each verifying what the
source actually says, classifying its bearing on each relevant principle
(confirm / support / contradict / extend / new), and judging uniqueness
against the [principles SOTA mapping](2026-07-08-principles-sota-mapping.md).

## Headline

The bibliography is largely non-redundant with the SOTA mapping — the mapping
was engineering-practice all the way down, and the lineage supplies what it
lacked: an epistemic-limits counterweight, the decision-record academic
ancestry, the executable-specification root, argument theory, and domain
language. **No source contradicts any principle's scoped reading.** Every
contradiction found operates at a maximal reading and resolves into a scope
line the principle should state. Two entries are honest pointers with little
independent content (Fowler ADR, Fowler GivenWhenThen); one belongs in a
different lineage entirely (Chang → the review/consult skill layer, not the
principles).

## The counterweight cluster (the mapping's missing voice)

- **Naur 1985** — the principled limit: a program's theory "could not
  conceivably be expressed" in documentation; his compiler case is a
  *phantom-free, accurate* record that still failed to transfer design
  insight. EXTENDS enforcement-claims-are-facts (truthful docs are necessary,
  not sufficient — confining claims to what mechanisms demonstrably do is
  retreating to the only territory text can hold). CONTRADICTS maximal
  readings of fix-the-detector (Ryle's regress: a detector encodes instances
  of a class, never the judgment that delimits it; detectors themselves rot
  when their theory-holders leave — the causal account under the detector-
  lifecycle theme) and reality-is-filable (some reality is unfilable in
  principle). Trap: the "documentation helps the next programmer build a
  theory" consolation appended to the common PDF is Cockburn 2002, not Naur.
- **Polanyi 1966** — supplies the **stopping rule** reality-is-filable lacks:
  taxonomy defects (filable in principle — fix the taxonomy) vs tacit residue
  (unfilable in principle — not a defect, stop). Indwelling gives the
  structural mechanism beneath the automation-complacency evidence: a reader
  attends *from* the record *to* the work, so phantom claims are absorbed in
  the normal mode of trust — verification must be a separate focal gate.
  Exemplar transmission grounds the acceptance-corpus clause: the corpus,
  not the rule text, carries the class.
- **Bjarnason 2022 / Goedecke 2026** — the agent-era application: **every
  agent session is maximal churn** (a fresh team member with no retained
  theory and no possible apprenticeship). This simultaneously explains the
  record-centric bet (there is no colleague to absorb theory from) and bounds
  it: the record's realistic promise is *cheaper theory reconstruction per
  session*, not theory persistence. Goedecke is the bibliography's only
  contrarian on the bet itself (documentation "a little" help; his remedy is
  model-level), and supplies the mechanism behind "gates are how agents
  self-correct": agents visibly run hypothesise/test/adjust loops, and a
  deterministic detector is a falsifier inserted into the loop's exact slot.
  Session-amnesia also freshens every same-change discipline: with an
  amnesiac workforce there is no one for whom "later" exists.
- **Martin 2002** — same diagnosis (unmaintained documents drift into lies —
  the phantom-gate argument 24 years early), opposite remedy (minimise
  documents, code is the design). The record must own this contradiction;
  the answer is agent consumption plus the same-change/lint machinery that
  neutralises the drift argument. Also: reactive OCP ("take the first
  bullet, then close against that kind of change") is pre-Spolsky prior art
  for fix-the-detector; *viscosity* names the paved-road causal driver;
  needless-complexity contradicts loud-staging's licence while confirming
  why the lease bound is load-bearing.

**Scope line these four jointly demand (candidate record statement):** the
record carries decisions and checkable facts; the theory is rebuilt every
session; gates exist because the rebuild is fallible. The ten principles
claim fact-currency, never theory-transmission.

## Decision-record lineage (with corrections)

- **Jansen & Bosch 2005** — "knowledge vaporization" names the rationale leg
  of reality-is-filable (Bowker & Star covers the taxonomy leg), and adds the
  causal chain the principle's Why lacks: unfilable knowledge actively causes
  erosion. The Archium-lineage adoption post-mortem (heavyweight decision
  tooling failed; lightweight text ADRs won) is a natural experiment
  endorsing the repo's markdown-ledger choices.
- **Tyree & Akerman 2005** — the rejected-alternatives discipline is T&A
  lineage, **not Nygard** (Nygard's template has no alternatives field).
  Actionable: a `related_principles:` ADR frontmatter field (their Related
  Principles element), now that principles/ exists. Caution: a 2026
  controlled comparison eliminated their full 14-field template at expert
  screening — adopt fields selectively.
- **Nygard 2011** — beyond the existing citation: supersession-never-deletion
  plus never-reuse-numbers is identifier retirement achieved by construction
  (retire-the-name's mark half without a lint); the blind-acceptance /
  blind-reversal dilemma is the citable reader-side Why for
  reality-is-filable. Successor pointers ("Superseded by ADR-N") are
  adr-tools/MADR practice, not the 2011 post.
- **Fowler ADR bliki** — redundant pointer; sole new atom is the
  **confidence-level field** on decisions. Keep as secondary/recency anchor.
- **Dagdeviren** — the genre split (proposal states and decision states get
  separate status vocabularies) extends reality-is-filable; **Decided ≠
  Implemented** is a ready-made value set for the brief row's wiring-status
  field and precisely the state whose absence produced the empty-shipped/
  incident. Read at solo scale, the post argues against adopting a separate
  RFC genre (its value driver is people to collect feedback from).
- **Bezos 2015 letter** — the reversibility taxonomy is the missing rationale
  for ratchet-not-big-bang (a ratchet converts gate adoption from a Type 1
  into a sequence of Type 2 decisions; the survivorship footnote grounds the
  non-baselinable mandatory-zero tier) and the canonical citation for the
  ADR-promotion threshold ("expensive to reverse"). Cite the 2015 letter for
  Type 1/2 only; decide-at-70% / disagree-and-commit is the 2016 letter.
- **Bryar & Carr 2021** — "good intentions don't work; mechanisms do" is the
  management-canon anchor for the promotion path itself. Two new design
  properties: **gatekeeper independence** (the bar raiser sits outside the
  pressured team — relevant to who approves baseline regenerations and lint
  escapes) and the **correlation-audit loop** (detector green while the
  finding class recurs ⇒ revise the detector's definition — extends the
  lifecycle theme beyond FP budgets and kill criteria). PR/FAQ is
  review-cadence-anchored rather than write-and-decay; the spec-first camp
  assignment stands with that one clause.

## Executable-specification root

- **Adzic 2011** — direct ancestor (CONFIRM) of enforcement-claims-are-facts
  and spec-moves-with-the-surface; Martraire descends from it, so the
  lineage root moves back to 2011 with the only multi-team empirical base
  (50+ case studies) the living-documentation claim has. Two absorptions:
  **key-examples corpus curation** (minimal, distinct-facet,
  boundary-inclusive — not every instance dumped in) and **automate
  validation without changing the specification** (the checker adapts to the
  human artifact; a record-lint must never force the brief to be rewritten
  gate-friendly). The phrase "specification by example" is Fowler 2004;
  Adzic contributed the pattern language and "living documentation".
- **Fowler GivenWhenThen** — thin pointer; one substantive line
  (then-clauses free of side-effects — an assertion phase that mutates can
  mask the very write a refusal fixture rules out) plus the attribution
  chain (Terhorst-North & Matts for GWT; Wake for AAA; Meszaros for
  four-phase).
- **Mavin et al. 2009 (EARS)** — the strongest under-recognised fit. Adds
  the **specification-side leg of the guard story**: every
  guards-prove-themselves source tests a guard that exists; EARS makes the
  refusal a first-class testable spec item before code — its core empirical
  claim is that unwanted-behaviour handling is the dominant omission class,
  which is why the capture typo-seam was unspecified. Authoring discipline
  to absorb: every mutating verb's brief row carries at least one If/Then
  unwanted-behaviour clause. Also 2009 industrial prior art for the brief's
  fixed-field rows (gently constrained natural language beats free prose
  and formal schemas), with a live bridge: Kiro emits acceptance criteria in
  EARS notation and spec-kit has an open EARS integration request. Its
  evolve-the-ruleset methodology is an independent lineage for
  reality-is-filable's expand-on-observed-need rule.

## Domain language and argument theory

- **Evans 2003** — bears on four principles and forces the originality
  correction below. Rename-propagates-to-code-immediately is 2003 doctrine
  (earliest prior art for retire-the-name's first half; the same-change
  *ban* coupling remains original). The anticorruption layer distinguishes
  sanctioned boundary-translation from forbidden intra-context copies — and
  the surface front doors onto the transport-agnostic core are structurally
  ACLs. The context map (as-is shared cartography) is a second,
  mechanically distinct support for reality-is-filable. Lineage fact: the
  brief's per-context glossary with forbidden synonyms is an operating DDD
  implementation — ACKNOWLEDGEMENTS-grade if formally adopted.
- **Fowler BoundedContext 2014** — where canonicality stops: one canon per
  context plus an explicit map; legitimate divergence must be *recorded*,
  not merely permitted. New bound for retire-the-name: **banlist entries
  name their context** — a term retired in one context may persist
  legitimately in another; a global ban over polysemous terms is a
  false-positive generator.
- **Toulmin 1958** — unifies three of the mapping's independently-derived
  refinement themes under one 1958 anchor: bidirectionality (warrant and
  rebuttal jointly mandatory), negative corpus examples (rebuttal conditions
  are part of the proof), contracts-stated (the qualifier must be explicit).
  New refinement: enforcement claims disclose their **exception set**
  (allowlists, baselines, skipped paths), not just blocking semantics.
  Format suggestion: an explicit revisit-when (rebuttal) field in ADRs; the
  research notes' evidence tiers are operationalised qualifiers already in
  use. Also a sixth independent lineage for reality-is-filable (working
  logic vs the analytic ideal, 1958). The classroom "Toulmin method" is
  Ehninger & Brockriede packaging, not the 1958 text.
- **Nonaka & Takeuchi 1995** — the theoretical warrant for the record
  itself: externalisation is possible (contra strong Polanyi readings),
  bounded by the 2009 continuum concession — cite 1995 together with Nonaka
  & von Krogh 2009 and Gourlay 2006, or it overstates. Fix-the-detector is
  a SECI cycle in miniature (review→externalise→combine→internalise), and
  the theory predicts the observed failure: hand-fixing without
  externalising skips the conversion, so the knowledge evaporates. *Ba* is
  Nonaka & Konno 1998, not the 1995 book. The "redundancy" enabling
  condition concerns team context overlap, not artifacts — not an anti-DRY
  argument.
- **Chang 2023** — weak fit to the ten by its own reviewer's judgment; its
  home is the review/consult skill lineage (CRIT's
  claim→evidence→recurse→score→rivals recipe). Do not add to per-principle
  source lists.

## Correction applied to the SOTA mapping

Both DDD reviewers independently flagged that the mapping overclaimed
one-canonical-primitive's boundary litmus as having "no published
equivalent": the may-legitimately-diverge half is bounded contexts (Evans
2003) in substance; the must-change-together half is DRY-as-knowledge (Hunt
& Thomas). What remains original is the **conjunction as a copy/consolidate
decision rule**. The mapping note has been amended accordingly.

## Attribution traps (verified this pass; respect in any future citation)

Cockburn 2002, not Naur, for "docs help the next programmer build a theory";
*ba* is 1998, not the 1995 book; the 2002 Martin book does not contain the
word "SOLID" (Feathers, later) and collects Meyer's OCP / Liskov's LSP /
Reeves' code-is-design; "specification by example" as a phrase is Fowler
2004; ADR successor pointers are adr-tools/MADR, not Nygard 2011; the
rejected-alternatives ADR field is Tyree & Akerman 2005, not Nygard;
"make work visible" is Benson & DeMaria Barry / DeGrandis, not DeMarco;
high-velocity decision doctrine is the 2016 Bezos letter, not 2015;
"Polanyi's paradox" is Autor's 2014 coinage; "pave the cowpaths" is a
W3C/UX principle, not lint prior art.

## Candidate absorptions (recording seeds, not commitments)

1. A record-scope statement: facts and decisions, not theory; gates fence
   the per-session rebuild (Naur/Polanyi/Bjarnason/Goedecke).
2. reality-is-filable gains the tacit-residue stopping rule and the
   closed-vs-open taxonomy distinction (Polanyi; Diátaxis).
3. retire-the-name bans become context-scoped and successor-naming
   (BoundedContext; faillint/Vale pattern already mapped).
4. Brief rows: EARS-style If/Then unwanted-behaviour clause per mutating
   verb; Decided/Implemented wiring-status vocabulary (Mavin; Dagdeviren).
5. Acceptance corpora curated as key examples with negative cases (Adzic;
   Toulmin's rebuttal).
6. ADR format: `related_principles:` frontmatter, a confidence qualifier,
   and a revisit-when line (Tyree & Akerman; Fowler; Toulmin).
7. Enforcement claims disclose exception sets alongside blocking semantics
   (Toulmin).
