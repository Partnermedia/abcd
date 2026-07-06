---
id: itd-28
slug: rp-reviews-into-flow
spec_id: fn-2-move-repoprompt-review-artifacts-into
kind: standalone
suggested_kind: null
reclassification_history: []
created: 2026-05-07
updated: 2026-05-11
prd_path: null
prd_grandfathered: true
grandfathered: true
grandfathered_at_phase: phase-3-intent
glossary_terms_used:
  - core/brief
  - core/intent
  - core/lifeboat
  - core/oracle
  - core/persona
  - core/phase
  - core/spec
  - core/transport
---

# Spec-Tied Reviews Live Next To The Spec They Reviewed

## Press Release

> **abcd lands every flow-next review next to the spec it reviewed, pinned to the commit it reviewed, with a sanitisation pass before commit.** When `/flow-next:plan-review` or `/flow-next:impl-review` finishes, an abcd-side post-processor (fired by Claude Code's `Stop`/`SubagentStop` hook) captures the receipt JSON via the existing `REVIEW_RECEIPT_PATH` contract and lands a canonical JSON sidecar (`review.json`) and rendered Markdown view (`review.md`) in a per-review directory at `.flow/reviews/<spec-id>/<NNNN>-<slug>-<ref>/`. The sidecar carries the full sanitised review body, structured findings, and a `review_of_commit` SHA pin so future agents can detect when a review's findings have gone stale. Raw transcripts land in a per-review `raw/` subdirectory (gitignored). A two-stage redaction scheme applies: Stage 1 is a write-time sanitiser (AWS, GCP, Azure, Cloudflare, GitHub, Anthropic, OpenAI, Stripe, Slack, JWT, PEM) before any file is written; Stage 2 is a `gitleaks protect --staged --redact=100` detect-and-block gate that rejects commits if secrets survive Stage 1. An on-demand `scripts/abcd/reviews_index.py --spec <spec-id>` regenerates `INDEX.md` + `INDEX.json`; CI runs `--check` mode to catch drift without ever writing to the working tree.
>
> "Reviews used to die on the laptop they were generated on, in a folder I had to know about," said Frank, SRE. "Spec-tied reviews live where the spec lives. When I clone the repo six months later, every review comes with it — and `staleness: 14_commits` tells me at a glance which ones I should re-run."

## Why This Matters

flow-next today writes review artifacts to `~/Library/Application Support/RepoPrompt/...` (RP backend) or `/tmp/*-receipt.json` (Codex backend). Reviews are not portable, not scoutable, not survivable across `git clone`, and not linked to the spec they reviewed. Three failure modes follow:

1. **Confidence laundering** — future agents read a review with no SHA pin and treat its conclusions as ground truth, even when the code has since changed.
2. **Bias propagation** — re-reviews shown the prior verdict anchor-bias toward it ([arXiv 2603.18740][bias]).
3. **Context-window pollution** — scouts pulling unbounded review history burn 5–30% of their window on stale reviews from completed specs ([Liip][liip]).

This intent commits abcd to a per-spec, SHA-pinned, hybrid commit/ignore review storage with a redaction safety net. The implementation is **purely additive** — no upstream patches to flow-next or RP — using two existing extension contracts: the receipt JSON written via `REVIEW_RECEIPT_PATH`, and Claude Code's `Stop`/`SubagentStop` hook (the same mechanism flow-next itself uses for `ralph-receipt-guard.sh` and `ralph-verbose-log.sh`).

This is the **`press-release`-shaped commitment** behind spec `fn-2-move-repoprompt-review-artifacts-into` (already specced in `.flow/specs/`), which decomposes into 5 implementation tasks.

### Carve-out: spec-tied vs unscoped RP chats

This intent covers **spec-tied reviews only** — the flow-next plan-review, impl-review, and (future) spec-completion-review artifacts that have a known spec ID. The brief's existing `reviews_backends/repoprompt.py` design (`05-internals/02-adapters.md`, `05-internals/03-configuration.md`) covers a **different job**: lifeboat-style storage of unscoped ad-hoc RP chats (the "I spent 30 minutes brainstorming the next intent in RP" case) into `.abcd/development/activity/reviews/`. Two stores, two clear jobs:

| Store | Purpose | Cadence | Format | Lifeboat consumes |
|---|---|---|---|---|
| `.flow/reviews/<spec>/` | Per-spec engineering audit trail | Push at write-time (fn-2 Stop hook) | Canonical JSON sidecar (`review.json`) + derived MD render (`review.md`) | Yes |
| `.abcd/development/activity/reviews/` | Lifeboat-style storage of unscoped RP transports | Pull at sync time (`abcd dev-sync`) | Verbatim MD per transport | Yes |

The brief's `reviews_backends/repoprompt.py` survives but its scope narrows: it sweeps unscoped transports only, since spec-tied ones are already landed by fn-2's Stop hook (`.claude/hooks/flow-next-postprocess.sh`).

## What's In Scope

- **abcd-side post-processor** (`scripts/abcd/review_postprocess.py`) that captures receipt JSON via `REVIEW_RECEIPT_PATH` and writes a canonical JSON sidecar (`review.json`) + rendered MD (`review.md`) into a per-review directory under `.flow/reviews/<epic-id>/<NNNN>-<slug>-<ref>/`. Atomic write via staging-dir + `rename(2)`. Idempotent. Never deletes source until target write+verify confirmed.
- **Claude Code Stop/SubagentStop hook** (`.claude/hooks/flow-next-postprocess.json`) invokes the post-processor when `REVIEW_RECEIPT_PATH` is set. Mirrors `ralph-verbose-log.sh` precedent.
- **JSON sidecar schema** (committed at `docs/reference/review-schema.md`): required metadata (`review_of_commit`, `spec_path`, `spec_sha256`, `reviewer_model`, `reviewer_tool`, `verdict`, `generated_at`, `iteration`, `focus`, `review_type`, `target_id`, `reviewed_files`, `backend`, `pinning`, `allow_no_commit`), required content (`summary`, `body_markdown`, `findings`), required provenance (`sanitized_raw_artifact_sha256`), required truncation metadata (`truncated`, `truncation_method`, `omitted_bytes`, `body_max_bytes`, `render_max_bytes`); optional (`superseded_by`, `worktree_sha256`, `dirty`, `chat_id`, `session_id`, `receipt_path`).
- **Verdict enum locked**: `{SHIP, NEEDS_WORK, MAJOR_RETHINK}` — matches `ralph-receipt-guard.sh` upstream validation.
- **Sequence allocation**: `flock` with 5s timeout; up to 5 retries with 100ms exponential backoff on contention. No fallback filename variants — if all retries fail, exit non-zero with guidance.
- **Directory convention**: `<NNNN>-<slug>-<ref>/` where `<ref>` = 7-char short SHA (when `pinning: "commit"`) or literal `unpinned` (when `pinning: "none"`).
- **Two-stage redaction**: Stage 1 write-time sanitiser (`_review_lib.py::sanitise_text()`) strips home-dir paths, AWS/GCP/Azure/Cloudflare/GitHub/Anthropic/OpenAI/Stripe/Slack/JWT/PEM secrets before any file is written. Stage 2 detect-and-block: `verify_reviews.py` runs `gitleaks protect --staged --redact=100` and blocks commits if secrets survive Stage 1.
- **Size caps**: `body_max_bytes` (default 20 KiB, bounds `body_markdown`); `render_max_bytes` (default 24 KiB, bounds `review.md`). Both caps recorded in `review.json`.
- **`.gitignore`** allow-lists `.flow/specs/`, `.flow/tasks/`, `.flow/reviews/`; ignores `.flow/.checkpoint-*`, `.flow/memory/`, `.flow/bin/`, `.flow/config.json`, `.flow/meta.json`, `.flow/reviews/**/<NNNN>-*/raw/`.
- **Index discovery**: `scripts/abcd/reviews_index.py --epic <epic-id>` generates `INDEX.md` + `INDEX.json` on demand (no new flowctl command surface). CI verifier runs `--check` mode on PRs touching `.flow/reviews/**`. No git hooks — on-demand plus CI only.
- **Brief edits in this repo (abcdDev)**:
  - `05-internals/02-adapters.md`: clarify `reviews.py` dispatcher's two stores (spec-tied via Stop hook → `.flow/reviews/`; unscoped via `repoprompt.py` adapter → `.abcd/development/activity/reviews/`).
  - `05-internals/03-configuration.md`: clarify `dev_sync.reviews.enabled` only sweeps unscoped transports.
  - ~~fn-5-bsc artifact taxonomy~~ — **REMOVED post-audit (2026-05-07)**: this referenced a legacy `abcd`-repo concept that does not exist in `abcdDev`'s live brief or its archives. The two carve-out edits above (to `02-adapters.md` and `03-configuration.md`) are sufficient; no separate "fn-5-bsc taxonomy edit" is needed.
- **Phase-1 default**: post-processor ON by default; `ABCD_RP_POSTPROCESS=0` kill switch.
- **Dirty-tree policy**: capture `worktree_sha256` and tag `dirty: true`; do NOT block.
- **No-commit-yet policy**: refuse review on branch with zero commits; emit guidance; respect `--allow-no-commit` override.
- **Acknowledgements**: README "Acknowledgements" cites `gitleaks` (Apache-2.0), `joelparkerhenderson/architecture-decision-record` (CC-0, ADR `NNNN` naming convention), `REPPL/abcdZero` F-075 / F-037 (prior art for abcd-side wrapper architecture).

## What's Out of Scope

- **Upstream patches to flow-next or RP** — receipt JSON + Stop hook is sufficient. (Github-scout established that flow-next has no plugin SDK and gmickel rejects/supersedes large `feature` PRs that overlap his roadmap.)
- **One-shot Application Support import tool** — explicitly out of scope (documented in ADR-6 migration policy).
- **Cross-spec review aggregation, sigstore signing, IDE integration, scoring/trends** — out of scope.
- **Modifying ephemeral build artifacts** (`/tmp/review-prompt.md`, `/tmp/re-review.md`) — keep these in `/tmp/`.
- **Hand-rolling redaction regexes** — use gitleaks instead.
- **Replacing flow-next reviews** — this intent adds storage + governance, not a new review backend.
- **Automatic git-hook regeneration of INDEX.md** — practice-scout established this is the canonical anti-pattern (pre-commit#2240 re-stage loop, post-commit feedback loops). On-demand only + CI verifier.
- **Unscoped RP transport storage** — covered by the brief's existing `reviews_backends/repoprompt.py` design, narrowed scope after this intent.
- **`Reviewed-by:` git trailer auto-injection on implementation commits** — out of scope (nice-to-have bidirectional linkage).

## Acceptance Criteria

- **Given** a persona runs `/flow-next:plan-review fn-X` end-to-end on either RP or Codex backend, **when** the review completes, **then** a per-review directory lands at `.flow/reviews/fn-X/<NNNN>-<slug>-<ref>/` containing `review.json` (all required fields populated, `verdict` ∈ `{SHIP, NEEDS_WORK, MAJOR_RETHINK}`, non-empty `body_markdown` and `reviewed_files`) and `review.md` (mechanically rendered from `review.json`).
- **Given** the post-processor runs twice on the same receipt, **when** both invocations complete, **then** there is exactly one per-review directory in `.flow/reviews/fn-X/` (idempotent; the second invocation is a no-op).
- **Given** the post-processor is killed mid-write (`kill -9`), **when** the persona inspects the working tree, **then** no `.tmp` or partial files are visible to git.
- **Given** 5 concurrent post-processor invocations on the same spec, **when** they complete, **then** 5 distinct sequence numbers exist (no collisions).
- **Given** the persona sets `ABCD_RP_POSTPROCESS=0` and runs `/flow-next:plan-review`, **when** the review completes, **then** the post-processor exits 0 with no side effects and the review lands where flow-next puts it today.
- **Given** a staged `.flow/reviews/**` file containing a multi-cloud secret (AWS access key, fine-grained GitHub PAT, Anthropic key, JWT, or PEM private key) that was NOT caught by Stage 1, **when** the pre-commit hook runs `gitleaks protect --staged`, **then** the commit is blocked with the finding path/line/rule reported (never the raw secret value).
- **Given** a committed `.flow/reviews/**/*.md` file containing `AKIAIOSFODNN7EXAMPLE`, **when** the pre-commit hook runs, **then** the EXAMPLE-allowlisted value is NOT redacted.
- **Given** a review file with a `review_of_commit` SHA that fails `git rev-parse --verify`, **when** the pre-commit hook runs, **then** the commit is rejected with a clear error message.
- **Given** a `review.json` with `body_markdown` exceeding `body_max_bytes`, **when** the pre-commit verifier runs, **then** the commit is rejected with guidance (the post-processor truncates automatically; this acceptance captures the case where someone manually edits the sidecar to violate the cap).
- **Given** a CI run on a PR touching `.flow/reviews/**`, **when** `scripts/abcd/reviews_index.py --check --all` runs, **then** the workflow fails on drift with the exact remediation command and never writes back to the branch.
- **Given** the brief is updated, **when** a contributor reads `05-internals/02-adapters.md` and `05-internals/03-configuration.md`, **then** they find explicit text describing the two-store carve-out (spec-tied via fn-1 Stop hook; unscoped via `repoprompt.py` adapter).
- **Given** the README is updated, **when** a contributor reads the Acknowledgements section, **then** they find explicit citations of `gitleaks` (Apache-2.0) and `REPPL/abcdZero` F-075 / F-037 prior art.

## Open Questions

- ~~**Sibling-repo edit location for fn-5-bsc taxonomy**~~ — **RESOLVED post-audit (2026-05-07)**: the legacy `fn-5-bsc` reference does not match anything in `abcdDev`'s live brief. The carve-out edits to `02-adapters.md` and `03-configuration.md` (see "In Scope" above) are the actual brief edits needed.
- **Hook coverage for non-Claude-Code runtimes** (Codex CLI users): standalone CLI invocation `scripts/abcd/review_postprocess.py --from-receipt $REVIEW_RECEIPT_PATH --epic <id>` is the documented fallback. Should a Codex-specific wrapper also ship?
- **Stale review threshold for `staleness` column**: currently `<N>_commits` since `review_of_commit`. Should we add a "danger" threshold (e.g., `>20_commits` shown red)? Deferred polish.

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._

## References

- Linked spec: `fn-2-move-repoprompt-review-artifacts-into` (in this repo's `.flow/specs/`).
- Depends on: `itd-1` (acceptance gates).
- Coordinates with: `itd-7` (RP workspace portability) — both pull from RP Application Support but for different artifact types.
- Coordinates with: `itd-13` (scheduled dev-sync) — unscoped review pull would benefit from scheduled sync.
- Prior art: `REPPL/abcdZero` F-075 (planned, evaluated upstream-vs-wrapper, chose wrapper) and F-037 (completed v2.5.1, shipped patch-based wrapper).

[bias]: https://arxiv.org/html/2603.18740v1 "Confirmation Bias in `LLM`-Assisted Security Code Review"
[liip]: https://www.liip.ch/en/blog/preventing-context-pollution-for-%61i-agents "Liip — Preventing Context Pollution for `AI` Agents"
