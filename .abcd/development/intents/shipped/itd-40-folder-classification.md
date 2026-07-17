---
id: itd-40
slug: folder-classification
spec_id: spc-5
kind: standalone
suggested_kind: standalone
reclassification_history: []
severity: major
---

# abcd Knows What Kind of Folder You're In

## Press Release

> **abcd recognises what kind of folder you ran it in and tracks every repo it manages in one place.** Run `/abcd:ahoy` anywhere and it tells you plainly: this is a repo abcd manages, a git repo abcd hasn't adopted yet, or just a folder. You never declare a "central location" or wire up paths — abcd keeps a single machine-wide registry, the history store's `index.json` keyed on each repo's root-commit SHA, and bootstraps it silently the first time you install.
>
> "I used to keep a mental map of which folders had abcd set up and which didn't," said Bob, staff engineer. "Now I just run `/abcd:ahoy` and it tells me. When I cloned a new repo into my apps folder it said 'unmanaged git repo — run install to adopt,' and when I ran it in a plain folder it just said there was nothing to manage. I never typed a path."

## Why This Matters

`/abcd:ahoy` sets up abcd in a folder — but "a folder" is not one shape. It might be a repo abcd already manages, a git repo it hasn't adopted, or a plain folder that is nothing to act on. Earlier drafts asked the persona to understand a multi-scope shape (machine-wide / workspace / repo) with a layer that grouped several repos together. abcd lives in **one repository** (adr-28): there is no such grouping layer. The only scopes are **user** (`~/.abcd/`) and **repo** (in-tree `.abcd/`), and abcd bootstraps the user scope silently.

This intent makes `/abcd:ahoy` classify the folder reliably, report it plainly, and track every managed repo in one machine-wide registry — the history store's `index.json`, keyed on each repo's immutable root-commit SHA — so abcd never has to guess or ask the persona for a path.

## What's In Scope

- **Folder classification** — bare `/abcd:ahoy` classifies `cwd` into one of three kinds and **reports only** (never mutates):
  - `managed-repo` — an abcd-managed repository
  - `unmanaged-repo` — a git repository with no abcd markers; reachable via `/abcd:ahoy install`
  - `unmanaged-folder` — not a git repository; nothing to act on
- **Signal hierarchy** — classification keys on abcd-owned markers first, never on weak signals alone:
  - **Strong:** a CLAUDE.md/AGENTS.md marker block, an in-tree `.abcd/` directory, or an entry for `cwd`'s root-commit SHA in `~/.abcd/history/index.json`.
  - **Weak:** a `.git/` directory tells abcd it's *a* repo, not that it's *managed* — necessary, not sufficient.
- **The history-store registry** — `~/.abcd/history/index.json` is the sole user-scope registry: it records every repo abcd manages, keyed on the immutable root-commit SHA, with `name`, `github`, and `path` as mutable labels ahoy refreshes each run. Central-location discovery becomes a lookup in this file — no hardcoded path, no env var, no parent-directory walk.
- **First-run bootstrap** — if `~/.abcd/history/` does not exist, the first `/abcd:ahoy install` creates it (directory + `index.json`) before registering the repo.

## What's Out of Scope

- **Auto-adoption** — bare `/abcd:ahoy` never adopts an unmanaged repo. It classifies and reports; adoption is always the explicit `install` sub-verb.
- **A grouping layer** — abcd is one repository (adr-28); it does not group repos, and there is no `workspaces.json`. A folder that merely holds several repos is an `unmanaged-folder`, not a managed grouping.
- **Duplicating identity** — the history `index.json` is the single registry; root-SHA identity and lineage live there directly. See `05-internals/03-configuration.md`.
- **Classifying non-local folders** — remote repos, URLs, and not-yet-cloned projects are out of scope; classification is of a local `cwd`.

## Acceptance Criteria

> _BDD format, per the itd-1 discipline._

- **Given** a folder with an abcd CLAUDE.md marker block and an `.abcd/` directory, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `managed-repo` and mutates nothing.
- **Given** a git repository with no abcd markers, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `unmanaged-repo` and names `/abcd:ahoy install` as the way to adopt it — without adopting it.
- **Given** a folder that is not a git repository and has no abcd markers, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `unmanaged-folder`, reports there is nothing to act on, and mutates nothing.
- **Given** no `~/.abcd/history/` store exists, **when** the first `/abcd:ahoy install` runs, **then** the store is bootstrapped (directory + `index.json`) and the repo is registered in it by its root-commit SHA.
- **Given** a managed repo registered in the history `index.json`, **when** abcd needs the central/host location, **then** it is resolved by reading `index.json` — not by a hardcoded path or directory walk.

## Open Questions

- Should `/abcd:ahoy doctor` reconcile the history `index.json` against the filesystem (flagging registered-but-missing or present-but-unregistered repos)?
- When a classified `unmanaged-repo` is discovered, should the report suggest `install` inline, or stay purely descriptive?

## Audit Notes

<!-- abcd-review: OWED receipt=rcp-a3791f7dde2e -->
Fidelity review OWED (receipt rcp-a3791f7dde2e).
