# Install

You can use `abcd` by installing it as a [plugin](#plugin) *(for a compatible agent harness)*, download a command-line app [binary](#cli), or by [building](#build) it directly.

## Plugin

This repository is also its own plugin marketplace, so a compatible agent
harness can install the `/abcd:*` surface — the commands under
[`commands/`](https://github.com/Partnermedia/abcd/tree/main/commands/), the agents under [`agents/`](https://github.com/Partnermedia/abcd/tree/main/agents/) and the hook
wiring in [`hooks/`](https://github.com/Partnermedia/abcd/tree/main/hooks/) — straight from it. Add the marketplace, then
install the plugin:

```text
/plugin marketplace add Partnermedia/abcd
/plugin install abcd@abcd-marketplace
```

`abcd-marketplace` is the marketplace name declared in
[`.claude-plugin/`](https://github.com/Partnermedia/abcd/tree/main/.claude-plugin/); `abcd` is the single plugin it lists,
sourced from the repository root. Pull the current state of the marketplace with:

```text
/plugin update abcd
```

The marketplace is served from the repository itself, so an install tracks the
repository rather than a versioned artefact: the manifests here carry no version
key, and a release publishes the `abcd` binaries and their checksums, alongside
the source archives GitHub attaches for the tagged tree.

The plugin provisions its own binary; this repository commits none. The
verified artefact is kept once in the plugin's persistent per-plugin download
cache (`$CLAUDE_PLUGIN_DATA`), and a plugin update — which lands in a fresh,
empty plugin root — is provisioned by a re-verified copy out of that cache
rather than a fresh download. [`hooks/bootstrap.sh`](https://github.com/Partnermedia/abcd/blob/main/hooks/bootstrap.sh) runs
first at session start: when the cache already holds the artefact for the
resolved release it copies it into the plugin root with no network —
authenticating the cached hash against the release's published `checksums.txt`
when online, or noting in its success line that it provisioned from an
unauthenticated cache when offline. Only an empty, stale, or unavailable cache
falls back to downloading the release binary and `checksums.txt` and verifying
the binary's SHA-256 against the manifest. A mismatch, a manifest that doesn't
list the platform, or a platform outside the released matrix (darwin and linux
on amd64 and arm64) installs nothing and says why in plain language. A plugin
root that already holds the binary costs one file test and no network.

Session start is not the only chance. Every other live-session hook that needs
the binary resolves it the same way the command files do — the plugin root
first, then an `abcd` on `PATH` — and when the plugin root is empty it first
attempts the bootstrap itself, silently and at most once per ten-minute window.
Session end is the deliberate exception: it resolves the plugin root then
`PATH` but never downloads, because a fetch there would race the host's
shutdown and lose the very transcript it exists to capture — so it says in one
line if the transcript was not captured rather than blocking on a bootstrap. A
session where provisioning cannot succeed degrades loudly rather than noisily:
each affected hook says in one line what is inactive (the rules loader, the
shell guard, the transcript capture) and that the [install](#cli) one-liner
restores it — after which the hooks resolve the `PATH` binary with no session
restart needed.

That covers the hooks. For the `abcd` command in your own terminal, keep the
[install](#cli) below, or put the plugin-root binary on your `PATH` by
running it once by its absolute path — `'<plugin-root>/abcd' ahoy install`.
The path is absolute because `abcd` is not on your `PATH` yet, which is what
that one run fixes. `<plugin-root>` is the directory the agent harness unpacked
the abcd plugin into, with the binary sitting directly inside it as `abcd`; the
bootstrap's success notice prints that full binary path, so the shortest route
is to copy the command straight out of the notice. That notice appears once per
plugin root — later sessions take the fast path and stay silent — so if it has
scrolled away and you would rather not go looking for the directory, the
[install](#cli) one-liner below needs no plugin root at all and gets you to
the same place.

For a stronger root of trust than same-origin checksums, build from source —
`go build ./cmd/abcd` — and place the binary in the plugin root and on your
`PATH` yourself. A binary placed there by hand takes the same no-network fast
path. A plugin root provisioned from the cache carries no root-local
`.binary-meta`: its provenance lives in the data directory's
`cache/binary-meta`, and it records whichever release the bootstrap last cached
— so a hand-built binary reports that release's vintage. Replace or remove the
cached provenance you control if you want a hand-built binary to stop reporting
a release it did not come from.

## CLI

One line, checksum-verified, no administrator rights. Pick your operating
system: the command detects your architecture, downloads the binary and the
`checksums.txt` manifest from the latest release, verifies the binary's SHA-256
against the manifest (and refuses to install on any mismatch — or if the
manifest doesn't list the binary at all), then installs to `~/.local/bin`, the
single-user location.

### macOS

```sh
sh -c 'set -eu; cd "$(mktemp -d)"; arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; esac; b="abcd-darwin-$arch"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/$b"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | shasum -a 256 -c -; mkdir -p "$HOME/.local/bin"; install -m 0755 "$b" "$HOME/.local/bin/abcd"; "$HOME/.local/bin/abcd" version'
```

### Linux

```sh
sh -c 'set -eu; cd "$(mktemp -d)"; arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; aarch64) arch=arm64;; esac; b="abcd-linux-$arch"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/$b"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | sha256sum -c -; mkdir -p "$HOME/.local/bin"; install -m 0755 "$b" "$HOME/.local/bin/abcd"; "$HOME/.local/bin/abcd" version'
```

### Windows

Not yet. The released matrix is darwin and linux on amd64 and arm64; a Windows
build is planned, and this page will carry its command when it ships. Until
then, Windows users can run the Linux command inside WSL.

### Afterwards

If `abcd` isn't found by name afterwards, `~/.local/bin` isn't on your `PATH`.
Add this line to your shell profile:

```sh
export PATH="$HOME/.local/bin:$PATH"
```

The installer prints the same one-line fix when it notices the gap, and
writes its own `PATH` entry to `~/.local/bin` unless you point it elsewhere
with `--bin-dir`. abcd never escalates privileges: a
directory it can't write to is an error, not a prompt for your password.

Already have an `abcd` in a system directory from an earlier install? Delete it
(`rm /usr/local/bin/abcd`, with whatever rights put it there) — otherwise it
comes first on `PATH` and keeps answering instead of the new one. `abcd ahoy`
names it in a gap rather than removing it: abcd does not touch a binary it does
not own.

Prefer to inspect before running? The command is exactly what it says: two
downloads from [the latest release](https://github.com/Partnermedia/abcd/releases/latest),
a checksum verification, and a copy into a directory you own. You can do the
same by hand — grab the binary for your platform plus `checksums.txt` from the
releases page, run `shasum -a 256 -c` (or `sha256sum -c`) against the matching
line, and copy the binary anywhere on your `PATH`. Every release is built and
published by CI from the exact tagged commit, with the checksums generated
over the same bytes that are uploaded.

## Build

```bash
make preflight   # the pre-push gate: lint-reviews, record-lint and docs-lint,
                 # then build, vet, test and race (internal)
go run ./cmd/abcd            # bare status board for the current directory
go run ./cmd/abcd version    # print the version
make build                   # cross-compile bin/abcd-<goos>-<arch>
```
