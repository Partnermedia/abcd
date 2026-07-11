# iss-35 semantic release gate — SOTA-verified design (Option C, task 5)

**Status:** design recorded 2026-07-11, maintainer signed off "full SOTA incl.
signing". Reached via the [prefer-sota](../principles/prefer-sota.md) process:
an adversary challenged generic release-engineering SOTA for repo fit, then a
`sota-researcher` verdict was taken on the fit-surviving hypothesis. This doc is
the design of record for the semantic (Direction-A) half of the iss-35
graduation — the deterministic half shipped as the `surface_coverage` record-lint
rule.

## The decision

Wire the LLM brief↔surface cross-check (the semantic pass that verifies brief
*prose* still matches *binary behaviour* — flags, exit codes, schema fields,
counts) as a standing, enforced pre-tag release gate. It cannot run in CI
(GitHub Actions has no model access); the repo's enforcement plane is
deliberately split — deterministic gates in CI (`release.yml`), semantic gates
in the host/preflight environment where a model is present.

## Why this form (adversary → SOTA verdict)

The adversary rejected, each for a named repo preference:

- **policy-as-code (OPA/Rego)** — a new language + runtime (ask-first) and a
  second lint engine beside the one Go engine the repo paid to consolidate onto.
- **release-please / semantic-release** — bot commits to a branch, which
  `release.yml`'s no-branch-commit tripwire fails closed on; and it duplicates
  the adr-31 derived-versioning engine.
- **LLM-judge-in-CI** — violates host-delegated-by-default, needs a model key in
  a tag-triggered workflow (trust-boundary expansion), and is a non-convergent
  (flaky) blocking gate.
- **a naive `make release-gate`** — manufactures a false green: deterministic
  gates pass, the model gate silently no-ops, exit 0.

The `sota-researcher` verdict on the fit-surviving hypothesis: it is a
**Verification Summary Attestation (VSA)** — a named SLSA pattern. The VSA spec
*explicitly* decouples the verifier from the builder and sanctions verification
"at build time, at upload time, at download time, or via continuous monitoring"
by "a separate trusted party" — a host-sited verifier emitting a receipt that
release-time *checks* is the canonical VSA topology, not a workaround. The
closest exemplar is SLSA AMPEL's `prerelease` policy (validates prior VSAs
rather than re-running checks). So the design is SOTA-aligned; the only delta
from full SOTA is signing the receipt for unforgeability — which this repo
already has the primitive for (it runs SLSA build-provenance attestation in
`release.yml`).

## The design

Four parts, from the verdict's ranked recommendations (R1–R4):

1. **Semantic-pass receipt = a VSA-shaped predicate** (R1, R3). A JSON record
   keyed to the release commit digest, using the VSA field shape so it is
   recognisable and upgradeable: `subject.digest` (the commit sha),
   `verifier` (identity), `timeVerified`, `verificationResult` (PROMOTE/HOLD),
   `policy`/detector version, plus the eval-standard fields — pinned
   **judge-model snapshot** (e.g. `claude-opus-4-8`, never a floating alias),
   judge-prompt hash, brief version, per-category result, failing-example ids.
2. **Fail-closed verification before tag** (R1, R4). A Go `record-lint` rule
   (the existing engine, zero new deps) asserts a receipt exists for `HEAD`'s sha
   **with `verificationResult: PROMOTE`**; absent or HOLD ⇒ BLOCK. The *gate* is
   deterministic (receipt present + verdict) even though the *judge* that
   produced it is stochastic — the stochastic judgement stays upstream in the
   host, never re-derived at tag time.
3. **Runbook↔`release.yml` lockstep** (R2). The dev-facing pre-tag runbook and
   the CI workflow must both stay human-authored (neither is a clean projection
   of the other), so a consistency lint fails when their gate lists diverge —
   the repo's golden-check anti-drift pattern (`gofmt -l`, `go generate` +
   `git diff --exit-code`).
4. **Signed receipt (unforgeability)** (R1 caveat, maintainer chose full SOTA).
   A committed sha-keyed file is trust-on-write — any committer can forge a
   PROMOTE. Sign the receipt with the SLSA attestor already used in
   `release.yml` (Sigstore/cosign-class), and verify the signature in-Go at gate
   time — never with a Rego/CUE policy (that re-imports the rejected
   new-language cost).

## Homes (host-agnostic, dev-tier)

- **Detector** (the LLM cross-check workflow, host-run, never CI): committed as
  a dev-record artifact under `.abcd/development/` (excluded from the release
  artifact), sanitised first — drop the hardcoded absolute local path
  (privacy rule) and the stale skills→commands ground-truth (abcd ships zero
  skills). It is invoked host-side via the agent harness, not `make`/CI.
- **Receipts**: sha-keyed, under a committed work-tier path (candidate
  `.abcd/work/reviews/`), subject to the RD001–003 reviews charter.
- **Runbook**: a maintainer how-to **outside `docs/`** (developer-facing →
  `.abcd/development/`), host-agnostic (docs-lint applies), referencing
  `release.yml` as the source of truth for the deterministic half and owning
  only the semantic gates.

## Phased build (each phase its own gate)

1. Sanitise + commit the detector to its dev-tier home; author the runbook
   (owns semantic gates, references CI for the rest). No enforcement yet — mark
   the not-yet-armed enforcement per [loud-staging](../principles/loud-staging.md).
2. Receipt schema (VSA-shaped, R3 fields) + the fail-closed `record-lint`
   verification rule (TDD, mirror `surface_coverage`); arm it and watch it fire
   on a missing/HOLD receipt.
3. Runbook↔`release.yml` lockstep consistency lint (TDD).
4. Receipt signing + in-Go signature verification; wire the release-time check
   into `release.yml`.

**Trust boundary:** phases 2–4 touch signing, attestation verification, and the
release workflow — the `security-reviewer` agent runs before presenting, and a
BLOCK verdict stops the change (per AGENTS.md).

## What this is NOT

Not an in-CI LLM judge, not policy-as-code, not a release-automation tool, not a
`make` target that re-runs the judge. The judge runs once, host-side; the gate
verifies its signed evidence.
