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

A single Go binary that carries the why from idea to shipped reality, usable as a plugin in compatible agent harnesses.

**[abcdev.app](https://abcdev.app)** is the front door: [who abcd is for](docs/explanation/rationale.md), the [roles](docs/explanation/roles.md), the [artefacts](docs/explanation/artefacts.md) and the [process](docs/explanation/process.md), rendered from the pages in [`docs/`](docs/README.md). The [development record](.abcd/development/README.md) — every decision, intent, spec and issue — is rendered there too.

# Install

```sh
sh -c 'set -eu; cd "$(mktemp -d)"; os=$(uname -s | tr "[:upper:]" "[:lower:]"); arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; aarch64) arch=arm64;; esac; b="abcd-$os-$arch"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/$b"; curl -fsSLO "https://github.com/Partnermedia/abcd/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | if command -v sha256sum >/dev/null; then sha256sum -c -; else shasum -a 256 -c -; fi; mkdir -p "$HOME/.local/bin"; install -m 0755 "$b" "$HOME/.local/bin/abcd"; "$HOME/.local/bin/abcd" version'
```

Checksum-verified, no administrator rights, installs to `~/.local/bin`. The plugin route, building from source, and what to do when `abcd` is not found afterwards: [`docs/how-to/install.md`](docs/how-to/install.md).

# Contributing

- [`AGENTS.md`](AGENTS.md) — build, test and checks; the working-tree layout; the definition of done.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — how changes land, the publish surface, attribution.
- [`SECURITY.md`](SECURITY.md) — report vulnerabilities privately.
- [`ACKNOWLEDGEMENTS.md`](ACKNOWLEDGEMENTS.md) — the ideas, tools and writing abcd stands on.

# Resources

- [`cmd/abcd/`](cmd/abcd/) — CLI entry point.
- [`internal/`](internal/) — the engine (`core/`) and front doors (`surface/`);
  see [`internal/README.md`](internal/README.md).
- [`commands/`](commands/), [`.claude-plugin/`](.claude-plugin/) — the plugin
  surface (auto-loaded).
- [`.abcd/`](.abcd/) — the development record and working files (present in
  every repository checkout, marketplace installs and release source archives
  included; never in the released binaries).
