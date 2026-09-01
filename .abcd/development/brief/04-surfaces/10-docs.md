# `/abcd:docs` — Documentation Currency and Citations

`/abcd:docs` runs abcd's documentation-currency checks over the current repo and
maintains the citation record those checks enforce. `lint` is **strictly
read-only** — it performs zero writes and touches no network — and it is the
deterministic half of the docs release gate (the semantic half is the
`docs-currency-reviewer` agent). The `cite` sub-tree is the writing half: it is
the only place abcd reaches the network on behalf of documentation, and it runs
when a maintainer asks.

## Sub-verbs

> _Machine-checked (`surface_coverage`, spc-27): each row records the verb's
> adr-40 bucket (`lint` / `review` / `audit` / `gate`, or `—` for a
> non-assessment verb) and its existence (`shipped` / `staged`), verified
> against the committed command-tree snapshot in both directions._

| Verb | Bucket | Status |
|---|---|---|
| `cite` | — | shipped |
| `cite confirm` | — | shipped |
| `cite refresh` | — | shipped |
| `lint` | lint | shipped |


- **`/abcd:docs lint`** — lint this repo's documentation for currency and print
  the findings. The plugin command (`commands/docs.md`) invokes
  `abcd docs lint --json` and summarises the result.
- **`abcd docs cite refresh`** — fetch every cited URL once and rewrite the
  committed citation baseline. Each URL gets exactly one bounded attempt with no
  retries, and no response body is read: liveness is judged from the status line.
  Sources that refuse automated fetchers are printed as a manual checklist rather
  than recorded as broken.
- **`abcd docs cite confirm <url>… | --receipt <file>`** — record that a human
  verified a citation the fetcher could not read. Both forms assemble one receipt
  schema, so the generated checklist page that lands later is a different
  producer of the same input rather than a second pathway. Only URLs the
  documentation cites can be confirmed.

Bare `abcd docs` prints command usage — it does **not** render a status board;
the [surfaces index](README.md) carries the one enumeration of where the
bare-status convention holds, and `docs` is not on it. The global `--json` flag emits the
machine-readable payload. The working verbs — `lint`, `cite refresh`, and
`cite confirm` — accept `--config` (path to the
`docs-lint.json` it loads, default `<root>/.abcd/docs-lint.json`) and `--root`
(repo root, default the current working directory); the bare `cite` parent
only routes to its sub-verbs and takes neither flag. So the refresh fetches
exactly the set the gate demands receipts for. `docs lint` additionally accepts
`--release-gate`, which promotes an overdue citation from a warning to a blocker;
the flag is release machinery's, so an ordinary commit is never blocked by the
calendar.

## What it checks

- **Change-narration** — prose that narrates a change rather than describing
  present state. Unambiguous change-narration **blocks** (`previously`,
  `formerly`, `renamed from`, `has been replaced`, `we switched`, `to be
  implemented`); phrases that can also describe present state **warn** advisorily
  rather than block (`deprecated`, `no longer`, `migrated from`). Docs are
  present tense: what *is*, never what *was superseded*.
- **Broken relative links** — every relative link must resolve to a file in the
  tree.
- **Stray root markdown** — no stray markdown at the repo root (it belongs under
  `docs/`; the allowed root files are the fixed set — README, CHANGELOG,
  CONTRIBUTING, etc.).
- **Citations** — where a repo arms the rules: footnote markers and definitions
  in bijection, every crosswalk table row carrying a footnote, well-formed cited
  URLs and DOIs, refused source domains, and the committed baseline at
  `.abcd/citations-baseline.json` — no cited URL without a receipt, none recorded
  broken, none whose recorded final address has drifted from what the page cites,
  and a staleness warning past 180 days. All of it reads committed files; the
  fetching lives in `cite refresh`.
- **Host-agnostic prose** — user-facing docs must not name a specific agent
  harness or bundled tool. This repo's `.abcd/docs-lint.json` defines a family of
  `harness/*` banned tokens (each a **blocker**) that catch such names, so the
  published surface stays host-agnostic; the `<!-- docs-lint: allow -->` escape
  covers the sanctioned exception (attribution).

## Output

The `--json` payload carries `blockers` (a count; any blocker fails the gate) and
`findings` (each with `File`, `Line`, `RuleID`, `Severity`, `Message`). A
`blockers` value of zero means the docs are currency-clean. The command exits
non-zero when a blocker is present, so it composes directly into CI and the
release gate.

## Composition

`/abcd:docs lint` is the deterministic, fast, always-runnable currency check.
The `docs-currency-reviewer` agent is its semantic complement — it verifies that
every user-facing claim still matches the code, which a structural lint cannot.
The release gate runs both: `docs lint` (deterministic) and the reviewer
(semantic) must each pass before a tag.

`cite refresh` composes with the gate by separation: the gate is deterministic
because the fetching happens elsewhere and arrives as a committed record a
reviewer reads in a diff. The baseline's age surfaces at `abcd ahoy` and in the
`abcd launch --dry-run` preflight, which names entries approaching the staleness
blocker while a release still cuts, and refuses on ones past it.

## References

- Plugin command: [`commands/docs.md`](../../../../commands/docs.md)
- Lint engine: `internal/core/lint`
- The documentation invariants it enforces: [`../02-constraints`](../02-constraints)
