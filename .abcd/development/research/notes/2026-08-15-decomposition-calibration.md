# Decomposition calibration — graded hand-runs of the itd-84 protocol

The calibration corpus for the
[itd-84](../../intents/disciplines/itd-84-intent-decomposition.md) hand-run
decomposition protocol (the four-piece table in the `/abcd:intent` surface
page). Every hand-run appends one graded entry; roughly **50 graded captures**
is the recorded threshold before the automated capture-time validator may be
built, and the grading follows the
[itd-81](../../intents/disciplines/itd-81-judge-calibration.md) recipe: the
initial routing is the prediction, the human-confirmed routing is the label.

## Entry format

Per hand-run, append:

- **Date, proposal** (one line), and the session context.
- **Initial routing table** — part | type | home, plus typed links and any
  reversal flags, exactly as proposed before the human ruled.
- **Confirmed routing** — what the human adopted, with edits marked.
- **Verdict** — FILE-AS-IS / SPLIT / HOLD, and whether the initial verdict
  survived confirmation (the graded outcome).
- **Notes** — over-flags, missed parts, taxonomy ambiguities (feeds the
  open-question on the enum).

## Graded entries

### 2026-07-13 — the auto-merge proposal (founding case, graded retrospectively)

- **Proposal:** a single "`--auto-merge` feature" intent.
- **Initial routing (as filed at the time):** one part | intent | intents/ —
  a monolith.
- **Confirmed routing (from the 2026-07-13 review):** four parts — the
  auto-merge *experience* | intent | intents/; the *trust rule* on what may
  merge unattended | ADR + brief invariant | decisions/adrs/ + brief; the
  additive-vs-editing *stance* | principle | principles/; the eligibility
  *plumbing* | brief | brief.
- **Verdict:** SPLIT. The initial (implicit FILE-AS-IS) routing did **not**
  survive review — the case that motivated itd-84.
- **Notes:** graded from the review record, not a live protocol run; counts
  toward the corpus as the canonical SPLIT exemplar.

### 2026-08-15 — itd-111 planning interview (first live hand-run)

- **Proposal:** itd-111 (staleness detection), an already-filed draft at its
  planning interview — the decomposition ran over the draft before promotion.
- **Initial routing:** five parts — staleness detection + refusal
  (capability | this intent); the network trichotomy (trust rule | ADR +
  brief invariant — flagged as system-binding: "no version-discovery request
  exists anywhere in abcd" outlives the feature); anti-wallpaper micro-prompt
  (capability seed | already extracted to iss-230); vintage-comparison seam
  (plumbing | brief via the spec); platform parity (verification | itd-109
  calibration set, `refines`). One advisory reversal flag: the SessionStart
  staleness notice vs the iss-206 skew-notice retirement (install-experience
  decision 7).
- **Confirmed routing:** adopted unchanged. The reversal flag was ruled a
  **scoped replacement** (`refines`, not `reverses`): steady-state machinery
  stays retired; itd-111 covers the non-steady states.
- **Verdict:** SPLIT (network posture → adr-38 + brief invariant 7),
  proposed and confirmed — the initial verdict survived confirmation.
- **Notes:** no over-flags; the one reversal candidate was genuinely
  ambiguous and worth the human ruling (advisory-only behaved as designed).
  Taxonomy: "trust rule → ADR + brief invariant" fit cleanly; no enum
  ambiguity encountered.

### 2026-08-16 — the multi-harness proposal (hand-run at user confirmation)

- **Proposal:** "make abcd available on further harnesses, as native as
  possible, mapping concepts where the host has them, MCP as the fallback,
  never double-writing per host" — plus a host-tier statement (MCP floor for
  any harness; the current plugin host as the assumed SOTA surface for the
  time being; an open-source harness as the eventual default).
- **Initial routing:** four parts — the host-tier policy incl. the
  default-flip gate (trust/strategy rule | ADR | decisions/adrs/); the shared
  adaptor machinery — host profile seam, ladder semantics, parity suite
  (capability | reworked itd-22, renamed `harness-portability`); the MCP
  front door (capability, the floor | new intent); per-host adoptions incl.
  concept mapping (capability | one intent per host, DEFERRED to explicit
  adoption decisions, nameless until then). Typed links: the 2026-08-15
  DECISIONS entry `reverses` the out-of-scope annotation on itd-22 (human
  had already confirmed that reversal); adr-39 `refines` that decision;
  itd-22 rework `supersedes` its own prior single-host framing (id kept).
- **Confirmed routing:** adopted structurally unchanged; two content edits at
  confirmation — the reference-host rationale reworded to "assumed SOTA
  surface for the time being", and the routing's own justification prose
  ordered kept OUT of the record.
