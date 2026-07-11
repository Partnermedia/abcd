---
name: consult
description: Consult the local sources corpus (the user-level home's sources store, default ~/.abcd/sources) and record source→decision provenance in its append-only ledger. Use when the user says "consult sources", "check the corpus", "what do my sources say", or when a design/research decision arises where prior literature or private working material plausibly matters. Confidential sources are NEVER cited or named in public artifacts.
---

# Consult sources

A local-only corpus at `~/.abcd/sources/` holds source documents (working
papers, private-repo notes, PDFs, books) the agent may **consult** but must
never **cite** publicly. Metadata lives in `sources.json` (CSL-JSON; the
`custom` block carries `confidential`, `permission_status`, `keywords`,
`aliases`). Full details: `~/.abcd/sources/README.md`. If the corpus is
absent, say so and stop — this command never creates it.

## Hard rule (overrides convenience, always)

For any entry with `custom.confidential: true`: never write its title, author
names, aliases, or any identifying string into anything tracked by git or sent
anywhere external — commits, commit messages, PR/issue text, docs, code
comments, published artifacts, pasted output. This covers **identifying
paraphrase** too: do not describe a confidential source so specifically that a
reader could identify it ("a forthcoming paper showing X beats Y on Z", a
private repo's distinctive architecture) — the mechanical guards catch literal
strings only; this rule is the paraphrase layer. Refer generically ("a working
paper on X", "prior private work"). The ledger holds the real reference;
citation is the user's manual decision, made when permission exists (public
citation requires BOTH the source's permission_status AND the ledger line's
cited_publicly flag).

Conversation with the user is fine — discuss confidential sources freely there.

## Consult

1. Search the corpus: `grep -ril "<term>" ~/.abcd/sources/confidential/
   ~/.abcd/sources/public/` — try keywords, author surnames, and CSL keys.
   The path IS the classification: any hit under `confidential/` falls under
   the hard rule above. Each source is a folder (`<class>/<key>/`) holding
   `original.<ext>`, `text.md`, and any summaries/notes.
2. Read matched files freely for conversation; the folder's class governs
   what may leave it.
3. If nothing relevant surfaces, say so — do not pad.

To add a source, use `/abcd:ingest`.

## Record influence (the ledger)

Whenever a source **meaningfully influences a decision** (supports it,
contradicts it, supplies a method, or shapes background understanding — not
mere incidental reading), append ONE line to
`~/.abcd/sources/ledger/<repo>.jsonl`:

```json
{"ts":"<UTC ISO-8601>","repo":"<repo>","decision_ref":"<DECISIONS.md date | ADR id | intent id | free text>","claim":"<what was decided/claimed>","source_key":"<CSL id>","locator":"<pp./§ if known>","influence":"supports|contradicts|method|background","used_in":["<repo-relative path(s) of the consuming document(s)>"],"cited_publicly":false}
```

`used_in` makes acknowledgment machine-readable in both directions: an idea
is traced to its source even when the consuming document only paraphrases
(public sources) or must stay silent (confidential sources). Fill it whenever
the influence landed in an identifiable document, not just a conversation.

Then commit in the corpus repo:
`git -C ~/.abcd/sources add -A && git -C ~/.abcd/sources commit -m "ledger(<repo>): <source_key> → <short decision>"`

The ledger is append-only: corrections are new lines, never edits.
`cited_publicly` is always written `false`; only the user flips it, by hand.

Always tell the user in conversation which key was recorded against which
decision, so they can decide about citing.

## Guard wiring

- On first use in a repo (and after any confidential source is added), run
  `~/.abcd/sources/bin/sync-banlist <repo-root>`. It maintains a generated
  block in the repo's untracked `.abcd/.work.local/private-names.txt`, which
  the repo's pre-commit guard reads — leakage is then blocked mechanically,
  not just by this command's rule.
- Before any document that drew on confidential material is committed, posted,
  or otherwise shared, run `~/.abcd/sources/bin/cite-guard <file>` (exit 1 =
  confidential identifier present; its report names only the CSL key, so the
  report itself is safe to relay).
