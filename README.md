<div align="center">

  <img src="docs/assets/img/logo.png" alt="abcd logo" width="150">

  <h1>Agent-Based Configuration for Development</h1>

  <p>A host-agnostic configuration layer for intent-driven development.</p>

  <img src="https://img.shields.io/badge/status-experimental-orange" alt="Status: experimental">
  <a href="https://github.com/Partnermedia/abcd/releases"><img src="https://img.shields.io/github/v/release/Partnermedia/abcd?cacheSeconds=300" alt="Release"></a>
  <img src="https://img.shields.io/github/last-commit/Partnermedia/abcd?cacheSeconds=300" alt="Last commit">
  <br />
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26">
  <a href="https://claude.ai/claude-code"><img src="https://img.shields.io/badge/Built_with-Claude_Code-3B5CE7?logo=anthropic&logoColor=white" alt="Built with Claude Code"></a> <!-- docs-lint: allow -->
  <br />
  <img src="https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white" alt="macOS">
  <img src="https://img.shields.io/badge/Linux-core%20CI--tested-FCC624?logo=linux&logoColor=black" alt="Linux: core CI-tested">

</div>


## What it is for

`abcd` is for people who know what they want to build and need help shipping it.

Agents are good at producing work and bad at remembering *why*. Ask one to build something and it will, then the next session starts again from the artefact rather than the argument. What was decided, what was rejected and on what evidence lives in a transcript nobody reads twice, so the reasoning is spent as fast as it is made.

`abcd` keeps that reasoning as records the agent and the human both read: the intent that says what shipping looks like, the decision that says what was chosen and what was refused, the specification, the issue ledger. They are plain files in the repository, checked by gates that refuse rather than warn, so what the record claims about the product stays true as the product moves.


## Built in the open

`abcd` is built with `abcd`, and its own record is public and complete: Every decision, intent, specification and issue, from the first commit onward, with
the reasoning attached rather than summarised.

That record is the demonstration. It is not a claim that the approach works but the trail of a real product being built this way, including the parts that were
wrong: Reversed decisions, abandoned designs, and defects found by the gates and recorded before they were fixed. Read it at **[abcdev.app](https://abcdev.app)**, alongside [who abcd is for](docs/explanation/rationale.md), the [roles](docs/explanation/roles.md), the [artefacts](docs/explanation/artefacts.md) and the [process](docs/explanation/process.md), rendered from the pages in [`docs/`](docs/README.md) and the [development record](.abcd/development/README.md).

`abcd` is **experimental** and the surface still moves: Verbs change and it's currently only available for a single harness. What's not experimental is the record and neither is the discipline it's being built with.


## Install

The easiest route to get started is to install `abcd` as a [Claude Code](https://claude.ai/claude-code) plugin. <!-- docs-lint: allow -->

Run these two, in order, from a session in that harness:

```
/plugin marketplace add Partnermedia/abcd
/plugin install abcd@abcd-marketplace
```

The first registers this repository as a plugin marketplace; the second installs `abcd` from it. Restart the session afterwards so the hooks load, then check what you got:

```
/abcd:version
```

Later, `/plugin update abcd` fetches a newer release. It tracks published releases rather than the main branch, so a change lands for you when it ships, not when it merges. Reloading plugins is not the same thing: that re-reads what is already on disk, so it refreshes the commands and skills while leaving the binary as it was.

Support for other harnesses will follow.

### As a CLI

If your harness isn't supported yet, `abcd` runs from a terminal in any repository, with no harness involved:

```sh
sh -c 'set -eu; cd "$(mktemp -d)"; os=$(uname -s | tr "[:upper:]" "[:lower:]"); arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; aarch64) arch=arm64;; esac; b="abcd-$os-$arch"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/$b"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | if command -v sha256sum >/dev/null; then sha256sum -c -; else shasum -a 256 -c -; fi; mkdir -p "$HOME/.local/bin"; install -m 0755 "$b" "$HOME/.local/bin/abcd"; "$HOME/.local/bin/abcd" version'
```

Checksum-verified, no administrator rights, installs to `~/.local/bin`. The [install guide](docs/how-to/install.md) covers building from source and what to do when `abcd` isn't found afterwards. Consult the [verb reference](docs/reference/cli/commands.md) for what `abcd` can do.


## Resources

- [`SECURITY.md`](SECURITY.md): report a vulnerability privately.
- [`ACKNOWLEDGEMENTS.md`](ACKNOWLEDGEMENTS.md): the ideas, tools and writing `abcd` stands on.
