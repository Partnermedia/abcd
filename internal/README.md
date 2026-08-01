# internal/

abcd's Go implementation. The organising rule is a **transport-agnostic core**:
all behaviour lives in `core/` as functions that take a structured request and
return a structured result, and the front doors under `surface/` only marshal
those results for their transport. This is what lets the CLI, the markdown
plugin surface, and a future MCP server share one engine.

## Package map

- **`core/`** — the engine. One package per capability; no stdout, prompt text,
  or transport coupling. Currently: identity/version and the read-only status
  snapshot. Grows per phase (ahoy, launch, capture, memory, intent, brief,
  review, spec, run, lifeboat, history, changelog).
- **`core/changelog/`** — release derivation. Holds `impact`, the one product
  judgement a record declares, and derives the next SemVer from the records that
  entered the terminal folders since the anchor tag. It owns the enum so the
  lints that GATE the judgement (`core/lint`), the ledger reader that VALIDATES
  it (`core/capture`), and the derivation that CONSUMES it cannot drift apart.
- **`core/surface/`** — the compatibility surface as DATA: the snapshot of every
  command, flag, and manifest entry a consumer binds to, and the diff that names
  what a release narrowed. It shares a word with the `surface/` front-door tier
  and nothing else — that tier is about transports, this package is about what
  those transports expose — so it is cobra-free like the rest of `core/`. The
  walk that reads the live command tree needs cobra and therefore lives in
  `surface/cli`, which hands its result in; the dependency never points back.
- **`core/cite/`** — the live half of the citation gate: the bounded fetcher and
  the refresh that writes the committed baseline `core/lint` then enforces with
  zero network. It is the only place abcd dials out on behalf of documentation,
  and it holds the seam a specialist link checker would later slot into — the
  baseline schema and the lint rules are the contract, the fetcher is a
  replaceable producer.
- **`core/lifeboat/`** — the brief↔lifeboat contract. `mapping.go` is the single
  source of truth for which brief section a lifeboat fills from which source
  tier, and it is rendered into the brief's `00-meta.md` with a test asserting
  the two agree. The table is a *hypothesis*: `abcd disembark probe` measures the
  same sections against real repositories in the same `grounded`/`partial`/`blank`
  vocabulary, and the evidence is expected to revise it (adr-35, itd-88).
- **`core/guard/`** — the shell-hazard registry. Bundled hazard entries (id,
  command-position pattern over shell tokens, blocker/warn tier, safe successor,
  plain-language why) plus the deterministic allow/warn/block decision a harness
  hook calls before a command runs. Matching is token-aware and command-position
  only, so a hazard named inside a quoted argument never fires; every bundled
  entry ships known-bad and known-good fixtures, and the admission gate test
  holds each one to its own corpus (100% true-negative floor, at least 40%
  known-good). An allow means no entry matched, never that a command is safe: a
  hazard reached another way — a string handed to an interpreter (`eval`,
  `sh -c`), a launcher outside the `wrappers` set, a launcher inside it carrying
  a value-taking flag `wrapperValueFlags` does not name, an `ArgPaths` root
  segment the host serves under a prefix (a GitHub Enterprise Server `/api/v3/`
  mount; `pathOf` normalises a `scheme://host` URL, not a mount point), a
  backtick substitution, or a form no entry describes — is not seen. Each limit
  is stated here and in
  the `abcd guard check` reference doc, never left implicit. `Load` reads the
  working tree, so a `disabled: true` takes effect before it is reviewed; the
  front door compensates by making a disabled registry loud rather than silent.
  Fail-open-loud on a broken guard belongs to the hook shim (`hooks/hooks.json`)
  and the `abcd guard hook` adapter, not here.
- **`core/banlist/`** — the two banned-names stores (itd-74, spc-20). The public
  layer is managed IN the docs-lint `banned_tokens` family under a `names/` id
  prefix: one banned-token primitive, and the prefix is the ownership boundary a
  removal respects. The private layer is the gitignored per-machine store, whose
  FIRST line declares its format — keyed (`KEY<space-or-tab>PATTERN`) or legacy (one
  whole-line pattern per line) — which the committed `.githooks/pre-commit` guard
  parses identically. Three shared fixture corpora under `testdata/` are read by both
  parsers, so their agreement on the keyed format, the legacy format, and every
  unusable line class is checked rather than assumed. Enforcement is NOT here: the
  hook is the private layer's enforcement point and `core/lint` the public one, so
  this package owns the stores, their formats, and their editing discipline. It does
  VALIDATE against each layer's own engine, though — a private pattern goes to `grep`
  on stdin, a public one through the linter's compile path — because a pattern checked
  against a third engine is stored as healthy while it matches nothing. Redaction is
  structural — the exported private entry type carries no pattern field, so no
  rendering can leak a value — and edits are surgical (a line for the private store,
  byte surgery on the located array for the config), never a whole-file re-marshal.
  Scaffolding is NOT here either: `core/ahoy` writes the five artefacts (the two
  committed guard hooks, the `.gitattributes` line keeping them at LF, the public
  family, and the gitignored stub) into a repo it configures, importing this package
  for the paths and for the one canonical statement of the private layer's reach.
- **`surface/cli/`** — the default front door: a Cobra command tree that calls
  `core` and formats results as text or `--json`. Holds no business logic.
- **`surface/mcp/`** *(later)* — an additive front door exposing the same core
  verbs as `mcp:abcd:*` tools. Added once a surface is worth exposing; no core
  rework required because the core is transport-agnostic.

## Planned seams (added when a phase consumes them, never as dead scaffolding)

Per the project rule "wired or it isn't done," the pluggable adapter seams are
introduced by the phase that first uses them, not pre-emptively:

- **`adapter/oracle/`** — LLM review. Native default = host-delegated (the host
  runs subagents); opt-in plug-ins: claude CLI, Anthropic API, MCP oracle.
- **`adapter/history/`** — transcripts. Native default = redacted local store;
  opt-in: private companion/remote, specstory cloud.
- **`adapter/spec/`** — spec/task store. Native minimal default; opt-in: the companion harness
  `ccpm` at the convention level.
- **`adapter/run/`** — autonomous run. Native thin loop fallback; host backends:
  Claude Workflows, the companion harness's agent loop.
- **`adapter/scanner/`** — secret/PII. Native default; opt-in: gitleaks,
  trufflehog.
- **`registry/`, `config/`** — declarative wiring of chosen adapters.

The full rationale is in the plan and the design record under
`.abcd/development/`.
