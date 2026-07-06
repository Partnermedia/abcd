# ahoy: history store — manual scaffolding (design input)

Captured while re-founding the abcd dev repo for collaboration. Everything
below was done by hand; `/abcd:ahoy` should automate it. This note is the
de-facto spec for ahoy's history-store responsibilities.

## The problem ahoy solves here

Session transcripts (`.specstory/`) and RepoPrompt exports are *evidence about
a repo* — the raw material the lifeboat reviews. They must not live inside the
repo's own git history (privacy: a co-author can `git log` every transcript)
and must not scatter to wherever the agent was `cd`'d. They need a single,
stable home that survives repo renames and GitHub-handle changes.

## Identity: key on the root-commit SHA

A repo's name and GitHub URL are mutable. Its **root-commit SHA**
(`git rev-list --max-parents=0 HEAD`) is immutable under rename, remote move,
and username change. The history store is keyed on it. Names live in
`meta.json` as mutable labels — same pattern SpecStory uses (`workspace_id` +
`git_id`). A re-founded repo has a *new* root SHA; the link to its predecessor
is an explicit `supersedes` field, not a shared path.

## Layout

```
~/.abcd/history/
  index.json                  root-SHA -> {name, github, path, status, supersedes}
  <root-sha>/
    meta.json                 identity + lineage for one repo
    specstory/                live transcripts (SpecStory output_dir points here)
    prompt-exports/           RepoPrompt oracle reviews
```

## SpecStory redirect — the per-repo config shim

SpecStory resolves config user-level (`~/.specstory/cli/config.toml`) then
project-level (`<repo>/.specstory/cli/config.toml`). `output_dir` can be
redirected, but `<repo>/.specstory/` must still exist to hold that config.
So each abcd-managed repo keeps a thin, **gitignored** `.specstory/` shim:

- `.specstory/cli/config.toml` — sets `output_dir` to
  `~/.abcd/history/<root-sha>/specstory`
- no `history/` inside the repo — transcripts land in the store instead

## ahoy steps (what to automate)

1. `git rev-list --max-parents=0 HEAD` -> root SHA.
2. Create `~/.abcd/history/<root-sha>/{specstory,prompt-exports}/`.
3. Write `meta.json` (name, github, status, `supersedes` if re-founded).
4. Register the repo in `index.json`.
5. Install the gitignored `.specstory/cli/config.toml` redirect.
6. Install/refresh the visibility-driven `.gitignore` block.

## Re-founding (the case that motivated this note)

When a repo is re-created with clean history (e.g. to strip in-repo
transcripts before sharing), it is genuinely a new repo with a new root SHA.
ahoy must: register the new SHA, set `supersedes` -> old SHA, mark the old
entry `superseded_by` -> new SHA, and leave the old repo's corpus in place
under its own SHA dir for lifeboat review.

## Related

- `../../brief/04-surfaces/01-ahoy.md` — ahoy surface contract
- `../../brief/05-internals/03-configuration.md` — visibility-driven gitignore
