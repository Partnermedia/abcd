# commands/

The plugin command surface, auto-loaded by compatible agent harnesses. Each markdown
file is a slash command whose body instructs the host agent to invoke the `abcd`
binary and present the result — the markdown is the surface, the binary is the
engine.

- `abcd.md` → `/abcd` — the read-only where-am-i status board (`abcd --json`).
- `abcd/<verb>.md` → `/abcd:<verb>` — one file per verb. The shipped set is the
  13 files in [`abcd/`](abcd/): `ahoy`, `audit`, `capture`, `consult`,
  `disembark`, `docs`, `embark`, `history`, `ingest`, `launch`, `memory`,
  `prepare-this-repo`, `version`.

Commands stay thin: they call `abcd <verb> --json` and format the result; they
never reimplement behaviour that belongs in the core.
