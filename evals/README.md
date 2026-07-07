# evals — smoke harness

A self-discovering smoke test for the `abcd` binary. It builds the real binary,
walks the Cobra command tree **in-process** (via `cli.NewRootCommand()`) to
discover every command and flag, and exercises each against the built binary — so
a command added tomorrow is covered here with no edit.

## What it checks (v1)

- **Every** command and subcommand: `abcd <cmd> --help` exits 0, produces output,
  and never panics. This catches the failure unit tests miss — a command that
  compiles but crashes when actually invoked.
- **Read-only, no-argument verbs** (`version`, the bare status board) run for real
  to a graceful exit.
- **Flag hygiene:** an unknown flag is a clean non-zero error, not a panic.

## Running it

Gated behind the `smoke` build tag so it stays out of the fast unit-test lane:

```bash
make smoke
# or
go test -tags smoke ./evals/...
```

CI runs it as the dedicated `smoke` job, and the release workflow smokes the
binary built from the tagged commit before publishing.

## `data/` (reserved for v2)

Fixture-driven, per-command scenarios — user-specified and synthetic inputs the
harness auto-discovers to drive richer smokes (e.g. `memory ingest` over a sample
corpus, `capture` round-trips). Deferred; the generalisation into an abcd-managed
eval framework is captured as intent **itd-75**.
