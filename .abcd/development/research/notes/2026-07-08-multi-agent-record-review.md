# Multi-agent record review — 2026-07-08

A full-repo review run as an orchestrated multi-agent workflow (58 agents):
five code/docs review dimensions whose findings were each adversarially
verified by an independent refuter; three structure reviews (the development
record, the brief, the enforcement layer); state-of-the-art research on
in-repo development records; and best-practice extraction from two sibling
projects. Of 46 raw code/docs findings, 3 were refuted and dropped, 43
confirmed; the structure reviews produced 41 findings synthesised into 22
improvements.

This note is the durable distillation. The full evidence bundle (per-finding
verifier reasoning) is a runtime artefact and intentionally not committed;
the actionable content lives in the issue ledger and the principles named
below.

## Recording method: fix the detector

Findings were NOT filed one-per-defect. Per the
[fix-the-detector](../../principles/fix-the-detector.md) principle distilled
from this review, findings sharing a root cause were clustered into one
ledger issue that records three things: the instances, the proposed
*detector* (the gate, lint rule, or test convention that catches the class),
and the instances as the detector's acceptance corpus. Cleanup drains behind
the armed detector, never ahead of it.

## Cluster map (iss-29 .. iss-49)

Defects and test gaps:

- **iss-29 fail-closed-capture-surface** — typo'd mutating sub-verb files an
  issue; non-JSON `--json` errors. Detector: malformed-input test convention
  per mutating verb.
- **iss-30 memory-ingest-boundary** — unchecked HTTP status, tilde mangling,
  partial-failure misreport, CRLF parser split, untested success path.
  Detector: ingest-boundary test suite.
- **iss-31 launch-dogfood-gate** — identity-scanner false positive blocks the
  repo's own gates; payload omissions; resolver cache race. Detector: CI runs
  the launch dry-run on this repo.
- **iss-32 atomic-write-consolidation** — four divergent durable-write
  copies, two crash-unsafe. Detector: canonical-name lint + fsutil test
  suite.
- **iss-33 ahoy-verb-hygiene** — unvalidated prompt persistence, swallowed
  registration errors, untested read-only verbs, silent dead scaffolding.
  Detector: wired-verb behavioural test convention + coverage-plus-caller
  audit.
- **iss-34 untested-refusal-guards** — symlink/deny, quotation-budget,
  licence-detection guards at zero coverage. Detector: every refusal path
  ships a refusing test.

Drift and spec:

- **iss-35 brief-surface-reconciliation** (critical) — the brief asserts a
  surface the shipped code falsifies; live verbs have no spec home.
  Detector: commands/skills ↔ brief-row cross-check in record-lint.
- **iss-36 retired-name-banlist** (critical) — ~50 stale references that
  survived a dedicated consistency pass. Detector: `banned_tokens` entries
  per retired name; drain behind the bans.
- **iss-37 phantom-enforcement-claims** — five described gates that do not
  run. Detector: named gates resolve to live definitions.
- **iss-38 hand-maintained-index-drift** — four stale hand-kept indexes.
  Detector: enumerate-by-hand lint; derive or delete.
- **iss-39 record-schema-validation** — dangling ADR handles, one-way
  supersession. Detector: mechanical schema check (handles resolve,
  supersession bidirectional, lifecycle catch-all).
- **iss-40 glossary-unification** — two competing registries, stale index.
  Detector: single-registry glossary lint.
- **iss-41 lifecycle-delivery-state** — shipped capability unrepresentable in
  the lifecycle. Detector: CHANGELOG ↔ intent-stage cross-check.
- **iss-42 record-orientation-currency** — stale CONTEXT.md trust caveat.
  Detector: handoff checklist rule.
- **iss-43 readme-capability-currency** — README claims unshipped
  capabilities as native defaults. Detector: docs-currency scope extension.
- **iss-44 plugin-surface-parity** — CLI sub-verbs unreachable from the
  plugin surface. Detector: surface-parity check.
- **iss-45 surface-layer-boundary** — business logic in the CLI surface;
  duplicate front doors; skills bypassing the binary. Detector:
  boundary check; converges with iss-27.
- **iss-46 lint-scope-holes** — coverage matrix gaps, no baseline ratchet.
  Detector: lint scope matrix + ratchet support.

Seeds and residue: **iss-47** generated CLI reference with drift test,
**iss-48** behavioural e2e scenarios, **iss-49** cosmetics batch (fixed
directly per the one-off bound).

## Principles recorded from this review

