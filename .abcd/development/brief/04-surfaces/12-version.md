# `/abcd:version` — Print the Installed Version

`/abcd:version` reports the installed abcd version, install mode, and vintage
(the verb's own help summary). It performs **zero writes**, and it is
network-silent unless the opt-in `--check` flag is passed — that flag fetches
the latest release exactly once and compares, this command's **only network
touch** (abcd never fetches implicitly —
[adr-38](../../decisions/adrs/0038-implicit-checks-are-disk-only.md)), naming its
source and verdict in the render.

## Behaviour

```bash
abcd version --json
```

emits `{ "name": "abcd", "version": "<version>", "vintage": "<revision>",
"staleness": "<fresh|stale|unknown>" }`, plus `install_mode` when a PATH-entry
install mode is resolvable (the field is omitted when empty) and a `check`
object (latest, source, verdict) when `--check` was passed. The plugin command
(`commands/version.md`) reads the JSON and tells the user the `name`,
`version`, `install_mode`, `vintage`, and `staleness`. Without `--json`, bare
`abcd version` prints a short block — the version line (e.g. `abcd dev` in a
development build) followed by `install:` (when resolvable), `vintage:`, and
`staleness:` lines — not the version string alone, and it does **not** render
a full status board; the bare-status convention is scoped to
`ahoy`/`banlist`/`capture`/`identity`/`intent`/`memory`/`spec`
and bare `abcd`, not to `version`.

## Where the version comes from

The version is **derived, never hand-authored** — it is read from the shipped
build, not from a literal in the record ([adr-31](../../decisions/adrs/0031-derived-versioning-from-intents.md)).
`/abcd:launch` is the surface responsible for stamping the derived version into
the release artefact; `/abcd:version` only reports what is installed.

## References

- Plugin command: [`commands/version.md`](../../../../commands/version.md)
- Derived versioning: [`04-launch.md § 3`](04-launch.md#3-versioning--marketplace)
