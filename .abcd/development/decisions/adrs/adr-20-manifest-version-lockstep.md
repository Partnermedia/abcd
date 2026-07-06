---
id: adr-20
slug: manifest-version-lockstep
status: accepted
date: 2026-07-03
supersedes: null
superseded_by: null
related_intents: [itd-69]
related_rfcs: []
related_adrs: [adr-19]
---

# ADR-20: The two published manifests stay version-consistent by an anti-drift checker over a pinned path list; dev stays unversioned; `--allow-dirty` must never bypass it

## Context

[adr-19](./adr-19-plugin-json-version-carve-out.md) settled **where** the plugin
version lives: only in the published public snapshot, at the location the fn-77.1
decision artifact ([`.abcd/config/version-location.json`](../../../config/version-location.json))
selects — today `plugin.json` at pointer `/version`. adr-19 owns the *write*
location and the dev-stays-unversioned polarity.

What adr-19 did NOT settle is the **anti-drift invariant** between the two
published manifests. The version is single-sourced by the render path
(`scripts/abcd/public_manifest.render_public_manifests`, driven by
`version-location.json`), and the published marketplace metadata
(`.claude-plugin/marketplace.json`) records the same release's version and
changelog entry (`04-launch.md` §4 step 3). Because two files describe one
release, a broken or partial publish can leave a **half-state** — a version in
one manifest and not the other, or a `present-null` where an absent key was
expected. `version-location.json` alone encodes only the *selected primary
location*; it does not enumerate the *full set* of version/changelog-bearing
fields the consistency check must inspect. That enumeration is the gap this ADR
closes.

Three coupled questions had to be settled:

1. **What exactly is checked, per tree?** The dev tree and the public tree have
   opposite polarity (adr-19: dev unversioned, public versioned). A heuristic
   "find every version-bearing field" scan is brittle and untestable. We need an
   explicit, pinned path list the checker reads — never discovered.
2. **Does `--allow-dirty` bypass the check?** The launch pre-flight suite
   (fn-79/fn-80) exposes an `--allow-dirty` escape hatch for a dirty worktree.
   Manifest consistency is a correctness invariant of the *published output*, not
   a worktree-cleanliness concern, so `--allow-dirty` must not weaken it.
3. **What shape is a published-marketplace changelog entry?** `04-launch.md` §4
   step 3 ("records the version + changelog entry in the marketplace metadata")
   was under-specified. fn-80's bump step is the stated consumer that writes and
   validates the entry, so the entry needs a machine-checkable contract.

## Decision

### The lockstep policy

The plugin version is **single-source-written** by the render path
(`render_public_manifests`, driven by `version-location.json`); no other code
writes a version into a published manifest. The anti-drift invariant — the two
published manifests, plus the `version-location.json` contract, describe the same
release consistently — is proven by a **read-only checker**
(`scripts/abcd/manifest_lockstep.py` + its CLI). The checker never writes, never
bumps, and **has no bypass/skip flag at its own layer**.

The dev tree stays **unversioned** (adr-19): the version keys the checker inspects
are expected ABSENT in dev mode. The public tree is **versioned**: the version at
the contract location must exist and every pinned secondary location must agree.

### The pinned per-mode path list (the contract this ADR owns)

The checker reads this list; it never discovers version-bearing fields
heuristically. Each entry is `(file, RFC-6901 pointer, meaning)`. The primary
location is not hard-coded here — it is read from `version-location.json`
(`manifest_path` + `json_pointer`) so an adr-19 relocation is absorbed without
editing this list. The secondary locations ARE pinned here, because
`version-location.json` does not enumerate them.

**PUBLIC tree — every pinned location must carry the release version, and all
must AGREE:**

| # | File | Pointer | Meaning |
|---|------|---------|---------|
| P1 | *(from `version-location.json`)* `manifest_path` | *(from `version-location.json`)* `json_pointer` | The primary version location adr-19 selects (today `.claude-plugin/plugin.json` `/version`). |
| P2 | `.claude-plugin/marketplace.json` | `/plugins/0/version` | The marketplace plugin entry's published version — must equal P1. |
| P3 | `.claude-plugin/marketplace.json` | `/plugins/0/changelog` | The marketplace plugin entry's changelog location/entry (validated against the changelog-entry schema, see below); its `version` field must equal P1. |

**DEV tree — every pinned location's version key must be ABSENT (adr-19
polarity):**

| # | File | Pointer | Meaning |
|---|------|---------|---------|
| D1 | *(from `version-location.json`)* `manifest_path` | *(from `version-location.json`)* `json_pointer` | The primary location must NOT carry a version in dev. |
| D2 | `.claude-plugin/marketplace.json` | `/plugins/0/version` | The marketplace entry must NOT carry a version in dev. |
| D3 | `.claude-plugin/marketplace.json` | `/plugins/0/changelog` | No changelog entry in the dev marketplace. |

