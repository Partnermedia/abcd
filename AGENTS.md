# AGENTS.md

<!-- BEGIN ABCD -->
<!--
  Managed by abcd (Agent-Based Configuration for Development).
  Do NOT hand-edit content inside the abcd-managed fences — `/abcd:ahoy`
  silently overwrites this block on drift (per itd-3). Per-repo rule
  customisation goes in <repo>/.abcd/rules.json instead.
-->

## abcd rule loader

This repository uses the abcd modular rules loader. On `UserPromptSubmit`, a hook
recall-matches the prompt against keyword triggers declared in the plugin-bundled
default domains and `<repo>/.abcd/rules.json`, and injects only the matched
domain rules into context — instead of force-loading the full ruleset every turn.
A prompt that matches no domain injects nothing (zero added tokens).

- Inspect rules: `abcd rules` renders the active set; `abcd rules <DOMAIN>`
  (case-insensitive) scopes to one domain.
- Per-repo overrides: edit `<repo>/.abcd/rules.json`. It is
  `{"schema_version": 1, "disabled": false, "domains": {}}` — add a domain key to
  override a default per-field (e.g. `{"ROADMAP": {"state": "dormant"}}` silences
  it while keeping its rules) or to declare a custom domain
  (`{"recall": [...], "rules": [...]}`).
- Kill switch: set `"disabled": true` at the top of `.abcd/rules.json`.
- Explicit activation: start a prompt with `*<DOMAIN>` (e.g. `*COMMITTING`,
  `*PII`) to inject that domain unconditionally — overrides a `dormant` state,
  but never the kill switch.

### Default domains

`COMMITTING`, `DOCUMENTATION`, `ROADMAP`, `ISSUES`, `INTENTS`, `LIFEBOAT`, `PII`,
`OPINIONS`. Each carries recall keywords and its rules, bundled in the abcd
binary; a repo overrides them per-field via `.abcd/rules.json`. `OPINIONS`
points at the canonical conventions under `.abcd/development/principles/` rather
than copying them.

### Reset triggers

`SessionStart` and `PreCompact` clear the per-session dedup ledger, so a matched
domain re-injects on the next prompt (the event-driven refresh that recovers
after compaction). Within a session the hook does not re-inject unchanged rules.

For internals see `.abcd/development/brief/05-internals/03-configuration.md`.

<!-- END ABCD -->

abcd (Agent-Based Configuration for Development) is a Go CLI and an agent-harness
plugin: a host-agnostic **configuration layer for development**. A single
`abcd` binary holds all behaviour in a transport-agnostic core; the CLI, the
markdown plugin surface, and (later) an MCP server are thin front doors onto it.

Start with the plan and the design record:

- Design record (the specification): [`.abcd/development/`](.abcd/development/) —
  brief, roadmap/intents, decisions/adrs, research.
- Package map: [`internal/README.md`](internal/README.md).

## Build, test, and checks

Run from the repo root.

```bash
make preflight      # the pre-push gate: build + vet + test + race (internal)
make build          # cross-compiles bin/abcd-<goos>-<arch> (there is no plain bin/abcd)
gofmt -l .          # format gate: any output names a file needing `gofmt -w`
go vet ./...        # static checks
go test ./...       # unit tests
go test ./internal/core/                 # a single package
go test -run TestStatus ./internal/core/ # a single test
```

CI (`.github/workflows/ci.yml`) runs the same `check` job on macOS + Linux, plus
full-history secret scanning (`gitleaks`) and a workflow audit (`zizmor`).

## Working-tree layout (three tiers under `.abcd/`)

Development material lives under `.abcd/`; `docs/` is user-facing only.

- `.abcd/development/` — **durable record** (committed): brief, intents, ADRs,
  plans, research. Excluded from the release artifact.
- `.abcd/work/` — **shared working** (committed): `CONTEXT.md` (current
  orientation) and `DECISIONS.md` (append-only decision log; architecture-shaping
  decisions graduate to ADRs under `.abcd/development/decisions/adrs/`).
- `.abcd/.work.local/` — **local ephemeral** (gitignored): `NEXT.md` handover,
  `scratch/`, `logs/`. Per-worktree, so it never merge-conflicts.

## Boundaries

- **Transport-agnostic core.** `internal/core` never writes to stdout or knows a
  transport; front doors under `internal/surface/*` format its results.
- **Wired or it isn't done.** Every verb is reachable from both the CLI and the
  plugin markdown surface and demonstrably executes there — no dead scaffolding.
- **Host-delegated by default.** LLM review/agent work is delegated to the host;
  native/CLI/API/MCP oracles are opt-in adapters.
- **Single repo, curated release.** `.abcd/**` stays in-tree but is excluded from
  the release artifact by packaging; the repo is the plugin marketplace.
- **Never commit or push without being asked.** Substantive work goes on a branch
  and PR; new dependencies need explicit sign-off before `go get`.

## Definition of done

- `make preflight` is clean (build, `gofmt -l .` empty, `go vet ./...`,
  `go test ./...`, and `go test -race ./internal/...`).
- Every new behaviour has a test watched fail before the change and pass after.
- A CHANGELOG entry accompanies any user-facing change.

## Attribution and acknowledgements

- **AI-assisted commits carry an `Assisted-by:` trailer**, kernel format
  (`Assisted-by: Claude:claude-opus-4-8`) — disclosure, not authorship. Never
  `Co-Authored-By:` for AI (it asserts an authorship the tool does not hold and
  inflates the contributor graph). A human-only `Signed-off-by:` (DCO) is deferred
  to the public flip or the first outside contribution. The human is the author of
  record, responsible for all AI-assisted output. See `CONTRIBUTING.md`.
- **Naming a tool is confined to credit.** User-facing prose (`README.md`,
  `docs/`) stays host-agnostic — the `harness/*` docs-lint rules enforce it. The
  one sanctioned place to name a tool is attribution: the README badge and
  `ACKNOWLEDGEMENTS.md`, using the `<!-- docs-lint: allow -->` escape where a lint
  root is involved. Private, unpublished tool names never appear in any committed
  file.
- **`ACKNOWLEDGEMENTS.md`** credits ideas, tools, and writing in three parts —
  development, inspirations, references. Add an entry in the same change that lands
  it (adopts a pattern, cites a source in an ADR, integrates a tool), never later.
