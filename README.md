# abcd

**Agent-Based Configuration for Development** — a host-agnostic configuration
layer for development, delivered as a single Go binary that is also a Claude Code
and the companion harness plugin.

abcd holds all behaviour in a transport-agnostic core; the CLI is the reliable
default front door, the markdown plugin surface shells out to it, and an MCP
server can be added on the same core. It depends on no external tools — the
capabilities it once borrowed (transcript capture, review oracle, spec/task
engine, autonomous run) each ship as a native default with an optional external
plug-in for more power.

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
  surface (auto-loaded by Claude Code and the companion harness).
- [`.abcd/`](.abcd/) — the development record and working files (never shipped).

Contributor guidance: [`AGENTS.md`](AGENTS.md).

## Licence

MIT. See [`LICENSE`](LICENSE).
