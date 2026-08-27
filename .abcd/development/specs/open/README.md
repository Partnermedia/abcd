# Open

Specs in flight — minted by `/abcd:intent plan <itd-N>` and not yet closed. Each
file is the design record for one intent's promised capability, and each carries
the `intent: itd-N` back-link the record lint checks in both directions.

Presence in this directory IS the open status: no spec carries a `status:`
frontmatter field, and the lint refuses one. A spec leaves for
[`../closed/`](../closed) when `abcd spec close <spc-N>` renames it, which in the
same call reconciles the linked intent `planned/` → `shipped/`.

A spec whose `## Summary` still holds the minted placeholder is unwritten, and
`abcd intent ready <itd-N>` reports its intent not ready until an author replaces
it.

See the [specs charter](../README.md) for the filename grammar, how ids are
minted, and the `spc-N` namespace caveat.
