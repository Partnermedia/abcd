# Closed

Delivered specs. `abcd spec close <spc-N>` renames a file here from
[`../open/`](../open) and, in the same synchronous call, reconciles the linked
intent `planned/` → `shipped/` and emits the OWED fidelity-review receipt into
that intent's `## Audit Notes`.

Presence in this directory IS the closed status: no spec carries a `status:`
frontmatter field, and the lint refuses one. Closing moves the file and changes
nothing else — a closed spec keeps its `spc-N` and its `intent: itd-N`
back-link, and stays in place as the design record the intent's fidelity audit
measures delivered reality against.

See the [specs charter](../README.md) for the filename grammar, how ids are
minted, and the `spc-N` namespace caveat.
