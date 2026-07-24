# Next implementation-run queue (2026-07-24)

**Status:** the backlog for the next autonomous implementation run, consumed by
the generic protocol at
[`2026-07-12-abcd-run-protocol.md`](2026-07-12-abcd-run-protocol.md). This file
supersedes the pick-up role of
[`2026-07-18-next-drain-run-queue.md`](2026-07-18-next-drain-run-queue.md)
(itd-93's readiness gates listed there are now cleared — see the 2026-07-24
DECISIONS.md entries; that file remains the record of why they were gates).

Run contract, unchanged: `make preflight` is the sole gate; ledger is
capture-only; correctness + security reviews before each PR; one PR per item;
never merge, never commit to main. Prioritisation lens (maintainer,
2026-07-24): abcd is designed while being built — prefer work that speeds
abcd's own build loop or gets new functionality into early dogfooded use.

## STOP conditions (this run)

Hitting one means stop and report, never push through (per the playbook):

1. **itd-93 PRD synthesis ambiguity** — the PRD is synthesised from the
   2026-07-24 grill decisions (DECISIONS.md). If synthesis requires an
   interpretation the grill did not settle, STOP; do not invent the bar.
2. **receipt_gate schema** — iss-122 adds manifest-hash and tier fields to the
   receipt contract. Any change beyond adding those fields (renaming, removing,
   re-gating existing fields) is a STOP.
3. **Missing or ambiguous plan/spec file** — fail closed; never synthesise a
   substitute record.
4. **Any fix ahead of an armed detector** — record the finding first
   (capture-only), then fix behind it.

## Track 1 — loop friction + grill consequences (small, mechanical)

1. **iss-117** — `capture resolve` / `intent seedDraft` gain the `impact`
   field so abcd's own verbs satisfy its own validators. Root-cause class:
   the write paths, not the validators.
2. **iss-115 / iss-120** — one allocator fix for the class: record-id minting
   (`iss-` / `itd-` / `spc-`) must not collide across parallel branches.
   Detectors are armed (see iss-120); fix the class, not the site.
3. **iss-101 / iss-102** — lock the lockless load-modify-write paths
   (history index registerRepo; capture orphan sweep vs commit).
4. **iss-75** — always-latest dev install mode, so the dogfooded binary is
   the tip build.
5. **iss-35 → resolved/** — via `abcd capture resolve`; the graduation gate
   it stayed open for is shipped and armed (decision 2026-07-24).
6. **iss-122** — implement the pinned crosscheck gate: committed input
   manifest under `.abcd/development/release-gate/` (doc list, directions,
   checker count, prompt hash); tiered depth (full for feature/breaking,
   Direction-B shallow for patch); receipt echoes manifest hash + tier;
   receipt_gate refuses on tier/manifest mismatch or undispositioned
   findings — procedural refusal only, findings themselves route to the
   maintainer (verifier-selects-gates-decide).
7. **itd-93 amend + promote** — fold the four settled design decisions into
   the intent (launch sub-verb; self-scaffold parity; built-in rehearsal
   mode; changelog-heading seam), synthesise the PRD from the grill,
   `abcd intent plan itd-93` → planned/, kind standalone, severity minor.
   **Promotion only — implementation is NOT in this run** (feature-class;
   warrants its own focused run per the 2026-07-18 shape note).

## Track 2 — itd-94 implement-readiness gate

`abcd intent ready <itd-N>` per spc-9: the four checks (bucket, acceptance
criteria, spec link, spec body), strict exit-code contract, refusal with
remedy. Self-contained, zero open questions. Lands the machine version of the
run-curation discipline this very queue applies by hand.

## Track 3 — probe hardening, then the itd-88 coverage experiment

1. Walk fixes first, so the experiment's verdicts are trustworthy:
   **iss-111** (open-question marker false positives), **iss-112** (unbounded
   ReadDir), **iss-114** (O(entries×depth) walk), **iss-116** (skip-set
   misses non-Go/Node ecosystems).
2. **itd-88** (spc-3) — `/abcd:disembark probe` + `coverage`, read-only,
   cite-or-be-dropped. Closing acceptance: a probe run over abcd-cli itself,
   and at least one foreign repo, with the coverage report read as the
   experiment's finding — the packer is built to whatever section list
   survives.

## Queued for the following run (not this one)

- **itd-28** — implement against the 2026-07-24 adapter decision: Stage 2
  through the scanner seam, native default, gitleaks opt-in adapter, engine
  reported loudly, CI gitleaks as backstop; iss-96 pattern parity in scope.
- **itd-93 implementation** — focused run once promoted.
- **Launch chain (itd-66 → itd-65 → itd-72)** — grilled and PRD-ready;
  hardens shipping, exercised once per release; sequenced after the loop
  friction and learning items by the 2026-07-24 lens.

## Also open (recorded so the run sees them)

- Ledger follow-ups from the 2026-07-18 file remain: iss-104, iss-105
  (itd-93-adjacent), iss-106 — drain candidates once triaged.
- `planned/README.md`'s file listing has drifted from the directory
  (iss-38 class); fix opportunistically when touching that directory.
