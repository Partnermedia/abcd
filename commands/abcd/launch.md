---
name: launch
description: Preview the public launch — the file bundle, the secret/PII scan, and the release gates — in dry-run mode, and cut a release by deriving its version and composing its changelog. The preview performs zero writes; `ship` writes the dated CHANGELOG heading and never publishes.
argument-hint: "[dry-run | ship]"
---

# `/abcd:launch` release preview and release cut

Two flows over the abcd binary, kept apart on purpose:

- **preview** (`dry-run`) — the bundle, the scan, and the gates. **Zero writes.**
- **ship** — the release cut: derive the version from what shipped, compose the
  changelog prose, write the dated heading. It writes **one** file, `CHANGELOG.md`,
  and **never publishes**.

Neither flow publishes. Tagging is `.github/workflows/auto-release.yml`'s job, and
it reads the dated heading `ship` writes.

## Preview (`dry-run`)

Run:

```bash
abcd launch --dry-run --json
```

Then summarise the JSON for the user:

- `version` — the version the release would carry.
- `files` — how many files the bundle would include.
- `scan.hard_fails` — secret/PII findings that would block the release.
- `smoke.ok` — whether the payload would install: both plugin manifests parse,
  the marketplace source resolves, and every declared command, agent, skill and
  hook path is carried. `smoke.findings` names any path that is not.
- `would_publish` — whether every gate passes.
- `would_refuse_on` — if non-empty, the gates that would refuse, so the user
  knows what to fix before a real launch.

This is preview-only: publishing is not driven from this command.

## Ship — the release cut

A release cut is **three steps over two Go entry points**, with a host-run agent in
the middle. It is the `disembark` synthesis shape: a deterministic step, a
delegated composition, a validating ingest.

### 1. Emit the cut (deterministic, writes nothing)

```bash
abcd launch ship --json
```

The binary derives everything the release is allowed to be: the base tag, the
`next_tag`, the deciding `impact`, the record set (`added` and `removed`), and the
surface guardrail's verdict. Read-only preview of the same thing:
`abcd changelog --json`.

Exit codes gate the flow:

- **0** — the cut is ready. Continue to step 2.
- **1** — the cut **REFUSES**. Render the whole report to the user and **stop**.
  Every refusal names the specific record, version, or surface that blocks it — a
  release in flight, a merged feature whose intent still sits in `planned/`, a
  missing surface baseline, a surface break with no `breaking` record. A refusal is
  a result to relay, not a crash, and not something to work around.
- **2** — a structural fault (the repository could not be read). Relay it and stop.

### 2. Compose the prose (host-delegated)

Run the **`release-changelog-composer`** agent
(`agents/release-changelog-composer.md`) over the emitted cut and the records it
names. It returns the changelog payload: `schema_version`, `prompt_version`,
`next_tag` echoed verbatim, and `entries[{section, records, text}]`. The agent owns
the **wording** and the **Keep a Changelog section**; the version, the date, the
heading, the section order, and the inclusion set stay the binary's.

**LOUD STAGE — if the composer cannot run in this context, the flow STOPS here.**
No fallback exists and none may be improvised:

- Do **not** hand-write the changelog lines yourself. Hand-written prose is the
  exact thing this flow abolishes, and prose written outside the composer carries a
  `prompt_version` that traces to no prompt — a provenance lie the payload has no
  way to express.
- Do **not** write a partial section, and do **not** invoke the ingest step with a
  payload covering some of the records "for now". The bijection would refuse it
  anyway; a partial cut is not a smaller release, it is a false one.
- Do **not** edit `CHANGELOG.md` by hand to unblock the release.

Say plainly that the composer is unreachable, that **nothing was written**, and
that the cut from step 1 is still valid and can be shipped once it is reachable.

### 3. Ingest the prose and write the heading

Write the agent's payload to a file and hand it back to the binary:

```bash
abcd launch ship --changelog-json <path>   # or - for stdin
abcd launch ship --changelog-json <path> --payload-dir <dir>   # also stage the payload
```

With `--payload-dir` the binary additionally stages the release payload in that
directory — an empty directory outside the repository — with the derived version
stamped into the payload's copies of `plugin.json` and `marketplace.json`. The
repository's own manifests are never touched: they carry no version, and the
version belongs to the artefact. The staged payload is proved consistent before
the command returns, so a stamp that missed a pinned location is a refusal rather
than a published half-state. Every refusal the staging step can make is checked
BEFORE the dated heading is written, and a refusal that slips past that check
rolls the heading back — so a ship that exits non-zero leaves no release record
behind for the next attempt to trip over. Without the flag nothing is staged;
`--payload-dir` on its own (no `--changelog-json`) is an operand error, because
only a completed cut has a version to stamp.

The binary re-derives the cut, then proves the prose describes it — the
**completeness bijection**: the set of record ids the payload cites must equal
`(added ∪ removed)` minus the records marked `in_changelog: false`. On any
mismatch it writes **nothing** and names three groups apart, because the fix
differs for each:

- **MISSING** — a shipped record no line cites; the release record would lie by
  omission.
- **INVENTED** — a cited id that is not in this cut at all.
- **INTERNAL** — a cited id that *is* in the cut but declares `impact: internal`;
  those records earn no changelog line.

On success it splices the dated section directly beneath `## [Unreleased]` — the
insertion anchor, which must **exist** and be **empty** (a derived cut never folds
hand-written prose into a generated section) — in one atomic write.

Exit codes, same shape as step 1:

- **0** — written. Report `heading`, `path`, `lines`, and `cited` (the proof set).
- **1** — the **cut** refuses. Nothing was written, whatever the payload said.
- **2** — the payload is unusable, including a failed bijection. Relay the report;
  the file is byte-identical to what it was.

Then show the user the written heading and the diff, so a human reviews the release
record before it is committed. This command never commits, tags, or publishes.

If the `abcd` binary is not on `PATH`, fall back to
`go run ./cmd/abcd launch ...` from the repo root, or tell the user to build it
with `make build`.

**User input:** $ARGUMENTS
