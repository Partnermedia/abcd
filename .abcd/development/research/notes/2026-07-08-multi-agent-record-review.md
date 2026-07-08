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