- **Verdict:** SPLIT, proposed and confirmed — the initial verdict survived.
- **Notes:** the stance piece ("never double-write per host") routed into the
  ADR rather than a standalone principle — it restates one-canonical-primitive
  at the host boundary, so a new principle file would have been a second copy;
  flagged here for the ~50-capture calibration review. Rename executed with a
  retire-the-name ban (`retired-itd-22-slug`).

### 2026-08-16 — itd-114 collision-proof ids (hand-run at capture)

- **Proposal:** abcd mints collision-proof record ids across parallel agents —
  native time+content-hash default, optional GitHub forge backend for capture.
- **Initial routing:** four parts — the collision-proof minting capability
  (capability | intent itd-114); the native-scheme + forge-adapter seam (SOTA
  declaration path-2 | inside the intent's SOTA section); the
  "collision-proof-by-construction, native-default forge-optional" stance
  (stance | embodies existing basics-built-in / adapter-over-native-default /
  prefer-sota principles — no new principle); and the id-FORMAT change
  (architecture decision | a future ADR, only if adopted). One advisory reversal
  flag: a pure time+hash id reverses the human-sequential-readable property of
  today's iss-N.
- **Confirmed routing:** adopted (author-confirmed at capture). No separate
  principle or ADR filed now — both are conditional-on-adoption. Reversal flag
  handled by writing an acceptance criterion that FORBIDS the readability
  regression, leaving the exact scheme (time-sortable / git-style dual id /
  forge number) to the grill.
- **Verdict:** FILE-AS-IS with flags (not SPLIT — nothing separate to file yet).
  Initial verdict survived.
- **Notes:** typed link `refines` iss-80 (the resolved issue whose body deferred
  this exact "SOTA under research" scheme). The reversal flag was the valuable
  part — it forced the readability requirement into the ACs rather than letting
  a hash-only scheme ship. Filed while the peer was concurrently minting iss-N
  in a parallel session: a live instance of the collision class the intent
  addresses (no actual collision — different family).

### 2026-08-16 — itd-115 managed-repo merge policy (hand-run at capture)

- **Proposal:** abcd-managed repos merge PRs without the out-of-sync churn —
  merge queue by default, protocol-level auto-update fallback, strict preserved.
- **Initial routing:** four parts — the merge policy capability (intent
  itd-115); the adopt-merge-queue SOTA + merge_group CI trigger + rung-1 fallback
  (SOTA path-1 declaration | inside the intent); the "strict is the duplicate-id
  gate, never relaxed until ids are collision-proof" trust rule (already recorded
  as iss-172's invariant — RESPECT, do not re-declare); the ruleset/CI-trigger
  wiring (plumbing | itd-92/itd-106 onboarding surface).
- **Confirmed routing:** adopted. The trust-rule part is the interesting one — it
  is ALREADY a recorded invariant (iss-172), so the decomposition's job was to
  route the intent to RESPECT an existing record rather than file a new one. The
  first-pass design (relax strict) would have reversed that invariant; reading
  iss-172 before filing caught it and the design was corrected — the value of the
  decomposition's interdependency pass.
- **Verdict:** FILE-AS-IS with flags. NO reversal flag (the corrected design
  respects iss-172; the rejected relax-strict alternative would have reversed it).
  Initial verdict survived after the correction.
- **Notes:** typed links refines iss-172 / itd-92 / itd-106 / itd-114 (itd-114
  unlocks the future relax-strict option). A clean case of the interdependency
  pass preventing a record-contradicting design.

### 2026-08-16 — itd-119 capture promote (hand-run at planning, process plan intent 1)

- **Proposal:** `abcd capture promote <iss-N>` — an issue graduates into an
  intent without retyping; `promoted_to` stamped in the same invocation (the
  first walkability intent of the process-coherence plan).
- **Initial routing:** four parts — the promote capability (capability | this
  intent); the "capture now, decide later" stance (already embodied in the
  Which-ledger note + `PromotedTo` schema — cite, file nothing); the docs
  corrections (capture.md / 04-naming.md / 04-surfaces | ride the sweep); the
  intent-side reciprocal edge field (spec detail | spec body). Typed link:
  `refines` iss-245 (its `promoted_to` half). No reversal flags.
- **Confirmed routing:** adopted structurally unchanged. Two AC-level edits at
  the walk/grill: the proposal's only-open-issues restriction was struck
  (promotion ruled orthogonal to fix-status — the maintainer's challenge), and
  the seed moved from body-copy to a by-id pointer + first line (SSOT) after
  the maintainer floated the link alternative mid-grill.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** the interesting failure was an *over-restriction*, not an
  over-flag: the proposed AC invented a status gate the schema never implied.
  Grill surfaced a real unimplementability (two-store atomicity) that became
  the mint-first + `--intent` repair contract. Filed as itd-119/spc-24.

### 2026-08-16 — itd-120 resolve provenance (hand-run at planning, process plan intent 2)

- **Proposal:** `abcd capture resolve` writes `resolved_by` via `--intent` /
  `--spec` / `--commit` (the second walkability intent of the
  process-coherence plan).
- **Initial routing:** four parts — the provenance-write capability
  (capability | this intent); validation depth (spec detail | spec body);
  docs (capture.md + issues README | ride the sweep); the two-evidence-
  standards observation (motivation | press-release prose, no new record).
  Typed link: `refines` iss-245 (its `resolved_by` half). No reversal flags.
- **Confirmed routing:** adopted unchanged. Rulings at the walk: provenance
  optional (never defaulted, never demanded); existence-checked ids,
  shape-checked sha; backfill of already-resolved issues ruled OUT of scope.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** clean run; the only near-part was the backfill idea, which the
  grill surfaced and the maintainer scoped out rather than routed. Filed as
  itd-120/spc-25.

### 2026-08-16 — itd-121 record-id dispatch (hand-run at planning, process plan intent 3)

- **Proposal:** `abcd <id>` — dispatch on a record id, report what it is and
  the next move (the third walkability intent of the process-coherence plan).
- **Initial routing:** four parts — the dispatch capability (capability |
  this intent); the next-move mapping (spec detail | spec body, one Go
  table); the bare-vs-id mental model (already recorded in the plan — cite);
  docs (root surface page + 04-surfaces | ride the sweep). Duplicate check
  vs itd-86 (cold reading) and itd-112 (banner): negative. No reversal
  flags; SD001-safe by construction.
- **Confirmed routing:** adopted unchanged. Rulings at the walk: `adr-N`
  joins the dispatch read-only (a scope widening the proposal had left out);
  the next-move mapping gains an anti-drift test asserting every recommended
  verb resolves in the cobra tree.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** the maintainer widened scope (adr-N) where the proposal had
  narrowed to the plan's three families — the second time this session the
  human's edit was toward *less* restriction than proposed. Filed as
  itd-121/spc-26.

### 2026-08-16 — itd-122 sub-verb tables (hand-run at planning, process plan intent 4)

- **Proposal:** sub-verb tables in every `04-surfaces/` file plus an extended
  `surface_coverage` checking each row against registered sub-commands both
  ways (the coherence keystone of the process plan; adr-40 decision 6).
- **Initial routing:** five parts — the tables + extended check (capability |
  this intent); the population pass (same change, per adr-40's consequence);
  the four-bucket vocabulary (already adr-40 — cite); the bucket-enum
  registration in 04-naming.md (rides the sweep, VR001); the no-`partial`
  ruling (recorded in draft prose). Typed link: `refines` iss-246. No
  reversal flags — implements a recorded decision.
- **Confirmed routing:** adopted unchanged. Rulings at the walk/grill:
  exemptions are explicit config like `bare_command` (host-delegated
  surfaces, operator-internal verbs); three arguable bucketings pre-ruled
  binding (identity = audit, launch changelog guardrail = gate, guard check
  = gate) so the unattended build never faces the ambiguity STOP on a known
  case.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** the grill's value was schedule-shaped: pre-ruling the arguable
  buckets moves a predictable overnight STOP into the human session. The
  committed `surface.json` snapshot dissolved the expected layering problem
  (core/lint never imports the cobra tree). Filed as itd-122/spc-27.

### 2026-08-16 — itd-123 intent-audit rename (hand-run at planning, process plan intent 5)

- **Proposal:** `abcd intent review` → `abcd intent audit`, breaking, with
  the agent and task-class token moving (adr-40's first named rename).
- **Initial routing:** four parts — the rename sweep (capability, breaking |
  this intent); the vocabulary decision (already adr-40 — cite); the
  sub-verb table row flip (same change, gated by itd-122); the no-aliases
  stance (settled — cite, not reopened). No reversal flags — implements a
  recorded decision.
- **Confirmed routing:** adopted unchanged. Rulings at the walk: the agent
  becomes `intent-auditor` (maintainer's counter-proposal over the proposed
  `intent-fidelity-auditor` — the intent-grain auditor of the three audit
  grains, mirroring the verb); the sweep boundary excludes historical/dated
  records, which keep the old name as history.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived; one naming edit at confirmation.
- **Notes:** the live sweep enumerates to 23 files (the plan's ~37 counted
  historical records the boundary ruling excludes). Filed as itd-123/spc-28.

### 2026-08-16 — itd-124 audit→lint rename (hand-run at planning, process plan intent 6)

- **Proposal:** `abcd audit` → `abcd lint`, breaking; `/abcd:audit` returns
  to itd-16's reservation (adr-40's second named rename; name ruled in
  planning question 1).
- **Initial routing:** four parts — the rename sweep (capability, breaking |
  this intent); the reservation return (already recorded in 04-naming.md —
  the sweep reinstates it); the chosen name (ruled today — recorded in
  draft); no-aliases (settled — cite). No reversal flags.
- **Confirmed routing:** adopted unchanged. The grill's package question
  (`repolint` vs merging the two lint engines) went to adversarial review at
  the maintainer's direction: rename-only won — the engines have different
  rule models and blast radii, and adr-40 rules implementation orthogonal to
  bucket. The merge is captured as iss-251, a deliberate future intent.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** the maintainer's "adversarially review both" instruction is a
  useful grill pattern: the review moved the ruling from taste to facts
  (importer counts, existing cross-import, contract differences). Filed as
  itd-124/spc-29.

### 2026-08-16 — itd-125 disembark-review rename (hand-run at planning, process plan intent 7)

- **Proposal:** `disembark oracle` → `disembark review`, breaking — the
  seventh intent, created by the session's own investigation of planning
  question 2 (the plan had recommended keep-the-verb).
- **Initial routing:** three parts — the rename sweep incl. artefacts and
  agent (capability, breaking | this intent); the adr-40 §5 amendment
  (decision | home ruled by the maintainer: dated in-place amendment in
  adr-40 + a DECISIONS.md line); the oracle seam itself (adr-25 — cite,
  untouched). **One reversal flag, the session's first genuine one:**
  reverses adr-40 §5, investigated and confirmed by the maintainer before
  routing.
- **Confirmed routing:** adopted unchanged. Rulings: agent becomes
  `lifeboat-reviewer` and artefacts `review/review-<manifest12>.*`
  (target-grain + role, self-describing basenames); older lifeboats get
  clean replacement across the rename (the grill's AC — one manifest never
  holds two verdicts).
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived.
- **Notes:** the reversal path worked as itd-84 designed it: flagged
  advisory, investigated against the code (the verb never invokes the seam
  — compute-or-ingest only), ruled by the human, homed before filing. Filed
  as itd-125/spc-30.

### 2026-08-16 — itd-116 GitHub-issue adoption into the ledger (hand-run at capture)

- **Proposal:** a verb that selects validated GitHub issues (the bughunt's
  filings), reviews them, and captures them into the ledger — "bug-adoption
  owner" made a tool.
- **Initial routing:** four parts — the adoption capability (intent itd-116);
  the host-delegated select/review + fail-closed ingest trust rule (already
  adr-25 + the core boundary — spec constraint, no new ADR); the
  human-adoption-gate stance (already `verifier-selects-gates-decide` — cite,
  don't re-file); provenance/dedupe plumbing (spec detail; consumes itd-87
  later). Duplicate check against itd-82 (drain) came back negative: opposite
  pipeline directions (GitHub→ledger vs ledger→PRs).
- **Confirmed routing:** adopted, with one live design exchange at capture —
  the author floated `drain --github-issues` as an alternative surface; set
  aside (drain is an unbuilt draft already stubbing itd-46; different blast
  radius) in favour of a capture extension, recorded in the draft as
  considered-and-rejected. Mint stays capture-only (one-canonical-primitive).
- **Verdict:** FILE-AS-IS (no reversal flags). Initial verdict survived; the
  surface-spelling exchange narrowed open question 1 rather than changing the
  routing.
- **Notes:** typed links builds_on itd-4; refines the DECISIONS.md
  bughunt-hybrid line ("adoption is a downstream human/fix step"); prose
  cross-refs itd-82 (downstream consumer) and itd-87 (recurrence semantics).
  Filed the same day the bughunt's first validated issue (#250) landed —
  itself the first adoption candidate.

### 2026-08-16 — itd-76 sources corpus (hand-run at planning)

- **Proposal:** the existing itd-76 draft — personal corpus + provenance
  ledger + guards, team share/ingest, and paper reconstruction in one
  intent, with three open questions and no acceptance criteria.
- **Initial routing:** six parts — personal core verbs
  (capability | itd-76, narrowed); team share/ingest (capability | new
  draft); paper reconstruction (capability | new draft); "documents and
  ledgers never travel; citation requires both gates" (trust rule | ADR +
  brief invariant); consult-freely-cite-deliberately (stance | principle);
  store formats and folder-as-classification (plumbing | brief/spec).
  Typed links: the split drafts `refines` itd-76; itd-76 `builds_on`
  itd-74 (shipped) and itd-77 (draft, non-gating — default path works).
  No reversal flags.
- **Confirmed routing:** adopted unchanged, plus three interview rulings the
  proposal had surfaced as open questions: the share surface is the existing
  `references.csl.json` research store, not a second `.abcd/work` exchange
  file (one committed bibliography); the "Dogfood (already running)" claim
  was stale and is rewritten as target (no corpus exists on the development
  machine); multi-machine ledger ownership is explicitly deferred. Seven
  Given-When-Then criteria authored and accepted in full; persona corrected
  Maya→Alice per the intents discipline.
- **Verdict:** SPLIT, proposed and confirmed — the initial verdict survived.
- **Notes:** filed as itd-76/spc-31 (planned), itd-126 and itd-127 (drafts,
  seeded criteria marked unconfirmed), adr-41 (proposed), brief invariant 9,
  and the consult-freely-cite-deliberately principle. Both prior share/ingest
  open questions moved to itd-126 with a third (share vs the store's
  admission criteria) added at the walk.

### 2026-08-19 — the itd-92 capability-ladder extension — EXCLUDED from the corpus

- **Excluded from the calibration count:** the human directed the home before
  any table ran ("draft the extension to itd-92"), so there is no blind
  prediction to grade — the label preceded the prediction. Recorded for the
  protocol's history, not as a sample toward the ~50-capture threshold.
- **Proposal:** bring the collaborator fence (built by hand on this repo,
  2026-08-19) to abcd-managed repos gracefully for users without GitHub or
  an organisation — tiers, per-piece verdicts, loud degradation.
- **Initial routing (as first filed):** four parts — doctor + apply + gate
  (capability | itd-92 in place); never-mutate-a-remote-uninvited (stance |
  claimed as carried by loud-staging); verdict mechanics (plumbing | intent
  body); evidence (research | the field note). Two adversarial reviews the
  same day refuted the routing: loud-staging does not carry the
  remote-mutation rule (it is a trust rule, homeless as filed), and a second
  trust rule — identity from caller-local facts, already binding the shipped
  launch scanner (iss-283) — was never routed at all.
- **Corrected routing:** read-only doctor + drift + tier report (capability |
  itd-92, extended in place); apply-on-request (capability | future intent,
  named in itd-92's out-of-scope); both trust rules (ADR |
  [adr-44](../../decisions/adrs/0044-remote-mutation-and-caller-identity-trust-rules.md),
  proposed, brief invariant at adoption); gate refusal policy (deferred to
  its own decision record after doctor field experience — out of itd-92's
  press release entirely); verdict mechanics (plumbing | spec at planning);
  evidence (research | the field note). Record relations: itd-92 will close
  iss-277 when shipped; iss-281 and iss-282 remain open gaps the fence
  tracks; iss-283 is resolved and serves as evidence only. No reversal flags.
- **Verdict:** SPLIT (capability kept in itd-92; trust rules to adr-44; apply
  and gate descoped) — reached only after adversarial review overturned the
  initial FILE-AS-IS. Not a graded sample (see exclusion above), but the
  overturn itself is calibration signal: a pre-directed routing survived
  neither reviewer.
- **Notes:** acceptance criteria rebuilt against what a read-only probe can
  decide per caller; the draft stays in `drafts/` for the planning interview.

### 2026-08-19 — one canonical YAML scalar resolver (itd-128)

- **Proposal:** from the PR #294 commissioned review's F6 — the fix widened
  two of four independent YAML scalar decoders; extract one exported
  resolution helper and have every decoder delegate.
- **Initial routing (proposed before the human ruled):** capability (the
  exported resolver, all four decoders delegating) | intent | itd-128;
  trust rule (ledger-canonical issue store, per-field ownership — decided the
  same session but a distinct part) | decision log now, ADR at graduation;
  stance (fix the class, not the instance) | already carried by the existing
  one-canonical-primitive principle, no new filing; plumbing (where quoting
  normalisation lives, whether the dump-path null list folds in) | intent
  open questions, settled at planning. Typed links: itd-128 closes iss-285;
  iss-287 is evidence. No reversal flags.
- **Confirmed routing:** adopted unchanged (human confirmation 2026-08-19,
  after filing — the prediction preceded the label).
- **Verdict:** FILE-AS-IS — survived confirmation unedited.
- **Notes:** the stance slot resolving to an *existing* principle (no
  artefact filed) is a case the four-piece table handles but the entry format
  under-describes: "home" here is a citation, not a filing. Worth a line in
  the protocol page when the enum question is next revisited.

### 2026-08-20 — `abcd update` in one verb (itd-130)

- **Proposal:** from the update-mechanism investigation — automatic updates
  for abcd across its two install shapes (plugin, standalone CLI), covering
  the self-update verb, plugin-side delivery, Homebrew, and the bootstrap
  script's future.
- **Initial routing (proposed before the human ruled):** capability (the
  `abcd update` verb: fetch/verify/swap, dispatch, refusals, progress) |
  intent | itd-130; trust rules ("never touches a plugin-root binary",
  "never ambient") | citations of existing records, not filings — adr-38
  tier 3 and itd-108's one-cut coherence carry both, no new ADR; stance
  (notify-only-plus-explicit-verb over silent auto-update) | already carried
  by adr-38, no new principle; decision (Homebrew parked until the verb
  exists; install-channel refusal ships with the verb) | dated DECISIONS.md
  entry 2026-08-20; plumbing (bootstrap.sh demoted to cold-start trampoline)
  | itd-130 open question (sequencing + delegation hardening), brief
  internals when it lands; defect discovered en route (stranded PATH symlink
  classifies foreign) | iss-345, fixed test-first the same day, independent
  of the verb. Typed links: itd-130 builds_on itd-105, itd-108. No reversal
  flags.
- **Confirmed routing:** adopted unchanged (human confirmation 2026-08-20,
  in-session, after two adversarial reviews had already reshaped the draft's
  scope — the reviews moved the trampoline from scope commitment to open
  question before the human saw the table).
- **Verdict:** FILE-AS-IS — survived confirmation unedited.
- **Notes:** like itd-128's stance slot, three of the six parts resolved to
  citations of existing records rather than filings; the four-piece table
  keeps working when "home" means "already recorded there". The defect slot
  (a bug found while investigating a capability) is a fifth part-type the
  table absorbs under plumbing-adjacent routing but the protocol page does
  not name; second data point for the enum revisit.

### 2026-08-20 — itd-114 planning interview (re-run after two adversarial reviews)

- **Proposal:** itd-114 (collision-proof record ids) at its planning
  interview, after the two-reviewer prerequisite.
- **Initial routing (the 2026-08-16 table):** FILE-AS-IS with flags — one
  capability, format change deferred to "a future ADR at adoption", stance
  claimed as carried by existing principles.
- **Confirmed routing:** the initial verdict did **not** survive. Both
  reviewers overturned it on the adr-44 precedent: the id format and the
  forge network posture are trust rules extracted to adr-45 + brief
  invariant 11 AT planning; the stance citation was corrected (the claimed
  principle did not exist — it reduces to brief invariant 2 + prefer-sota);
  the forge option was reframed from store to allocator with a typed
  builds_on edge to itd-129, whose ledger-canonical decision the original
  draft had silently reversed. Maintainer confirmed the corrected routing at
  the interview and ruled the four open forks (timestamp-numeric; captures
  first then all families; loud native fallback; detectors kept).
- **Verdict:** SPLIT (capability in itd-114/spc-33; trust rules in
  adr-45 + invariant 11), reviewer-proposed and maintainer-confirmed — the
  initial FILE-AS-IS was overturned. Counts as a graded sample: the 08-16
  table was the blind prediction, the interview the label.
- **Notes:** second consecutive FILE-AS-IS overturned by the two-reviewer
  prerequisite (itd-92 was the first) — the prerequisite is earning its
  cost; the calibration question for the eventual automated pre-pass is
  whether "trust rule parked in intent prose" is mechanically detectable.

### 2026-08-21 — managed-repo identity gate, routine extension (itd-131)

- **Proposal:** from a session finding — ahoy/new-repo setup should establish
  the human git identity so autonomous-routine commits pass the attribution
  gate (a routine defaults to `Claude <noreply@anthropic.com>` and its PR
  reads as unmergeable). Captured as `iss-…367948`
  (ahoy-setup-routine-git-identity).
- **Initial routing (proposed before the human ruled):** capability (guarantee
  the human git identity on every commit incl. routines) | intent — but the
  grill found iss-62 (managed-repo-identity-gate) ALREADY owns this capability,
  so the routing was PROMOTE iss-62, not file-new; the new capture refines it
  (cloud-routine instance). Trust rule ("human is author of record; AI only in
  the trailer) | already a convention (AGENTS.md § Attribution + check-attribution),
  no new ADR. Stance (how AI is acknowledged) | itd-91, cited not duplicated.
  Plumbing (local attribution mirror; lint the pushed tree) | the two sibling
  captures filed the same session. Typed links: new capture refines iss-62;
  intent refines/relates itd-91; adr-44 adjacent.
- **Confirmed routing:** adopted (human confirmation 2026-08-21) — PROMOTE
  iss-62 to itd-131, fold in the routine refinement, cite itd-91.
- **Verdict:** SPLIT/HOLD reached, NOT file-as-is — the decomposition caught
  that a fresh intent would duplicate iss-62. This is the discipline working:
  the grill's "already owned by iss-62" is exactly the miss itd-84 exists to
  prevent, and it landed before an intent was minted, not after.
- **Notes:** third session data point (with itd-128, itd-130) where the grill
  redirected a would-be new intent to an existing owner — twice to a citation
  (itd-128, itd-130), once to a promote (itd-131). Worth a line in the protocol
  page: "promote an existing seed" is a distinct routing outcome from "file new"
  and "cite existing," and the four-piece table should name it.

### 2026-08-21 — hook binary to persistent data dir (itd-132)

- **Proposal:** from the plugin-update post-mortem (session 8db3dbd6) — the
  hook binary lives in the harness's re-cloned, GC'd plugin cache dir, so
  every update re-downloads it, an update-then-quit cancels the SessionEnd
  bootstrap and loses the transcript, and the pinned PATH symlink dangles
  when the orphaned dir is collected. Five ledger captures
  (iss-2608210934566221..225) preceded the routing.
- **Initial routing (proposed, human confirmed same session):** capability
  (binary + .binary-meta relocate to the harness persistent data dir; PATH
  symlink retargets; refresh only on release-tag change) | intent | itd-132,
  absorbing iss-…222; trust rule (persistence must not weaken the spc-21
  verification posture; SessionEnd performs no network work) | ADR + brief
  invariant, drafted alongside the spec; stance (store durable state where
  the platform documents it survives — never fight the harness lifecycle
  inside the cache dir) | principle candidate; plumbing (pluginBinaryPath
  consumers, hooks.json exec order, meta relocation, migration) | spec
  detail; independent defect (SessionEnd bootstraps at exit) | iss-…223,
  fixed test-first on its own branch, no intent; seeds (missed-capture
  recovery sweep, statusline verb) | iss-…224/225 held in the ledger.
  Typed links: itd-132 builds_on itd-105; supersedes spc-21's cache-dir
  fast-path contract (flagged for the human — spc-21's "update into fresh
  cache dir heals by re-fetch" AC is deliberately reversed to "update never
  re-fetches unless the release changed"). No other reversal flags.
- **Confirmed routing:** SPLIT, human-confirmed 2026-08-21 in-session.
  Sequencing caveat for calibration: confirmation preceded the two-reviewer
  prerequisite (reviews launched after, per the interview protocol) — the
  reverse of itd-130's order. The planning interview is the final label;
  relabel here if the reviewers or the interview overturn the table.
- **Verdict:** SPLIT (blind prediction, confirmed pre-review).
- **Notes:** the defect slot recurs (iss-…223 mirrors itd-130's iss-345:
  a bug found en route, fixed test-first, routed outside the intent) —
  third data point for naming that part-type in the protocol enum. The
  reversal flag here is of a shipped spec's acceptance criterion rather
  than an invariant; the table's supersedes link carried it naturally.
  **Interview label (2026-08-21, same day):** the SPLIT table survived the
  two-reviewer prerequisite and the interview unchanged — routing homes and
  typed links held; the reviews instead overturned the *capability's
  internal shape* (relocate-execution → download-cache + per-root copy +
  owned PATH regular file, maintainer-ruled) and forced the record fixes
  (builds_on written, supersession flag recorded, SPLIT parts captured as
  iss-2608210934566226/227 before planning). Counts as a graded sample:
  routing prediction correct; shape prediction wrong. Planned same session
  as itd-132/spc-35.

### 2026-08-21 — the visual-identity proposal (itd-133)

- **Proposal:** one visual identity for abcd — a block-pixel duckling mascot,
  an official logo spelling a-b-c-d in true maritime signal flags, and
  lifeboat art scoped to the lifeboat verbs — rendered on any CLI and as
  HTML/SVG (forge README and web). Live session; the artwork was iterated
  interactively before filing.
- **Initial routing:** three parts — the identity assets and their
  multi-surface rendering (capability | this intent, filed as itd-133); the
  pixel-grid single source of truth with ANSI/SVG generators (plumbing |
  spec body at planning); the role assignment duckling=mascot /
  flags=logo / lifeboat=lifeboat-verbs (intent content, not a separate
  record). Typed link: itd-133 `refines` itd-112 — the banner draft
  explicitly leaves its "small colour logo — an obvious object" open, and
  this supplies it. No reversal flags.
- **Confirmed routing:** adopted unchanged — the maintainer directed the
  filing ("record it as an intent") after choosing the three assets
  in-session.
- **Verdict:** FILE-AS-IS, proposed and confirmed — the initial verdict
  survived confirmation.
- **Notes:** first aesthetic/identity capability in the corpus; the
  stance-shaped part (role assignment) routed *into* the intent rather than
  to a principle because it is the capability's content, not a standing
  rule — a data point for the enum's stance boundary. Confirmation preceded
  the two-reviewer prerequisite (as with itd-132); the planning interview is
  the final label.
  **Interview label (2026-08-21, same day):** the two reviews
  (design/feasibility + record-discipline) overturned the routing in part:
  the terminal-rendering slice (ANSI, colour detection, fallbacks) was
  re-routed *out* of itd-133 to itd-112, which already claimed those
  obligations — so FILE-AS-IS was optimistic; the honest retro-label is
  SPLIT-at-interview. The role-assignment part additionally graduated to a
  decision-log entry (the reviewers' RD4: an "official logo" designation
  needs a durable home beyond the intent), refining the stance-boundary
  data point above. Typed links held but had to be made machine-visible
  (frontmatter `related_intents` written; prose-only `refines` was
  invisible to the lexical pass), and a missed relation to itd-102
  surfaced. A reversal-adjacent flag the initial table missed: itd-133
  forecloses itd-112's object-vs-text-logo open question
  (maintainer-confirmed at interview). Graded: routing prediction partly
  wrong (scope boundary and one link missed); the interview corrected it.
  Planned same session.
### 2026-08-22 — the abcdev.app website proposal

- **Proposal:** "give abcd a website generated from the repository" — the
  full 2026-08-21 site plan (landing page, record explorer, install.sh,
  README migration, `site` verb family, deploy pipeline) arriving as one
  body of investigation.
- **Initial routing table:** capability — the landing page | intent |
  itd-135 (umbrella); capability — the record explorer, split further per
  decompose-before-filing because one press release could not carry both
  the reading surface and the visual chart | intents | itd-136 (record
  pages, contributors, references) + itd-137 (relationship chart,
  genealogy), `builds_on` itd-135; capability — install.sh | intent |
  itd-138, `builds_on` itd-135; capability (gated) — the generic explorer
  on a second instance | intent | itd-139, `builds_on` itd-135/136, held
  in drafts pending its fixture demonstration; trust rules — the
  single-source rule + the adr-30 amendment ("never bundled, rendered
  read-only") | ADR | adr-47; trust rule — deploy-from-tag only | ADR |
  adr-48; stance — the generic/specific boundary (genericity demonstrated,
  never asserted; working-tier crossing by opt-in) | discipline | itd-140;
  plumbing — README migration + the `site` verb family | brief |
  05-internals/10-site.md. Typed links as recorded in the files; no
  reversal flags (the adr-30 boundary change was routed as an amendment
  recorded inside adr-47, not a reversal).
- **Confirmed routing:** the coarse routing (three website intents +
  plumbing-to-brief + two ADRs) was human-dictated in the facilitation
  instructions before the run — the label preceded the prediction for
  those parts. The explorer split (136/137), the discipline, and the gated
  adoption intent follow the two ideate verdicts
  (abcdev-site: survives; record-explorer-generalisation: reframed) and
  await the interview label.
- **Verdict:** SPLIT (eight parts across four record types), part-confirmed
  in advance; the session-added parts pending interview confirmation.
- **Notes:** first hand-run where an ideate reframing minted a discipline
  (the boundary stance came out of the adversarial leg, not the proposal);
  the "capability gated on demonstration" part-type (itd-139) is new to
  the enum — a capability whose filing is confirmed but whose planning is
  explicitly blocked on evidence.


### 2026-08-22 — itd-112 planning interview (post-grill hand-run)

- **Proposal:** itd-112 (the bare-abcd banner), already grilled 2026-08-21
  (scope split to itd-134 confirmed there), at its planning interview after
  two adversarial reviews (design/feasibility and record-discipline).
- **Initial routing:** five parts — the banner capability (this intent);
  the managed-repo generator (already split to itd-134 at the grill); the
  emission-discipline trust rules — TTY-only decoration plus the termsafe
  carve-out, one boundary — (trust rule | one ADR + one brief invariant,
  adr-49 + invariant 13, not two records); the colour ladder and TTY seam
  (plumbing, but a shared primitive with a declared second consumer —
  named in-scope as exported primitives, with itd-110 gaining
  `builds_on: [itd-112]` rather than minting its own record, per
  one-canonical-primitive); the identity bake (plumbing | spec). Typed
  links: itd-112 `builds_on` itd-133, `refines` itd-102; itd-134 carries
  its own `refines` itd-112 (the coined `refined-by` direction was struck —
  not in the enum; the corpus records the pair from the child's side).
- **Confirmed routing:** adopted; the maintainer additionally ruled the
  slug rename (retire-the-name ban on the old slug) and eight design forks
  (identity baked at build time; half-blocks on a painted panel; five-row
  shade-block mono; tagline-below layout; truecolor rung + pinned tables;
  root-local --no-color; dev-build version render; Windows scoped out by
  the release matrix).
- **Verdict:** SPLIT overall (the generator half left at the grill; the
  trust rules left at the interview) — proposed and confirmed in stages
  across the two sessions.
- **Notes:** first hand-run where the reviews moved parts the grill had
  already settled *in prose* into their record homes — evidence for the
  "grill settles, interview routes" division of labour. The
  shared-primitive part (colour ladder) deliberately did NOT get its own
  record: no user moment, so the routing is an in-scope export plus a
  dependency edge — a taxonomy data point for plumbing-with-two-consumers.
