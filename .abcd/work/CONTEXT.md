# CONTEXT

Shared team/agent orientation — what you need to know *right now* to be useful.
Short and pointer-heavy; durable design truth lives in
[`../development/`](../development/), personal session state in
`../.work.local/NEXT.md` (local).

## What this repo is

`abcd-cli` is the from-scratch **Go** rebuild of abcd as a host-agnostic
configuration layer for development — a single `abcd` binary with a
transport-agnostic core, usable as a Claude Code plugin and in the the companion harness
ecosystem, depending on no external tools. It supersedes the Python
implementation (frozen, read-only, in the sibling `abcdDev` repo).

## Current phase

**Phase 0 — Foundations.** Scaffolding the repo: git/CI/build carried from
`ferry`, the design record carried from `abcdDev`, and the Go core + CLI + plugin
surface skeleton. Exit: `make preflight` green and a verb round-tripping
CLI → core → JSON. After Phase 0 we pick the cadence for Phase 1+.

**Next:** Phase 0.5 — a full up-front reconciliation of the copied design record
against the current architecture decisions (single-repo launch, native-minimal
spec, host-delegated LLM, Workflows-not-Ralph, the three-tier layout), before any
feature code.

## Live constraints / sharp edges

- The copied `.abcd/development/` record still describes the *old* architecture
  (flow-next required, overlay/abstraction-boundary, two-repo launch, RP/codex
  oracles). It is the starting spec, not current truth, until Phase 0.5
  reconciles it. Do not treat it as authoritative before then.
- Single repo, curated release (no dev→public mirror). `.abcd/**` never ships.
- Never commit/push without the maintainer asking; new deps need sign-off.
