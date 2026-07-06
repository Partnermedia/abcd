---
id: adr-6
slug: rp-review-storage-and-architecture
status: superseded
date: 2026-05-10
supersedes: null
superseded_by: adr-25
related_intents: [itd-28]
related_rfcs: []
related_adrs: [adr-5]
---

# ADR-6: RP Review Storage and Architecture

> Superseded by [ADR-25](0025-host-delegated-llm-default.md) — oracle capture is
> now host-delegated, and the redaction model is salvaged into
> [ADR-29](0029-native-transcript-corpus.md).

> **Terminology note.** The *how* layer is named the **spec**. This ADR's prose
> was updated by the spec-terminology-rename ADR
> ([adr-11](0011-spec-terminology-rename.md)).

## Context

### Why a Storage Decision Is Needed

The `fn-2` spec introduces automatic capture of RP and Codex review artifacts (plan-review, impl-review, completion-review) into the repository. Where those artifacts land, how much of them is committed, and what stays gitignored are decisions with long-term consequences for repo health, portability, and security.

### Why an Architecture Decision Is Needed

The capability described above requires integrating with flow-next (the orchestration layer that drives reviews) without modifying flow-next itself. Three integration approaches were considered:

- **Upstream PR**: Submit a PR to `gmickel/flow-next` to add native storage.
- **Downstream wrapper (abcd-side)**: Intercept review output via hooks and post-process it in abcd's own scripts.
- **Stand-alone tool**: A separate CLI that users run manually after reviews.

This ADR documents both the storage decision and the architectural decision together because they are coupled: the architecture determines *how* artifacts are captured, which constrains the storage model.

### Prior Art Consulted

- **F-075** (`REPPL/abcdZero/docs/development/roadmap/features/planned/F-075-flow-next-local-review.md`) — explicit evaluation of upstream-vs-wrapper for a prior review-storage capability; concluded wrapper. That decision was provisional; this ADR revisits with fn-2's concrete requirements.
- **F-037** (`REPPL/abcdZero/docs/development/roadmap/features/completed/F-037-multi-model-review.md`) — v2.5.1 shipped a patch-based wrapper for multi-model reviews; no upstream changes required.

---

## Decisions

### Decision 1: Hybrid Commit/Ignore Storage Model

**Commit**: structured JSON sidecar (`review.json`) and rendered Markdown view (`review.md`), bounded by `body_max_bytes` (default 20 KiB) and `render_max_bytes` (default 24 KiB).

**Gitignore**: raw transcript files (full RP chat export / Codex stdout). Raw transcripts land in `.flow/reviews/<spec-id>/<NNNN>-<slug>-<ref>/raw/<sha>.json` (per-review subdirectory) and are listed in `.gitignore` under `.flow/reviews/**/<NNNN>-*/raw/`.

**Per-review directory structure**:
```
.flow/reviews/<spec-id>/
├── INDEX.md          — human-readable index (generated)
├── INDEX.json        — machine-readable index (generated)
└── <NNNN>-<slug>-<ref>/
    ├── review.json   — canonical JSON sidecar (committed)
    ├── review.md     — rendered Markdown view (committed)
    └── raw/          — raw transcript / oracle export (gitignored)
        └── <sha>.json
```

**Alternatives considered**:

| Alternative | Why Rejected |
|---|---|
| Commit everything (raw transcripts included) | Repo bloat; raw transcripts may contain PII, home paths, or credential-adjacent context that survives Stage 1 sanitisation. |
| Gitignore everything | Defeats portability goal — the point of fn-2 is to make review history a first-class repo artifact. |
| MD as canonical record | Conflicts with abcd's JSON-internal / MD-render invariant (brief § 5). JSON is machine-readable; MD is derived. |

### Decision 2: abcd-Side Wrapper Architecture

**Implement** the capture layer inside abcd's own scripts (`scripts/abcd/`), using Claude Code's `Stop` / `SubagentStop` hook as the interception point. **Do not** submit an upstream PR to `gmickel/flow-next`.

**Rationale**:

1. **flow-next has no plugin SDK.** There is no documented extension point for "write these artifacts to disk." Any upstream change would require modifying internal flow-next code.

2. **Receipt JSON is the established integration contract.** flow-next already emits receipt JSON (impl-review, plan-review, completion-review receipts in `scripts/ralph/runs/.../receipts/`). abcd reads this contract. No new contract is needed.

3. **Stop/SubagentStop hook precedent.** `gmickel/flow-next/plugins/flow-next/hooks/hooks.json` already uses a Stop hook for Ralph's post-task processing. The hook pattern is canonical in this repo.

4. **Upstream rejection risk is concrete.** PR #65 (MCP backend) was open for 3.5+ months without merge. PR #86 (Copilot integration) was superseded after 3 months. A PR adding storage hooks would stall the fn-2 timeline and might be rejected on scope grounds. The wrapper avoids this risk entirely.

5. **F-075/F-037 prior art.** Both prior evaluations reached the same conclusion. This ADR adopts their reasoning for fn-2's concrete requirements.

---

## Two-Stage Redaction Model

Review artifacts may contain credentials, home-directory paths, or other sensitive material that arrived in the reviewed text. Two stages of protection apply:

### Stage 1: Write-Time Sanitiser

`scripts/abcd/_review_lib.py::sanitise_text()` runs *before* any file is written to disk or staged. It applies pattern-based replacements:

