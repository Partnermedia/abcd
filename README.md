<div align="center">

  <img src="docs/assets/img/logo.png" alt="abcd logo" width="150">

  <h1>Agent-Based Configuration for Development</h1>

  <p>A host-agnostic configuration layer for intent-driven development.</p>

  <img src="https://img.shields.io/badge/status-experimental-orange" alt="Status: experimental">
  <a href="https://github.com/intentdriven/abcd/releases"><img src="https://img.shields.io/github/v/release/intentdriven/abcd?cacheSeconds=300" alt="Release"></a>
  <img src="https://img.shields.io/github/last-commit/intentdriven/abcd?cacheSeconds=300" alt="Last commit">
  <br />
  <img src="https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white" alt="Go 1.26">
  <a href="https://claude.ai/claude-code"><img src="https://img.shields.io/badge/Built_with-Claude_Code-3B5CE7?logo=anthropic&logoColor=white" alt="Built with Claude Code"></a> <!-- docs-lint: allow -->
  <br />
  <img src="https://img.shields.io/badge/macOS-000000?logo=apple&logoColor=white" alt="macOS">
  <img src="https://img.shields.io/badge/Linux-core%20CI--tested-FCC624?logo=linux&logoColor=black" alt="Linux: core CI-tested">

</div>


## What it is for

`abcd` is for people who know what they want to build and need help shipping it. While agents are good at producing work, they're not very good at remembering *why* it was built: *What was decided*, *what was rejected*, and on *what evidence*. All of that reasoning lives in transcripts nobody reads.

`abcd` keeps that reasoning as structured records agents and humans *do* read: The *intent* that says what shipping looks like, the *decision* that says what was chosen and what was refused, the *specification*, and the *issue ledger*. They are plain files in the repository, checked by gates that refuse rather than warn, so what these records claim about the product stays true as the product moves.


## How it works

Two mechanisms carry the record into the work:

**Rules reach the agent selectively.** On each prompt a hook matches what you
typed against the keyword triggers every rule domain declares, and injects
only the domains that matched: A prompt about committing gets the commit conventions,
and a prompt that matches nothing adds no tokens at all.

**Gates refuse rather than warn.** A warning is advice, and advice is what a
hurried author skips. The gates run before the push leaves the machine, and they
stop it:

```text
$ git push
check-issue-resolution: RS001 commit c1e2c4c3d915 declares 'Resolves: iss-317', but iss-317 does not enter .abcd/work/issues/resolved/ or .abcd/work/issues/wontfix/ in origin/main..HEAD. Resolve it in this change (abcd capture resolve iss-317 ...) or drop the trailer.
check-issue-resolution: FAILED — 1 violation(s)
```


## Built in the open

`abcd` is built with `abcd`, and its own record is public and complete: Every decision, intent, specification and issue, from the first commit onward, with
the reasoning attached.

That record *is* the demonstration. It is not a claim that the approach works but the trail of a real product being built this way, including the parts that were wrong: Reversed decisions, abandoned designs, and defects found by the gates and recorded before they were fixed.

You can interrogate `abcd`'s development process at [abcdev.app/record/](https://abcdev.app/record/), alongside more details on [who abcd is for](docs/explanation/rationale.md), [roles](docs/explanation/roles.md), [artefacts](docs/explanation/artefacts.md) and the [process](docs/explanation/process.md).

At this stage, `abcd` remains **experimental**, and you should expect the surface to keep moving: Verbs change, and the plugin surface currently targets a single harness (the command-line app is host-agnostic and runs in any repository with no harness at all). What's not experimental is the [record](https://abcdev.app/record/).


## Install

The easiest route to get started is to install `abcd` as a [Claude Code](https://claude.ai/claude-code) plugin. <!-- docs-lint: allow -->

Run these two from a session in that harness. The first registers this repository as a plugin marketplace; the second installs `abcd` from it.

```
/plugin marketplace add intentdriven/abcd
```

Wait for the confirmation that the marketplace was added, then:

```
/plugin install abcd@abcd-marketplace
```

Restart the session afterwards so the hooks load, then check what you got:

```
/abcd:version
```

Later, `/plugin update abcd` pulls the marketplace's current state.


### As a CLI

Support for other harnesses will follow. If your harness isn't supported yet, `abcd` runs from a terminal in any repository, with no harness involved:

```sh
sh -c 'set -eu; unset HTTPS_PROXY https_proxy HTTP_PROXY http_proxy ALL_PROXY all_proxy CURL_HOME CURL_CA_BUNDLE SSL_CERT_FILE SSL_CERT_DIR; cd "$(mktemp -d)"; os=$(uname -s | tr "[:upper:]" "[:lower:]"); arch=$(uname -m); case "$arch" in x86_64) arch=amd64;; aarch64) arch=arm64;; esac; b="abcd-$os-$arch"; curl -q --proto =https --proto-redir =https -fsSLO "https://github.com/intentdriven/abcd/releases/latest/download/$b"; curl -q --proto =https --proto-redir =https -fsSLO "https://github.com/intentdriven/abcd/releases/latest/download/checksums.txt"; grep " $b$" checksums.txt | if command -v sha256sum >/dev/null; then sha256sum -c -; else shasum -a 256 -c -; fi; mkdir -p "$HOME/.local/bin"; install -m 0755 "$b" "$HOME/.local/bin/abcd"; "$HOME/.local/bin/abcd" version'
```

Checksum-verified, no administrator rights, installs to `~/.local/bin`. The [install guide](docs/how-to/install.md) covers building from source and what to do when `abcd` isn't found afterwards. Consult the [verb reference](docs/reference/cli/commands.md) for what `abcd` can do.


## First run

Point `abcd` at a repository you care about and ask where you are. The status
board is read-only, so it is safe on any tree:

```text
$ abcd
abcd — /path/to/your-repo
  git repo:   true
  record:     true
  work tiers: [development work work.local]
```

The plugin surface adds three commands the CLI does not carry, all host-delegated markdown with no Go verb behind them. The one that matters
first: `/abcd:prepare-this-repo` gives a repository with no record yet the three-tier `.abcd/` layout, an `AGENTS.md` router, and the commit
gates. (`/abcd:consult` and `/abcd:ingest` drive a local sources corpus.)

From there, three verbs cover most of a first session. In a plugin session, `/abcd:lint` checks the repository against the working conventions and names what is missing. `/abcd:capture "…"` files a half-formed observation to the issue ledger so it survives the session that noticed it. `/abcd:intent "…"` opens a user-facing change as a press-release intent, which is where a shipping change starts.

*The [verb reference](docs/reference/cli/commands.md) lists the rest.*


## Citation

If you use `abcd` in your work, please cite it. The form below is rendered from
[`CITATION.cff`](CITATION.cff), which is the record GitHub's own *Cite this
repository* button reads:

```bibtex
@software{reppel_abcd,
  author  = {Reppel, Alex},
  title   = {{abcd}: {Agent-Based} {Configuration} for {Development}},
  url     = {https://github.com/intentdriven/abcd},
  license = {MIT}
}
```


## Resources

- [`SECURITY.md`](SECURITY.md): Report a vulnerability privately.
- [`ACKNOWLEDGEMENTS.md`](ACKNOWLEDGEMENTS.md): The ideas, tools and writing `abcd` stands on.
