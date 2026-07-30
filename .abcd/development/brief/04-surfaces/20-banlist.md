# `/abcd:banlist` — Banned Names, Two Layers

`/abcd:banlist` maintains the names a repo must not publish. Bare invocation and
`list` are **strictly read-only**; `add` and `remove` are the write paths, and each
names its layer explicitly.

Two failure modes share one root. A tool that claims to be host-agnostic
undermines itself the moment its published docs name a specific harness. Worse, a
private collaborator's, project's, or machine's name leaking into a public repo is
a confidentiality breach a history rewrite cannot fully undo — merged-PR diffs and
cached views persist server-side. Both are cheap to prevent at authoring time and
expensive to remediate afterwards.

## Why two layers

A deterministic CI gate is the right tool for a *public* banned name and the wrong
place for a *private* one, because the rule would have to contain the very string
it forbids. Splitting enforcement by sensitivity resolves the tension without
compromise.

| | public layer | private layer |
|---|---|---|
| store | `.abcd/docs-lint.json`, the `banned_tokens` family | `.abcd/.work.local/private-names.txt`, gitignored |
| enforced by | `abcd docs lint` in CI, per-line escape | the committed `.githooks/pre-commit` guard |
| reach | every clone, every pull request | only machines that have opted in |
| visibility | entries render in full | entries render **by key only** |

The public layer is not a new mechanism. It is the `banned_tokens` family that
already gates this repo's harness names — one canonical primitive, so an entry a
verb writes and an entry a human hand-curated are enforced by the same engine with
the same escape hatch. Verb-written entries carry the `names/` id prefix, which is
the ownership boundary: `list` shows the whole family, and a removal is refused for
anything outside that namespace.

## The private entry format

One entry per line, `KEY<whitespace>PATTERN`:

```text
lab-host   alice-laptop\.example\.com
lab-ip     192\.0\.2\.17
```

`KEY` is a stable, non-sensitive handle (`[A-Za-z0-9][A-Za-z0-9._/-]*`) and the
only part of an entry that reaches any output. `PATTERN` is a POSIX extended
regular expression matched case-insensitively. Machine identifiers — hostnames,
IPv4/IPv6 addresses, CIDR prefixes, MAC addresses, device names — are ordinary
entries, matched exactly as a name is (the fixture values above are RFC 5737 and
persona-derived, per [`examples-use-reserved-identifiers`](../../principles/examples-use-reserved-identifiers.md)).

The key charset excludes every regular-expression metacharacter, which is what
makes the format backward compatible: a line that does not parse as key + pattern
is read as a bare pattern under the synthetic key `entry-<line-number>`, so a store
written in the older one-pattern-per-line format keeps blocking exactly what it
blocked before. Protection never weakens because the format grew a column.

## The guard's output contract

On a match the guard refuses the commit and names **the key alone**. The matched
string and the pattern value never reach stdout, stderr, or a log — a refusal that
echoed the string would defeat the layer at the moment it worked. A pattern the
regex engine refuses is itself a refusal, naming its line number and nothing else:
an unusable entry is never skipped, because a banlist that cannot be read must not
look like a banlist that found nothing.

An absent store prints a loud `INACTIVE` warning and lets the commit through. The
layer protects machines that opted in, and silence must never impersonate
protection — which is why the read surface reports `present` as a distinct state
rather than rendering an empty list.

## Honest reach

`abcd banlist` states plainly that CI cannot enforce the private layer. That is not
a limitation to be fixed but the design: a pattern in CI config is a published
pattern. Wiring the same statement into the status and report surfaces, and
`ahoy` scaffolding of the guard hook, gitignore entry, and a documented stub with
reserved-value examples, are the remaining slices of this surface.

## References

- Plugin command: [`commands/abcd/banlist.md`](../../../../commands/abcd/banlist.md)
- Spec: [`spc-20`](../../specs/open/spc-20-name-banlist.md)
- Intent: [`itd-74`](../../intents/planned/itd-74-name-banlist.md)
- Public-layer gate: [`10-docs.md`](10-docs.md)
- Install surface: [`01-ahoy.md`](01-ahoy.md)
