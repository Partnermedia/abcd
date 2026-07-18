---
id: itd-93
slug: abcd-scaffolds-a-hardened-changelog-driven-release-gate-into
spec_id: null
kind: null
suggested_kind: null
reclassification_history: []
builds_on: []
related_adrs: [adr-37]
severity: minor
---

# abcd Scaffolds a Release Gate That Works on the First Try

## Press Release

> _Facilitator-seeded draft — the product thinker owns the final press-release
> framing._

> **A repo abcd manages gets a release process that is correct the day it goes
> public — no self-inflicted first-release failure.** Ask abcd to set up
> releases and it lands a changelog-driven release gate: rolling `[Unreleased]`
> into a dated version in a reviewed PR is the release decision, and on merge the
> automation tags exactly that commit and publishes. The gate that verifies the
> release is armed against the *reviewed content commit*, so the very first
> public release cannot hit the receipt-vs-tag self-reference that once blocked
> abcd's own.
>
> "I flipped my repo public and cut a release the same afternoon — it just
> worked," said Alice, a solo founder. "I didn't have to discover, the hard
> way, that my release gate could never be satisfied. abcd gave me the version
> abcd itself only reached after a day of untangling."

## Why This Matters

When abcd-cli went public and cut its first release, the semantic release gate
— which had never run end-to-end while the repo was private — **fail-closed by
construction**: it armed against the tagged commit and read the reviewer
receipts from that commit's own tree, but a receipt names the commit its
reviewer read and can only live in a *later* commit, so a receipt naming the
tagged commit can never sit in the tagged commit's tree. The fix
([PR #99](https://github.com/REPPL/abcd-cli/pull/99), recorded in
[iss-108]) was to arm against the reviewed *content* commit (`HEAD^2^` on the
auto-release merge path, `HEAD^` on a direct tag) and to structure the release
branch as two commits — the changelog roll, then the receipts naming it.

That flaw was abcd-cli's own CI, and it did **not** reach managed repos —
abcd ships and scaffolds no release workflow today ([iss-108] verified this:
`launch-payload.json` excludes `.github/`, ahoy/launch write no CI). But that is
exactly the gap this intent closes: abcd is a *configuration layer for
development*, and "how you cut a correct, gated release" is configuration a
managed repo should be able to inherit rather than re-derive — and re-derive
*with the same latent self-reference bug baked in*. The lesson abcd paid for
once should ship as a working default, not a trap every managed repo rediscovers
at its first public release.

## What's In Scope

- **A scaffold path** (verb TBD — see Open Questions) that writes, into a
  managed repo that lacks them, the fixed release machinery: a `release.yml`
  (verify → build → publish, gate armed against the reviewed content commit) and
  an `auto-release.yml` (detect newest dated CHANGELOG version → tag that commit
  → call `release.yml`), both `GITHUB_TOKEN`-only and injection-safe.
- **The adr-37 policy, carried as a per-repo runbook**: the CHANGELOG is the
  release instrument; rolling `[Unreleased]` → `## [X.Y.Z] - <date>` in a
  reviewed PR is the release decision; the two-commit release-branch shape
  (roll → receipts) is documented so `HEAD^2^` resolution holds.
- **The receipt/charter interop already fixed in abcd-cli**: the sha-keyed
  receipt-dir convention plus the `check-reviews` (RD001) exemption, so the two
  in-repo review conventions do not collide.
- **Wiring to the repo's own facts**: the required-status-check contexts and the
  release-gate's required detectors are derived from the target repo's actual CI
  job names / configured gates, not hard-coded to abcd-cli's.
- **Idempotent + fail-safe scaffolding**: re-running is a no-op when the
  machinery is current; it never overwrites a workflow the operator hand-edited
  without a transparent-confirm; it refuses rather than half-writing.

## What's Out of Scope

- **The semantic detectors themselves** (`docs-currency-reviewer`,
  `brief↔surface cross-check`) — they are host-run LLM passes, not scaffolded CI;
  a managed repo opts into them (or runs none) and the gate's required-detector
  list reflects that. Scaffolding must degrade cleanly to the deterministic
  gates alone when no semantic detector is configured.
- **Signing/attestation infrastructure** beyond what the built-in
  `GITHUB_TOKEN` + `actions/attest` already provide (no new dependency, no PAT).
- **Non-GitHub forges** — this scaffold targets GitHub Actions; other CI hosts
  are a later concern.
- **Choosing the version number** — that remains adr-31/itd-73 (derived from
  intent impact); this intent scaffolds the *cutting and gating* machinery.

## Acceptance Criteria

> _BDD format, per the itd-1 discipline. Facilitator-seeded from the abcd-cli
> fix; the product thinker should confirm the bar._

- **Given** a managed GitHub repo with no release workflow, **when** the
  operator runs the release-scaffold verb, **then** `release.yml`,
  `auto-release.yml`, and the release runbook are written, wired to the repo's
  own CI check names, and pass the repo's workflow audit (e.g. zizmor) with no
  injection or duplicate-key findings.
- **Given** a repo with the scaffolded gate, **when** it cuts its first public
  release (roll `[Unreleased]` → dated heading, merge), **then** the release
  publishes — the gate is armed against the reviewed content commit and does
  **not** hit the receipt-vs-tag self-reference (a test exercises the merge path
  and asserts a published release, not a fail-closed gate).
- **Given** a repo that configures **no** semantic detector, **when** it
  releases, **then** the deterministic gates alone admit the release and the
  receipt gate requires nothing (no host-run pass is silently treated as
  missing).
- **Given** the machinery is already present and current, **when** the scaffold
  verb runs again, **then** it reports a no-op and mutates nothing; **and given**
  an operator hand-edited a scaffolded workflow, **when** the verb runs, **then**
  it refuses or transparent-confirms rather than clobbering.
- **Given** the scaffolded `check-reviews`/RD001 charter, **when** sha-keyed
  receipt directories exist, **then** they are exempt from the dated-review-dir
  shape (the abcd-cli collision does not recur in the managed repo).

## Open Questions

- **Which surface scaffolds it?** An `ahoy install` step, a new
  `abcd launch setup` / `abcd release init` sub-verb, or an `embark`-time
  record family? Leans toward an explicit, opt-in verb (release CI is a
  deliberate, outward-facing addition, not part of bare install).
- **How much is templated vs. copied?** A verbatim copy of abcd-cli's fixed
  workflows risks drift as abcd's own evolve; a template needs per-repo
  substitution (check-name list, module path, Go version). Prefer a template
  with a lockstep test against abcd-cli's own workflows so the shipped pattern
  can't silently diverge from the proven one.
- **Private→public activation.** abcd-cli's gate was dormant while private and
  activated on the public flip — which is *how the flaw stayed hidden*. Should
  the scaffolded gate be exercisable while private (e.g. a dry-run mode) so a
  managed repo discovers problems before going public?
- **Relationship to itd-73** (derived versioning) — the scaffold should leave a
  clean seam for the derived-version number to feed the CHANGELOG roll.

## References

- [adr-37](../../decisions/adrs/0037-changelog-driven-releases.md) — the
  changelog-driven release policy this scaffolds.
- iss-108 — the self-reference flaw, its abcd-cli fix (PR #99), and the verified
  finding that no release machinery currently reaches managed repos.
- `.abcd/development/release-gate/` — abcd-cli's own runbook + detectors, the
  proven pattern to generalise.
