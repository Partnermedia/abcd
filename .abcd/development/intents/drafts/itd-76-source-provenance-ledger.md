---
id: itd-76
slug: source-provenance-ledger
spec_id: null
kind: standalone
suggested_kind: null
reclassification_history: []
severity: major
blocked_by: [itd-74]
builds_on: [itd-77]
---

# abcd Lets You Consult Sources You Cannot Cite — and Remembers Every Debt

## Press Release

> **Consult any source freely; cite only by deliberate human choice — abcd keeps the ledger in between.** A developer-researcher often works from material they are not free to name in public: working papers under review, a collaborator's private repository, notes shared under an NDA. The agent should be able to *read* that material when it bears on a decision — and must never *cite* it in anything published. abcd manages this as three coupled pieces. A **local-only corpus** (abcd's user-level home, `~/.abcd/sources/` by default, itself a no-remote git repository) holds the documents and a machine-readable bibliography (CSL-JSON; a `custom` block carries `confidential`, `permission_status`, retrieval keywords, and banned aliases). An **append-only provenance ledger** (JSONL, one file per consuming repo) records every meaningful influence — which source, which decision, what claim, what kind of influence — gated twice: the source's `permission_status` grants the *right* to cite, and each ledger line's `cited_publicly` flag *exercises* it, flipped only by the human. And the **guardrails are mechanical for what mechanism can catch**: the confidential entries *generate* the pattern block in each repo's untracked private-names banlist, the pre-commit guard refreshes and enforces it on every commit, and a `cite-guard` scan clears any text before it is shared. Published sources travel the other way: their *citation data* (never the documents, never the ledger) is shared into the repo, so a whole team can ingest and build on one bibliography. Later, the same machine-readable trail — sources, decisions, claims — reconstructs an academic paper, publicly rendered only through a structurally filtered bibliography and a post-render check.
>
> "I could never let an agent near my working papers before, because one helpful footnote could burn a collaborator's trust," said Maya, a researcher-developer. "Now it reads everything, records what influenced what, and cites nothing. When the paper behind a decision is finally published, I flip one flag and the citation appears in the next render — with the whole influence trail already written." Their teammate Bob never sees their corpus at all: "I just ingest the repo's shared references and every public source Maya worked from is in my own local store, ready to consult."

## Why This Matters

Automatic citation is a virtue that becomes a breach the moment a source is confidential: an agent that helpfully names "the working paper this design follows" in a commit message has leaked something no history rewrite fully recalls. The naive fix — keep the material away from the agent — throws away exactly the context that makes its design work good. The resolution is to split *consultation* from *citation* and put a durable, machine-readable record between them: influence is captured eagerly and automatically (cheap, local, append-only), citation happens lazily and manually (when permission exists). The ledger is also the seed of something bigger: a paper whose claims trace to decisions and whose decisions trace to sources is *reconstructable* rather than rewritten from memory. And because a repo is a team surface even when every corpus is personal, the public slice of the bibliography must flow through the repo — citation data is shareable; documents and influence trails are not.

This composes three existing abcd designs rather than inventing new machinery: the two-layer name banlist ([itd-74](itd-74-name-banlist.md)) supplies the leak guard; the append-only audit chain ([itd-16](itd-16-hash-chain-merkle-audit.md)) is a possible later integrity backend for the ledger — the corpus repo's git history carries tamper-evidence until then, and nothing here depends on itd-16 shipping; and the provenance substrate ([`09-provenance-substrate.md`](../../brief/05-internals/09-provenance-substrate.md)) already defines citation blocks, a source registry, and an NDA-aware publish gate for *ingested* content — this intent extends the same stance to *consulted* content.

## What It Looks Like

- **`abcd source add`** registers a document: CSL-JSON entry (permission status, keywords, aliases), original and extracted text stored in a per-source folder whose *location* — `confidential/<key>/` or `public/<key>/` — is the classification, declared once at ingestion. Derived artifacts (summaries, notes) live in the same folder and inherit the class; declassification is a visible move, never a silent flag edit. Consultation is plain search over the folders — no index, no service. The corpus lives in abcd's **user-level home** (`~/.abcd/`, path configurable; relocation is [itd-77](itd-77-relocatable-user-home.md)) — the first user-tier surface alongside the repo's `.abcd/` tiers.
- **`abcd source ledger`** appends an influence record — `{decision_ref, claim, source_key, locator, influence, cited_publicly}` — and commits it; corrections are new lines, never edits. A public citation requires **both** gates: the source permits (`permission_status`) *and* the line is flipped (`cited_publicly: true`).
- **`abcd source share` / `abcd source ingest`** move *citation data* through the repo: `share` writes a public entry into the committed `.abcd/work/references.json` (refusing any `confidential: true` entry mechanically); `ingest` imports the repo's shared entries into the local corpus. Documents and ledgers never travel — only bibliography. The hand-curated references registry in the development record can later derive from this file.
- **`abcd source sync-banlist`** projects every confidential entry's identifying strings — title and aliases always; author names only by per-source opt-in (`ban_authors: true`, for the rare collaboration that is itself secret — the common confidential types, one's own submitted work, purchased reports, and private repos, are protected by title and aliases, and banning their authors would mostly ban legitimate names, including one's own) — into the repo's untracked private-names banlist (the itd-74 private layer). The pre-commit guard **auto-refreshes** this block when the corpus is present (no-op otherwise), so the guard is only ever as stale as the current commit. **`abcd source cite-check`** scans any text and reports offending entries by key only — its output is safe to relay.
- **Paper reconstruction** walks the ledger: claims grouped by decision, citations resolved from the bibliography, rendered to PDF and HTML from one markdown source. The public render is proven clean twice, independently: structurally (it renders from a *generated* bibliography that omits confidential entries, so an unpermitted key fails the build) and by a deterministic post-render check of the output's citations against both gates.

## What It Cannot Enforce

The mechanical layer blocks **literal identifying strings**. It cannot detect a paraphrase that identifies a source without naming it — "a forthcoming paper shows X beats Y", or a description of a private repository's distinctive architecture. That residual risk is handled behaviourally (the consultation skill forbids identifying description, not just naming) and by the human review that gates every publish; abcd states this boundary plainly rather than implying coverage it does not have. Durability is likewise bounded: a no-remote corpus survives disk loss only via machine backup and offline `git bundle` snapshots — abcd documents that discipline; it cannot perform it.

## Dogfood (already running)

The convention-first prototype is live for this repo's development: the corpus, ledger, guard scripts, and an agent-side consultation skill exist as a user-level scaffold, and the generated banlist block feeds the same `.githooks/pre-commit` guard that itd-74 generalises. The feature is to lift that scaffold into abcd verbs so any managed repo inherits it. See the plan ([`2026-07-08-confidential-sources-scaffold.md`](../../plans/2026-07-08-confidential-sources-scaffold.md)) and the SOTA survey ([`2026-07-08-confidential-sources-provenance-sota.md`](../../research/notes/2026-07-08-confidential-sources-provenance-sota.md)).

## Open Questions

- Ledger ownership once work spans machines: per-repo files in the user-level corpus (current) work for one machine; deferred until a second machine actually exists.
- `share`/`ingest` conflict shape when two teammates share the same source with divergent metadata — last-write, merge, or key-ownership.
- Whether ingest should mark provenance on imported entries (who shared, from which repo) so a team bibliography stays auditable.
