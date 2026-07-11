# CONTEXT

Shared team/agent orientation — what you need to know *right now* to be useful.
Short and pointer-heavy; durable design truth lives in
[`../development/`](../development/), personal session state in
`../.work.local/NEXT.md` (local).

This file is **status-free** (see `DECISIONS.md`, 2026-07-10): what is being
worked on, what is next, and what has shipped are read live from the surfaces
built for that — the issue ledger (`abcd capture list --open`), the roadmap
dashboard (`../development/roadmap/README.md`), and the intent buckets — never
written here, where they would only go stale. A record-lint rule
(`context_status_free`) enforces this.

## What this repo is

`abcd-cli` is the from-scratch **Go** rebuild of abcd as a host-agnostic
configuration layer for development — a single `abcd` binary with a
transport-agnostic core, usable as a Claude Code plugin and in the companion
harness's ecosystem, depending on no external tools. It supersedes the frozen
Python reference implementation.

## Live constraints / sharp edges

- Parts of the copied design record may still describe superseded
  architecture; the brief-vs-surface reconciliation (iss-35 in the ledger) is
  the open cross-check. Where the record and the binary disagree, verify
  against the binary before trusting either.
- Single repo, curated release (no dev→public mirror). `.abcd/**` never ships.
- Never commit/push without the maintainer asking; new deps need sign-off.
