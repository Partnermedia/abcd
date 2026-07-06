---
id: itd-40
slug: folder-classification
spec_id: spc-15-folder-classification-workspacesjson
kind: standalone
suggested_kind: standalone
reclassification_history: []
---

# abcd Knows What Kind of Folder You're In

## Press Release

> **abcd recognises what kind of folder you ran it in and tracks everything it manages in one place.** Run `/abcd:ahoy` anywhere and it tells you plainly: this is a repo abcd manages, a workspace holding several repos, a git repo abcd hasn't adopted yet, or just a folder. You never declare a "central location" or wire up paths — abcd keeps a single machine-wide registry, `~/.abcd/workspaces.json`, and the only choice it ever asks you to make is the one you actually think in: *is this folder one repo, or a home for several (a workspace)?*
>
> "I used to keep a mental map of which folders had abcd set up and which didn't," said Bob, staff engineer. "Now I just run `/abcd:ahoy` and it tells me. When I cloned a new repo into my apps folder it said 'unmanaged git repo — run install to adopt,' and when I ran it at the apps folder itself it offered to make it a workspace. I never typed a path."

## Why This Matters

`/abcd:ahoy` sets up abcd in a folder — but "a folder" is not one shape. The folder might be a single repo, or a workspace holding a public/dev repo pair, or the persona's central development root. Earlier drafts asked the persona to understand a four-scope shape (machine-wide / development / workspace / repo). That is abcd's *internal* plumbing — it should not be the persona's mental shape.

The persona thinks in two states: **this folder is one repo**, or **this folder holds several related repos (a workspace)**. Everything else — the host store, the machine-wide directory — is plumbing abcd bootstraps silently. This intent makes `/abcd:ahoy` classify the folder reliably, report it plainly, and track every managed folder in one machine-wide registry so abcd never has to guess or ask the persona for a path.

## What's In Scope

- **Folder classification** — bare `/abcd:ahoy` classifies `cwd` into one of five kinds and **reports only** (never mutates):
  - `managed-repo` — an abcd-managed repository
  - `managed-workspace` — an abcd-managed workspace (holds repos)
  - `unmanaged-workspace` — a folder with sibling repo-shaped subdirs and no abcd markers; reachable via `/abcd:ahoy install`
  - `unmanaged-repo` — a git repository with no abcd markers; reachable via `/abcd:ahoy install`
  - `unmanaged-folder` — neither; nothing to act on
- **Signal hierarchy** — classification keys on abcd-owned markers first, never on weak signals alone:
  - **Strong:** a CLAUDE.md/AGENTS.md marker block, an `.abcd/` directory, an entry in `~/.abcd/workspaces.json`.
  - **Weak:** a `.git/` directory tells abcd it's *a* repo, not that it's *managed* — necessary, not sufficient.
  - **Heuristic:** naming/structure (sibling repo-shaped subdirs) only disambiguates repo-vs-workspace, never decides managed-vs-unmanaged.
- **The two-valued persona choice** — `/abcd:ahoy install` asks at most one structural question: *make this a repo, or a workspace?* The internal scopes are bootstrapped without the persona naming them.
- **`~/.abcd/workspaces.json`** — a machine-wide registry of every workspace and repo abcd manages: `repo|workspace` tag, mutable path, and for repos a reference to the root-SHA in the history `index.json`. Host/central-location discovery becomes a lookup in this file — no hardcoded path, no env var, no parent-directory walk.
- **First-run bootstrap** — if `~/.abcd/workspaces.json` does not exist, the first `/abcd:ahoy install` creates it alongside the host history store.

## What's Out of Scope

- **Auto-adoption** — bare `/abcd:ahoy` never adopts an unmanaged repo. It classifies and reports; adoption is always the explicit `install` sub-verb.
- **Duplicating identity** — `workspaces.json` records human structure (groupings, paths). Root-SHA identity and lineage stay in the history `index.json`; a repo entry *references* the root-SHA, never copies the identity fields. See `05-internals/03-configuration.md`.
- **Classifying non-local folders** — remote repos, URLs, and not-yet-cloned projects are out of scope; classification is of a local `cwd`.
- **Repo-scope `.abcd/`** — a repo is a classification result and an install target, not an `.abcd/` scope. It carries only the marker block + `.specstory/` shim.

## Acceptance Criteria

> _BDD format, per the itd-1 discipline._

- **Given** a folder with an abcd CLAUDE.md marker block and an `.abcd/` directory, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `managed-repo` (or `managed-workspace`) and mutates nothing.
- **Given** a git repository with no abcd markers, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `unmanaged-repo` and names `/abcd:ahoy install` as the way to adopt it — without adopting it.
- **Given** a folder containing sibling repo-shaped subdirectories with no abcd markers and no `~/.abcd/workspaces.json` entry, **when** the persona runs bare `/abcd:ahoy`, **then** it reports `unmanaged-workspace` and names `/abcd:ahoy install` as the way to adopt it — without adopting it.
- **Given** a folder containing sibling repo-shaped subdirectories, **when** `/abcd:ahoy install` runs, **then** the persona is asked the single repo-vs-workspace question and the choice is recorded in `~/.abcd/workspaces.json`.
- **Given** no `~/.abcd/workspaces.json` exists, **when** the first `/abcd:ahoy install` runs, **then** the registry is created and the managed folder is recorded in it.
- **Given** a managed repo registered in `workspaces.json`, **when** abcd needs the central/host location, **then** it is resolved by reading `workspaces.json` — not by a hardcoded path or directory walk.

## Open Questions

- Should `workspaces.json` record a repo's parent workspace explicitly, or derive it from path containment? (Explicit survives a repo moving; derived can't drift.)
- When a classified `unmanaged-repo` sits inside a `managed-workspace`, should the report suggest registering it under that workspace specifically, or just generic `install`?
- Does `/abcd:ahoy doctor` reconcile `workspaces.json` against the filesystem (flagging registered-but-missing or present-but-unregistered folders)?

## Audit Notes

_Empty. Populated by intent-fidelity-reviewer when intent moves to shipped/._
