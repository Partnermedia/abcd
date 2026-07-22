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
- **`core/lifeboat/`** — the brief↔lifeboat contract. `mapping.go` is the single
  source of truth for which brief section a lifeboat fills from which source
  tier, and it is rendered into the brief's `00-meta.md` with a test asserting
  the two agree. The table is a *hypothesis*: `abcd disembark probe` measures the
  same sections against real repositories in the same `grounded`/`partial`/`blank`
  vocabulary, and the evidence is expected to revise it (adr-35, itd-88).
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
