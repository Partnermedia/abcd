# commands/

The plugin command surface, auto-loaded by Claude Code and the companion harness. Each markdown
file is a slash command whose body instructs the host agent to invoke the `abcd`
binary and present the result — the markdown is the surface, the binary is the
engine.

- `abcd.md` → `/abcd` — the read-only where-am-i status board (`abcd --json`).
- `abcd/<verb>.md` → `/abcd:<verb>` — one file per verb. Currently `version`.

Commands stay thin: they call `abcd <verb> --json` and format the result; they
never reimplement behaviour that belongs in the core.
