# commands/

The plugin command surface, auto-loaded by compatible agent harnesses. Each markdown
file is a slash command whose body instructs the host agent to invoke the `abcd`
binary and present the result — the markdown is the surface, the binary is the
engine.

- `abcd.md` → `/abcd` — the read-only where-am-i status board (`abcd --json`).
- `abcd/<verb>.md` → `/abcd:<verb>` — one file per verb, one verb per file:
  <!-- index: commands -->
  `ahoy`, `audit`, `banlist`, `capture`, `consult`, `disembark`, `docs`,
  `embark`, `guard`, `history`, `ideate`, `identity`, `ingest`, `intent`,
  `launch`, `memory`, `prepare-this-repo`, `version`.
  <!-- /index -->

Commands stay thin: they call `abcd <verb> --json` and format the result; they
never reimplement behaviour that belongs in the core.

The verb list above is gated rather than trusted: the `index_drift` record-lint
rule holds the marked region to the contents of [`abcd/`](abcd/), so a verb file
added, renamed, or removed without the same edit here fails the record gate.
