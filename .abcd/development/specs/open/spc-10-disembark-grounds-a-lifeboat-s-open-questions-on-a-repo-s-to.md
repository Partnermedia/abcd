---
id: spc-10
slug: disembark-grounds-a-lifeboat-s-open-questions-on-a-repo-s-to
intent: itd-95
---
# disembark-grounds-a-lifeboat-s-open-questions-on-a-repo-s-to

## Summary

spc-10 delivers itd-95: a conventions-tier `Source` adapter that grounds
`evidence/open-questions` on the in-code work markers a team left in its source
tree, plus the bounded recursive-walk primitive the adapter needs. A repository
with no `.abcd/` record but a codebase full of `TODO:` and `FIXME:` now reports
that section non-blank, citing `file:line` for every marker it found, instead of
falling through to a blank a rescuer has to fill from nothing.

## Scope

- **Core primitive** (`internal/core/lifeboat/probe.go`):
  `(*SourceContext).WalkFiles(rel string) (paths []string, truncated bool)` — a
  bounded, contained recursive walk of regular files under a repo-relative root.
  It is the missing primitive both itd-95 and itd-96 raise, landed once here
  (one-canonical-primitive) and consumed by itd-96's internals adapter.
- **Adapter** (`internal/core/lifeboat/sources_conventions.go`):
  `convOpenQuestionsSource`, a sibling of `convGlossarySource`, grounding
  `evidence/open-questions` at `TierConventions`.
- **Registration** (`conventionSources()` in `probe.go`): one line.
- **No mapping change.** The `evidence/open-questions` row already names "TODO
  and FIXME markers" as a read and predicts a *partial* conventions status; this
  spec delivers the adapter that makes the prediction true, and honours the
  prediction as a ceiling rather than editing the contract.
- **No new dependency.** `regexp`, `io/fs`, and the existing `SourceContext`
  surface only.

## Approach

### The `WalkFiles` primitive

`SourceContext` today exposes a non-recursive `ListDir` and a bounded
`ReadFile`; there is no recursive walk. `WalkFiles` adds one, mirroring
`embark.go`'s `walkLifeboatFiles` — the repository's existing canonical
`fs.WalkDir(root.FS(), …)` walk — rather than inventing a second traversal
idiom.

- **Containment.** The walk runs over `c.root.FS()`, so `os.Root` refuses any
  component that escapes the repository root, symlinked intermediates included.
  A `nil` root (unopenable repository) returns `(nil, false)`.
- **Symlinks are skipped, never followed.** Unlike embark — where a symlink in a
  packed lifeboat is a trust violation and therefore fatal — a probe reads an
  arbitrary foreign tree where symlinks are ordinary. The walk skips symlinked
  entries (files and directories) and continues; it never errors on one.
- **Skip set.** Directory names never descended into: `.git`, `node_modules`,
  `vendor`, `generated`. These are dependency, VCS-internal, and generated
  trees — never a team's own open questions, and the dominant cost of an
  unfiltered walk.
- **Cap.** `maxWalkFiles` mirrors `maxDirEntries` (50 000). On reaching it the
  walk stops and returns `truncated = true`. Truncation is *reported*, never
  silent (loud-staging): an adapter that hits the cap says so in its cited
  evidence.
- **Non-regular files** (FIFOs, devices, sockets) are skipped, so the walk
  cannot hand an adapter a path whose read would block.
- **Output** is repo-relative POSIX paths, sorted, so every consumer is
  deterministic.

`WalkFiles` returns paths only; content is read through the existing
`ReadFile`, which already enforces the containment root, the
`maxProbeReadBytes` per-file cap, regular-file-only, and the non-blocking open.
The primitive therefore adds no second read path to audit.

### The marker adapter

`convOpenQuestionsSource.Probe` walks the tree from `.`, reads each file through
`ReadFile`, and records every recognised marker as a `file:line` citation.

- **Recognised markers:** `TODO`, `FIXME`, `XXX`, `HACK`, `BUG` — uppercase
  only, matched by
  `(^|[^A-Za-z0-9_])(TODO|FIXME|XXX|HACK|BUG)(:|\(|\s|$)`. The leading class
  is the word boundary that stops `TODO` matching inside `TODOS` or
  `todo_list`; the trailing class admits the two conventional spellings
  (`TODO:` and `TODO(alice):`) plus a bare word, while rejecting the common
  redaction placeholder shape (`XXX-XXX-XXX`). `NOTE` and `OPTIMIZE` are *not*
  recognised: `NOTE` marks explanation rather than unfinished work, and
  `OPTIMIZE` is rare enough that its false-positive cost exceeds its value.
