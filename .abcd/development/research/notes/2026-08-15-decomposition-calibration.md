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
