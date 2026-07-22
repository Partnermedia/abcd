# CLI command reference

This page is generated from the abcd command tree by `GenerateReference` in
`internal/surface/cli`. It is a derived artefact: do not edit it by hand. A
drift test regenerates the tree and fails the build whenever this page and the
tree disagree, so the reference can never silently go stale. Regenerate it with
`go generate ./internal/surface/cli`.

Every user-facing command is listed with its usage line, summary, and flags;
the operator-internal hook entrypoints are omitted.

## `abcd`

Agent-based configuration for development

**Usage:** `abcd`

**Flags:**

```
      --json   emit machine-readable JSON
```

### `abcd ahoy`

Install/update abcd in this repo; bare invocation is read-only status

**Usage:** `abcd ahoy`

#### `abcd ahoy doctor`

Report every gap read-only, including user-scope state (never mutates)

**Usage:** `abcd ahoy doctor`

#### `abcd ahoy dry-run`

Render the detection-result JSON envelope; never mutates

**Usage:** `abcd ahoy dry-run`

#### `abcd ahoy identity-check`

Exit non-zero if the git commit identity does not match .abcd/config/identity.json

**Usage:** `abcd ahoy identity-check`

#### `abcd ahoy install`

Install or update abcd in this repo (idempotent)

**Usage:** `abcd ahoy install [flags]`

**Flags:**

```
      --adopt                   adopt an unmanaged repo without prompting
      --docs-target string      marker target: claude_md | agents_md | both | skip
      --oracle-backend string   oracle backend: host-delegated | native | cli | api | mcp
      --refuse-adopt            decline to adopt an unmanaged repo
      --scan-deep string        enable deep scan: true | false
      --visibility string       repo visibility: private | public
      --yes                     approve every resolvable change category without prompting
```

#### `abcd ahoy uninstall`

Remove the marker block and owned PATH symlink (leaves .abcd/ intact)

**Usage:** `abcd ahoy uninstall`

### `abcd audit`

Check this repo against the working conventions (read-only)

**Usage:** `abcd audit [flags]`

**Flags:**

```
      --root string   repo root to audit (default: current working directory)
```

### `abcd capture`

Capture issues to the ledger; bare invocation is read-only status

**Usage:** `abcd capture [text] [flags]`

**Flags:**

```
      --blocked-by string     comma-separated iss-ids this issue is blocked by
      --category string       issue category (default observation)
      --found-at string       optional repo-relative path or conceptual location
      --found-during string   session/command context (default manual-capture)
      --severity string       severity: nitpick | minor | major | critical (default minor)
      --slug string           override the slug derived from the text
      --source string         surfacing channel (default user-observation)
```

#### `abcd capture list`

List issues by state (one of --open/--resolved/--wontfix/--all required)

**Usage:** `abcd capture list [flags]`

**Flags:**

```
      --all        issues across all three states
      --open       issues currently in open/
      --resolved   issues currently in resolved/
      --wontfix    issues currently in wontfix/
```

#### `abcd capture resolve`

Mark an open issue resolved (open/ -> resolved/)

**Usage:** `abcd capture resolve <iss-N> <note>`

#### `abcd capture wontfix`

Record an explicit non-action decision (open/ -> wontfix/)

**Usage:** `abcd capture wontfix <iss-N> <reason>`

### `abcd changelog`

Preview the next release cut — derived version, records, guardrail (read-only, no prose)

**Usage:** `abcd changelog`

### `abcd disembark`

Lifeboat tooling: coverage probe, pack dry-run, and out-of-tree pack

**Usage:** `abcd disembark`

#### `abcd disembark coverage`

Aggregate probe reports into the cross-repo section×repo coverage table

**Usage:** `abcd disembark coverage <report.json>...`

#### `abcd disembark graveyard`

Validate host-produced lesson JSON against a packed lifeboat and write the survivors (cite-or-be-dropped)

**Usage:** `abcd disembark graveyard <lifeboat-dir> --lessons-json <file|-> [flags]`

**Flags:**

```
      --lessons-json string   path to the host-produced lesson JSON (or - for stdin)
```

#### `abcd disembark oracle`

Audit a packed lifeboat against its source repo — a registered verdict and cited findings (deterministic, or validate host-produced audit JSON)

**Usage:** `abcd disembark oracle <lifeboat-dir> <source-repo> [--oracle-json <file|->] [flags]`

**Flags:**

```
      --oracle-json string   path to host-produced audit JSON (or - for stdin); absent runs deterministic mode
```

#### `abcd disembark pack`

Pack a lifeboat from a repository into a destination directory (writes <dest>, never the source)

**Usage:** `abcd disembark pack <repo> <dest>`

#### `abcd disembark plan`

Show the full lifeboat file set a pack would write, without writing anything (dry run)

**Usage:** `abcd disembark plan [repo]`

#### `abcd disembark press-release`

Compose the lifeboat's press release (deterministic from the brief/spine, or validate host-produced press-release JSON)

**Usage:** `abcd disembark press-release <lifeboat-dir> [--press-release-json <file|->] [flags]`

**Flags:**

```
      --press-release-json string   path to host-produced press-release JSON (or - for stdin); absent runs deterministic mode
```

