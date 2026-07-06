# Build Sequence

The build comprises the **brief Phase 0 — Foundation (fn-1 in flow-next)**, the **phased intents**, and the **brief-defined plumbing phases (Pass A/B/C, embark, launch)** below. The dependency DAG (not a linear list) governs what can run when. The intents are bundled across the roadmap phases — the canonical intent set and its execution order live in the phase docs and the intent index, not here: see [`roadmap/phases/README.md`](../../roadmap/phases/README.md) and [`roadmap/intents/README.md`](../../intents/README.md).

> **Canonical sequencing lives in [`roadmap/phases/`](../../roadmap/phases/README.md).** Per [adr-9](../../decisions/adrs/adr-9-phase-as-product-layer.md), the project's ordered build plan is the phase set — each phase doc bundles a set of intents and the plumbing phases below, and opens with a product `## Expectation`. This file remains the canonical *plumbing-phase detail* (what Pass A/B/C, embark, and launch actually do); the phase docs reference these sections via their `## Maps against`. The phase docs (Phase 0 — Substrate & disciplines, Phase 1 — ahoy, Phase 2 — capture, Phase 3 — intent, Phase 4 — lifeboat, Phase 5 — round-trip) are the live plan and the authoritative intent order (per adr-9).
>
> **Two numberings, deliberately distinct.** This file's `## 2`–`## 10` headers number the *brief plumbing phases* (Foundation, ahoy flow, adapters, Pass A/B/C, embark, launch). The roadmap's Phase 0–5 are the *product phases*. They are not the same axis and not meant to align one-to-one — a product phase doc bundles one or more brief plumbing phases via its `## Maps against`. The plumbing-phase numbers below are stable; do not renumber them to chase the roadmap.

## 1. Intent execution order

Per [adr-9](../../decisions/adrs/adr-9-phase-as-product-layer.md), the authoritative intent set and its execution order — including per-intent dependencies and the bundling into product phases — live in the phase docs and the intent index, not in this file. See [`roadmap/phases/README.md`](../../roadmap/phases/README.md) for the ordered phase plan (each phase doc names the intents it bundles via its `## Maps against`) and [`roadmap/intents/README.md`](../../intents/README.md) for the intent corpus with kinds and lifecycle state. The dependency DAG governs what can run when; abcd's autonomous loop picks ready intents up in dependency order, with parallelism where dependencies allow.

## 2. Phase 0 — Foundation (fn-1 in flow-next)

Five prerequisite tasks before any user-facing intent makes confident decisions. Already plan-reviewed and ready to ship.

