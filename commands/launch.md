---
name: launch
description: Preview the public launch — the file bundle, the secret/PII scan, and the release gates — in dry-run mode, cut a release by deriving its version and composing its changelog, and scaffold the changelog-driven release gate into a managed repo. The preview performs zero writes; `ship` writes the dated CHANGELOG heading and never publishes; `scaffold` writes the release workflows and never publishes.
argument-hint: "[dry-run | ship | scaffold]"
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
"${CLAUDE_PLUGIN_ROOT}/abcd" launch --dry-run --json
```

Then summarise the JSON for the user:

- `version` — the version the release would carry.
- `bundle.files` — the files the bundle would include (an array; report its length as the count).
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

Those three steps write the CHANGELOG heading. They do **not** finish the
release. Two host-run semantic passes must also run and record receipts, and the
release branch has to carry them in a second commit — see *Semantic receipts*
below. A branch that skips them merges and tags cleanly and then fails at
`release.yml`'s fail-closed receipt gate, which is the most expensive place to
find out: the tag is already created by then, and the workflow never moves a tag.

### 1. Emit the cut (deterministic, writes nothing)

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" launch ship --json
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
"${CLAUDE_PLUGIN_ROOT}/abcd" launch ship --changelog-json <path>   # or - for stdin
"${CLAUDE_PLUGIN_ROOT}/abcd" launch ship --changelog-json <path> --payload-dir <dir>   # also stage the payload
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

## Semantic receipts — the second half of the cut

`release.yml` arms `receipt_gate` fail-closed against the release **content**
commit. It refuses unless every required gate has a PROMOTE receipt naming that
exact commit, pinning a judge model, and declaring the matching detector. CI
cannot produce these: they spawn LLM agents, so they are host-run, and an un-run
pass is never a silent pass.

**The release branch is exactly two commits.** A receipt names the commit its
reviewer read and must live in a LATER commit — it can never sit in the tree of
the commit it names, because adding it would change that commit's sha. So:

1. **The CHANGELOG roll** — the release-content commit, written by the three
   steps above. This is what the reviewers read.
2. **The receipts** — a commit recording the semantic verdicts that name commit 1.

On merge, `release.yml` derives the content commit as `<merge>^2^` and finds its
receipts in the released tree. A one-commit branch breaks this: the single commit
is taken as the receipts commit, the gate arms against whatever preceded it, and
no receipt names that commit.

### Running the passes

Run both in the agent harness against commit 1, then commit their receipts:

- **`docs-currency-reviewer`** — verifies every user-facing claim still matches
  the code. The agent is `agents/docs-currency-reviewer.md`.
- **`iss35-brief-surface-crosscheck`** — the brief's surface prose against the
  shipped binary's actual behaviour. Its scope, depth and prompt are pinned by
  [`.abcd/development/release-gate/manifest.json`](../.abcd/development/release-gate/manifest.json),
  and a receipt echoes that file's sha256 as `manifestHash`.

Receipts live at `.abcd/work/reviews/<content-sha>/<gate>.json`. The shape is
[`.abcd/development/release-gate/receipt.example.json`](../.abcd/development/release-gate/receipt.example.json);
the full procedure, including the tiered depth a release's impact class requires,
is the adr-37 runbook at
[`.abcd/development/release-gate/README.md`](../.abcd/development/release-gate/README.md).

A receipt is bound to its gate by its `policy.detector` value, not its filename,
and a mismatched, malformed, HOLD, model-less or wrong-detector receipt blocks.
Report a HOLD to the user and stop: a HOLD is a result, not an obstacle to route
around, and the receipts cannot be hand-written to unblock a release.

## Scaffold — the release-gate scaffolder

`scaffold` writes the changelog-driven release machinery into a **managed repo that
lacks it** — a different job from `ship` (which cuts a release in a repo that
already has the machinery). It **never publishes**.

```bash
"${CLAUDE_PLUGIN_ROOT}/abcd" launch scaffold --json
```

It writes three files, wired to the repo's own default branch and Go version:

- `.github/workflows/release.yml` — verify → build → publish, the verify gate
  armed against the reviewed **content** commit (`HEAD^2^` on the auto-release
  merge path, `HEAD^` on a direct tag), so the first public release cannot hit the
  receipt-vs-tag self-reference.
- `.github/workflows/auto-release.yml` — newest dated CHANGELOG heading → tag that
  commit → call `release.yml`. `GITHUB_TOKEN`-only, no personal access token.
- `.abcd/development/release-gate/README.md` — the adr-37 runbook.

The workflows come from one embedded template that abcd-cli's own release
workflows are regenerated from (self-scaffold parity), so every abcd release
exercises the exact machinery a managed repo receives. The scaffolded `release.yml`
carries a `workflow_dispatch` **rehearsal**: run it green once before the first
real release — it arms the full gate against a simulated changelog roll and
reviewed-content commit, proves the gate admits, and publishes nothing (no tag,
Release, or attestation).

It is idempotent and fail-safe. Exit codes gate the flow:

- **0** — every file written, or already current (a no-op re-run). Report the
  per-file disposition.
- **1** — a file exists and **differs** (hand-edited or stale); the report names
  it and **nothing was written**. Relay it; re-run with `--confirm` to overwrite.
- **2** — a structural fault (the repository or a template could not be read).

A refusal is a result to relay, not a crash. Never hand-edit the workflows to work
around it: re-run with `--confirm` when the operator intends to replace the drift.

**Binary resolution.** Run `"${CLAUDE_PLUGIN_ROOT}/abcd"` — a plugin install
provisions the binary into the plugin root, so this is the rung that fires for a
plugin user. If that path does not exist, try `abcd` on `PATH`; if that fails
too, you are in a source checkout of this repo, where — and only there —
`go run ./cmd/abcd` works, the published payload carrying no `cmd/`. To put a
binary on `PATH`, run `ahoy install` through whichever rung just resolved:
`"${CLAUDE_PLUGIN_ROOT}/abcd" ahoy install`, `abcd ahoy install`, or
`go run ./cmd/abcd ahoy install` in a source checkout.

**User input:** $ARGUMENTS