`present-null` (`"key" in data` but value `null`) and `absent-key` (`"key" not in
data`) are **distinguished**: dev mode requires absent-key at the version
pointers (a `present-null` is drift, not compliance); public mode requires a
present non-null agreeing value.

### Exit codes and report format (pinned)

The checker's contract, consumed by fn-79/fn-80 wiring:

| Exit | Meaning |
|------|---------|
| 0 | consistent |
| 1 | drift detected |
| 2 | contract unreadable (a pinned file missing/unparseable, or `version-location.json` malformed) |

`stdout` carries a machine-parseable report: one `DRIFT <tree> <file><pointer>:
<detail>` line per drifting field, an `OK <tree>` line when consistent, and an
`UNREADABLE <detail>` line for exit 2. The `--tree dev|public` argument is
**explicit** — no auto-detection.

### `--allow-dirty` must NOT bypass manifest consistency (wiring policy)

When fn-79/fn-80 wire this checker into the launch pre-flight suite, the manifest
lockstep gate MUST run and MUST be honoured **even under `--allow-dirty`**.
`--allow-dirty` relaxes the worktree-cleanliness pre-flight, not the
published-output-correctness invariant. **This is a POLICY this ADR records; it is
not enforced by fn-83.** fn-83 ships only the flag-less checker (which structurally
cannot be bypassed at its own layer) plus this policy record. The rule becomes
enforced when fn-79/fn-80 lands its wiring test asserting the gate fires under
`--allow-dirty`. Until then it is **advisory**.

### The fn-80 handoff

fn-80 (Tier B publishing) is the caller. After it writes the public manifests via
`render_public_manifests` and records the changelog entry, its bump step invokes
the checker in `--tree public` mode against the staged public tree and refuses to
publish on a non-zero exit. fn-80 also validates the changelog entry it writes
against the changelog-entry schema (below). Registering the gate in the pre-flight
suite is fn-79/fn-80 territory, not an fn-83 edit — fn-83 delivers the callable
module + CLI + tests and this handoff record.

### The published-marketplace changelog-entry schema

**FORM decision (the deciding factor is fn-80's consumption mode):** fn-80's bump
step validates the changelog entry **programmatically** before publishing (the
stated consumer writes it from the git-tag + launch-report auto-changelog and must
prove its shape). Because the entry is machine-written and machine-validated — not
human-authored prose — the schema is a **committed schema file**
(`scripts/abcd/schemas/changelog-entry.schema.json`), not ADR-section prose. The
ADR-prose form is reserved for the human-authored-and-never-machine-validated case,
which does not apply here.

The changelog entry records, per published release: the `version` (must equal P1),
the bump `tier` (`patch|minor|major`), the human `reason` string, the release
`date`, and the `source_sha` the release was cut from. Its authoritative field
list lives in the committed schema; see the schema `description` for the binding
detail.

## Alternatives Considered

- **Heuristic version-field discovery.** Rejected: scanning every manifest for
  "version-looking" keys is brittle, untestable, and would silently miss a new
  version location or false-positive on an unrelated `version` field (e.g. a
  `$schema` URL). A pinned list is the testable contract.
- **Auto-detect the tree from the manifest content.** Rejected: the dev and public
  trees have identical file names and opposite version polarity — content-based
  detection would be circular (it would need the very version state it is checking
  to decide which polarity to enforce). An explicit `--tree` argument removes the
  ambiguity.
- **Let the checker also write/repair drift.** Rejected: writes belong to the
  single-source render path (adr-19 + fn-77/fn-80). A read-only checker keeps the
  write authority in one place and makes the gate safe to run anywhere.
- **Changelog entry as ADR-section prose only.** Rejected under the stated
  criterion: fn-80 validates the entry programmatically, so a committed schema is
  required; prose could not be loaded by the bump step.

## Consequences

- The anti-drift invariant is machine-checkable and pinned; a half-state publish
  is caught before it ships. The check direction is correct per tree without
  heuristics.
- A future adr-19 relocation of the primary version location is absorbed: the
  checker reads P1/D1 from `version-location.json`, so only the decision artifact
  changes, not this list. Relocating a *secondary* location (P2/P3/D2/D3) is a
  deliberate edit to this ADR's pinned table plus the checker.
- fn-79/fn-80 gains a callable gate and a recorded obligation: wire it into
  pre-flight and honour it under `--allow-dirty`, proven by their wiring test.
- fn-80 gains a committed changelog-entry schema to validate against, closing the
  `04-launch.md` §4 step 3 under-specification.
- `04-launch.md` §4 is reconciled to adr-19: the stale "writes the same bumped
  version into the corresponding dev manifest" wording contradicted adr-19's
  dev-stays-unversioned rule and the `render_public_manifests` contract (which
  leaves dev manifests untouched). The doc now states the present-state rule: only
  the public snapshot is versioned, and the anti-drift note points here.
