# Runbook — dogfood abcd locally as a plugin (private, always-latest)

Maintainer runbook (developer-facing, so it names the harness — unlike `docs/`,
which stays host-agnostic). Goal: install and use abcd yourself from a local
checkout, **without publishing anything**, and have it always track live
development so a broken build shows up immediately.

abcd has **two halves** and you install both:

1. the **`abcd` Go binary** the slash-commands shell out to, and
2. the **plugin surface** (`commands/`, `agents/`, `hooks/`) the harness loads.

## 1. The binary — an always-latest wrapper

The commands call `abcd` on `PATH`. To make that always the newest source (and to
surface breakage the moment it happens), put a wrapper on `PATH` that rebuilds from
the checkout on every call. `go build` is incremental, so an unchanged call is fast;
a source change that no longer compiles fails loudly here instead of running a stale
binary.

Wrapper at `~/.local/bin/abcd` (any `PATH` dir works; this file is local machine
setup, never committed — so an absolute checkout path is fine *there*):

```sh
#!/bin/sh
set -e
REPO=<absolute-path-to-your-abcd-cli-checkout>
BIN="${TMPDIR:-/tmp}/abcd-dogfood-bin"
go build -C "$REPO" -o "$BIN" ./cmd/abcd
exec "$BIN" "$@"
```

```sh
chmod +x ~/.local/bin/abcd
abcd            # from the checkout → the status board; git repo: true
abcd version    # from any dir → builds latest, runs it
```

Two properties this buys: it works from **any** working directory (it does not rely
on a `go.mod` in the current directory, so it runs against another repo too), and a
red build anywhere in the module stops the wrapper with the compiler error.

To uninstall: `rm ~/.local/bin/abcd`. For a faster, pinned install instead, use
`make build` and symlink the produced `bin/abcd-<goos>-<arch>` onto `PATH`, or
`abcd ahoy install` — at the cost of manual rebuilds.

## 2. The plugin — a local, private marketplace

The repo already declares itself a single-plugin marketplace
(`.claude-plugin/marketplace.json`, `source: "./"`). A filesystem marketplace
publishes nothing.

In the harness:

```
/plugin marketplace add <path-to-your-abcd-cli-checkout>
/plugin install abcd@abcd-marketplace
```

Choose **local** scope when prompted (writes `.claude/settings.local.json`, which is
gitignored) to keep it to yourself and out of the repo. User scope
(`~/.claude/settings.json`) also avoids a per-project workspace-trust prompt.

Confirm it loaded:

```
/plugin list
/plugin details abcd
```

`commands/abcd/<verb>.md` become `/abcd:<verb>` (namespaced by the plugin name).

## 3. Iterating

- **Binary:** nothing to do — the wrapper rebuilds each call.
- **Command / agent markdown:** a local marketplace is copied to a cache on install,
  so edits to the checkout are **not** live until you refresh it:

  ```
  /plugin marketplace update abcd-marketplace
  /reload-plugins
  ```

  Command bodies often re-read immediately after the update; agents and hooks need
  the reload.

Because `plugin.json` carries no `version`, the harness treats each commit as a new
version, which suits active development.

## 4. Testing it on another repo you own

The point of a host-agnostic core is that abcd manages repositories other than its
own. To exercise that:

1. From a repo you own that is **not yet abcd-managed**, run
   `/abcd:prepare-this-repo` — it refuses on repos you do not own, audits the repo,
   then adopts the three-tier `.abcd/` layout and the conventions section.
2. Author a run PLAN under that repo's `.abcd/development/plans/` (the `/abcd:loop`
   contract — see the loop's own plan template).
3. Run the loop against it.

The binary wrapper already works from any directory, so no reinstall is needed to
switch repos — only the target repo's own `.abcd/` scaffold and plan.

## 5. Gotchas

- **Cache staleness** is the main friction — after editing commands, run
  `/plugin marketplace update` + `/reload-plugins`, or the old copy keeps serving.
- **Wrapper latency** is one incremental `go build` per call (sub-second when
  unchanged). Swap for a pinned binary if that ever bites.
- **No remote or commit required** — a working tree on disk is enough; nothing is
  pushed.
- **Trust prompt** appears once per project at project scope; user/local scope
  sidesteps it.
