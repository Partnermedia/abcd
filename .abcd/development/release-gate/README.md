# Release gate — the pre-tag procedure

Every `v*` release must clear this gate before the tag is pushed. The gate spans
**two enforcement planes**, by design: deterministic checks run in CI (they need
no model); semantic checks run **host-side in the agent harness** (they need a
model, which CI does not have). This runbook is the single human enumeration of
both planes; it owns the semantic half and defers to
[`release.yml`](../../../.github/workflows/release.yml) as the source of truth
for the deterministic half.

Design of record: [`../plans/2026-07-11-iss35-semantic-release-gate.md`](../plans/2026-07-11-iss35-semantic-release-gate.md).

## Deterministic gates (CI-enforced)

The [`release.yml`](../../../.github/workflows/release.yml) `verify` job runs
these, in order, on macOS + Linux. `release.yml` is authoritative; this list is
the human-readable mirror.

1. Format (gofmt)
2. Build
3. Vet
4. Test
5. Test (race, internal)
6. Record-lint (design-record drift gate)
7. Docs-lint (docs-currency gate)
8. Reviews-charter discipline (RD001-RD003)
9. Smoke every command (self-discovering harness)

This list is machine-checked: the `gate_lockstep` `record-lint` rule blocks if it
diverges from `release.yml`'s `verify` job steps (setup steps excepted). Edit both
together — the mirror cannot silently drift.

## Semantic gates (host-run, before the tag)

CI cannot run these — they spawn LLM agents. Run them in the agent harness
against the exact commit to be tagged:

1. **`docs-currency-reviewer`** — verifies every user-facing claim still matches
   the code (the semantic complement of `docs lint`; see
   [`../brief/04-surfaces/10-docs.md`](../brief/04-surfaces/10-docs.md)).
2. **Brief↔surface cross-check** — [`brief-surface-crosscheck.js`](brief-surface-crosscheck.js),
   the Direction-A semantic half of the iss-35 graduation: the brief's surface
   *prose* (flags, sub-verbs, exit codes, schema fields, counts) vs. the shipped
   binary's actual behaviour. The deterministic Direction-B half is the
   `surface_coverage` `record-lint` rule and already runs in CI. Its scope and
   depth are pinned by [`manifest.json`](manifest.json) — see *Pinned inputs and
   tiered depth* below.

## Recording the semantic verdict

**A receipt names the commit its reviewer READ, and lives in a LATER commit** —
it can never sit in the tree of the commit it names, because adding it would
change that commit's sha. Under changelog-driven auto-release (adr-37) the
release branch is therefore exactly two commits: the CHANGELOG roll (the
release-content commit the reviewers read), then the semantic receipts naming
it. On merge, the released tree carries the receipts, and `release.yml` arms the
gate with the *content* commit (`<merge>^2^` for an auto-release merge, `<tag>^`
for a manual tag on the receipts commit) — not the merge commit — so
`subject.digest.gitCommit` still matches the armed commit exactly and the gate
stays strict. (Before this, the gate armed with the tagged merge commit, whose
tree can never hold a receipt naming itself — an unsatisfiable self-reference
that fail-closed every public release; it was dormant until the public flip, so
it was never exercised.)

Each semantic pass records its outcome as a **commit-sha-keyed receipt** — a
Verification Summary Attestation (VSA) shape carrying `verificationResult`
(PROMOTE / HOLD), the pinned judge-model snapshot, the detector version, and the
failing categories. Receipts live at `.abcd/work/reviews/<commit-sha>/<gate>.json`;
[`receipt.example.json`](receipt.example.json) is the concrete shape. The
`receipt_gate` `record-lint` rule verifies, before a tag, that each required gate
has a PROMOTE receipt whose subject digest is the target commit, whose
`policy.detector` names that gate, and which pins a judge model; a missing,
mismatched, malformed, HOLD, model-less, or wrong-detector receipt **blocks** the
release (fail-closed — an un-run semantic pass is never a silent pass).

A receipt is bound to its gate by its `policy.detector` value, not by its
filename: the `<gate>.json` at `.abcd/work/reviews/<sha>/` must carry
`policy.detector` equal to `<gate>`. This stops one genuine PROMOTE receipt from
being copied across every gate's path to satisfy them all — each gate needs its
own receipt from its own detector.

