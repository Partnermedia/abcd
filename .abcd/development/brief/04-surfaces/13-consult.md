# `/abcd:consult` — Confidential Sources Corpus

`/abcd:consult` searches a local-only corpus of source documents — working
papers, private-repo notes, PDFs, books — that the agent may **consult** but must
never **cite** publicly, and records source→decision provenance in an append-only
ledger. It is **host-delegated**: no Go verb backs it, there is no `abcd consult`
binary sub-verb and no bare-status render — the workflow runs entirely in the host
agent, orchestrating the corpus with `grep`, file reads, and `git` in the corpus
repo.

## What it does

The corpus lives at `~/.abcd/sources/` (the user-level home's sources store), with
each source held as a folder `<class>/<key>/` under `confidential/` or `public/`.
**The path IS the classification**: any hit under `confidential/` falls under the
hard rule. Metadata lives in `sources.json` (CSL-JSON; the `custom` block carries
`confidential`, `permission_status`, `keywords`, `aliases`). If the corpus is
absent, the command says so and stops — it never creates it.

## Flow

1. **Search** the corpus with `grep -ril "<term>"` across `confidential/` and
   `public/` — trying keywords, author surnames, and CSL keys.
2. **Read** matched files freely for conversation; the folder's class governs what
   may leave it.
3. **Record influence** — whenever a source meaningfully shapes a decision
   (supports, contradicts, supplies a method, or informs background), append ONE
   line to `~/.abcd/sources/ledger/<repo>.jsonl` and commit it in the corpus repo.
   The ledger is append-only: corrections are new lines, never edits. `used_in`
   traces the influence to the consuming document in both directions.

## The hard rule

For any entry with `custom.confidential: true`, never write its title, authors,
aliases, or any identifying string into anything tracked by git or sent anywhere
external — commits, PR/issue text, docs, code comments, published artifacts. This
covers **identifying paraphrase** too: the mechanical guards catch literal strings
only, so this rule is the paraphrase layer. Refer generically ("a working paper on
X"). Public citation requires BOTH the source's `permission_status` AND the ledger
line's `cited_publicly` flag — and `cited_publicly` is only ever flipped by the
user, by hand. Conversation with the user is exempt: confidential sources may be
discussed freely there.

## Guard wiring

The command's rule is backed mechanically, not trusted alone:

- `~/.abcd/sources/bin/sync-banlist <repo-root>` maintains a generated block in
  the repo's untracked `.abcd/.work.local/private-names.txt`, which the repo's
  pre-commit guard reads — run on first use in a repo and after any confidential
  source is added.
- `~/.abcd/sources/bin/cite-guard <file>` runs before any document that drew on
  confidential material is committed or shared (exit 1 = confidential identifier
  present; its report names only the CSL key, so the report itself is safe to
  relay).

## Composition

`/abcd:consult` is the read-and-record side of the sources system; `/abcd:ingest`
is the write side that adds a source to the corpus. The two share the corpus at
`~/.abcd/sources/` and its ledger.

## References

- Plugin command: [`commands/abcd/consult.md`](../../../../commands/abcd/consult.md)
- Corpus store and its schema: `~/.abcd/sources/README.md`
- The confidentiality invariants it enforces: [`../02-constraints`](../02-constraints)
