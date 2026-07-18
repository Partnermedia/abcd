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
   `surface_coverage` `record-lint` rule and already runs in CI.

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
