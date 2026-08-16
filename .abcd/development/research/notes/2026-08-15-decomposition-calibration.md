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
