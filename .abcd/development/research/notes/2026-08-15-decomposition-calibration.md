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
