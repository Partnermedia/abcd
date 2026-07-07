<div align="center">

  <img src="docs/assets/img/logo.png" alt="abcd logo" width="150">

  <h1>Agent-Based Configuration for Development</h1>

  <p>An opinionated, intent-driven development framework for <a href="https://x.com/signulll/status/2030404483897815089">product thinkers</a>.</p>

  <!-- Static badges only: shields.io cannot read a private repo's API, so dynamic
       github/* badges (release, last-commit) render as errors while private. Re-add
       them at the public flip, when the API is readable. -->
  <a href="https://github.com/REPPL/abcd-cli/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green" alt="License: MIT"></a>
  <img src="https://img.shields.io/badge/status-experimental-orange" alt="Status: experimental">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25">
  <a href="https://claude.ai/claude-code"><img src="https://img.shields.io/badge/Built_with-Claude_Code-3B5CE7?logo=anthropic&logoColor=white" alt="Built with Claude Code"></a> <!-- docs-lint: allow — attribution names the tool by design (see ACKNOWLEDGEMENTS.md) -->
  <br />
  <img src="https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white" alt="macOS">
  <img src="https://img.shields.io/badge/Linux-core%20CI--tested-FCC624?logo=linux&logoColor=black" alt="Linux: core CI-tested">

</div>

---

**Agent-Based Configuration for Development** — a host-agnostic configuration
layer for development, delivered as a single Go binary.

abcd holds all behaviour in a transport-agnostic core; the CLI is the reliable
default front door, the markdown plugin surface shells out to it, and an MCP
server can be added on the same core. It depends on no external tools — the
capabilities it once borrowed (transcript capture, review oracle, spec/task
engine, autonomous run) each ship as a native default with an optional external
plug-in for more power.

<div align="center">
  <img src="docs/assets/img/intro.png" alt="abcd — a product thinker holds the why; a facilitator translates it into work AI agents can act on">
</div>

## Status

Early development (Phase 0 — foundations). This is the from-scratch Go rebuild;
the prior Python implementation is frozen and read-only.

## Build

```bash
make preflight   # build + vet + test + race (the pre-push gate)
go run ./cmd/abcd            # bare status board for the current directory
go run ./cmd/abcd version    # print the version
make build                   # cross-compile bin/abcd-<goos>-<arch>
```

## Layout

- [`cmd/abcd/`](cmd/abcd/) — CLI entry point.
- [`internal/`](internal/) — the engine (`core/`) and front doors (`surface/`);
  see [`internal/README.md`](internal/README.md).
- [`commands/`](commands/), [`.claude-plugin/`](.claude-plugin/) — the plugin
  surface (auto-loaded).
- [`.abcd/`](.abcd/) — the development record and working files (never shipped).

Contributor guidance: [`AGENTS.md`](AGENTS.md).

## Licence

MIT. See [`LICENSE`](LICENSE).