Ten files under [`principles/`](../../principles/): fix-the-detector,
enforcement-claims-are-facts, retire-the-name, ratchet-not-big-bang,
one-canonical-primitive, guards-prove-themselves,
unrecognized-input-never-writes, spec-moves-with-the-surface, loud-staging,
reality-is-filable. Three further findings sharpen existing doctrine rather
than adding to it: slug derivation in the surface violates the ratified
transport-agnostic-core boundary; hand-kept indexes are adr-5 lacking a gate;
surface parity is the standing wired-or-it-is-not-done rule, refined by
loud-staging.

## Independent cross-check (same day)

As a test of a second review channel, the brief and record were re-reviewed
by an independent single-pass oracle over a workspace-context selection
(~195k tokens: full brief, decisions/ADRs, index READMEs, principles,
CONTEXT.md), deliberately blind to the findings above. Outcome:

- **Corroboration.** Every structural cluster above that fell inside its
  selection was independently re-found: the retired issue-ledger paths
  (iss-36), the adr-12/adr-32 supersession gap (iss-39; the oracle rates it
  critical where the review said major), the stale intents/README listings
  and out-of-scope drift (iss-38, iss-41), the zero-skills and six-commands
  claims (iss-35), the competing glossaries and the glossary's own stale
  index (iss-40), and the stale CONTEXT.md trust caveat (iss-42). No
  confirmed finding was contradicted.
- **New instances, verified against the filesystem and added to the
  acceptance corpora of existing clusters** — notably, every one of them
  falls inside an already-captured detector class, which is itself evidence
  for the fix-the-detector recording model:
  - "three disciplines (itd-1, itd-5, itd-37)" in 01-product/04-scope.md and
    03-mental-model.md, while disciplines/ holds four (itd-79) → iss-38.
  - development/README.md "issues graduate … rather than a ledger" vs
    ADR-32's ledger → iss-36/iss-38.
  - glossary/core/brief.md defines the brief as "immutable once approved"
    vs ADR-5's living record → iss-40.
  - adr-7 cites a retired terminology home and carries no ADR frontmatter
    → iss-39 and iss-36.
  - 04-surfaces/06-capture.md marks behaviours LIVE via predecessor spc-N
    provenance, against the brief README's own provenance rule → iss-37.
  - 04-surfaces/04-launch.md's release include-list names agents/ and
    hooks/, which do not exist → iss-37/iss-31 corpus.
  - 04-surfaces/09-reflect.md references commands/abcd/reflect.md and an
    agents/ roster file, neither of which exists → iss-35 (unshipped
    surface without staging status).
  - 05-internals/07-memory.md "agent count stays at 15" vs the 16-agent
    roster in 01-agents.md → iss-38 (hand-maintained counts).
- **One refuted claim.** The oracle's other critical — itd-49 present in
  both planned/ and superseded/ simultaneously — is false; itd-49 exists
  only in superseded/. Single-pass oracle output requires filesystem
  verification before recording: of its two criticals, one was real and
  pre-known, one was fabricated.

Channel verdict: useful as a cheap corpus-expander and corroborator on a
large single-context selection; not trustworthy unverified, and it found no
new *class* — the clustered detectors above already cover everything it
surfaced.

## Sibling-project practices worth adopting

From a sibling Go CLI project: baseline-diff ratchet lints for architectural
invariants; generated CLI docs gated by a drift test inside the normal test
run; module boundaries enforced twice (dependency rules with reasons, plus
in-tree regression tests); a canonical-primitives table and a "what NOT to
build" section in every plan; ADRs that name their load-bearing invariant.
Its cautionary tale: root-level markdown sprawl from ungated working notes —
validation of the tiered `.abcd/` layout, with enforcement as the gap.

From a sibling multi-language project: non-negotiable invariants enumerated
in the agent front door, each paired with a test and a mandatory review
trigger; a periodically re-verified half-done-features register (its audit
found ten built-but-unreachable features — the same species as this repo's
staged ship path); a fail-closed release preflight that refuses its own
escape hatches; schema-first cross-language contracts with parity tests as
the enforcement.

## State of the art, in brief

The conventions this record already matches: MADR-style ADRs, Diátaxis for
user docs, docs-as-code with lint gates, an agent front door (AGENTS.md).
Where SOTA is ahead of this repo: teams automate terminology and name bans
(prose linters with per-term rules), link integrity across all committed
markdown, and freshness gates that regenerate-and-diff rather than trust;
spec-driven flows keep the spec in the same change as the surface. All four
map onto issues above (iss-36, iss-46, iss-47, iss-35).
