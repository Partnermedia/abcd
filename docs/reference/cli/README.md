# CLI reference

The full command reference lives in [`commands.md`](commands.md): every
user-facing `abcd` command with its usage line, summary, and flags.

That page is generated from the Cobra command tree in `internal/surface/cli`, so
it always matches the binary. A drift test regenerates the tree on every `go test`
run and fails the build whenever the committed page and the tree disagree, so the
reference stays in step with the code.

To refresh the page after changing a command, run:

```bash
go generate ./internal/surface/cli
```

For interactive help, the binary also documents itself: `abcd --help` and
`abcd <verb> --help` (e.g. `abcd disembark --help`).