#### `abcd disembark principles`

Distil principles from a packed lifeboat (deterministic from the ADRs, or validate host-produced principle JSON)

**Usage:** `abcd disembark principles <lifeboat-dir> [--principles-json <file|->] [flags]`

**Flags:**

```
      --principles-json string   path to host-produced principle JSON (or - for stdin); absent runs deterministic mode
```

#### `abcd disembark probe`

Report which brief sections a lifeboat could ground from a repository (read-only)

**Usage:** `abcd disembark probe [repo]`

### `abcd docs`

Documentation-currency checks for this repo

**Usage:** `abcd docs`

#### `abcd docs lint`

Lint docs for change-narration, broken links, and stray root markdown

**Usage:** `abcd docs lint [flags]`

**Flags:**

```
      --config string   path to docs-lint.json (default: <root>/.abcd/docs-lint.json)
      --root string     repo root to lint (default: current working directory)
```

### `abcd embark`

Unpack a lifeboat's record families back into a target repo (probe read-only; from writes)

**Usage:** `abcd embark`

#### `abcd embark from`

Write a lifeboat's record families into a target repo; refuses on any conflict

**Usage:** `abcd embark from <lifeboat-dir> [target-dir]`

#### `abcd embark probe`

Report what a lifeboat would write into a target, read-only (coverage blanks first)

**Usage:** `abcd embark probe <lifeboat-dir> [target-dir]`

### `abcd history`

Manage the native session-transcript store

**Usage:** `abcd history`

#### `abcd history capture`

Redact and store a raw session transcript (reads a file or stdin)

**Usage:** `abcd history capture [<transcript-file>|-] [flags]`

**Flags:**

```
      --kind string      source kind: native | specstory-import (default native)
      --session string   session id for the record (default: transcript filename; required for stdin)
```

#### `abcd history list`

List stored transcripts for this repo, newest first

**Usage:** `abcd history list`

#### `abcd history show`

Show one stored transcript's metadata and redacted body

**Usage:** `abcd history show <session-id-or-filename>`

### `abcd intent`

Intent lifecycle; bare invocation is read-only status, quoted text files a draft

**Usage:** `abcd intent [text]`

#### `abcd intent link`

Link a planned intent to an existing spec (writes the intent's spec_id)

**Usage:** `abcd intent link <itd-N> <spc-N>`

#### `abcd intent new`

Deprecated alias for `abcd intent "<text>"` (files a draft from the text)

**Usage:** `abcd intent new <text>`

#### `abcd intent plan`

Plan a draft intent: mint its spec, link both sides, move drafts -> planned

**Usage:** `abcd intent plan <itd-N>`

#### `abcd intent ready`

Report whether an intent is ready to implement (planned + AC + written spec); exit 1 when not

**Usage:** `abcd intent ready <itd-N>`

#### `abcd intent review`

Fidelity review: re-emit a shipped intent's request, or ingest a verdict

**Usage:** `abcd intent review [<itd-N>]`

##### `abcd intent review ingest`

Ingest an intent-fidelity verdict JSON into the shipped intent's Audit Notes

**Usage:** `abcd intent review ingest --verdict-json <path> [flags]`

**Flags:**

```
      --verdict-json string   path to the intent-fidelity verdict JSON
```

### `abcd launch`

Preview the public launch bundle and release gates (read-only)

**Usage:** `abcd launch [flags]`

**Flags:**

```
      --dry-run   preview the launch bundle and gates without publishing
```

#### `abcd launch ship`

Cut a release: derive the version and the record set from what shipped (exit 1 when the cut refuses)

**Usage:** `abcd launch ship [--changelog-json <file|->] [flags]`

**Flags:**

```
      --changelog-json string   path to the host-composed changelog JSON (or - for stdin); absent runs the deterministic emit step
```

### `abcd memory`

Curated knowledge substrate; bare invocation is read-only status

**Usage:** `abcd memory`

#### `abcd memory ask`

Query memory and synthesise a cited answer

**Usage:** `abcd memory ask <question> [flags]`

**Flags:**

```
      --file-back          file the synthesised answer back as a new memory page
      --page-json string   the answer page dict as JSON (file path, or - for stdin)
      --top-n int          retrieval depth (0 uses the pinned default)
```

#### `abcd memory ingest`

Distil an external source into cited memory pages

**Usage:** `abcd memory ingest <path-or-url> [flags]`

**Flags:**

```
      --keep-original       store the original at .abcd/memory/sources/<sha256>.<ext>
      --pages-json string   DistilledPage JSON array (file path, or - for stdin)
```

#### `abcd memory lint`

Curator health-check over the whole memory store

**Usage:** `abcd memory lint`

### `abcd rules`

Render the active rule set; a positional DOMAIN scopes to one (read-only)

**Usage:** `abcd rules [domain]`

### `abcd spec`

Native spec store; bare invocation is read-only status

**Usage:** `abcd spec`

#### `abcd spec close`

Close a spec (open/ -> closed/) and ship its linked intent (planned/ -> shipped/)

**Usage:** `abcd spec close <spc-N>`

### `abcd version`

Print abcd's version

**Usage:** `abcd version`
