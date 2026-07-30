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

**The store's first line declares its format**, and that one line decides the whole
file. A store whose first line is exactly `# abcd-banlist: keyed` is a *keyed*
store: every non-comment, non-blank line must parse as `KEY<space-or-tab>PATTERN`.

```text
# abcd-banlist: keyed
lab-host   alice-laptop\.example\.com
lab-ip     192\.0\.2\.17
```

`KEY` is a stable, non-sensitive handle (`[A-Za-z0-9][A-Za-z0-9._/-]*`) and the
only part of an entry that reaches any output. `PATTERN` is a POSIX extended
regular expression matched case-insensitively — the engine is the guard's `grep
-iE`, so `(?i)` is never needed and Perl escapes such as `\d`, `\w` and `\b` are
not available. Machine identifiers — hostnames, IPv4/IPv6 addresses, CIDR prefixes,
MAC addresses, device names — are ordinary entries, matched exactly as a name is
(the fixture values above are RFC 5737 and persona-derived, per
[`examples-use-reserved-identifiers`](../../principles/examples-use-reserved-identifiers.md)).

A store **without** that first line is a *legacy* store, the format the guard
shipped with: every non-comment, non-blank line is one whole-line pattern under the
synthetic key `entry-<line-number>`, and no line is ever split. An old store
therefore keeps matching exactly what it always matched, and no part of any line can
be printed. That is the whole reason the declaration exists. Deciding per line
whether a first field "looks like a key" did two harmful things at once: it printed
part of a legacy line as a key — and on this layer a pattern *is* the secret — and it
narrowed an old whole-line pattern to the remainder after its first field. The
declaration makes both unrepresentable, at the cost of one line a user adds by hand.
`add` and `remove` refuse a non-empty legacy store for exactly that reason: writing a
keyed line into it would change what every *other* line means.

Leading and trailing ASCII spaces and tabs are stripped, and nothing else is — the
Go parser and the shell hook strip the same set, byte for byte. A whitespace class
that differs between the two readers is a line one of them silently ignores while
the other reports it as live.

## The guard's output contract

The guard checks the **content of every staged file**, read out of the index
(`git show ":0:<path>"`, stage-explicit so a path shaped like a revision cannot
redirect the read), not the text of a diff. That is the question it is actually
asking — is this name in what I am about to commit? — and unlike diff text it has no
shape to route around: a content line beginning `++`, a blob containing a NUL, a
committed `.gitattributes` carrying `-diff`, and a rename all hide a name from a
diff-text reading. Binary blobs are scanned like anything else, because a name in a
binary file is in history just the same.

On a match the guard refuses the commit and names **the key alone**. The matched
string and the pattern value never reach stdout, stderr, or a log — a refusal that
echoed the string would defeat the layer at the moment it worked. The pattern reaches
grep on stdin, never in argv, for the same reason. A line that does not parse, and a
pattern the engine refuses, are each themselves a refusal naming a line number and
nothing else: an unusable entry is never skipped, because a banlist that cannot be
read must not look like a banlist that found nothing. **Any** git step that fails
refuses the commit too — a check that could not run must never be indistinguishable
from a check that passed.

An absent store prints a loud `INACTIVE` warning and lets the commit through, and a
store that exists but yields no entries prints an equally loud `NO ENTRIES` warning:
it checks exactly as much. The layer protects machines that opted in, and silence
must never impersonate protection — which is why the read surface reports `present`
as a distinct state rather than rendering an empty list.

## Two ways an entry fails, reported apart

`abcd banlist list --private` distinguishes a line the guard's engine **cannot use**
from one it **accepts and reads differently**, because the two need opposite
responses. An unusable line stops every commit until it is fixed. An inert line — a
Perl-style escape, an inline flag group — stops nothing: grep may read it
differently than written, so the name can go unguarded while the store looks healthy. `add --private` refuses both up
front, screening the constructs POSIX ERE does not implement and then asking grep
itself, with the pattern on stdin, whether the expression is usable. A private
pattern is therefore checked against the engine that enforces it rather than
against Go's, which accepted `\d` and `(?i)` as healthy (grep reads them
differently) and accepted `[a-z-.]`, which grep refuses (its fail-safe branch then
blocks every commit).

Because the store's safety rests entirely on its being untracked, `add --private`
refuses outright when git does not ignore the store's path: the guard cannot catch
its own source.

## Scaffolded, not hand-wired

`abcd ahoy` writes all three artefacts into any repo it configures, so a repo
becomes name-safe by being abcd-managed:

| artefact | where | note |
|---|---|---|
| guard hook | `.githooks/pre-commit` | committed, so every clone inherits it; a clone arms it once with `git config core.hooksPath .githooks` |
| public family | `.abcd/docs-lint.json` | an **empty** `banned_tokens` array — abcd cannot know which names a repo may not publish, and a ban nobody declared would fail a build over a word the maintainer never chose |
| private stub | `.abcd/.work.local/private-names.txt` | inside the gitignored local tier, which the installed `.gitignore` fence covers under either visibility |

Every write is create-if-absent. A hook, a CI-gating config, and above all a
populated private store are the maintainer's, and re-seeding one would delete work
abcd cannot see. The stub is written only after the fence is on disk: a stub git
would track is the hazard, not the remedy — so the gap that asks for it is
advertised as resolvable only once the fence covers it.

The stub's worked examples are **all commented out**, so a fresh scaffold parses to
zero entries and the guard says so loudly at commit time rather than looking like
protection. Every illustrative value in it is a reserved documentation value (RFC
5737, RFC 3849, RFC 2606, RFC 7042) or a persona-derived fixture host, per
[`examples-use-reserved-identifiers`](../../principles/examples-use-reserved-identifiers.md);
the scaffold test judges that with the repo's own network-identifier detector, and
proves the detector is armed with a control value first.

## Honest reach

`abcd banlist` states plainly that CI cannot enforce the private layer, and so does
every other surface that describes it: `abcd ahoy`'s status board and the detection
envelope both carry the sentence beside the state it qualifies. It lives once, as
`banlist.PrivateReachNote`, because a human line, a verb, and a machine consumer
reading the envelope's booleans must not be able to disagree about what "hook
installed" beside a present store means — and the reading a machine consumer takes
on its own is the wrong one.

That is not a limitation to be fixed but the design: a pattern in CI config is a
published pattern.

## References

- Plugin command: [`commands/abcd/banlist.md`](../../../../commands/abcd/banlist.md)
- Spec: [`spc-20`](../../specs/closed/spc-20-name-banlist.md)
- Intent: [`itd-74`](../../intents/shipped/itd-74-name-banlist.md)
- Public-layer gate: [`10-docs.md`](10-docs.md)
- Install surface: [`01-ahoy.md`](01-ahoy.md)
