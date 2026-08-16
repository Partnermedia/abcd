# Consult Freely, Cite Deliberately

**The rule.** An agent may read any source the human has admitted to their
corpus — confidential or not — whenever it bears on the work. It may never
cite, name, or identifiably describe a source in any artefact that leaves the
machine unless a human has deliberately released that citation. Between the
two sits a durable record: influence is captured eagerly and automatically
(cheap, local, append-only); citation happens lazily and manually (when
permission exists and a human flips the flag).

**Why.** Automatic citation is a virtue that becomes a breach the moment a
source is confidential — one helpful footnote can leak what no history rewrite
fully recalls. The naive fix, keeping material away from the agent, throws
away exactly the context that makes its design work good. Splitting
consultation from citation keeps both: full context in, nothing out without a
deliberate human act. The mechanical half of the rule is
[adr-41](../decisions/adrs/0041-corpus-trust-boundary.md) (documents and
ledgers never travel; both gates required); this principle carries the
behavioural half — identifying *description* is as much a citation as a name,
so the consultation discipline forbids it even where no banned string appears.

**In practice.** itd-76 builds the personal corpus, ledger, and guards;
itd-126 moves citation data through the repo; itd-127 renders the ledger into
a paper through a structurally filtered bibliography. Every one of those
surfaces treats a citation as a human release, never an agent convenience.