## Pinned inputs and tiered depth

The cross-check runs against a **committed input manifest**,
[`manifest.json`](manifest.json), which pins the reproducibility inputs: the
22-document brief-doc list, the two directions (A: brief→surface, B:
surface→brief), the checker count (27 = 22 brief docs + 5 surfaces), and the
prompt (context plus both direction templates, with the prompt's own
`sha256`). Two honest runs of the same tier therefore mean the same thing —
the maintainer no longer chooses the scope per run. The
[`brief-surface-crosscheck.js`](brief-surface-crosscheck.js) detector consumes
this manifest as its input rather than composing an ad-hoc list.

Depth is **tiered by the release's impact class**:

- **`full`** — both directions over the whole brief-doc list — is required for a
  **feature (additive) or breaking** release.
- **`shallow`** — Direction B only — is sufficient for a **patch (fix/internal)**
  release.

The impact class is derived at gate time from the shipped records — the same
signal the version itself derives from (changelog-driven versioning, adr-37).
The version *number* alone cannot carry it while abcd is pre-1.0: an additive
feature and a fix both bump the patch component at 0.x, so only the records'
`impact` frontmatter tells a feature release apart from a patch one. The gate
reads the strongest impact in the cut between the newest release tag older than
the current CHANGELOG version and `HEAD`.

A manifest-era receipt therefore carries two further fields alongside the ones
above:

- **`manifestHash`** — the `sha256` of `manifest.json` the run pinned its inputs
  against.
- **`tier`** — `full` or `shallow`, the depth the run used.

`receipt_gate` adds three **procedural** refusals once the manifest exists in the
release's content tree:

1. `manifestHash` does not equal `manifest.json`'s actual hash — the run's pinned
   inputs are not this release's.
2. `tier` is insufficient for the release's impact class — a shallow receipt on a
   feature or breaking release.
3. A `failing` entry carries no `disposition` — an un-triaged finding.

None of these judges a finding's **content or severity**. Confirmed findings
route to the maintainer, whose PROMOTE with recorded dispositions is the gate
(verifier-selects-gates-decide); the gate does not hard-block on a confirmed
major and keeps no never-worse ratchet across tiers.

**Era gating.** The manifest's presence in the armed content tree is the era
marker. The three refusals above arm only when `manifest.json` is present at the
gate's armed commit, so receipts written before the manifest existed — the
receipts committed under `.abcd/work/reviews/` for earlier releases — stay valid
for their own commits, judged by the checks that predate the manifest.

The `receipt_gate` rule is **disabled by default** — it must never fire on
ordinary PRs/pushes, only at release time — and is armed by `release.yml`, which
supplies the tagged commit and the required-gate list from the workflow (the
trust root), not the in-tree config: `record-lint --release-gate <sha>
--require-gate <name>…`. `release.yml` then signs the receipts with
`actions/attest` (predicate `.../semantic-release-gate/v1`) and verifies the
attestation with `gh attestation verify` — no new dependency (the same attest
family + `gh` the binary provenance already uses).

> **Dormant until the public flip.** Artifact attestation is a public-repo
> feature, so the whole gate is gated `if: !github.event.repository.private` —
> exactly like the binary attestation — and does nothing on a private repo,
> activating on the public flip. And the signature is **auditable release
> provenance, not committer-forgery-proof**: a receipt forged and committed
> before the tag would be signed too; that residual is bounded by the iss-62
> identity gate + branch protection. Stated per
> [`../principles/loud-staging.md`](../principles/loud-staging.md), not implied
> away. Full forgery-prevention would need host-side signing at receipt
> production — a later step if the threat model warrants it.

## Procedure

1. Land all work on the release commit; open the release the normal way (branch
   → PR → merge). The `verify` job gates the merge.
2. On the merged commit, run the two semantic gates above in the harness.
3. Record each verdict as a receipt keyed to that commit's sha.
4. Tag `vX.Y.Z` on the commit. Once the fail-closed verify rule is armed, the
   tag is rejected unless every semantic receipt is present and PROMOTE.
