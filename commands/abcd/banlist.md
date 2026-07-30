---
name: banlist
description: Maintain the two banned-names layers — the committed CI-enforced public list and the gitignored per-machine private list — by invoking the abcd binary. Bare invocation is a read-only render; add/remove act on one named layer.
argument-hint: "[list --private|--public] | add --private|--public <key> <pattern> [--severity blocker|warn] [--successor <text>] | remove --private|--public <key>"
---

# `/abcd:banlist` — banned names, two layers

Some names must not appear in what a repo publishes: a specific agent harness (so
the surface stays host-agnostic), a partner's product, a private project whose very
name is confidential — and, most sensitively, the user's own machine identifiers:
hostnames, device names, and the addresses or prefixes of their private network.

Enforcement splits by sensitivity, because a deterministic CI gate is the right
tool for a public banned name and the **wrong** place for a private one: the rule
would have to contain the very string it forbids.

| layer | store | enforced by | visibility |
|---|---|---|---|
| public | `.abcd/docs-lint.json` (the `banned_tokens` family) | `abcd docs lint` in CI, with a per-line escape | entries render in full |
| private | `.abcd/.work.local/private-names.txt` (gitignored) | the committed `.githooks/pre-commit` guard, on this machine only | entries render **by key only** |

## Render both layers (bare)

```bash
abcd banlist --json
```

Summarise the JSON: for `private`, its `present` flag and each entry's `key` —
**never** ask for or repeat a private pattern value, which the binary does not
emit; for `public`, each entry's `id`, `severity`, `pattern`, and whether it is
`managed` (verb-owned, id under `names/`) or hand-curated. State plainly that CI
cannot enforce the private layer: it protects only machines that have opted in.
When `private.present` is false, say that the layer is inactive on this machine —
an absent store checks nothing, and silence must never look like protection.

`list` is the same render with an optional scope:

```bash
abcd banlist list --private --json      # or --public
```

## Add an entry

```bash
abcd banlist add --private <key> "<pattern>" --json
abcd banlist add --public  <key> "<pattern>" --json
```

The layer is **never** guessed: an add with no layer flag, or with both, exits 2.
`<key>` is a stable, non-sensitive handle (`[A-Za-z0-9][A-Za-z0-9._/-]*`) and
`<pattern>` a POSIX extended regular expression matched case-insensitively.
Machine identifiers — hostnames, IPv4/IPv6 addresses, CIDR prefixes, MAC
addresses, device names — are ordinary private entries.

A public add takes `--severity` (`blocker`, the default, or `warn`) and
`--successor` (the replacement the finding cites; default "a generic term"), and
writes one entry into the committed config under the `names/` id namespace. Report
the entry `id` and remind the user to commit it: the public layer gates everyone.

A private add writes to the gitignored per-machine store and reports the key
alone. **Do not echo the pattern back to the user** — it is the value the whole
layer exists to keep out of transcripts, logs, and history.

## Remove an entry

```bash
abcd banlist remove --private <key> --json
abcd banlist remove --public  <key> --json
```

A public removal is refused for a hand-curated entry (an id outside `names/`):
those are edited in the config by a human, in a reviewable commit. An unknown key
is refused rather than treated as a no-op.

## Fallback

If the `abcd` binary is not on `PATH`, fall back to `go run ./cmd/abcd banlist …`
from the repo root, or tell the user to build it with `make build`.

**User input:** $ARGUMENTS
