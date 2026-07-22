# `/abcd:docs` — Documentation-Currency Lint

`/abcd:docs` runs abcd's documentation-currency checks over the current repo and
presents the result. It is **strictly read-only** — it performs zero writes, it
only reports. It is the deterministic half of the docs release gate (the semantic
half is the `docs-currency-reviewer` agent).

## Sub-verbs

The shipped verb surface is one sub-verb, `lint`:

- **`/abcd:docs lint`** — lint this repo's documentation for currency and print
  the findings. The plugin command (`commands/abcd/docs.md`) invokes
  `abcd docs lint --json` and summarises the result.

Bare `abcd docs` prints command usage (its one sub-verb, `lint`) — it does **not**
render a status board; the bare-status convention is scoped to `ahoy`/`capture`/
`memory`/`intent`/`spec` and bare `abcd`, not to `docs`. The global `--json` flag
emits the machine-readable finding list, and `docs lint` additionally accepts two
local flags: `--config` (path to the `docs-lint.json` it loads, default
`<root>/.abcd/docs-lint.json`) and `--root` (repo root to lint, default the
current working directory).

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

`/abcd:docs` is the deterministic, fast, always-runnable currency check.
The `docs-currency-reviewer` agent is its semantic complement — it verifies that
every user-facing claim still matches the code, which a structural lint cannot.
The release gate runs both: `docs lint` (deterministic) and the reviewer
(semantic) must each pass before a tag.

## References

- Plugin command: [`commands/abcd/docs.md`](../../../../commands/abcd/docs.md)
- Lint engine: `internal/core/lint`
- The documentation invariants it enforces: [`../02-constraints`](../02-constraints)