**Output destinations**: study artefacts (predecessor-notes, transcript-sampling, idelphi-rescue-study) go to `.abcd/development/research/phase/<N>/` (committed in private repos — these are *design inputs* future phases consume). Ephemeral acceptance-check logs (e.g., `claude plugin validate` stdout) go to `.abcd/logbook/phase/<N>/` (per the visibility rule [`05-internals/03-configuration.md § 1`](../05-internals/03-configuration.md#1-visibility-driven-gitignore-policy); same tracking semantics either way, but `research/` vs `logbook/` is a semantic split — design-input vs run-log).

1. **Read predecessors (a skim, not a deep read).** Skim BOTH `~/ABCDevelopment/Autonomous/abcdZero/` (older first attempt) and `~/ABCDevelopment/Autonomous/abcdSubZero/` (Python CLI iteration). Notes file at `.abcd/development/research/phase/0/predecessor-notes.md`. Patterns to learn from, NOT a port plan. Same pass also confirms current state of Codex CLI/MCP for the oracle backend wiring ([`05-internals/01-agents.md` § Oracle backend resolution](../05-internals/01-agents.md#oracle-backend-resolution)).
2. **Sample idelphiDev transcripts.** Open 5–10 `.specstory/` transcripts at random + 5 from time windows around known spec boundaries. Output: signal-density notes (focused / wandering / off-topic). Informs Pass B design.
3. **Study idelphi rescue.** Two-part read: (a) `idelphiDev/.work/lifeboat/rescue/extraction.md` to understand what a *good lifeboat captures*; (b) directory diff `iDelphiZero/iDelphi/` ↔ `idelphiDev/iDelphi/` Swift source trees. Output: `.abcd/development/research/phase/0/idelphi-rescue-study.md` with sections "good lifeboat patterns" and "rescue diff observations", with anti-patterns flagged.
4. **Scaffold abcdDev minimum.** Create `.claude-plugin/plugin.json`, `marketplace.json`, `README.md`, `.gitignore`, `scripts/abcd-cli` bash wrapper (named `abcd-cli` because `scripts/abcd/` package pre-exists), stub `abcd_cli.py`. Plugin loads cleanly via `claude --plugin-dir`.
5. **Define harness.py interface.** Write the harness shim signatures (AskUserQuestion, agent dispatch, MCP calls, background, scheduling) before any command logic. Locks the abstraction boundary day one.

**Acceptance:** see fn-1 spec (`.flow/specs/fn-1-phase-0-foundation-predecessor-study.md`).

## 3. Capability slices (intents itd-1..itd-7)

Each **standalone or bundle-member** intent has its own spec under `.flow/specs/<fn-N>-<intent-id>-<slug>.md` once `/abcd:intent plan <itd-N>` runs (bundle-members share a spec with their bundle-mates). **Discipline** intents get no *dedicated 1:1* spec — the *rule* registers by living in `disciplines/` and is enforced by being inherited as an acceptance gate on every other spec (per itd-34). But this does not mean a discipline ships without code: the **enforcement machinery a discipline specifies** — lint checks (the `IL`/`MG`/`VR` families in `intent_lint.py`, the prompt-quality checks in `lint_prompts.py`), intent/spec template sections, and the `intent-fidelity-reviewer` roles that run the discipline checks — is *plumbing*, and ships through brief-phase plumbing specs like all other plumbing (per § 2 below and the plumbing-phase sections). So a discipline has two homes: the rule lives in `disciplines/` with no spec; the machinery that makes the rule biting is built by whichever plumbing spec owns the lint and reviewer infrastructure. All intent kinds carry their own acceptance criteria (per itd-1). The brief doesn't duplicate intent specs; see the intent files in `.abcd/development/roadmap/intents/{drafts,planned,disciplines}/`.

Post-brief additions (itd-27, itd-28, itd-34) are not enumerated here; they ship per their own dependencies — see [`roadmap/phases/`](../../roadmap/phases/README.md) for the phase plan and [`intents/README.md`](../../intents/README.md) for the intent index.

Notable inter-intent + intent-vs-phase couplings:

- **itd-3 (rules-loader) precedes ahoy** so ahoy ships with the right marker block.
- **itd-4 (capture) writes to the issue ledger** that subsequent work references. Once shipped, all future "discovered issue" prose in the brief and intents references the structured ledger, not free-form `.work/issues.md`.
- **itd-5 (prompt-quality) is a cross-cutting acceptance rule.** Every Pass A/B/C agent spec's acceptance includes the itd-5 gates (prompt_version, self-improvement preflight, injection-canary fixture for agents reading untrusted input).
- **itd-6 (RP MCP) and itd-7 (RP workspace) ship after the lifeboat pipeline establishes the dev-sync + adapter foundation.**

## 4. Phase 1 — `/abcd:ahoy` end-to-end

Runs after itd-2 (in-session), itd-3 (rules-loader), itd-4 (capture).

1. `abcd init --json` and `abcd config get|set`.
2. `/abcd:ahoy install` command flow (Steps 0–12 from [`../04-surfaces/01-ahoy.md`](../04-surfaces/01-ahoy.md)).
3. Probe-only stubs: `/abcd:disembark probe`, `/abcd:embark probe <path>`, `/abcd:launch dry-run`, bare `/abcd:intent` (renders intent corpus per universal bare-command-as-render discipline; see [`../02-constraints/04-naming.md`](../02-constraints/04-naming.md)), `/abcd:capture list` (filtered query — earned sub-verb).

**Acceptance:** see [`../04-surfaces/01-ahoy.md § 1`](../04-surfaces/01-ahoy.md#1-acceptance).

## 5. Phase 2 — Settled-artefact adapters

Runs after Phase 1.

All adapters in [`../05-internals/02-adapters.md`](../05-internals/02-adapters.md). Interactive source confirmation when assets or ambiguous docs found.

**Acceptance:**
- **Given** each corpus repo, **when** `/abcd:disembark to home --no-agents` runs, **then** the current repo's `.abcd/lifeboat/` is produced with all verbatim sections (specs/, ADRs, memory, oracle reviews) but no synthesised files; all schemas validate.
- **Given** a repo where an adapter source is missing or sparse, **when** the adapter probes, **then** the result is documented in the report (empty + reason) and downstream agents handle the absence gracefully.

**Three orthogonal disembark preview shapes** (post bare-as-help refactor): `probe` sub-verb (adapter probes only, list sources, write nothing — ultra-light), `dry-run` sub-verb (full plan: what files would be written, which agents would be dispatched, estimated tokens — write nothing), `to <path> --no-agents` (the action with a flag-shaped modifier: write the verbatim/deterministic parts, skip LLM synthesis — used here for adapter validation).

## 6. Phase 3 — Pass A spine agents

Runs after Phase 2.

`flow-essence`, `decision-archaeologist`, `review-collator`, `code-rescuer` (principle-only).

**Acceptance:**
- **Given** a corpus repo with `.flow/specs/` and ADRs present, **when** Pass A runs, **then** `spec-essence.json` shows correct supersession chain and `decisions-timeline.json` references all ADRs.
- **Given** Pass A outputs, **when** the oracle audit runs, **then** the verdict is "sufficient" with specific findings.
- **Given** Pass A outputs, **when** Pass B inputs are validated, **then** the round-trip gate passes (no parse errors, schemas validate).
- **Given** itd-5 is in force, **when** any Pass A agent's spec closes, **then** the agent has `prompt_version: 1.0.0`, the self-improvement pre-flight is recorded in `agents/CHANGELOG.md`, and the injection-canary fixture (for agents reading untrusted input — `decision-archaeologist`, `review-collator`, `code-rescuer`) passes.

## 7. Phase 4 — Pass B targeted chat retrieval

Runs after Phase 3.

Time-window index from `.flow/specs/` git-blame + override via `.abcd/lifeboat/spec-windows.json`. `chat-distiller` per spine entry. `--full-distill` opt-in for exhaustive map-reduce.

**Acceptance:**
- **Given** a spine entry with a time window, **when** `chat-distiller` is dispatched, **then** the agent receives only the time-windowed transcript subset (no full corpus dump).
- **Given** a Pass B run on each corpus repo, **when** complete, **then** `rationale-fills.json` cites transcript filenames and the oracle gate passes.
- **Given** the round-trip gate to Pass C, **when** Pass B outputs are consumed, **then** schemas validate without errors.
- **Given** itd-5 is in force, **when** `chat-distiller`'s spec closes, **then** the injection-canary fixture (for transcript content) passes.
- **Given** the input corpus, **when** Pass B starts, **then** files matching pattern `you-are-running-one*` are excluded before any LLM call.
- **Given** a time-windowed transcript subset, **when** signal density is computed, **then** the denominator is total user messages in the window AND density ≥ 15% triggers VIABLE; <15% triggers MITIGATIONS-REQUIRED; the spec-formula metric is NOT the gating signal.
- **Given** a transcript file >100 KB after time-window extraction, **when** Pass B processes it, **then** within-transcript chunking (2000-line segments) and map-reduce aggregation precede any `chat-distiller` call.

## 8. Phase 5 — Pass C principles, compose, audit + `/abcd:disembark to <path>` end-to-end

Runs after Phase 4.

**Prerequisite:** `dev-sync` ([`../05-internals/03-configuration.md § 2`](../05-internals/03-configuration.md#2-abcddevelopmentactivity-namespace-and-dev-sync)) must run first as Phase 0 of disembark — Pass C agents (`principle-distiller`, `press-release-composer`) consume curated `.abcd/development/activity/` and `.abcd/memory/` content. If dev-sync fails on any source, Pass C runs in degraded mode with clear notes in the report.

`principle-distiller`, `artefact-curator`, `brief-composer`, `press-release-composer` (with embedded oracle product audit), `lifeboat-oracle`, `documentation-auditor` (subagent pre-pack). Re-run flow with overwrite confirmation + `.bak` safety net. (The earlier `--apply-audit` flag was deprecated in the bare-as-help refactor; re-running `disembark to <path>` against a stale lifeboat applies the same Pass B+C re-execution.)

**Acceptance:** see [`../04-surfaces/02-disembark.md § 7`](../04-surfaces/02-disembark.md#7-acceptance).

## 9. Phase 6 — `/abcd:embark from <path>`

Runs after Phase 5.

Source lookup (`from home` for round-trip → `from <path>` → `scan` discovers candidates), emptiness check, scaffolder agent, conflict bulk prompt, asset curation per classification. Documentation-auditor runs post-scaffold.

**Acceptance:** see [`../04-surfaces/03-embark.md § 6`](../04-surfaces/03-embark.md#6-acceptance).

## 10. Phase 7 — `/abcd:launch ship`

Runs after Phase 6.

Scan stack (gitleaks + Presidio + custom regex + optional TruffleHog + OWASP/security check + documentation-auditor pre-promotion), payload manifest + `.abcd/launch.allow`, mirror modes, version bump + marketplace.json update.

**Acceptance:** see [`../04-surfaces/04-launch.md § 7`](../04-surfaces/04-launch.md#7-acceptance).

## 11. Validation cadence

After **every phase**, run `/abcd:disembark to home` (or the relevant preview sub-verb — `probe` for adapter-only inspection, `dry-run` for the full plan without writes) against the full validation corpus. Catch regressions early. Acceptance recorded in `.abcd/logbook/phase/<phase-id>/` (per the logbook layout in [`../05-internals/04-universal-patterns.md § 6`](../05-internals/04-universal-patterns.md#6-abcdlogbook-layout)).

## 12. OpenCode portability

**Comes in a later phase as itd-22.** abcd ships Claude Code only with the `harness.py` shim ready for a second implementation later (the harness shim is the early commitment that makes the OpenCode port tractable). Same applies to the `memory_backends/opencode.py` and `reviews_backends/codex.py` stubs — interfaces declared, implementations come later.
