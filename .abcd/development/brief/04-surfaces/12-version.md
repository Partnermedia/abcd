# `/abcd:version` — Print the Installed Version

`/abcd:version` reports the installed abcd version. It is **strictly read-only**
— it performs zero writes.

## Behaviour

```bash
abcd version --json
```

emits `{ "name": "abcd", "version": "<version>" }`. The plugin command
(`commands/abcd/version.md`) reads the JSON and tells the user the `name` and
`version`. Without `--json`, bare `abcd version` prints the version string only
(e.g. `abcd dev` in a development build) — it does **not** render a status board;
the bare-status convention is scoped to `ahoy`/`capture`/`memory` and bare
`abcd`, not to `version`.

## Where the version comes from

The version is **derived, never hand-authored** — it is read from the shipped
build, not from a literal in the record ([adr-31](../../decisions/adrs/0031-derived-versioning-from-intents.md)).
`/abcd:launch` is the surface responsible for stamping the derived version into
the release artefact; `/abcd:version` only reports what is installed.

## References

- Plugin command: [`commands/abcd/version.md`](../../../../commands/abcd/version.md)
- Derived versioning: [`04-launch.md § 3`](04-launch.md#3-versioning--marketplace)