- AWS access and secret keys → `[REDACTED-aws_access_key]` / `[REDACTED-aws_secret_key]`
- Anthropic API keys (`sk-ant-...`) → `[REDACTED-anthropic_key]`
- OpenAI API keys (`sk-...`) → `[REDACTED-openai_key]`
- GitHub PATs and fine-grained tokens → `[REDACTED-github_token]`
- GCP service account private keys and API keys (`AIza...`) → `[REDACTED-gcp_service_account_key]` / `[REDACTED-gcp_api_key]`
- Azure storage account keys and client secrets → `[REDACTED-azure_storage_key]` / `[REDACTED-azure_client_secret]`
- Cloudflare API tokens → `[REDACTED-cloudflare_token]`
- Stripe secret and restricted keys (`sk_live_...`, `rk_test_...`) → `[REDACTED-stripe_secret_key]` / `[REDACTED-stripe_restricted_key]`
- Slack bot/app tokens (`xox*-...`) and webhook paths → `[REDACTED-slack_token]` / `[REDACTED-slack_webhook]`
- JWT tokens (three base64url segments) → `[REDACTED-jwt_token]`
- PEM private key blocks (`-----BEGIN ... PRIVATE KEY-----`) → `[REDACTED-pem_private_key]`
- Bearer tokens → `[REDACTED-bearer_token]`
- Home-directory paths (`/Users/<name>/`, `/home/<name>/`) → `~/`

The sanitiser is intentionally conservative: only patterns it is highly confident about are rewritten. Unknown secrets survive Stage 1 and are caught by Stage 2.

### Stage 2: Pre-Commit gitleaks Verifier

`scripts/abcd/hooks/verify_reviews.py` runs in the pre-commit hook context. It invokes:

```
gitleaks protect --staged --redact=100 --config=.gitleaks.toml
```

This is a detect-and-block gate: if gitleaks finds any remaining secrets in staged `.flow/reviews/` files, the commit is blocked. Output policy: `--redact=100` ensures gitleaks never prints raw secret bytes; this script surfaces only `<path>:<line>:<rule-id>` triples.

### Why Single-Stage gitleaks-as-Rewriter Is Unimplementable

`gitleaks --redact=75` (or any redact level) redacts the **output log**, not the staged files themselves. gitleaks is a read-only detector; it does not write back to the working tree or the staging area. There is no gitleaks mode that rewrites staged files in place. This is why Stage 1 must be a separate write-time mutator rather than delegating rewriting to gitleaks.

---

## Three Failure Modes Mitigated

### 1. Confidence Laundering

**Risk**: Future agents trust stale verdicts — a SHIP verdict from six months ago is treated as valid for the current codebase.

**Mitigation**: The `review_of_commit` field pins each review to a specific commit SHA. The index `staleness` column is computed at generation time from `reviewed_files`: the number of commits since each reviewed file was last touched (or `"unpinned"` for commit-unpinned reviews). Agents MUST check staleness before acting on a verdict.

### 2. Bias Propagation

**Risk**: Re-reviews are shown the prior verdict and anchor-bias toward it, inflating agreement even when the implementation has changed.

**Mitigation**: The `--diff-since-last-review` discipline (documented for the future re-review skill) feeds reviewers only the delta since the last review, not the full history. Reviewers receive the `review_of_commit` pointer but not the prior verdict text. This limits the anchoring surface.

### 3. Context-Window Pollution

**Risk**: Scouts or collator agents pull 30 stale reviews into context, exhausting the window before useful work begins.

**Mitigation**: The `superseded_by` chain marks older reviews as superseded when a new review covers the same artifact. `INDEX.json` carries the chain, and consumers SHOULD filter `superseded_by == null` to load only the active review. `INDEX.md` renders the chain for human inspection.

---

## Migration Policy

Existing reviews stored in Application Support (from ad-hoc RP chat sessions before fn-2) stay where they are. New reviews generated after fn-2 ships land in `.flow/reviews/<spec-id>/` automatically via the Stop hook.

A one-shot import utility that migrates Application Support reviews into `.flow/reviews/` is **explicitly deferred to v0.6**. This deferral is tracked as a `.work/issues.md` entry (per project standing convention for cross-version deferrals). No import tool ships in v1.

Migration is **opt-in**: users who do not run any future import tool are unaffected. All new review traffic routes through the new storage automatically.

---

## Pocock's Three-Clause ADR Test

An ADR is warranted when the decision is: (1) hard to reverse, (2) surprising, (3) a real trade-off.

| Clause | Evaluation |
|---|---|
| Hard to reverse | Yes — the schema (review.json sidecar + review.md render + directory layout) is referenced by the index generator, the pre-commit hook, future collators, and the review skills. Changing the schema requires migrating all existing review directories. |
| Surprising | Yes — most projects do not store review transcripts (even sanitised summaries) in-repo. A new contributor would reasonably ask "why is this here?". |
| Real trade-off | Yes — the alternatives (Application Support only, separate repo, commit-everything) each have different portability, security, and maintenance profiles. The hybrid commit/ignore model is not the obvious default. |

All three clauses pass. This ADR is warranted.

---

## External Acknowledgements

Per abcd's External Acknowledgement Policy:

- **gitleaks** (Apache-2.0, `zricethezav/gitleaks`) — used as Stage 2 detection backend in `verify_reviews.py`.
- **Pocock's three-clause ADR test** (described in ADR methodology literature; MIT-licensed tooling in the GitHub Spec Kit) — applied above.

---

## Related Documentation

- [`.flow/reviews/README.md`](../../../../.flow/reviews/README.md) — two-store split overview
- [ADR-5: Brief Is Current State](0005-brief-is-current-state.md) — JSON-internal / MD-render invariant
- [F-075 flow-next-local-review prior art](../../../../REPPL/abcdZero/docs/development/roadmap/features/planned/F-075-flow-next-local-review.md) — prior wrapper evaluation (external repo)
- [F-037 multi-model-review prior art](../../../../REPPL/abcdZero/docs/development/roadmap/features/completed/F-037-multi-model-review.md) — v2.5.1 patch-based wrapper (external repo)