- **Binary files are skipped** by a NUL byte in the first 8 KiB — the
  conventional heuristic, and dependency-free. No extension allow-list is
  maintained.
- **Caps.** Files walked: `maxWalkFiles` (inherited). Bytes per file:
  `maxProbeReadBytes` (inherited — an oversized file is skipped by `ReadFile`).
  Citations reported: `maxMarkerCitations` (200); beyond it the scan keeps
  counting but stops citing, and says so.
- **Output shape** mirrors `convADRsSource`: a headline count
  (`"N work marker(s) across M file(s)"`) followed by up to
  `maxMarkerCitations` `path:line (TODO)` entries, all passed through
  `dedupeSorted` so repeated identical citations collapse and the order is
  stable.
- **Status ceiling: `StatusPartial`.** Markers are a thread, not a framed set of
  open questions — a `TODO` says something is unfinished, not what the question
  is. This also honours the mapping row's conventions prediction. Confidence
  carries the strength instead: `ConfidenceMedium` at ten or more markers,
  `ConfidenceLow` below that.
- **No markers → a blank** carrying `Searched` (the marker set and the skip
  set) and the human `Question`, exactly as every other adapter's blank
  contract requires.
- **Read-only by construction.** The adapter opens nothing for writing and
  touches nothing outside the containment root; a byte-for-byte tree-invariance
  test proves it.

### Resolved open questions (itd-95 § Open Questions)

| Question | Decision |
|---|---|
| Which markers? | `TODO`, `FIXME`, `XXX`, `HACK`, `BUG`; uppercase only; word-boundary anchored, trailing `:`/`(`/space/EOL. `NOTE`, `OPTIMIZE` excluded. |
| Which tier — conventions or git? | **Conventions.** A working-tree file scan through the `SourceContext` file surface. The adapter never touches git, so it grounds a bare snapshot as readily as a working tree. |
| Scan scope and the missing primitive | Option (a): add a bounded recursive-walk primitive, `WalkFiles`, to `SourceContext`. It is shared with itd-96, so the walk lands once. |
| Which files to scan | Every regular file the walk yields, minus the skip set (`.git`, `node_modules`, `vendor`, `generated`), minus symlinks, minus binaries (NUL-byte heuristic), minus oversized files (`ReadFile`'s cap). |
| Per-repo caps | `maxWalkFiles` = 50 000 files (mirrors `maxDirEntries`), `maxProbeReadBytes` per file, `maxMarkerCitations` = 200 citations. Truncation is reported in the cited evidence. |
| Output shape and framing | Headline count + up to 200 `path:line (MARKER)` citations, `dedupeSorted`. |
| Dedup | `dedupeSorted` on the rendered citation string, so an identical `path:line (MARKER)` appears once. Multiple distinct markers in one file each keep their own line. |
| Status and confidence thresholds | Ceiling `StatusPartial`. `ConfidenceMedium` at ≥ 10 markers, else `ConfidenceLow`. Never `StatusGrounded`. |
| Interaction with the native adapter | Unchanged: richest-tier-wins. `nativeOpenQuestionsSource` displaces the marker evidence on a repository carrying both, deterministically, and the coverage report names the winning tier. Merging across tiers stays out of scope. |

## Acceptance-criteria satisfaction

- **Record-less repo with markers → non-blank, cites the files** —
  `convOpenQuestionsSource` over a fixture repo carrying `TODO`/`FIXME`;
  asserted non-blank, `Tier == TierConventions`, evidence carrying `file:line`.
- **No markers → an honest blank** — a marker-free fixture asserts
  `StatusBlank` with a populated `Searched` and a non-empty `Question`.
- **Read-only** — a byte-for-byte tree-invariance test (walk the fixture,
  hash every file, probe, re-hash) proves the probe writes nothing.
- **Pathological tree stays inside the caps** — a fixture exceeding
  `maxWalkFiles` asserts the walk returns `truncated = true` and terminates; a
  fixture with a symlink escaping the root asserts nothing outside the
  containment root is ever read.
- **Both tiers present → one deterministic result** — the existing
  `beats`/`tierRank` reduction is unchanged and already tested; the new
  adapter adds no second winner.
